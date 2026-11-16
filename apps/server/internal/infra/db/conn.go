package db

import (
	"context"
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stephenafamo/bob"
	bobpgx "github.com/stephenafamo/bob/drivers/pgx"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

type Database struct {
	Pool *pgxpool.Pool
	Exec bob.Executor // Pass this into Bob generated queries
}

func ConnectAndMigrate(ctx context.Context, databaseURL string) (*Database, error) {
	// 1. Create pgx connection pool
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create pgxpool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	// 2. Run Goose migrations automatically using stdlib *sql.DB wrapper
	sqlDB := stdlib.OpenDBFromPool(pool)
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, err
	}
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		return nil, fmt.Errorf("goose migrate up: %w", err)
	}
	_ = sqlDB.Close() // Closes the stdlib wrapper without closing the underlying pgxpool

	// 3. Wrap pgxpool for Bob
	bobExec := bobpgx.NewPool(pool)

	return &Database{
		Pool: pool,
		Exec: bobExec,
	}, nil
}

func (d *Database) Close() {
	d.Pool.Close()
}