package application

import (
	"context"
	"encoding/json"

	"github.com/SirNacou/mirage/apps/server/internal/db/gen/models"
	"github.com/SirNacou/mirage/apps/server/internal/domain"
	"github.com/SirNacou/mirage/packages/sync-core/chunker"
	"github.com/aarondl/opt/omit"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/types"
)

type CommitFileCommand struct {
	RootID       string
	RelativePath string
	Size         int64
	MtimeUnix    int64
	IsDeleted    bool
	Chunks       []chunker.ChunkMeta
}

type CommitFileHandler struct {
	exec bob.Executor
}

func NewCommitFileHandler(exec bob.Executor) *CommitFileHandler {
	return &CommitFileHandler{exec: exec}
}

func (h *CommitFileHandler) Handle(ctx context.Context, cmd CommitFileCommand) (int64, error) {
	safePath, err := domain.ParseSafeRelPath(cmd.RelativePath)
	if err != nil {
		return 0, err
	}

	if err := domain.ValidateManifest(cmd.Size, cmd.IsDeleted, cmd.Chunks); err != nil {
		return 0, err
	}

	rootID := cmd.RootID
	if rootID == "" {
		rootID = "default"
	}

	chunks := cmd.Chunks
	if chunks == nil {
		chunks = []chunker.ChunkMeta{}
	}
	chunksJSON, err := json.Marshal(chunks)
	if err != nil {
		return 0, err
	}

	file, err := models.Files.Insert(
		&models.FileSetter{
			RootID:       omit.From(rootID),
			RelativePath: omit.From(safePath.String()),
			Size:         omit.From(cmd.Size),
			MtimeUnix:    omit.From(cmd.MtimeUnix),
			IsDeleted:    omit.From(cmd.IsDeleted),
			Chunks:       omit.From(types.NewJSON(json.RawMessage(chunksJSON))),
		},
		im.OnConflict(models.Files.Columns.RootID, models.Files.Columns.RelativePath).DoUpdate(
			im.SetExcluded(
				models.Files.Columns.Size.Name(),
				models.Files.Columns.MtimeUnix.Name(),
				models.Files.Columns.IsDeleted.Name(),
				models.Files.Columns.Chunks.Name(),
			),
			im.SetCol(models.Files.Columns.SyncSeq.Name()).
				To(psql.Raw("nextval('global_sync_seq')")),
			im.SetCol(models.Files.Columns.UpdatedAt.Name()).To(psql.Raw("NOW()")),
		),
		im.Returning(models.Files.Columns.SyncSeq),
	).One(ctx, h.exec)
	if err != nil {
		return 0, err
	}

	return file.SyncSeq, nil
}
