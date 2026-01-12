// Package metric provides Prometheus metrics server and utilities.
package metric

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/pprof"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/ssgohq/goten-core/logx"
)

var (
	once          sync.Once
	started       atomic.Bool
	defaultServer *Server
)

// Server is a standalone HTTP server for Prometheus metrics.
type Server struct {
	config Config
	mux    *http.ServeMux
	routes []string
	ready  atomic.Bool
}

// NewServer creates a new metrics server.
func NewServer(cfg Config) *Server {
	cfg.SetDefaults()
	return &Server{
		config: cfg,
		mux:    http.NewServeMux(),
	}
}

func (s *Server) addRoutes() {
	s.handleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(s.routes); err != nil {
			logx.Errorw("Failed to encode routes", "error", err)
		}
	})

	s.handleFunc(s.config.HealthPath, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(s.config.HealthResponse))
	})

	s.handleFunc(s.config.ReadyPath, func(w http.ResponseWriter, _ *http.Request) {
		if s.ready.Load() {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ready"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("not ready"))
		}
	})

	if s.config.EnableMetrics {
		s.handleFunc(s.config.MetricsPath, promhttp.Handler().ServeHTTP)
	}

	if s.config.EnablePprof {
		s.handleFunc("/debug/pprof/", pprof.Index)
		s.handleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		s.handleFunc("/debug/pprof/profile", pprof.Profile)
		s.handleFunc("/debug/pprof/symbol", pprof.Symbol)
		s.handleFunc("/debug/pprof/trace", pprof.Trace)
	}
}

func (s *Server) handleFunc(pattern string, handler http.HandlerFunc) {
	s.mux.HandleFunc(pattern, handler)
	s.routes = append(s.routes, pattern)
}

// SetReady marks the service as ready for traffic.
func (s *Server) SetReady(ready bool) {
	s.ready.Store(ready)
}

// Start starts the metrics server in a goroutine.
func (s *Server) Start() {
	s.addRoutes()
	go func() {
		addr := s.config.Addr()
		logx.Infow("Starting metrics server",
			"addr", addr,
			"metrics", s.config.MetricsPath,
			"health", s.config.HealthPath,
			"ready", s.config.ReadyPath,
		)
		server := &http.Server{
			Addr:              addr,
			Handler:           s.mux,
			ReadHeaderTimeout: 10 * time.Second,
		}
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logx.Errorw("Metrics server error", "error", err)
		}
	}()
}

// Stop gracefully shuts down the metrics server.
func (s *Server) Stop(ctx context.Context) error {
	// For now, nothing to do since we start a new server each time
	return nil
}

// Name returns the service name for lifecycle management.
func (s *Server) Name() string {
	return "metrics-server"
}

// StartAgent starts the metric server if enabled.
// This is a singleton that will only start once.
func StartAgent(c Config) {
	if !c.IsEnabled() {
		return
	}

	once.Do(func() {
		defaultServer = NewServer(c)
		defaultServer.Start()
		started.Store(true)
	})
}

// SetReady marks the default metric server as ready for traffic.
func SetReady(ready bool) {
	if defaultServer != nil {
		defaultServer.SetReady(ready)
	}
}

// IsStarted returns true if the metric server has been started.
func IsStarted() bool {
	return started.Load()
}