package middleware

import (
	"context"
	"os"
	"strings"

	"github.com/OmnTeam/ppanel-pro/pkg/tool"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

const (
	// DefaultJWTSecret is the default JWT secret key
	DefaultJWTSecret = "your-secret-key-change-in-production"

	// Authorization header key
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "
)

// JWTAuth returns a JWT authentication middleware
func JWTAuth() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// Get JWT secret from environment or use default
			secret := os.Getenv("JWT_SECRET")
			if secret == "" {
				secret = DefaultJWTSecret
			}

			// Extract token from Authorization header
			if tr, ok := transport.FromServerContext(ctx); ok {
				token := tr.RequestHeader().Get(authorizationHeader)

				// Remove "Bearer " prefix if present
				if strings.HasPrefix(token, bearerPrefix) {
					token = strings.TrimPrefix(token, bearerPrefix)
				}

				// If token is empty, continue without authentication
				// This allows public endpoints to work
				if token == "" {
					return handler(ctx, req)
				}

				// Parse JWT token
				claims, err := tool.ParseJWT(token, secret)
				if err != nil {
					return nil, errors.Unauthorized("UNAUTHORIZED", "Invalid or expired token")
				}

				// Extract UserId from claims
				userID, ok := claims["UserId"].(float64)
				if !ok {
					return nil, errors.Unauthorized("UNAUTHORIZED", "Invalid token claims: missing UserId")
				}

				sessionID, _ := claims["SessionId"].(string)

				// Store in context
				ctx = context.WithValue(ctx, userIDKey, int64(userID))
				if sessionID != "" {
					ctx = context.WithValue(ctx, sessionIDKey, sessionID)
				}
			}

			return handler(ctx, req)
		}
	}
}
