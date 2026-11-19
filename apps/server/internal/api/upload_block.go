package api

import (
	"net/http"

	"github.com/SirNacou/mirage/apps/server/internal/app"
	"github.com/labstack/echo/v5"
)

func registerUploadBlock(router *echo.Group, handler *app.UploadBlockHandler) {
	router.PUT("/blocks/:hash", func(c *echo.Context) error {
		err := handler.Handle(c.Request().Context(), app.UploadBlockRequest{
			RawHash: c.Param("hash"),
			Body:    c.Request().Body,
		})
		if err != nil {
			return err
		}
		return c.NoContent(http.StatusNoContent)
	})
}
