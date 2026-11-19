package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/SirNacou/mirage/apps/server/internal/app"
	"github.com/danielgtaylor/huma/v2"
)

type downloadBlockInput struct {
	Hash string `path:"hash" minLength:"64" maxLength:"64" doc:"64-char hex BLAKE3 chunk hash"`
}

func registerDownloadBlock(api huma.API, handler *app.DownloadBlockHandler) {
	huma.Register(api, huma.Operation{
		OperationID: "download-block",
		Method:      http.MethodGet,
		Path:        "/blocks/{hash}",
		Summary:     "Download a raw binary file chunk",
		Description: "Streams a content-addressed BLAKE3 chunk so clients can reassemble files.",
		Tags:        []string{"Sync"},
	}, func(ctx context.Context, input *downloadBlockInput) (*huma.StreamResponse, error) {
		stream, err := handler.DownloadBlock(ctx, input.Hash)
		if errors.Is(err, os.ErrNotExist) {
			return nil, huma.Error404NotFound("block not found")
		}
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to open block")
		}

		return &huma.StreamResponse{
			Body: func(hctx huma.Context) {
				defer stream.BlockStream.Body.Close()

				hctx.SetHeader("Content-Type", "application/octet-stream")
				hctx.SetHeader("Content-Length", strconv.FormatInt(stream.BlockStream.Size, 10))
				hctx.SetHeader("Cache-Control", "public, max-age=31536000, immutable")

				_, _ = io.Copy(hctx.BodyWriter(), stream.BlockStream.Body)
			},
		}, nil
	})
}
