// Package middleware provides common HTTP middleware for Hertz servers.
package middleware

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
)

// CORSConfig represents CORS middleware configuration.
type CORSConfig struct {
	// AllowOrigins is a list of origins that may access the resource.
	// Default: ["*"]
	AllowOrigins []string `yaml:"allowOrigins,omitempty" json:"allowOrigins,omitempty"`

	// AllowMethods is a list of methods allowed for the resource.
	// Default: ["GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"]
	AllowMethods []string `yaml:"allowMethods,omitempty" json:"allowMethods,omitempty"`

	// AllowHeaders is a list of headers that are allowed in a request.
	// Default: ["Origin", "Content-Type", "Accept", "Authorization"]
	AllowHeaders []string `yaml:"allowHeaders,omitempty" json:"allowHeaders,omitempty"`

	// ExposeHeaders is a list of headers that are safe to expose.
	ExposeHeaders []string `yaml:"exposeHeaders,omitempty" json:"exposeHeaders,omitempty"`

	// AllowCredentials indicates whether credentials can be shared.
	AllowCredentials bool `yaml:"allowCredentials,omitempty" json:"allowCredentials,omitempty"`

	// MaxAge indicates how long the preflight request can be cached (in seconds).
	// Default: 86400 (24 hours)
	MaxAge int `yaml:"maxAge,omitempty" json:"maxAge,omitempty"`
}

// SetDefaults applies default values.
func (c *CORSConfig) SetDefaults() {
	if len(c.AllowOrigins) == 0 {
		c.AllowOrigins = []string{"*"}
	}
	if len(c.AllowMethods) == 0 {
		c.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	}
	if len(c.AllowHeaders) == 0 {
		c.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	}
	if c.MaxAge == 0 {
		c.MaxAge = 86400
	}
}

// CORS returns a CORS middleware handler.
func CORS(cfg CORSConfig) app.HandlerFunc {
	cfg.SetDefaults()

	allowMethods := joinStrings(cfg.AllowMethods)
	allowHeaders := joinStrings(cfg.AllowHeaders)
	exposeHeaders := joinStrings(cfg.ExposeHeaders)

	return func(ctx context.Context, c *app.RequestContext) {
		origin := string(c.Request.Header.Peek("Origin"))

		// Check if origin is allowed
		allowed := false
		for _, o := range cfg.AllowOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			if cfg.AllowOrigins[0] == "*" && !cfg.AllowCredentials {
				c.Header("Access-Control-Allow-Origin", "*")
			} else {
				c.Header("Access-Control-Allow-Origin", origin)
			}

			c.Header("Access-Control-Allow-Methods", allowMethods)
			c.Header("Access-Control-Allow-Headers", allowHeaders)

			if exposeHeaders != "" {
				c.Header("Access-Control-Expose-Headers", exposeHeaders)
			}

			if cfg.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}

			if string(c.Request.Method()) == "OPTIONS" {
				c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", cfg.MaxAge))
				c.AbortWithStatus(204)
				return
			}
		}

		c.Next(ctx)
	}
}

func joinStrings(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += ", " + strs[i]
	}
	return result
}