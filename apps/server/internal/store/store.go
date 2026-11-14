package store

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/SirNacou/mirage/apps/server/internal/db"
	"github.com/SirNacou/mirage/apps/server/internal/db/gen/models"
	"github.com/SirNacou/mirage/packages/sync-core/protocol"
	"github.com/aarondl/opt/omit"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/zeebo/blake3"
)

type SyncStore struct {
	db       *db.Database
	blockDir string
}

func NewSyncStore(database *db.Database, dataDir string) (*SyncStore, error) {
	blockDir := filepath.Join(dataDir, "blocks")
	if err := os.MkdirAll(blockDir, 0o755); err != nil {
		return nil, err
	}
	return &SyncStore{db: database, blockDir: blockDir}, nil
}

// 1. FindMissingBlocks uses Bob's type-safe Where/In filter to check which hashes exist.
func (s *SyncStore) FindMissingBlocks(ctx context.Context, hashes []string) ([]string, error) {
	if len(hashes) == 0 {
		return []string{}, nil
	}

	existingRows, err := models.Chunks.Query(
		models.SelectWhere.Chunks.Hash.In(hashes...),
	).All(ctx, s.db.Exec)
	if err != nil {
		return nil, err
	}

	existing := make(map[string]struct{}, len(existingRows))
	for _, row := range existingRows {
		existing[row.Hash] = struct{}{}
	}

	var missing []string
	for _, h := range hashes {
		if _, ok := existing[h]; !ok {
			missing = append(missing, h)
		}
	}
	return missing, nil
}

// 2. WriteBlock streams the chunk to a temp file, verifies BLAKE3, renames atomically, and inserts via Bob.
func (s *SyncStore) WriteBlock(ctx context.Context, expectedHash string, r io.Reader) error {
	prefixDir := filepath.Join(s.blockDir, expectedHash[:2])
	if err := os.MkdirAll(prefixDir, 0o755); err != nil {
		return err
	}
	finalPath := filepath.Join(prefixDir, expectedHash)

	tmpFile, err := os.CreateTemp(prefixDir, "upload-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	hasher := blake3.New()
	written, err := io.Copy(io.MultiWriter(tmpFile, hasher), r)
	if err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()

	actualHash := hex.EncodeToString(hasher.Sum(nil))
	if actualHash != expectedHash {
		return fmt.Errorf("hash mismatch: expected %s, got %s", expectedHash, actualHash)
	}

	if err := os.Rename(tmpFile.Name(), finalPath); err != nil {
		return err
	}

	// Insert into chunks table using Bob (ignoring duplicates if already uploaded)
	_, err = models.Chunks.Insert(
		&models.ChunkSetter{
			Hash: omit.From(expectedHash),
			Size: omit.From(written),
		},
	).Exec(ctx, s.db.Exec)
	// Or if concurrent uploads of the same hash happen, fallback to ON CONFLICT DO NOTHING:
	if err != nil {
		_, err = s.db.Pool.Exec(
			ctx,
			`INSERT INTO chunks (hash, size) VALUES ($1, $2) ON CONFLICT (hash) DO NOTHING`,
			expectedHash,
			written,
		)
	}
	return err
}

func (s *SyncStore) BlockPath(hash string) string {
	return filepath.Join(s.blockDir, hash[:2], hash)
}

// 3. CommitFile upserts the file recipe and increments global_sync_seq.
func (s *SyncStore) CommitFile(ctx context.Context, req protocol.CommitFileReq) (int64, error) {
	chunksJSON, err := json.Marshal(req.Chunks)
	if err != nil {
		return 0, err
	}

	var seq int64
	err = s.db.Pool.QueryRow(ctx, `
		INSERT INTO files (root_id, relative_path, size, mtime_unix, is_deleted, chunks, sync_seq, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, nextval('global_sync_seq'), NOW())
		ON CONFLICT (root_id, relative_path) DO UPDATE SET
			size = EXCLUDED.size,
			mtime_unix = EXCLUDED.mtime_unix,
			is_deleted = EXCLUDED.is_deleted,
			chunks = EXCLUDED.chunks,
			sync_seq = nextval('global_sync_seq'),
			updated_at = NOW()
		RETURNING sync_seq
	`, req.RootID, req.RelativePath, req.Size, req.MtimeUnix, req.IsDeleted, string(chunksJSON)).Scan(&seq)

	return seq, err
}

// 4. GetChangesSince uses Bob's generated File query builder to fetch changes ordered by sync_seq.
func (s *SyncStore) GetChangesSince(
	ctx context.Context,
	rootID string,
	sinceSeq int64,
) (protocol.SyncChangesResp, error) {
	files, err := models.Files.Query(
		models.SelectWhere.Files.RootID.EQ(rootID),
		models.SelectWhere.Files.SyncSeq.GT(sinceSeq),
		sm.OrderBy(models.Files.Columns.SyncSeq).Asc(),
	).All(ctx, s.db.Exec)
	if err != nil {
		return protocol.SyncChangesResp{}, err
	}

	resp := protocol.SyncChangesResp{LatestSeq: sinceSeq}
	for _, f := range files {
		ev := protocol.FileChangeEvent{
			SyncSeq:      f.SyncSeq,
			RelativePath: f.RelativePath,
			Size:         f.Size,
			MtimeUnix:    f.MtimeUnix,
			IsDeleted:    f.IsDeleted,
		}
		// Unmarshal the JSONB column from Bob's types.JSON[json.RawMessage]
		_ = json.Unmarshal(f.Chunks.Val, &ev.Chunks)

		if f.SyncSeq > resp.LatestSeq {
			resp.LatestSeq = f.SyncSeq
		}
		resp.Changes = append(resp.Changes, ev)
	}
	return resp, nil
}
