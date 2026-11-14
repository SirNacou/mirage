package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Env struct {
	PORT        int    `env:"SERVER_PORT" envDefault:"8080"`
	DatabaseURL string `env:"DATABASE_URL" envDefault:"postgres://mirage:mirage@localhost:5432/mirage?sslmode=disable"`
}

func LoadEnv() (*Env, error) {
	var e Env
	if err := env.Parse(&e); err != nil {
		return nil, fmt.Errorf("failed to load env: %w", err)
	}

	return &e, nil
}
