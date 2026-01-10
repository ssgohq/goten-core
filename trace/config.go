// Package trace provides OpenTelemetry tracing integration.
// It simplifies setting up distributed tracing with support for
// OTLP, Jaeger, and stdout exporters.
package trace

import "time"

// Config represents the tracing configuration.
type Config struct {
	// Name is the service name for tracing.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`

	// Endpoint is the collector endpoint URL.
	// For OTLP: http://localhost:4318 or grpc://localhost:4317
	// For Jaeger: http://localhost:14268/api/traces
	Endpoint string `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`

	// Exporter specifies the exporter type: "otlp", "jaeger", or "stdout"
	// Default: "otlp"
	Exporter string `yaml:"exporter,omitempty" json:"exporter,omitempty"`

	// Protocol specifies the protocol for OTLP: "grpc" or "http"
	// Default: "http"
	Protocol string `yaml:"protocol,omitempty" json:"protocol,omitempty"`

	// SampleRate is the sampling rate (0.0 to 1.0).
	// Default: 1.0 (sample everything)
	SampleRate float64 `yaml:"sampleRate,omitempty" json:"sampleRate,omitempty"`

	// Enabled explicitly enables or disables tracing.
	// Default: true when Name and Endpoint are set.
	Enabled *bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`

	// Insecure disables TLS for the connection.
	Insecure bool `yaml:"insecure,omitempty" json:"insecure,omitempty"`

	// Headers are additional headers to send with traces.
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`

	// BatchTimeout is the maximum time to wait before sending a batch.
	// Default: 5s
	BatchTimeout time.Duration `yaml:"batchTimeout,omitempty" json:"batchTimeout,omitempty"`

	// ExportTimeout is the maximum time to wait for export.
	// Default: 30s
	ExportTimeout time.Duration `yaml:"exportTimeout,omitempty" json:"exportTimeout,omitempty"`

	// MaxExportBatchSize is the maximum number of spans to export in a batch.
	// Default: 512
	MaxExportBatchSize int `yaml:"maxExportBatchSize,omitempty" json:"maxExportBatchSize,omitempty"`
}

// IsEnabled returns true if tracing should be enabled.
func (c Config) IsEnabled() bool {
	if c.Enabled != nil {
		return *c.Enabled
	}
	return c.Name != "" && c.Endpoint != ""
}

// SetDefaults applies default values to the config.
func (c *Config) SetDefaults() {
	if c.Exporter == "" {
		c.Exporter = "otlp"
	}
	if c.Protocol == "" {
		c.Protocol = "http"
	}
	if c.SampleRate == 0 {
		c.SampleRate = 1.0
	}
	if c.BatchTimeout == 0 {
		c.BatchTimeout = 5 * time.Second
	}
	if c.ExportTimeout == 0 {
		c.ExportTimeout = 30 * time.Second
	}
	if c.MaxExportBatchSize == 0 {
		c.MaxExportBatchSize = 512
	}
}