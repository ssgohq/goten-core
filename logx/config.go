// Package logx provides a structured logging solution based on zap.
// It offers a simple API for application logging with support for
// different output formats, log levels, and contextual fields.
package logx

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config represents the logger configuration.
type Config struct {
	// Level is the minimum enabled logging level.
	// Supported values: debug, info, warn, error, dpanic, panic, fatal
	Level string `yaml:"level,omitempty" json:"level,omitempty"`

	// Format specifies the output format: "json" or "console"
	Format string `yaml:"format,omitempty" json:"format,omitempty"`

	// Development enables development mode with DPanic panic behavior
	// and more human-readable stack traces.
	Development bool `yaml:"development,omitempty" json:"development,omitempty"`

	// DisableCaller stops annotating logs with the calling function's file name and line number.
	DisableCaller bool `yaml:"disableCaller,omitempty" json:"disableCaller,omitempty"`

	// DisableStacktrace disables automatic stacktrace capturing.
	DisableStacktrace bool `yaml:"disableStacktrace,omitempty" json:"disableStacktrace,omitempty"`

	// OutputPaths is a list of paths to write logging output to.
	// Default: ["stdout"]
	OutputPaths []string `yaml:"outputPaths,omitempty" json:"outputPaths,omitempty"`

	// ErrorOutputPaths is a list of paths to write internal logger errors to.
	// Default: ["stderr"]
	ErrorOutputPaths []string `yaml:"errorOutputPaths,omitempty" json:"errorOutputPaths,omitempty"`

	// InitialFields are fields to add to every log entry.
	InitialFields map[string]interface{} `yaml:"initialFields,omitempty" json:"initialFields,omitempty"`
}

// DefaultConfig returns the default production configuration.
func DefaultConfig() Config {
	return Config{
		Level:         "info",
		Format:        "json",
		Development:   false,
		OutputPaths:   []string{"stdout"},
		InitialFields: nil,
	}
}

// DevelopmentConfig returns the default development configuration.
func DevelopmentConfig() Config {
	return Config{
		Level:             "debug",
		Format:            "console",
		Development:       true,
		DisableStacktrace: true,
		OutputPaths:       []string{"stdout"},
	}
}

// ConfigFromEnv creates a Config from environment variables.
// Supported variables:
//   - LOG_LEVEL: log level (debug, info, warn, error)
//   - LOG_FORMAT: output format (json, console)
//   - LOG_DEV: enable development mode (true, false)
func ConfigFromEnv() Config {
	cfg := DefaultConfig()

	if level := os.Getenv("LOG_LEVEL"); level != "" {
		cfg.Level = strings.ToLower(level)
	}

	if format := os.Getenv("LOG_FORMAT"); format != "" {
		cfg.Format = strings.ToLower(format)
	}

	if dev := os.Getenv("LOG_DEV"); dev == "true" || dev == "1" {
		cfg.Development = true
		if cfg.Level == "info" {
			cfg.Level = "debug"
		}
		if cfg.Format == "json" {
			cfg.Format = "console"
		}
	}

	return cfg
}

// toZapConfig converts Config to zap.Config.
func (c *Config) toZapConfig() zap.Config {
	level := zap.NewAtomicLevel()
	switch strings.ToLower(c.Level) {
	case "debug":
		level.SetLevel(zapcore.DebugLevel)
	case "info":
		level.SetLevel(zapcore.InfoLevel)
	case "warn", "warning":
		level.SetLevel(zapcore.WarnLevel)
	case "error":
		level.SetLevel(zapcore.ErrorLevel)
	case "dpanic":
		level.SetLevel(zapcore.DPanicLevel)
	case "panic":
		level.SetLevel(zapcore.PanicLevel)
	case "fatal":
		level.SetLevel(zapcore.FatalLevel)
	default:
		level.SetLevel(zapcore.InfoLevel)
	}

	outputPaths := c.OutputPaths
	if len(outputPaths) == 0 {
		outputPaths = []string{"stdout"}
	}

	errorOutputPaths := c.ErrorOutputPaths
	if len(errorOutputPaths) == 0 {
		errorOutputPaths = []string{"stderr"}
	}

	var encoderConfig zapcore.EncoderConfig
	if c.Development {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	} else {
		encoderConfig = zap.NewProductionEncoderConfig()
	}
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	encoding := "json"
	if c.Format == "console" {
		encoding = "console"
	}

	return zap.Config{
		Level:             level,
		Development:       c.Development,
		DisableCaller:     c.DisableCaller,
		DisableStacktrace: c.DisableStacktrace,
		Encoding:          encoding,
		EncoderConfig:     encoderConfig,
		OutputPaths:       outputPaths,
		ErrorOutputPaths:  errorOutputPaths,
		InitialFields:     c.InitialFields,
	}
}