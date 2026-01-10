// Package middleware provides common middleware for Kitex RPC servers.
package middleware

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/cloudwego/kitex/pkg/endpoint"

	"github.com/ssgohq/goten-core/logx"
)

// Recovery returns a middleware that recovers from panics.
// It logs the panic with stack trace and returns an internal error.
func Recovery() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req, resp interface{}) (err error) {
			defer func() {
				if r := recover(); r != nil {
					stack := debug.Stack()
					logx.Errorw("Panic recovered in RPC handler",
						"panic", fmt.Sprintf("%v", r),
						"stack", string(stack),
					)
					err = fmt.Errorf("internal server error")
				}
			}()
			return next(ctx, req, resp)
		}
	}
}

// RecoveryWithHandler returns a recovery middleware with custom panic handler.
func RecoveryWithHandler(
	handler func(ctx context.Context, panicValue interface{}, stack []byte) error,
) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req, resp interface{}) (err error) {
			defer func() {
				if r := recover(); r != nil {
					stack := debug.Stack()
					err = handler(ctx, r, stack)
				}
			}()
			return next(ctx, req, resp)
		}
	}
}