package api

import (
	"context"
	"net/http"

	"github.com/SirNacou/mirage/apps/server/internal/app"
	"github.com/SirNacou/mirage/packages/sync-core/protocol"
	"github.com/danielgtaylor/huma/v2"
)

type listChangesInput struct {
	RootID   string `query:"root_id"`
	SinceSeq int64  `query:"since"`
}

type listChangesOutput struct {
	Body protocol.SyncChangesResp
}

func registerListChanges(api huma.API, handler *app.ListChangesHandler) {
	huma.Register(api, huma.Operation{
		OperationID: "list-changes",
		Method:      http.MethodGet,
		Path:        "/changes",
		Summary:     "List file changes since a sequence number",
	}, func(ctx context.Context, input *listChangesInput) (*listChangesOutput, error) {
		changes, err := handler.Handle(ctx, app.ListChangesQuery{
			RootID:   input.RootID,
			SinceSeq: input.SinceSeq,
		})
		if err != nil {
			return nil, err
		}
		return &listChangesOutput{Body: changes}, nil
	})
}
