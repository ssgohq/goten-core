package lifecycle

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ssgo/goten-core/logx"
)

// ServiceGroup manages multiple services with signal handling.
// It provides graceful shutdown on SIGINT/SIGTERM.
type ServiceGroup struct {
	services []Service
	config   LifecycleConfig
	manager  *Manager
	mu       sync.Mutex
}

// NewServiceGroup creates a new service group.
func NewServiceGroup(config LifecycleConfig) *ServiceGroup {
	return &ServiceGroup{
		services: make([]Service, 0),
		config:   config,
		manager:  NewManager(config),
	}
}

// Add adds a service to the group.
func (g *ServiceGroup) Add(svc Service) *ServiceGroup {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.services = append(g.services, svc)
	g.manager.Register(svc)
	return g
}

// Run starts all services and blocks until a shutdown signal is received.
// It handles SIGINT and SIGTERM for graceful shutdown.
func (g *ServiceGroup) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start services
	if err := g.manager.Start(ctx); err != nil {
		return err
	}

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	logx.Infow("Received shutdown signal", "signal", sig.String())

	// Stop services
	return g.manager.Stop(ctx)
}

// RunWithContext starts all services and blocks until context is cancelled
// or a shutdown signal is received.
func (g *ServiceGroup) RunWithContext(ctx context.Context) error {
	// Start services
	if err := g.manager.Start(ctx); err != nil {
		return err
	}

	// Wait for context cancellation or shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		logx.Infow("Context cancelled, shutting down")
	case sig := <-sigCh:
		logx.Infow("Received shutdown signal", "signal", sig.String())
	}

	// Stop services
	return g.manager.Stop(context.Background())
}

// Stop stops all services gracefully.
func (g *ServiceGroup) Stop() error {
	return g.manager.Stop(context.Background())
}