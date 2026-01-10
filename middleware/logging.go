package middleware

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"

	"github.com/ssgohq/goten-core/logx"
	"github.com/ssgohq/goten-core/trace"
)

// RequestID returns a middleware that adds a request ID to the context and response headers.
func RequestID() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// Check for existing request ID
		requestID := string(c.Request.Header.Peek("X-Request-ID"))
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set request ID in context and response
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next(ctx)
	}
}

// AccessLog returns a middleware that logs HTTP requests.
func AccessLog() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		path := string(c.Request.URI().Path())
		method := string(c.Request.Method())

		// Process request
		c.Next(ctx)

		// Log request
		duration := time.Since(start)
		status := c.Response.StatusCode()

		fields := []interface{}{
			"method", method,
			"path", path,
			"status", status,
			"duration", duration.String(),
			"duration_ms", duration.Milliseconds(),
			"client_ip", c.ClientIP(),
		}

		// Add request ID if present
		if requestID, exists := c.Get("requestID"); exists {
			fields = append(fields, "request_id", requestID)
		}

		// Add trace ID if present
		if traceID := trace.TraceIDFromContext(ctx); traceID != "" {
			fields = append(fields, "trace_id", traceID)
		}

		// Log based on status code
		if status >= 500 {
			logx.Errorw("HTTP request", fields...)
		} else if status >= 400 {
			logx.Warnw("HTTP request", fields...)
		} else {
			logx.Infow("HTTP request", fields...)
		}
	}
}

// Recovery returns a middleware that recovers from panics.
func Recovery() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				logx.Errorw("Panic recovered",
					"panic", fmt.Sprintf("%v", r),
					"stack", string(stack),
					"path", string(c.Request.URI().Path()),
					"method", string(c.Request.Method()),
				)
				c.AbortWithStatus(500)
			}
		}()
		c.Next(ctx)
	}
}

// LoggingConfig configures the logging middleware.
type LoggingConfig struct {
	// SkipPaths is a list of paths to skip logging.
	SkipPaths []string `yaml:"skipPaths,omitempty" json:"skipPaths,omitempty"`

	// SlowThreshold logs a warning for slow requests.
	SlowThreshold time.Duration `yaml:"slowThreshold,omitempty" json:"slowThreshold,omitempty"`
}

// AccessLogWithConfig returns a customized access log middleware.
func AccessLogWithConfig(cfg LoggingConfig) app.HandlerFunc {
	skipMap := make(map[string]bool, len(cfg.SkipPaths))
	for _, p := range cfg.SkipPaths {
		skipMap[p] = true
	}

	return func(ctx context.Context, c *app.RequestContext) {
		path := string(c.Request.URI().Path())

		// Skip logging for certain paths
		if skipMap[path] {
			c.Next(ctx)
			return
		}

		start := time.Now()
		method := string(c.Request.Method())

		// Process request
		c.Next(ctx)

		// Log request
		duration := time.Since(start)
		status := c.Response.StatusCode()

		fields := []interface{}{
			"method", method,
			"path", path,
			"status", status,
			"duration", duration.String(),
			"duration_ms", duration.Milliseconds(),
			"client_ip", c.ClientIP(),
		}

		// Add request ID if present
		if requestID, exists := c.Get("requestID"); exists {
			fields = append(fields, "request_id", requestID)
		}

		// Add trace ID if present
		if traceID := trace.TraceIDFromContext(ctx); traceID != "" {
			fields = append(fields, "trace_id", traceID)
		}

		// Check slow threshold
		if cfg.SlowThreshold > 0 && duration > cfg.SlowThreshold {
			logx.Warnw("Slow HTTP request", fields...)
		} else if status >= 500 {
			logx.Errorw("HTTP request", fields...)
		} else if status >= 400 {
			logx.Warnw("HTTP request", fields...)
		} else {
			logx.Infow("HTTP request", fields...)
		}
	}
}