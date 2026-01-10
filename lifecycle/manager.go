package lifecycle

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/ssgohq/goten-core/logx"
)

// Manager orchestrates the lifecycle of multiple services.
// It handles graceful startup and shutdown, executing hooks at appropriate times.
type Manager struct {
	config   LifecycleConfig
	services []Service
	hooks    map[HookPhase][]Hook
	state    State
	mu       sync.RWMutex
}

// NewManager creates a new lifecycle manager.
func NewManager(config LifecycleConfig) *Manager {
	if config.ShutdownTimeout == 0 {
		config.ShutdownTimeout = 30 * time.Second
	}
	if config.GracePeriod == 0 {
		config.GracePeriod = 5 * time.Second
	}
	return &Manager{
		config:   config,
		services: make([]Service, 0),
		hooks: map[HookPhase][]Hook{
			HookPhaseStartup:  {},
			HookPhaseShutdown: {},
		},
		state: StateIdle,
	}
}

// Register adds a service to be managed.
func (m *Manager) Register(svc Service) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.services = append(m.services, svc)
}

// AddHook adds a lifecycle hook.
func (m *Manager) AddHook(hook Hook) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hooks[hook.Phase] = append(m.hooks[hook.Phase], hook)
}

// State returns the current state.
func (m *Manager) State() State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

// Start starts all registered services in order.
// It executes startup hooks before and after starting services.
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	m.state = StateStarting
	m.mu.Unlock()

	// Execute pre-start hooks
	if err := m.executeHooks(ctx, HookPhaseStartup, "before_start"); err != nil {
		m.setState(StateError)
		return fmt.Errorf("pre-start hooks failed: %w", err)
	}

	// Start services
	for _, svc := range m.services {
		logx.Infow("Starting service", "name", svc.Name())
		if err := svc.Start(ctx); err != nil {
			m.setState(StateError)
			return fmt.Errorf("service %s failed to start: %w", svc.Name(), err)
		}
		logx.Infow("Service started", "name", svc.Name())
	}

	// Execute post-start hooks
	if err := m.executeHooks(ctx, HookPhaseStartup, "after_start"); err != nil {
		logx.Warnw("Post-start hooks failed", "error", err)
	}

	m.setState(StateRunning)
	return nil
}

// Stop stops all registered services in reverse order.
// It executes shutdown hooks before and after stopping services.
func (m *Manager) Stop(ctx context.Context) error {
	m.mu.Lock()
	m.state = StateStopping
	m.mu.Unlock()

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, m.config.ShutdownTimeout)
	defer cancel()

	// Execute pre-stop hooks
	if err := m.executeHooks(timeoutCtx, HookPhaseShutdown, "before_stop"); err != nil {
		logx.Warnw("Pre-stop hooks failed", "error", err)
	}

	// Stop services in reverse order
	var stopErr error
	for i := len(m.services) - 1; i >= 0; i-- {
		svc := m.services[i]
		logx.Infow("Stopping service", "name", svc.Name())
		if err := svc.Stop(timeoutCtx); err != nil {
			logx.Errorw("Service failed to stop", "name", svc.Name(), "error", err)
			if stopErr == nil {
				stopErr = err
			}
		} else {
			logx.Infow("Service stopped", "name", svc.Name())
		}
	}

	// Execute post-stop hooks
	if err := m.executeHooks(timeoutCtx, HookPhaseShutdown, "after_stop"); err != nil {
		logx.Warnw("Post-stop hooks failed", "error", err)
	}

	m.setState(StateStopped)
	return stopErr
}

// executeHooks executes hooks for the given phase and name.
func (m *Manager) executeHooks(ctx context.Context, phase HookPhase, name string) error {
	m.mu.RLock()
	hooks := make([]Hook, 0)
	for _, h := range m.hooks[phase] {
		if h.Name == name {
			hooks = append(hooks, h)
		}
	}
	m.mu.RUnlock()

	// Sort by priority
	sort.Slice(hooks, func(i, j int) bool {
		return hooks[i].Priority < hooks[j].Priority
	})

	for _, hook := range hooks {
		if err := hook.Fn(ctx); err != nil {
			return fmt.Errorf("hook %s failed: %w", hook.Name, err)
		}
	}
	return nil
}

func (m *Manager) setState(state State) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state = state
}