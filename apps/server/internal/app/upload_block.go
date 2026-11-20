package app

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/SirNacou/mirage/apps/server/internal/domain"
	"github.com/SirNacou/mirage/apps/server/internal/infra/db/gen/models"
	"github.com/aarondl/opt/omit"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/zeebo/blake3"
)

type UploadBlockRequest struct {
	RawHash string
	Body    io.Reader
}

type UploadBlockHandler struct {
	exec     bob.Executor
	blockDir string
}

func NewUploadBlockHandler(exec bob.Executor, dataDir string) (*UploadBlockHandler, error) {
	blockDir := filepath.Join(dataDir, "blocks")
	if err := os.MkdirAll(blockDir, 0o755); err != nil {
		return nil, err
	}
	return &UploadBlockHandler{exec: exec, blockDir: blockDir}, nil
}

func (h *UploadBlockHandler) Handle(ctx context.Context, cmd UploadBlockRequest) error {
	hash, err := domain.ParseBlockHash(cmd.RawHash)
	if err != nil {
		return err
	}

	finalPath := hash.DiskPath(h.blockDir)
	prefixDir := filepath.Dir(finalPath)
	if err := os.MkdirAll(prefixDir, 0o755); err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(prefixDir, "upload-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	hasher := blake3.New()
	written, err := io.Copy(io.MultiWriter(tmpFile, hasher), cmd.Body)
	if err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()

	actualHash := hex.EncodeToString(hasher.Sum(nil))
	if actualHash != hash.String() {
		return fmt.Errorf("hash mismatch: expected %s, got %s", hash, actualHash)
	}

	if err := os.Rename(tmpFile.Name(), finalPath); err != nil {
		return err
	}

	_, err = models.Chunks.Insert(
		&models.ChunkSetter{
			Hash: omit.From(hash.String()),
			Size: omit.From(written),
		},
		im.OnConflict().DoNothing(),
	).Exec(ctx, h.exec)

	return err
}
