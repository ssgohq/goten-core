package srpc

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/limit"
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	kitextracing "github.com/kitex-contrib/obs-opentelemetry/tracing"
	consul "github.com/kitex-contrib/registry-consul"

	"github.com/ssgo/goten-core/logx"
	"github.com/ssgo/goten-core/srpc/middleware"
)

// ServerBuilder helps construct Kitex server with common options.
type ServerBuilder struct {
	config   *ServerConfig
	options  []server.Option
	registry registry.Registry
}

// NewServerBuilder creates a new server builder with the given configuration.
func NewServerBuilder(config *ServerConfig) *ServerBuilder {
	config.SetDefaults()
	return &ServerBuilder{
		config:  config,
		options: make([]server.Option, 0),
	}
}

// Build returns all configured server options ready to pass to the generated
// Kitex service NewServer function.
//
// Example:
//
//	builder := srpc.NewServerBuilder(&config)
//	svr := userservice.NewServer(&impl, builder.Build()...)
func (b *ServerBuilder) Build() []server.Option {
	opts := make([]server.Option, 0, 10)

	// 1. Basic service info
	opts = append(opts, server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
		ServiceName: b.config.Name,
	}))

	// 2. Address binding
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", b.config.Host, b.config.Port))
	if err != nil {
		logx.Errorw("Failed to resolve server address", "host", b.config.Host, "port", b.config.Port, "error", err)
	} else {
		opts = append(opts, server.WithServiceAddr(addr))
	}

	// 3. Connection and QPS limits
	if b.config.MaxConnections > 0 || b.config.MaxQPS > 0 {
		opts = append(opts, server.WithLimit(&limit.Option{
			MaxConnections: b.config.MaxConnections,
			MaxQPS:         b.config.MaxQPS,
		}))
	}

	// 4. Service registry
	if reg := b.buildRegistry(); reg != nil {
		opts = append(opts, server.WithRegistry(reg))
		b.registry = reg
	}

	// 5. OpenTelemetry tracing suite
	if b.config.Trace.IsEnabled() {
		opts = append(opts, server.WithSuite(kitextracing.NewServerSuite()))
	}

	// 6. Recovery middleware
	if b.config.EnableRecovery {
		opts = append(opts, server.WithMiddleware(middleware.Recovery()))
	}

	// 7. Access logging middleware
	if b.config.EnableAccessLog {
		opts = append(opts, server.WithMiddleware(middleware.AccessLog()))
	}

	// 8. User-provided options
	opts = append(opts, b.options...)

	return opts
}

// WithOption adds a custom server option.
func (b *ServerBuilder) WithOption(opt server.Option) *ServerBuilder {
	b.options = append(b.options, opt)
	return b
}

// WithMiddleware adds a middleware to the server.
func (b *ServerBuilder) WithMiddleware(mw endpoint.Middleware) *ServerBuilder {
	b.options = append(b.options, server.WithMiddleware(mw))
	return b
}

// buildRegistry creates a service registry based on configuration.
func (b *ServerBuilder) buildRegistry() registry.Registry {
	switch b.config.Discovery.Type {
	case "consul":
		return b.buildConsulRegistry()
	case "etcd":
		return b.buildEtcdRegistry()
	default:
		return nil
	}
}

// buildConsulRegistry creates a Consul registry.
func (b *ServerBuilder) buildConsulRegistry() registry.Registry {
	cfg := b.config.Discovery.Consul

	r, err := consul.NewConsulRegister(cfg.Address)
	if err != nil {
		logx.Errorw("Failed to create Consul registry", "address", cfg.Address, "error", err)
		return nil
	}

	logx.Infow("Consul registry created", "address", cfg.Address)
	return r
}

// buildEtcdRegistry creates an etcd registry.
// Note: Requires github.com/kitex-contrib/registry-etcd
func (b *ServerBuilder) buildEtcdRegistry() registry.Registry {
	// TODO: Implement etcd registry when needed
	// cfg := b.config.Discovery.Etcd
	// r, err := etcd.NewEtcdRegistry(cfg.Hosts)
	logx.Warnw("Etcd registry is not yet implemented, falling back to no registry")
	return nil
}

// Server wraps a Kitex server with additional lifecycle management.
type Server struct {
	kitexServer server.Server
	config      *ServerConfig
	registry    registry.Registry //nolint:unused // reserved for future service deregistration
}

// NewServer creates a Server wrapper around a Kitex server.
// This enables additional lifecycle management like graceful shutdown.
//
// Example:
//
//	builder := srpc.NewServerBuilder(&config)
//	kitexSvr := userservice.NewServer(&impl, builder.Build()...)
//	svr := srpc.NewServer(kitexSvr, &config)
//	if err := svr.Run(); err != nil {
//	    log.Fatal(err)
//	}
func NewServer(kitexServer server.Server, config *ServerConfig) *Server {
	return &Server{
		kitexServer: kitexServer,
		config:      config,
	}
}

// Run starts the server and blocks until shutdown signal is received.
// It handles graceful shutdown automatically.
func (s *Server) Run() error {
	logx.Infow("Starting RPC server",
		"name", s.config.Name,
		"host", s.config.Host,
		"port", s.config.Port,
		"discovery", s.config.Discovery.Type,
	)

	return RunWithGracefulShutdown(s.kitexServer)
}

// Stop stops the server gracefully.
func (s *Server) Stop() error {
	return s.kitexServer.Stop()
}

// RunWithGracefulShutdown starts a Kitex server and handles graceful shutdown
// on SIGINT and SIGTERM signals.
func RunWithGracefulShutdown(svr server.Server) error {
	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		if err := svr.Run(); err != nil {
			errCh <- err
		}
	}()

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-sigCh:
		logx.Infow("Received shutdown signal", "signal", sig)
		return svr.Stop()
	}
}

// MustRun starts the server and panics if it fails.
// Useful for main() functions.
func MustRun(svr server.Server) {
	if err := RunWithGracefulShutdown(svr); err != nil {
		logx.Fatalw("Server failed", "error", err)
	}
}

// StartServer is a convenience function that creates a ServerBuilder,
// builds options, and returns them ready for use.
//
// Example:
//
//	opts := srpc.StartServer(&config)
//	svr := userservice.NewServer(&impl, opts...)
//	srpc.MustRun(svr)
func StartServer(config *ServerConfig) []server.Option {
	return NewServerBuilder(config).Build()
}

// WithTracing returns a Kitex server suite for OpenTelemetry tracing.
// This is automatically included when trace config has name and endpoint set.
func WithTracing() server.Option {
	return server.WithSuite(kitextracing.NewServerSuite())
}

// ShutdownHook represents a function to run during shutdown.
type ShutdownHook func(ctx context.Context) error

// RunWithHooks starts a server with custom shutdown hooks.
func RunWithHooks(svr server.Server, hooks ...ShutdownHook) error {
	errCh := make(chan error, 1)
	go func() {
		if err := svr.Run(); err != nil {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-sigCh:
		logx.Infow("Received shutdown signal, running hooks", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Run shutdown hooks
		for _, hook := range hooks {
			if err := hook(ctx); err != nil {
				logx.Warnw("Shutdown hook failed", "error", err)
			}
		}

		return svr.Stop()
	}
}