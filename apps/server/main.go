package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/SirNacou/mirage/apps/server/internal/api"
	"github.com/SirNacou/mirage/apps/server/internal/infra/config"
	"github.com/SirNacou/mirage/apps/server/internal/infra/db"
	"github.com/danielgtaylor/huma/v2/humacli"
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

	ctx := context.Background()
	database, err := db.ConnectAndMigrate(ctx, env.DatabaseURL)
	if err != nil {
		log.Fatalf("database init failed: %v", err)
	}

	router, err := api.NewRouter(database.Exec, env.DataDir)
	if err != nil {
		database.Close()
		log.Fatalf("api initialization failed: %v", err)
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", env.PORT),
		Handler: router,
	}

	h.OnStart(func() {
		slog.Info("server is running", slog.Int("port", env.PORT))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("error when starting server: %v", err)
		}
	})

	h.OnStop(func() {
		// Give the server 5 seconds to gracefully shut down, then give up.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("error when shutting down server: %v", err)
		}
		database.Close()
	})
}
