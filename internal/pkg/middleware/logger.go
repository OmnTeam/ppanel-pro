package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

const (
	// SlowRequestThreshold is the threshold for slow request logging (in milliseconds)
	SlowRequestThreshold = 3000
)

// Logging returns a logging middleware for HTTP requests
func Logging(logger log.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			start := time.Now()

			// Get transport information
			var method, path, clientIP, userAgent, requestURI, requestHost string
			if tr, ok := transport.FromServerContext(ctx); ok {
				if ht, ok := tr.(khttp.Transporter); ok {
					path = ht.Operation()
					method = ht.Request().Method
					clientIP = ht.Request().RemoteAddr
					userAgent = ht.Request().UserAgent()
					requestURI = ht.Request().RequestURI
					requestHost = ht.Request().Host

					// Try to get real IP from X-Forwarded-For or X-Real-IP
					if forwardedFor := ht.Request().Header.Get("X-Forwarded-For"); forwardedFor != "" {
						clientIP = strings.Split(forwardedFor, ",")[0]
					} else if realIP := ht.Request().Header.Get("X-Real-IP"); realIP != "" {
						clientIP = realIP
					}

					// Store User-Agent, Client IP, Request URI, and Request Host in context for later use
					ctx = context.WithValue(ctx, userAgentKey, userAgent)
					ctx = context.WithValue(ctx, clientIPKey, clientIP)
					ctx = context.WithValue(ctx, requestURIKey, requestURI)
					ctx = context.WithValue(ctx, requestHostKey, requestHost)
					// Note: gatewayMode is not available in the middleware, needs to be set elsewhere
				}
			}

			// Log request start
			log.Context(ctx).Infow(
				"[HTTP Request Started]",
				"method", method,
				"path", path,
				"client_ip", clientIP,
				"user_agent", userAgent,
			)

			// Call handler
			reply, err := handler(ctx, req)

			// Calculate duration
			duration := time.Since(start)

			// Determine log level based on error and duration
			if err != nil {
				// Log error
				log.Context(ctx).Errorw(
					"[HTTP Request Error]",
					"method", method,
					"path", path,
					"client_ip", clientIP,
					"duration", duration,
					"error", err.Error(),
				)
			} else if duration.Milliseconds() > SlowRequestThreshold {
				// Log slow request warning
				log.Context(ctx).Warnw(
					"[HTTP Slow Request]",
					"method", method,
					"path", path,
					"client_ip", clientIP,
					"duration", duration,
					"duration_ms", fmt.Sprintf("%d ms", duration.Milliseconds()),
				)
			} else {
				// Log success
				log.Context(ctx).Infow(
					"[HTTP Request Completed]",
					"method", method,
					"path", path,
					"client_ip", clientIP,
					"duration", duration,
					"duration_ms", fmt.Sprintf("%d ms", duration.Milliseconds()),
				)
			}

			return reply, err
		}
	}
}

// LoggingWithBody returns a logging middleware with request/response body logging
// Note: This middleware logs request and response bodies, use with caution as it may impact performance
func LoggingWithBody(logger log.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			start := time.Now()

			// Get transport information
			var method, path string
			var requestBody string

			if tr, ok := transport.FromServerContext(ctx); ok {
				if ht, ok := tr.(khttp.Transporter); ok {
					path = ht.Operation()
					method = ht.Request().Method

					// Read request body for POST/PUT/PATCH requests
					if method == "POST" || method == "PUT" || method == "PATCH" {
						requestBody = extractRequestBody(ht)
					}
				}
			}

			// Log request
			log.Context(ctx).Infow(
				"[HTTP Request]",
				"method", method,
				"path", path,
				"body", maskSensitiveFields(requestBody),
			)

			// Call handler
			reply, err := handler(ctx, req)

			// Calculate duration and log response
			duration := time.Since(start)

			// Log response
			if err != nil {
				log.Context(ctx).Errorw(
					"[HTTP Response Error]",
					"method", method,
					"path", path,
					"duration", duration,
					"error", err,
				)
			} else {
				log.Context(ctx).Infow(
					"[HTTP Response]",
					"method", method,
					"path", path,
					"duration", duration,
				)
			}

			return reply, err
		}
	}
}

// extractRequestBody extracts request body from HTTP transport
func extractRequestBody(ht khttp.Transporter) string {
	if ht.Request().Body == nil {
		return ""
	}

	// Read body
	body, err := io.ReadAll(ht.Request().Body)
	if err != nil {
		return ""
	}

	// Restore body for subsequent reads
	ht.Request().Body = io.NopCloser(bytes.NewBuffer(body))

	return string(body)
}

// maskSensitiveFields masks sensitive fields in request body
func maskSensitiveFields(body string) string {
	if body == "" {
		return body
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		// Not JSON, return as is
		return body
	}

	// Sensitive fields to mask
	sensitiveFields := []string{
		"password",
		"old_password",
		"new_password",
		"confirm_password",
		"secret",
		"token",
		"api_key",
	}

	// Mask sensitive fields
	for _, field := range sensitiveFields {
		if _, exists := data[field]; exists {
			data[field] = "***"
		}
	}

	masked, err := json.Marshal(data)
	if err != nil {
		return body
	}

	return string(masked)
}
