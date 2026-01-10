package middleware

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/golang-jwt/jwt/v5"
)

// JWTConfig represents JWT middleware configuration.
type JWTConfig struct {
	// Secret is the signing key for HS256 algorithm.
	Secret string `yaml:"secret,omitempty" json:"secret,omitempty"`

	// TokenLookup specifies where to find the token.
	// Format: "<source>:<name>" where source is "header", "query", or "cookie".
	// Default: "header:Authorization"
	TokenLookup string `yaml:"tokenLookup,omitempty" json:"tokenLookup,omitempty"`

	// AuthScheme is the auth scheme (e.g., "Bearer").
	// Default: "Bearer"
	AuthScheme string `yaml:"authScheme,omitempty" json:"authScheme,omitempty"`

	// ContextKey is the key to store claims in context.
	// Default: "jwt"
	ContextKey string `yaml:"contextKey,omitempty" json:"contextKey,omitempty"`

	// Claims is the claims type to use for parsing.
	// If nil, uses jwt.MapClaims.
	Claims jwt.Claims

	// Skipper determines whether to skip JWT validation.
	Skipper func(ctx context.Context, c *app.RequestContext) bool
}

// SetDefaults applies default values.
func (c *JWTConfig) SetDefaults() {
	if c.TokenLookup == "" {
		c.TokenLookup = "header:Authorization"
	}
	if c.AuthScheme == "" {
		c.AuthScheme = "Bearer"
	}
	if c.ContextKey == "" {
		c.ContextKey = "jwt"
	}
}

// Common errors
var (
	ErrMissingToken   = errors.New("missing JWT token")
	ErrInvalidToken   = errors.New("invalid JWT token")
	ErrTokenExpired   = errors.New("JWT token has expired")
	ErrMissingSecret  = errors.New("missing JWT secret")
	ErrInvalidLookup  = errors.New("invalid token lookup format")
)

// JWT returns a JWT authentication middleware.
func JWT(cfg JWTConfig) app.HandlerFunc {
	cfg.SetDefaults()

	// Parse token lookup
	parts := strings.SplitN(cfg.TokenLookup, ":", 2)
	if len(parts) != 2 {
		panic(ErrInvalidLookup)
	}
	source, name := parts[0], parts[1]

	return func(ctx context.Context, c *app.RequestContext) {
		// Check skipper
		if cfg.Skipper != nil && cfg.Skipper(ctx, c) {
			c.Next(ctx)
			return
		}

		// Extract token
		var tokenString string
		switch source {
		case "header":
			auth := string(c.Request.Header.Peek(name))
			if cfg.AuthScheme != "" {
				prefix := cfg.AuthScheme + " "
				if strings.HasPrefix(auth, prefix) {
					tokenString = strings.TrimPrefix(auth, prefix)
				}
			} else {
				tokenString = auth
			}
		case "query":
			tokenString = c.Query(name)
		case "cookie":
			tokenString = string(c.Cookie(name))
		}

		if tokenString == "" {
			c.AbortWithMsg(ErrMissingToken.Error(), 401)
			return
		}

		// Parse token
		var claims jwt.Claims = jwt.MapClaims{}
		if cfg.Claims != nil {
			claims = cfg.Claims
		}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(_ *jwt.Token) (interface{}, error) {
			if cfg.Secret == "" {
				return nil, ErrMissingSecret
			}
			return []byte(cfg.Secret), nil
		})

		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				c.AbortWithMsg(ErrTokenExpired.Error(), 401)
				return
			}
			c.AbortWithMsg(ErrInvalidToken.Error(), 401)
			return
		}

		if !token.Valid {
			c.AbortWithMsg(ErrInvalidToken.Error(), 401)
			return
		}

		// Store claims in context
		c.Set(cfg.ContextKey, token.Claims)
		c.Next(ctx)
	}
}

// GetClaims extracts JWT claims from the request context.
func GetClaims(c *app.RequestContext, key string) jwt.Claims {
	if key == "" {
		key = "jwt"
	}
	if claims, exists := c.Get(key); exists {
		if jwtClaims, ok := claims.(jwt.Claims); ok {
			return jwtClaims
		}
	}
	return nil
}

// GenerateToken generates a new JWT token.
func GenerateToken(secret string, claims jwt.MapClaims, expiry time.Duration) (string, error) {
	if claims == nil {
		claims = jwt.MapClaims{}
	}

	// Set expiration if not already set
	if _, ok := claims["exp"]; !ok && expiry > 0 {
		claims["exp"] = time.Now().Add(expiry).Unix()
	}

	// Set issued at if not already set
	if _, ok := claims["iat"]; !ok {
		claims["iat"] = time.Now().Unix()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}