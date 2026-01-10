// Package srpc provides a thin wrapper around CloudWeGo Kitex for RPC services.
// It offers simplified configuration, pre-built middleware suites, and integration
// with goten-core's infrastructure (logx, trace).
package srpc

import (
	"time"

	"github.com/ssgo/goten-core/trace"
)

// ServerConfig represents RPC server configuration.
type ServerConfig struct {
	// Name is the service name for registration and identification.
	Name string `yaml:"name" json:"name"`
	// Host is the address to bind to. Default: "0.0.0.0"
	Host string `yaml:"host,omitempty" json:"host,omitempty"`
	// Port is the port to listen on. Default: 8888
	Port int `yaml:"port,omitempty" json:"port,omitempty"`

	// Discovery configuration for service registration.
	Discovery DiscoveryConfig `yaml:"discovery,omitempty" json:"discovery,omitempty"`

	// Trace configuration for OpenTelemetry tracing.
	Trace trace.Config `yaml:"trace,omitempty" json:"trace,omitempty"`

	// Timeout settings for connections.
	Timeout TimeoutConfig `yaml:"timeout,omitempty" json:"timeout,omitempty"`

	// MaxConnections limits the maximum number of concurrent connections.
	// 0 means unlimited.
	MaxConnections int `yaml:"maxConnections,omitempty" json:"maxConnections,omitempty"`
	// MaxQPS limits the maximum queries per second. 0 means unlimited.
	MaxQPS int `yaml:"maxQps,omitempty" json:"maxQps,omitempty"`

	// EnableRecovery enables panic recovery middleware. Default: true
	EnableRecovery bool `yaml:"enableRecovery,omitempty" json:"enableRecovery,omitempty"`
	// EnableAccessLog enables request/response logging. Default: false
	EnableAccessLog bool `yaml:"enableAccessLog,omitempty" json:"enableAccessLog,omitempty"`
}

// SetDefaults applies sensible defaults to the configuration.
func (c *ServerConfig) SetDefaults() {
	if c.Host == "" {
		c.Host = "0.0.0.0"
	}
	if c.Port == 0 {
		c.Port = 8888
	}
	if c.Timeout.Read == 0 {
		c.Timeout.Read = 3 * time.Second
	}
	if c.Timeout.Write == 0 {
		c.Timeout.Write = 3 * time.Second
	}
	if c.Timeout.Idle == 0 {
		c.Timeout.Idle = 60 * time.Second
	}
	c.Discovery.SetDefaults()
}

// TimeoutConfig represents timeout settings.
type TimeoutConfig struct {
	// Read timeout for reading request.
	Read time.Duration `yaml:"read,omitempty" json:"read,omitempty"`
	// Write timeout for writing response.
	Write time.Duration `yaml:"write,omitempty" json:"write,omitempty"`
	// Idle timeout for idle connections.
	Idle time.Duration `yaml:"idle,omitempty" json:"idle,omitempty"`
}

// DiscoveryConfig represents service discovery configuration.
type DiscoveryConfig struct {
	// Type specifies the discovery backend: "consul", "etcd", "direct", or "none".
	// Default: "none"
	Type string `yaml:"type,omitempty" json:"type,omitempty"`
	// Consul configuration (used when Type is "consul").
	Consul ConsulConfig `yaml:"consul,omitempty" json:"consul,omitempty"`
	// Etcd configuration (used when Type is "etcd").
	Etcd EtcdConfig `yaml:"etcd,omitempty" json:"etcd,omitempty"`
}

// SetDefaults applies sensible defaults to the discovery configuration.
func (c *DiscoveryConfig) SetDefaults() {
	if c.Type == "" {
		c.Type = "none"
	}
	c.Consul.SetDefaults()
	c.Etcd.SetDefaults()
}

