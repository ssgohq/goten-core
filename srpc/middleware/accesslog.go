package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/rpcinfo"

	"github.com/ssgohq/goten-core/logx"
)

// AccessLog returns a middleware that logs RPC access information.
func AccessLog() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req, resp interface{}) error {
			start := time.Now()

			// Extract RPC info
			ri := rpcinfo.GetRPCInfo(ctx)
			var method, caller, service string
			if ri != nil {
				if ri.Invocation() != nil {
					method = ri.Invocation().MethodName()
					service = ri.Invocation().ServiceName()
				}
				if ri.From() != nil {
					caller = ri.From().ServiceName()
				}
			}

			// Execute the request
			err := next(ctx, req, resp)

			// Log the access
			duration := time.Since(start)
			fields := []interface{}{
				"method", method,
				"service", service,
				"caller", caller,
				"duration", duration.String(),
				"duration_ms", duration.Milliseconds(),
			}

			if err != nil {
				fields = append(fields, "error", err.Error())
				logx.Warnw("RPC access", fields...)
			} else {
				logx.Infow("RPC access", fields...)
			}

			return err
		}
	}
}

// AccessLogWithConfig returns an access log middleware with custom configuration.
type AccessLogConfig struct {
	// SkipMethods is a list of methods to skip logging.
	SkipMethods []string
	// SlowThreshold is the threshold for slow request logging.
	SlowThreshold time.Duration
}

// AccessLogWithConfig returns a customized access log middleware.
func AccessLogWithConfig(cfg AccessLogConfig) endpoint.Middleware {
	skipMap := make(map[string]bool, len(cfg.SkipMethods))
	for _, m := range cfg.SkipMethods {
		skipMap[m] = true
	}

	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req, resp interface{}) error {
			start := time.Now()

			// Extract RPC info
			ri := rpcinfo.GetRPCInfo(ctx)
			var method, caller, service string
			if ri != nil {
				if ri.Invocation() != nil {
					method = ri.Invocation().MethodName()
					service = ri.Invocation().ServiceName()
				}
				if ri.From() != nil {
					caller = ri.From().ServiceName()
				}
			}

			// Skip logging for certain methods
			if skipMap[method] {
				return next(ctx, req, resp)
			}

			// Execute the request
			err := next(ctx, req, resp)

			// Log the access
			duration := time.Since(start)
			fields := []interface{}{
				"method", method,
				"service", service,
				"caller", caller,
				"duration", duration.String(),
				"duration_ms", duration.Milliseconds(),
			}

			if err != nil {
				fields = append(fields, "error", err.Error())
				logx.Warnw("RPC access", fields...)
			} else if cfg.SlowThreshold > 0 && duration > cfg.SlowThreshold {
				logx.Warnw("RPC slow access", fields...)
			} else {
				logx.Infow("RPC access", fields...)
			}

			return err
		}
	}
}