package middleware

import (
	"context"
	"strings"

	"github.com/OmnTeam/npanel-pro/internal/conf"
	"github.com/OmnTeam/npanel-pro/pkg/tool"
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
func JWTAuth(authConfig *conf.Server_Auth) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 如果JWT认证被禁用，直接通过
			if authConfig != nil && !authConfig.EnableJwt {
				return handler(ctx, req)
			}

			// 检查是否为无需认证的路径
			if tr, ok := transport.FromServerContext(ctx); ok {
				operation := tr.Operation()
				if shouldSkipAuth(operation, authConfig) {
					return handler(ctx, req)
				}

				// Get JWT secret from config or use default
				secret := DefaultJWTSecret
				if authConfig != nil && authConfig.JwtSecret != "" {
					secret = authConfig.JwtSecret
				}

				token := tr.RequestHeader().Get(authorizationHeader)

				// Remove "Bearer " prefix if present
				if strings.HasPrefix(token, bearerPrefix) {
					token = strings.TrimPrefix(token, bearerPrefix)
				}

				// 对于需要认证的路径，token不能为空
				if token == "" {
					return nil, errors.Unauthorized("UNAUTHORIZED", "Authorization token required")
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

// shouldSkipAuth 检查是否应该跳过认证
func shouldSkipAuth(operation string, authConfig *conf.Server_Auth) bool {
	if authConfig == nil {
		return false
	}

	for _, pathPrefix := range authConfig.NoAuthPaths {
		if strings.HasPrefix(operation, pathPrefix) {
			return true
		}
	}
	return false
}
