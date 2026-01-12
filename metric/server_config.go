package metric

import "fmt"

// Config is config for the metric/observability server.
// This is an alias for compatibility with templates.
type Config struct {
	// Enabled enables/disables the metrics server.
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`

	// Host is the address to bind to. Default: "0.0.0.0"
	Host string `yaml:"host,omitempty" json:"host,omitempty"`

	// Port is the port to listen on. Default: 6060
	Port int `yaml:"port,omitempty" json:"port,omitempty"`

	// MetricsPath is the metrics endpoint path. Default: "/metrics"
	MetricsPath string `yaml:"metricsPath,omitempty" json:"metricsPath,omitempty"`

	// HealthPath is the health check endpoint path. Default: "/healthz"
	HealthPath string `yaml:"healthPath,omitempty" json:"healthPath,omitempty"`

	// ReadyPath is the readiness check endpoint path. Default: "/readyz"
	ReadyPath string `yaml:"readyPath,omitempty" json:"readyPath,omitempty"`

	// HealthResponse is the response body for health checks. Default: "OK"
	HealthResponse string `yaml:"healthResponse,omitempty" json:"healthResponse,omitempty"`

	// EnableMetrics enables Prometheus metrics endpoint.
	EnableMetrics bool `yaml:"enableMetrics,omitempty" json:"enableMetrics,omitempty"`

	// EnablePprof enables pprof debug endpoints.
	EnablePprof bool `yaml:"enablePprof,omitempty" json:"enablePprof,omitempty"`
}

// SetDefaults applies default values.
func (c *Config) SetDefaults() {
	if c.Host == "" {
		c.Host = "0.0.0.0"
	}
	if c.Port == 0 {
		c.Port = 6060
	}
	if c.MetricsPath == "" {
		c.MetricsPath = "/metrics"
	}
	if c.HealthPath == "" {
		c.HealthPath = "/healthz"
	}
	if c.ReadyPath == "" {
		c.ReadyPath = "/readyz"
	}
	if c.HealthResponse == "" {
		c.HealthResponse = "OK"
	}
}

// Addr returns the server address in host:port format.
func (c *Config) Addr() string {
	host := c.Host
	if host == "" {
		host = "0.0.0.0"
	}
	return fmt.Sprintf("%s:%d", host, c.Port)
}

// IsEnabled returns true if the metric server should start.
func (c *Config) IsEnabled() bool {
	return c.Enabled && c.Port > 0
}