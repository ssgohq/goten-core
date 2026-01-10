// Package redis provides Redis client configuration and connection utilities.
package redis

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Config represents Redis connection configuration
type Config struct {
	// Host is the Redis server host. If empty, Redis is disabled.
	Host string `yaml:"host,omitempty" json:"host,omitempty"`

	// Port is the Redis server port, default 6379
	Port int `yaml:"port,omitempty" json:"port,omitempty"`

	// Password for Redis authentication
	Password string `yaml:"password,omitempty" json:"password,omitempty"`

	// DB is the database number, default 0
	DB int `yaml:"db,omitempty" json:"db,omitempty"`
}

// IsEnabled returns true if Redis is configured
func (c Config) IsEnabled() bool {
	return c.Host != ""
}

// Addr returns the Redis address in host:port format
func (c Config) Addr() string {
	port := c.Port
	if port == 0 {
		port = 6379
	}
	return fmt.Sprintf("%s:%d", c.Host, port)
}

// Options returns go-redis Options
func (c Config) Options() *redis.Options {
	return &redis.Options{
		Addr:     c.Addr(),
		Password: c.Password,
		DB:       c.DB,
	}
}

// New creates a new Redis client
func New(c Config) *redis.Client {
	if !c.IsEnabled() {
		return nil
	}
	return redis.NewClient(c.Options())
}

// MustNew creates a new Redis client or panics
func MustNew(c Config) *redis.Client {
	client := New(c)
	if client == nil {
		panic("redis: config not enabled")
	}
	return client
}