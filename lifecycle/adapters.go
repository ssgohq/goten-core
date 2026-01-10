package lifecycle

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app/server"
	kitexserver "github.com/cloudwego/kitex/server"
)

// HertzAdapter wraps a Hertz server to implement the Service interface.
type HertzAdapter struct {
	name   string
	server *server.Hertz
}

// NewHertzAdapter creates a new Hertz service adapter.
func NewHertzAdapter(name string, h *server.Hertz) *HertzAdapter {
	return &HertzAdapter{
		name:   name,
		server: h,
	}
}

// Name returns the service name.
func (a *HertzAdapter) Name() string {
	return a.name
}

// Start starts the Hertz server.
func (a *HertzAdapter) Start(_ context.Context) error {
	go a.server.Spin()
	return nil
}

// Stop stops the Hertz server gracefully.
func (a *HertzAdapter) Stop(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}

// KitexAdapter wraps a Kitex server to implement the Service interface.
type KitexAdapter struct {
	name   string
	server kitexserver.Server
}

// NewKitexAdapter creates a new Kitex service adapter.
func NewKitexAdapter(name string, s kitexserver.Server) *KitexAdapter {
	return &KitexAdapter{
		name:   name,
		server: s,
	}
}

// Name returns the service name.
func (a *KitexAdapter) Name() string {
	return a.name
}

// Start starts the Kitex server.
func (a *KitexAdapter) Start(_ context.Context) error {
	go func() {
		_ = a.server.Run()
	}()
	return nil
}

// Stop stops the Kitex server gracefully.
func (a *KitexAdapter) Stop(_ context.Context) error {
	return a.server.Stop()
}

// FuncService creates a simple service from start/stop functions.
type FuncService struct {
	name    string
	startFn func(ctx context.Context) error
	stopFn  func(ctx context.Context) error
}

// NewFuncService creates a new function-based service.
//
// Example:
//
//	svc := lifecycle.NewFuncService("cleanup",
//	    func(ctx context.Context) error {
//	        // Start logic
//	        return nil
//	    },
//	    func(ctx context.Context) error {
//	        // Stop logic
//	        return nil
//	    },
//	)
func NewFuncService(
	name string,
	startFn func(ctx context.Context) error,
	stopFn func(ctx context.Context) error,
) *FuncService {
	return &FuncService{
		name:    name,
		startFn: startFn,
		stopFn:  stopFn,
	}
}

// Name returns the service name.
func (s *FuncService) Name() string {
	return s.name
}

// Start executes the start function.
func (s *FuncService) Start(ctx context.Context) error {
	if s.startFn != nil {
		return s.startFn(ctx)
	}
	return nil
}

// Stop executes the stop function.
func (s *FuncService) Stop(ctx context.Context) error {
	if s.stopFn != nil {
		return s.stopFn(ctx)
	}
	return nil
}