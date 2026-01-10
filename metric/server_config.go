package metric

// ServerConfig configures the Prometheus metrics server.
type ServerConfig struct {
	// Enabled enables/disables the metrics server.
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`

	// Host is the address to bind to. Default: "0.0.0.0"
	Host string `yaml:"host,omitempty" json:"host,omitempty"`

	// Port is the port to listen on. Default: 9090
	Port int `yaml:"port,omitempty" json:"port,omitempty"`

	// Path is the metrics endpoint path. Default: "/metrics"
	Path string `yaml:"path,omitempty" json:"path,omitempty"`

	// HealthPath is the health check endpoint path. Default: "/healthz"
	HealthPath string `yaml:"healthPath,omitempty" json:"healthPath,omitempty"`

	// ReadyPath is the readiness check endpoint path. Default: "/readyz"
	ReadyPath string `yaml:"readyPath,omitempty" json:"readyPath,omitempty"`
}

// SetDefaults applies default values.
func (c *ServerConfig) SetDefaults() {
	if c.Host == "" {
		c.Host = "0.0.0.0"
	}
	if c.Port == 0 {
		c.Port = 9090
	}
	if c.Path == "" {
		c.Path = "/metrics"
	}
	if c.HealthPath == "" {
		c.HealthPath = "/healthz"
	}
	if c.ReadyPath == "" {
		c.ReadyPath = "/readyz"
	}
}