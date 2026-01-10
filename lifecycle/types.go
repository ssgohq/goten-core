// Package lifecycle provides service lifecycle management for microservices.
// It handles graceful startup and shutdown, health checks, and lifecycle hooks.
package lifecycle

import (
	"context"
	"time"
)

// Service represents a managed service with start/stop lifecycle.
type Service interface {
	// Name returns the service name for logging and identification.
	Name() string
	// Start starts the service. It should block until the service is ready.
	Start(ctx context.Context) error
	// Stop stops the service gracefully.
	Stop(ctx context.Context) error
}

// HookPhase defines when a hook should be executed.
type HookPhase int

const (
	// HookPhaseStartup indicates a hook that runs during startup.
	HookPhaseStartup HookPhase = iota
	// HookPhaseShutdown indicates a hook that runs during shutdown.
	HookPhaseShutdown
)

// HookName is a string identifier for hooks.
type HookName = string

// Hook represents a lifecycle hook.
type Hook struct {
	// Name is a unique identifier for the hook.
	Name HookName
	// Phase indicates when this hook should run.
	Phase HookPhase
	// Priority determines execution order (lower runs first).
	Priority int
	// Fn is the hook function to execute.
	Fn func(ctx context.Context) error
}

// LifecycleConfig configures the lifecycle manager.
type LifecycleConfig struct {
	// ShutdownTimeout is the maximum time to wait for graceful shutdown.
	// Default: 30 seconds.
	ShutdownTimeout time.Duration `yaml:"shutdownTimeout,omitempty" json:"shutdownTimeout,omitempty"`
	// GracePeriod is the time to wait before forceful shutdown after timeout.
	// Default: 5 seconds.
	GracePeriod time.Duration `yaml:"gracePeriod,omitempty" json:"gracePeriod,omitempty"`
}

// State represents the current state of a service.
type State int

const (
	// StateIdle indicates the service is not running.
	StateIdle State = iota
	// StateStarting indicates the service is starting up.
	StateStarting
	// StateRunning indicates the service is running normally.
	StateRunning
	// StateStopping indicates the service is shutting down.
	StateStopping
	// StateStopped indicates the service has stopped.
	StateStopped
	// StateError indicates the service encountered an error.
	StateError
)

// String returns a string representation of the state.
func (s State) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateStarting:
		return "starting"
	case StateRunning:
		return "running"
	case StateStopping:
		return "stopping"
	case StateStopped:
		return "stopped"
	case StateError:
		return "error"
	default:
		return "unknown"
	}
}