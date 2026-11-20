package api

import (
	"context"
	"net/http"

	"github.com/SirNacou/mirage/apps/server/internal/app"
	"github.com/SirNacou/mirage/packages/sync-core/protocol"
	"github.com/danielgtaylor/huma/v2"
)

type commitFileInput struct {
	Body protocol.CommitFileReq
}

type commitFileOutput struct {
	Body protocol.CommitFileResp
}

func registerCommitFile(api huma.API, handler *app.CommitFileHandler) {
	huma.Register(api, huma.Operation{
		OperationID: "commit-file",
		Method:      http.MethodPost,
		Path:        "/files/commit",
		Summary:     "Commit a file manifest to the sync history",
	}, func(ctx context.Context, input *commitFileInput) (*commitFileOutput, error) {
		seq, err := handler.Handle(ctx, app.CommitFileCommand{
			RootID:       input.Body.RootID,
			RelativePath: input.Body.RelativePath,
			Size:         input.Body.Size,
			MtimeUnix:    input.Body.MtimeUnix,
			IsDeleted:    input.Body.IsDeleted,
			Chunks:       input.Body.Chunks,
		})
		if err != nil {
			return nil, err
		}
		return &commitFileOutput{Body: protocol.CommitFileResp{SyncSeq: seq}}, nil
	})
}
