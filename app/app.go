// Package app provides high-level application abstraction for services.
// It integrates lifecycle management, logging, tracing, and HTTP/RPC servers
// into a unified structure for building microservices.
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	hertztracing "github.com/hertz-contrib/obs-opentelemetry/tracing"

	"github.com/ssgohq/goten-core/lifecycle"
	"github.com/ssgohq/goten-core/logx"
	"github.com/ssgohq/goten-core/trace"
)

// HookName defines standard hook names for application lifecycle.
type HookName = lifecycle.HookName

const (
	// HookBeforeStart is called before the application starts.
	HookBeforeStart HookName = "before_start"
	// HookAfterStart is called after the application has started.
	HookAfterStart HookName = "after_start"
	// HookBeforeStop is called before the application stops.
	HookBeforeStop HookName = "before_stop"
	// HookAfterStop is called after the application has stopped.
	HookAfterStop HookName = "after_stop"
)

// Config represents the application configuration.
type Config struct {
	// Name is the application name (used for logging and tracing).
	Name string `yaml:"name" json:"name"`
	// Version is the application version.
	Version string `yaml:"version,omitempty" json:"version,omitempty"`
	// Env is the environment (development, staging, production).
	Env string `yaml:"env,omitempty" json:"env,omitempty"`

	// EnableTracing enables OpenTelemetry tracing.
	EnableTracing bool `yaml:"enableTracing,omitempty" json:"enableTracing,omitempty"`

	// Trace configuration
	Trace trace.Config `yaml:"trace,omitempty" json:"trace,omitempty"`

	// GracePeriod is the time to wait before forceful shutdown.
	GracePeriod time.Duration `yaml:"gracePeriod,omitempty" json:"gracePeriod,omitempty"`

	// StopTimeout is the maximum shutdown time.
	StopTimeout time.Duration `yaml:"stopTimeout,omitempty" json:"stopTimeout,omitempty"`
}

// SetDefaults applies default values to the configuration.
func (c *Config) SetDefaults() {
	if c.Env == "" {
		c.Env = os.Getenv("ENV")
		if c.Env == "" {
			c.Env = "development"
		}
	}
	if c.GracePeriod == 0 {
		c.GracePeriod = 5 * time.Second
	}
	if c.StopTimeout == 0 {
		c.StopTimeout = 30 * time.Second
	}
}

// App represents a goten application with integrated services.
type App struct {
	config        Config
	manager       *lifecycle.Manager
	services      []lifecycle.Service
	tracingEnabled bool
	traceShutdown func(context.Context) error
	mu            sync.Mutex
}

// New creates a new App with the given configuration.
func New(cfg Config) *App {
	cfg.SetDefaults()

	lc := lifecycle.LifecycleConfig{
		ShutdownTimeout: cfg.StopTimeout,
		GracePeriod:     cfg.GracePeriod,
	}
	return &App{
		config:   cfg,
		manager:  lifecycle.NewManager(lc),
		services: make([]lifecycle.Service, 0),
	}
}

// Name returns the application name.
func (a *App) Name() string {
	return a.config.Name
}

// Version returns the application version.
func (a *App) Version() string {
	return a.config.Version
}

// AddService adds a service to be managed by the application.
func (a *App) AddService(svc lifecycle.Service) *App {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.services = append(a.services, svc)
	a.manager.Register(svc)
	return a
}

// AddHook adds a lifecycle hook.
func (a *App) AddHook(name HookName, fn func(ctx context.Context) error) *App {
	a.manager.AddHook(lifecycle.Hook{
		Name:  name,
		Phase: lifecycle.HookPhaseStartup,
		Fn:    fn,
	})
	return a
}

// Run starts all services and blocks until shutdown.
func (a *App) Run(ctx context.Context) error {
	logx.Infow("Starting application",
		"name", a.config.Name,
		"version", a.config.Version,
		"env", a.config.Env,
	)

	// Initialize tracing if enabled
	if a.config.EnableTracing && a.config.Trace.IsEnabled() {
		shutdown, err := trace.StartAgent(a.config.Trace)
		if err != nil {
			return fmt.Errorf("failed to start trace agent: %w", err)
		}
		a.traceShutdown = shutdown
		a.tracingEnabled = true
		logx.Infow("Tracing enabled", "endpoint", a.config.Trace.Endpoint)
	}

	// Start all services
	if err := a.manager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logx.Infow("Shutdown signal received, stopping application...")

	// Stop all services
	if err := a.manager.Stop(ctx); err != nil {
		logx.Errorw("Error stopping services", "error", err)
	}

	// Shutdown tracing if it was enabled
	if a.tracingEnabled && a.traceShutdown != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if traceErr := a.traceShutdown(shutdownCtx); traceErr != nil {
			logx.Errorw("Trace shutdown error", "error", traceErr)
		}
	}

	logx.Infow("Application shutdown complete")
	return nil
}

// Stop stops all services gracefully.
func (a *App) Stop() error {
	return a.manager.Stop(context.Background())
}

// MustRun starts the application and exits on error.
func (a *App) MustRun(ctx context.Context) {
	if err := a.Run(ctx); err != nil {
		logx.Fatalw("Application error", "error", err)
	}
}

// HertzOption configures the Hertz server creation.
type HertzOption func(*hertzOptions)

type hertzOptions struct {
	enableTracing  bool
	maxRequestBody int
	serverOptions  []config.Option
}

// WithTracing enables OpenTelemetry tracing middleware on the Hertz server.
func WithTracing(enable bool) HertzOption {
	return func(o *hertzOptions) {
		o.enableTracing = enable
	}
}

// WithMaxRequestBody sets the maximum request body size in bytes.
func WithMaxRequestBody(size int) HertzOption {
	return func(o *hertzOptions) {
		o.maxRequestBody = size
	}
}

// WithServerOptions adds additional Hertz server options.
func WithServerOptions(opts ...config.Option) HertzOption {
	return func(o *hertzOptions) {
		o.serverOptions = append(o.serverOptions, opts...)
	}
}

// NewHertzServer creates a pre-configured Hertz HTTP server with optional
// tracing middleware. This is the recommended way to create a Hertz server
// as it handles all the boilerplate configuration automatically.
//
// Example:
//
//	h := app.NewHertzServer(":8080", app.WithTracing(true))
//	handler.RegisterHandlers(h, svcCtx)
//
//	app.New(cfg).AddHTTP("http", h, ":8080").MustRun(ctx)
func NewHertzServer(addr string, opts ...HertzOption) *server.Hertz {
	options := &hertzOptions{
		maxRequestBody: 20 << 20, // 20MB default
	}
	for _, opt := range opts {
		opt(options)
	}

	// Build base server options
	baseOpts := []config.Option{
		server.WithHostPorts(addr),
		server.WithMaxRequestBodySize(options.maxRequestBody),
	}

	// Append custom server options
	baseOpts = append(baseOpts, options.serverOptions...)

	// Add tracing if enabled
	if options.enableTracing {
		tracer, tracerCfg := hertztracing.NewServerTracer()
		baseOpts = append(baseOpts, tracer)
		h := server.Default(baseOpts...)
		h.Use(hertztracing.ServerMiddleware(tracerCfg))
		return h
	}

	// Create server without tracing
	return server.Default(baseOpts...)
}

// WithLogger initializes the logger with the standard logx configuration.
// This is a convenience method that sets up logging based on environment.
func WithLogger(_ interface{}) {
	// No-op, use logx.Init() directly
}