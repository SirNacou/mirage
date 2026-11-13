package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Env struct {
	PORT int `env:"SERVER_PORT" envDefault:"8080"`
}

func LoadEnv() (*Env, error) {
	var e Env
	if err := env.Parse(&e); err != nil {
		return nil, fmt.Errorf("failed to load env: %w", err)
	}

	return &e, nil
}
