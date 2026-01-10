// Package postgres provides PostgreSQL database configuration and connection utilities.
package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config represents PostgreSQL connection configuration
type Config struct {
	// DSN is the PostgreSQL connection string
	// Format: postgres://user:password@host:port/dbname?sslmode=disable
	DSN string `yaml:"dsn,omitempty" json:"dsn,omitempty"`

	// MaxConns is the maximum number of connections in the pool, default 10
	MaxConns int32 `yaml:"maxConns,omitempty" json:"maxConns,omitempty"`

	// MinConns is the minimum number of connections in the pool, default 2
	MinConns int32 `yaml:"minConns,omitempty" json:"minConns,omitempty"`
}

// IsEnabled returns true if PostgreSQL is configured
func (c Config) IsEnabled() bool {
	return c.DSN != ""
}

// New creates a new PostgreSQL connection pool
func New(ctx context.Context, c Config) (*pgxpool.Pool, error) {
	if !c.IsEnabled() {
		return nil, nil
	}

	config, err := pgxpool.ParseConfig(c.DSN)
	if err != nil {
		return nil, err
	}

	// Apply defaults
	if c.MaxConns > 0 {
		config.MaxConns = c.MaxConns
	} else {
		config.MaxConns = 10
	}
	if c.MinConns > 0 {
		config.MinConns = c.MinConns
	} else {
		config.MinConns = 2
	}

	return pgxpool.NewWithConfig(ctx, config)
}

// MustNew creates a new PostgreSQL connection pool or panics
func MustNew(ctx context.Context, c Config) *pgxpool.Pool {
	pool, err := New(ctx, c)
	if err != nil {
		panic(err)
	}
	if pool == nil {
		panic("postgres: config not enabled")
	}
	return pool
}