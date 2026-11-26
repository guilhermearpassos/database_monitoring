package config

import (
	"context"

	"github.com/guilhermearpassos/database-monitoring/internal/common/telemetry"
	"github.com/jmoiron/sqlx"
)

type PostgresConfig struct {
	Connstring string `toml:"connstring" yaml:"connstring"`
}

func (c PostgresConfig) Get(ctx context.Context) (*sqlx.DB, error) {
	return telemetry.OpenInstrumentedDB("postgres", c.Connstring)
}
