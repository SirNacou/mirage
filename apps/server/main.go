package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/SirNacou/mirage/apps/server/internal/config"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/labstack/echo/v5"
)

func main() {
	cli := humacli.New(setup)

	cli.Run()
}

func setup(h humacli.Hooks, o *struct{}) {
	env, err := config.LoadEnv()
	if err != nil {
		log.Fatalln(err)
	}

	router := echo.New()
	_ = humaecho.New(router, huma.DefaultConfig("Mirage API", "1.0.0"))

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", env.PORT),
		Handler: router,
	}

	h.OnStart(func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("error when starting server: %v", err)
		}
	})

	h.OnStop(func() {
		// Give the server 5 seconds to gracefully shut down, then give up.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("error when shutting down server: %v", err)
		}
	})
}
