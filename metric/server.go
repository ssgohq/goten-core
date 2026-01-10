// Package metric provides Prometheus metrics server and utilities.
package metric

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/ssgohq/goten-core/logx"
)

// Server is a standalone HTTP server for Prometheus metrics.
type Server struct {
	config ServerConfig
	server *http.Server
}

// NewServer creates a new metrics server.
func NewServer(cfg ServerConfig) *Server {
	cfg.SetDefaults()

	mux := http.NewServeMux()
	mux.Handle(cfg.Path, promhttp.Handler())
	mux.HandleFunc(cfg.HealthPath, healthHandler)
	mux.HandleFunc(cfg.ReadyPath, readyHandler)

	return &Server{
		config: cfg,
		server: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}
}

// Start starts the metrics server in a goroutine.
func (s *Server) Start() {
	go func() {
		logx.Infow("Starting metrics server",
			"addr", s.server.Addr,
			"path", s.config.Path,
		)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logx.Errorw("Metrics server error", "error", err)
		}
	}()
}

// Stop gracefully shuts down the metrics server.
func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// Name returns the service name for lifecycle management.
func (s *Server) Name() string {
	return "metrics-server"
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func readyHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}