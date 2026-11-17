package api

import (
	"context"
	"net/http"

	"github.com/SirNacou/mirage/apps/server/internal/app"
	"github.com/SirNacou/mirage/packages/sync-core/protocol"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v5"
	"github.com/stephenafamo/bob"
)

type blockCheckInput struct {
	Body protocol.BlockCheckReq
}

type blockCheckOutput struct {
	Body protocol.BlockCheckResp
}

type commitFileInput struct {
	Body protocol.CommitFileReq
}

type commitFileOutput struct {
	Body protocol.CommitFileResp
}

type listChangesInput struct {
	RootID   string `query:"root_id"`
	SinceSeq int64  `query:"since"`
}

type listChangesOutput struct {
	Body protocol.SyncChangesResp
}

// NewRouter initializes application handlers and registers all HTTP endpoints.
func NewRouter(exec bob.Executor, dataDir string) (*echo.Echo, error) {
	handlers, err := app.NewHandlers(exec, dataDir)
	if err != nil {
		return nil, err
	}

	router := echo.New()
	humaAPI := humaecho.New(router, huma.DefaultConfig("Mirage API", "1.0.0"))
	registerSyncRoutes(humaAPI, router, handlers)

	return router, nil
}

func registerSyncRoutes(humaAPI huma.API, router *echo.Echo, handlers *app.Handlers) {
	syncAPI := huma.NewGroup(humaAPI, "/api/sync")
	syncRouter := router.Group("/api/sync")

	huma.Register(syncAPI, huma.Operation{
		OperationID: "check-blocks",
		Method:      http.MethodPost,
		Path:        "/blocks/check",
		Summary:     "Find blocks that are missing from the server",
	}, func(ctx context.Context, input *blockCheckInput) (*blockCheckOutput, error) {
		missing, err := handlers.FindMissingBlocks.Handle(ctx, app.CheckBlocksQuery{
			RawHashes: input.Body.Hashes,
		})
		if err != nil {
			return nil, err
		}
		return &blockCheckOutput{Body: protocol.BlockCheckResp{Missing: missing}}, nil
	})

	huma.Register(syncAPI, huma.Operation{
		OperationID: "commit-file",
		Method:      http.MethodPost,
		Path:        "/files/commit",
		Summary:     "Commit a file manifest to the sync history",
	}, func(ctx context.Context, input *commitFileInput) (*commitFileOutput, error) {
		seq, err := handlers.CommitFile.Handle(ctx, app.CommitFileCommand{
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

	huma.Register(syncAPI, huma.Operation{
		OperationID: "list-changes",
		Method:      http.MethodGet,
		Path:        "/changes",
		Summary:     "List file changes since a sequence number",
	}, func(ctx context.Context, input *listChangesInput) (*listChangesOutput, error) {
		changes, err := handlers.ListChanges.Handle(ctx, app.ListChangesQuery{
			RootID:   input.RootID,
			SinceSeq: input.SinceSeq,
		})
		if err != nil {
			return nil, err
		}
		return &listChangesOutput{Body: changes}, nil
	})

	syncRouter.PUT("/blocks/:hash", func(c *echo.Context) error {
		err := handlers.UploadBlock.Handle(c.Request().Context(), app.UploadBlockRequest{
			RawHash: c.Param("hash"),
			Body:    c.Request().Body,
		})
		if err != nil {
			return err
		}
		return c.NoContent(http.StatusNoContent)
	})
}
