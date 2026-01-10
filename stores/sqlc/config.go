// Package sqlc provides database configuration utilities for sqlc-generated queries.
package sqlc

import (
	"context"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBType represents the database type
type DBType string

const (
	DBTypePostgres DBType = "postgres"
	DBTypeMySQL    DBType = "mysql"
)

// Config represents sqlc database configuration
// This provides a unified config for sqlc-generated queries
type Config struct {
	// Type is the database type: "postgres" or "mysql"
	Type DBType `yaml:"type,omitempty" json:"type,omitempty"`

	// DSN is the database connection string
	// PostgreSQL: postgres://user:password@host:port/dbname?sslmode=disable
	// MySQL: user:password@tcp(host:port)/dbname?parseTime=true
	DSN string `yaml:"dsn,omitempty" json:"dsn,omitempty"`

	// MaxConns is the maximum number of connections (pool size), default 10
	MaxConns int32 `yaml:"maxConns,omitempty" json:"maxConns,omitempty"`

	// MinConns is the minimum number of connections (only for PostgreSQL), default 2
	MinConns int32 `yaml:"minConns,omitempty" json:"minConns,omitempty"`
}

// IsEnabled returns true if database is configured
func (c Config) IsEnabled() bool {
	return c.DSN != ""
}

// DBTX is the interface for database/sql operations (used by sqlc)
type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

// NewPostgres creates a PostgreSQL connection pool for sqlc
func NewPostgres(ctx context.Context, c Config) (*pgxpool.Pool, error) {
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

// MustNewPostgres creates a PostgreSQL connection pool for sqlc or panics
func MustNewPostgres(ctx context.Context, c Config) *pgxpool.Pool {
	pool, err := NewPostgres(ctx, c)
	if err != nil {
		panic(err)
	}
	if pool == nil {
		panic("sqlc: postgres config not enabled")
	}
	return pool
}

// NewMySQL creates a MySQL connection for sqlc
func NewMySQL(c Config) (*sql.DB, error) {
	if !c.IsEnabled() {
		return nil, nil
	}

	db, err := sql.Open("mysql", c.DSN)
	if err != nil {
		return nil, err
	}

	// Apply defaults
	maxConns := c.MaxConns
	if maxConns == 0 {
		maxConns = 10
	}

	db.SetMaxOpenConns(int(maxConns))

	return db, nil
}

// MustNewMySQL creates a MySQL connection for sqlc or panics
func MustNewMySQL(c Config) *sql.DB {
	db, err := NewMySQL(c)
	if err != nil {
		panic(err)
	}
	if db == nil {
		panic("sqlc: mysql config not enabled")
	}
	return db
}