// ConsulConfig represents Consul-specific configuration.
type ConsulConfig struct {
	// Address is the Consul agent address. Default: "localhost:8500"
	Address string `yaml:"address,omitempty" json:"address,omitempty"`
	// Token is the ACL token for authentication.
	Token string `yaml:"token,omitempty" json:"token,omitempty"`
	// Datacenter specifies the datacenter to use.
	Datacenter string `yaml:"datacenter,omitempty" json:"datacenter,omitempty"`
	// HealthCheck enables gRPC health checking. Default: true
	HealthCheck bool `yaml:"healthCheck,omitempty" json:"healthCheck,omitempty"`
	// CheckTimeout is the health check timeout. Default: 5s
	CheckTimeout time.Duration `yaml:"checkTimeout,omitempty" json:"checkTimeout,omitempty"`
	// Interval is the health check interval. Default: 10s
	Interval time.Duration `yaml:"interval,omitempty" json:"interval,omitempty"`
	// DeregisterAfter is the duration after which a service is deregistered
	// if it's been critical. Default: 1m
	DeregisterAfter time.Duration `yaml:"deregisterAfter,omitempty" json:"deregisterAfter,omitempty"`
	// Tags are the service tags for filtering.
	Tags []string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// SetDefaults applies sensible defaults to the Consul configuration.
func (c *ConsulConfig) SetDefaults() {
	if c.Address == "" {
		c.Address = "localhost:8500"
	}
	if c.CheckTimeout == 0 {
		c.CheckTimeout = 5 * time.Second
	}
	if c.Interval == 0 {
		c.Interval = 10 * time.Second
	}
	if c.DeregisterAfter == 0 {
		c.DeregisterAfter = time.Minute
	}
}

// EtcdConfig represents etcd-specific configuration.
type EtcdConfig struct {
	// Hosts is the list of etcd endpoints. Default: ["localhost:2379"]
	Hosts []string `yaml:"hosts,omitempty" json:"hosts,omitempty"`
	// Username for authentication.
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	// Password for authentication.
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
}

// SetDefaults applies sensible defaults to the etcd configuration.
func (c *EtcdConfig) SetDefaults() {
	if len(c.Hosts) == 0 {
		c.Hosts = []string{"localhost:2379"}
	}
}

// ClientConfig represents RPC client configuration.
type ClientConfig struct {
	// ServiceName is the target service name for discovery.
	ServiceName string `yaml:"serviceName" json:"serviceName"`

	// Endpoints are direct endpoints when not using discovery.
	// Format: ["host:port", "host:port"]
	Endpoints []string `yaml:"endpoints,omitempty" json:"endpoints,omitempty"`

	// Discovery configuration for service resolution.
	Discovery DiscoveryConfig `yaml:"discovery,omitempty" json:"discovery,omitempty"`

	// Timeout settings for RPC calls.
	Timeout ClientTimeoutConfig `yaml:"timeout,omitempty" json:"timeout,omitempty"`

	// Retry configuration.
	Retry RetryConfig `yaml:"retry,omitempty" json:"retry,omitempty"`

	// CircuitBreaker configuration.
	CircuitBreaker CircuitBreakerConfig `yaml:"circuitBreaker,omitempty" json:"circuitBreaker,omitempty"`

	// LoadBalancer specifies the load balancing strategy.
	// Options: "roundrobin", "random", "weightedrandom", "consistenthash"
	// Default: "roundrobin"
	LoadBalancer string `yaml:"loadBalancer,omitempty" json:"loadBalancer,omitempty"`

	// Connection pool settings.
	// MaxIdlePerAddress is the maximum idle connections per address.
	MaxIdlePerAddress int `yaml:"maxIdlePerAddress,omitempty" json:"maxIdlePerAddress,omitempty"`
	// MaxIdleGlobal is the maximum idle connections globally.
	MaxIdleGlobal int `yaml:"maxIdleGlobal,omitempty" json:"maxIdleGlobal,omitempty"`
	// MaxIdleTimeout is the maximum duration a connection can be idle.
	MaxIdleTimeout time.Duration `yaml:"maxIdleTimeout,omitempty" json:"maxIdleTimeout,omitempty"`
}

// SetDefaults applies sensible defaults to the client configuration.
func (c *ClientConfig) SetDefaults() {
	c.Discovery.SetDefaults()
	c.Timeout.SetDefaults()
	c.Retry.SetDefaults()
	c.CircuitBreaker.SetDefaults()

	if c.LoadBalancer == "" {
		c.LoadBalancer = "roundrobin"
	}
	if c.MaxIdlePerAddress == 0 {
		c.MaxIdlePerAddress = 10
	}
	if c.MaxIdleGlobal == 0 {
		c.MaxIdleGlobal = 100
	}
	if c.MaxIdleTimeout == 0 {
		c.MaxIdleTimeout = 30 * time.Second
	}
}

// ClientTimeoutConfig represents client timeout settings.
type ClientTimeoutConfig struct {
	// RPC is the timeout for the entire RPC call. Default: 3s
	RPC time.Duration `yaml:"rpc,omitempty" json:"rpc,omitempty"`
	// Connect is the timeout for establishing connection. Default: 1s
	Connect time.Duration `yaml:"connect,omitempty" json:"connect,omitempty"`
	// ReadWrite is the timeout for read/write operations.
	ReadWrite time.Duration `yaml:"readWrite,omitempty" json:"readWrite,omitempty"`
}

// SetDefaults applies sensible defaults to the timeout configuration.
func (c *ClientTimeoutConfig) SetDefaults() {
	if c.RPC == 0 {
		c.RPC = 3 * time.Second
	}
	if c.Connect == 0 {
		c.Connect = time.Second
	}
}

// RetryConfig represents retry configuration.
type RetryConfig struct {
	// Enabled enables retry on failure. Default: false
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	// MaxRetries is the maximum number of retry attempts. Default: 2
	MaxRetries int `yaml:"maxRetries,omitempty" json:"maxRetries,omitempty"`
	// Delay is the initial delay between retries. Default: 100ms
	Delay time.Duration `yaml:"delay,omitempty" json:"delay,omitempty"`
	// MaxDelay is the maximum delay between retries. Default: 1s
	MaxDelay time.Duration `yaml:"maxDelay,omitempty" json:"maxDelay,omitempty"`
	// RetryOn specifies which error types to retry on.
	// Options: "timeout", "connection", "server_error"
	RetryOn []string `yaml:"retryOn,omitempty" json:"retryOn,omitempty"`
}

// SetDefaults applies sensible defaults to the retry configuration.
func (c *RetryConfig) SetDefaults() {
	if c.MaxRetries == 0 {
		c.MaxRetries = 2
	}
	if c.Delay == 0 {
		c.Delay = 100 * time.Millisecond
	}
	if c.MaxDelay == 0 {
		c.MaxDelay = time.Second
	}
}

// CircuitBreakerConfig represents circuit breaker configuration.
type CircuitBreakerConfig struct {
	// Enabled enables the circuit breaker. Default: false
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	// ErrorRate is the error rate threshold (0.0-1.0) to trip the breaker.
	// Default: 0.5 (50%)
	ErrorRate float64 `yaml:"errorRate,omitempty" json:"errorRate,omitempty"`
	// MinSamples is the minimum number of samples before the breaker can trip.
	// Default: 200
	MinSamples int64 `yaml:"minSamples,omitempty" json:"minSamples,omitempty"`
}

// SetDefaults applies sensible defaults to the circuit breaker configuration.
func (c *CircuitBreakerConfig) SetDefaults() {
	if c.ErrorRate == 0 {
		c.ErrorRate = 0.5
	}
	if c.MinSamples == 0 {
		c.MinSamples = 200
	}
}