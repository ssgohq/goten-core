package lifecycle

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/ssgohq/goten-core/logx"
)

// HealthStatus represents the health status of a component.
type HealthStatus string

const (
	// HealthStatusUp indicates the component is healthy.
	HealthStatusUp HealthStatus = "up"
	// HealthStatusDown indicates the component is unhealthy.
	HealthStatusDown HealthStatus = "down"
	// HealthStatusDegraded indicates partial functionality.
	HealthStatusDegraded HealthStatus = "degraded"
)

// HealthCheck is a function that returns the health of a component.
type HealthCheck func(ctx context.Context) HealthStatus

// ComponentHealth represents the health of a single component.
type ComponentHealth struct {
	Status    HealthStatus   `json:"status"`
	Details   map[string]any `json:"details,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// HealthResponse represents the overall health response.
type HealthResponse struct {
	Status     HealthStatus               `json:"status"`
	Components map[string]ComponentHealth `json:"components,omitempty"`
	Timestamp  time.Time                  `json:"timestamp"`
}

// HealthManager manages health checks for services.
type HealthManager struct {
	checks map[string]HealthCheck
	mu     sync.RWMutex
}

// NewHealthManager creates a new health manager.
func NewHealthManager() *HealthManager {
	return &HealthManager{
		checks: make(map[string]HealthCheck),
	}
}

// Register adds a health check for a component.
func (h *HealthManager) Register(name string, check HealthCheck) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks[name] = check
}

// Check runs all health checks and returns the overall health.
func (h *HealthManager) Check(ctx context.Context) HealthResponse {
	h.mu.RLock()
	defer h.mu.RUnlock()

	response := HealthResponse{
		Status:     HealthStatusUp,
		Components: make(map[string]ComponentHealth),
		Timestamp:  time.Now(),
	}

	for name, check := range h.checks {
		status := check(ctx)
		response.Components[name] = ComponentHealth{
			Status:    status,
			Timestamp: time.Now(),
		}

		// Update overall status
		if status == HealthStatusDown {
			response.Status = HealthStatusDown
		} else if status == HealthStatusDegraded && response.Status == HealthStatusUp {
			response.Status = HealthStatusDegraded
		}
	}

	return response
}

// HTTPHandler returns an HTTP handler for health checks.
// Returns 200 for healthy, 503 for unhealthy.
func (h *HealthManager) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		response := h.Check(ctx)

		w.Header().Set("Content-Type", "application/json")

		var statusCode int
		switch response.Status {
		case HealthStatusUp:
			statusCode = http.StatusOK
		case HealthStatusDegraded:
			statusCode = http.StatusOK
		default:
			statusCode = http.StatusServiceUnavailable
		}
		w.WriteHeader(statusCode)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			logx.Errorw("Failed to encode health response", "error", err)
		}
	}
}

// LivenessHandler returns an HTTP handler for liveness probe.
// Always returns 200 if the process is running.
func LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"up"}`))
	}
}

// ReadinessHandler returns an HTTP handler that uses the health manager.
func (h *HealthManager) ReadinessHandler() http.HandlerFunc {
	return h.HTTPHandler()
}