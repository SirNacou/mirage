package application

import (
	"context"
	"encoding/json"

	"github.com/SirNacou/mirage/apps/server/internal/db/gen/models"
	"github.com/SirNacou/mirage/packages/sync-core/protocol"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

type ListChangesQuery struct {
	RootID   string
	SinceSeq int64
}

type ListChangesHandler struct {
	exec bob.Executor
}

func NewListChangesHandler(exec bob.Executor) *ListChangesHandler {
	return &ListChangesHandler{exec: exec}
}

func (h *ListChangesHandler) Handle(
	ctx context.Context,
	q ListChangesQuery,
) (protocol.SyncChangesResp, error) {
	rootID := q.RootID
	if rootID == "" {
		rootID = "default"
	}

	files, err := models.Files.Query(
		models.SelectWhere.Files.RootID.EQ(rootID),
		models.SelectWhere.Files.SyncSeq.GT(q.SinceSeq),
		sm.OrderBy(models.Files.Columns.SyncSeq).Asc(),
	).All(ctx, h.exec)
	if err != nil {
		return protocol.SyncChangesResp{}, err
	}

	resp := protocol.SyncChangesResp{
		LatestSeq: q.SinceSeq,
		Changes:   make([]protocol.FileChangeEvent, 0, len(files)),
	}

	for _, f := range files {
		ev := protocol.FileChangeEvent{
			SyncSeq:      f.SyncSeq,
			RelativePath: f.RelativePath,
			Size:         f.Size,
			MtimeUnix:    f.MtimeUnix,
			IsDeleted:    f.IsDeleted,
		}
		if err := json.Unmarshal(f.Chunks.Val, &ev.Chunks); err != nil {
			return resp, err
		}
		if f.SyncSeq > resp.LatestSeq {
			resp.LatestSeq = f.SyncSeq
		}
		resp.Changes = append(resp.Changes, ev)
	}

	return resp, nil
}
