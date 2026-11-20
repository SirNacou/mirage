package api

import (
	"github.com/SirNacou/mirage/apps/server/internal/app"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v5"
	"github.com/stephenafamo/bob"
)

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

	registerCheckBlocks(syncAPI, handlers.FindMissingBlocks)
	registerDownloadBlock(syncAPI, handlers.DownloadBlock)
	registerCommitFile(syncAPI, handlers.CommitFile)
	registerListChanges(syncAPI, handlers.ListChanges)
	registerUploadBlock(syncRouter, handlers.UploadBlock)
}
