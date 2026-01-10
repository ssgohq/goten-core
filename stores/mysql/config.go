// Package mysql provides MySQL database configuration and connection utilities.
package mysql

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Config represents MySQL connection configuration.
type Config struct {
	// DSN is the MySQL connection string.
	// Format: user:password@tcp(host:port)/dbname?parseTime=true
	DSN string `yaml:"dsn,omitempty" json:"dsn,omitempty"`

	// MaxOpenConns is the maximum number of open connections, default 10.
	MaxOpenConns int `yaml:"maxOpenConns,omitempty" json:"maxOpenConns,omitempty"`

	// MaxIdleConns is the maximum number of idle connections, default 5.
	MaxIdleConns int `yaml:"maxIdleConns,omitempty" json:"maxIdleConns,omitempty"`

	// ConnMaxLifetime is the maximum connection lifetime, default 1 hour.
	ConnMaxLifetime time.Duration `yaml:"connMaxLifetime,omitempty" json:"connMaxLifetime,omitempty"`

	// ConnMaxIdleTime is the maximum idle connection lifetime, default 30 minutes.
	ConnMaxIdleTime time.Duration `yaml:"connMaxIdleTime,omitempty" json:"connMaxIdleTime,omitempty"`
}

// IsEnabled returns true if MySQL is configured.
func (c Config) IsEnabled() bool {
	return c.DSN != ""
}

// SetDefaults applies default values.
func (c *Config) SetDefaults() {
	if c.MaxOpenConns == 0 {
		c.MaxOpenConns = 10
	}
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = 5
	}
	if c.ConnMaxLifetime == 0 {
		c.ConnMaxLifetime = time.Hour
	}
	if c.ConnMaxIdleTime == 0 {
		c.ConnMaxIdleTime = 30 * time.Minute
	}
}

// New creates a new MySQL connection pool.
func New(c Config) (*sql.DB, error) {
	if !c.IsEnabled() {
		return nil, nil
	}

	c.SetDefaults()

	db, err := sql.Open("mysql", c.DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(c.MaxOpenConns)
	db.SetMaxIdleConns(c.MaxIdleConns)
	db.SetConnMaxLifetime(c.ConnMaxLifetime)
	db.SetConnMaxIdleTime(c.ConnMaxIdleTime)

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// MustNew creates a new MySQL connection pool or panics.
func MustNew(c Config) *sql.DB {
	db, err := New(c)
	if err != nil {
		panic(err)
	}
	if db == nil {
		panic("mysql: config not enabled")
	}
	return db
}