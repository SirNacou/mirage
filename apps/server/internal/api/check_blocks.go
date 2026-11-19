package api

import (
	"context"
	"net/http"

	"github.com/SirNacou/mirage/apps/server/internal/app"
	"github.com/SirNacou/mirage/packages/sync-core/protocol"
	"github.com/danielgtaylor/huma/v2"
)

type blockCheckInput struct {
	Body protocol.BlockCheckReq
}

type blockCheckOutput struct {
	Body protocol.BlockCheckResp
}

func registerCheckBlocks(api huma.API, handler *app.CheckBlocksHandler) {
	huma.Register(api, huma.Operation{
		OperationID: "check-blocks",
		Method:      http.MethodPost,
		Path:        "/blocks/check",
		Summary:     "Find blocks that are missing from the server",
	}, func(ctx context.Context, input *blockCheckInput) (*blockCheckOutput, error) {
		missing, err := handler.Handle(ctx, app.CheckBlocksQuery{
			RawHashes: input.Body.Hashes,
		})
		if err != nil {
			return nil, err
		}
		return &blockCheckOutput{Body: protocol.BlockCheckResp{Missing: missing}}, nil
	})
}
