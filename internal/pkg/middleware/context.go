package middleware

import "context"

// Context keys for storing user information
type contextKey string

const (
	userIDKey      contextKey = "userID"
	sessionIDKey   contextKey = "sessionID"
	userAgentKey   contextKey = "userAgent"
	clientIPKey    contextKey = "clientIP"
	requestURIKey  contextKey = "requestURI"
	requestHostKey contextKey = "requestHost"
	gatewayModeKey contextKey = "gatewayMode"
	queryParamsKey contextKey = "queryParams"
)

// WithUserID stores user ID in context
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// WithSessionID stores session ID in context
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDKey, sessionID)
}

// WithUserAgent stores user agent in context.
func WithUserAgent(ctx context.Context, userAgent string) context.Context {
	return context.WithValue(ctx, userAgentKey, userAgent)
}

// WithClientIP stores client IP in context.
func WithClientIP(ctx context.Context, clientIP string) context.Context {
	return context.WithValue(ctx, clientIPKey, clientIP)
}

// WithRequestURI stores request URI in context.
func WithRequestURI(ctx context.Context, requestURI string) context.Context {
	return context.WithValue(ctx, requestURIKey, requestURI)
}

// WithRequestHost stores request host in context.
func WithRequestHost(ctx context.Context, requestHost string) context.Context {
	return context.WithValue(ctx, requestHostKey, requestHost)
}

// WithGatewayMode stores gateway mode flag in context.
func WithGatewayMode(ctx context.Context, gatewayMode bool) context.Context {
	return context.WithValue(ctx, gatewayModeKey, gatewayMode)
}

// WithQueryParams stores the original query params in context.
func WithQueryParams(ctx context.Context, queryParams map[string]string) context.Context {
	if len(queryParams) == 0 {
		return context.WithValue(ctx, queryParamsKey, map[string]string{})
	}

	cloned := make(map[string]string, len(queryParams))
	for key, value := range queryParams {
		cloned[key] = value
	}
	return context.WithValue(ctx, queryParamsKey, cloned)
}

// GetUserID retrieves user ID from context
func GetUserID(ctx context.Context) int64 {
	if userID, ok := ctx.Value(userIDKey).(int64); ok {
		return userID
	}
	// Return default user ID if not found in context
	return 0
}

// GetSessionID retrieves session ID from context
func GetSessionID(ctx context.Context) string {
	if sessionID, ok := ctx.Value(sessionIDKey).(string); ok {
		return sessionID
	}
	return ""
}

// GetUserAgent retrieves User-Agent from context
func GetUserAgent(ctx context.Context) string {
	if userAgent, ok := ctx.Value(userAgentKey).(string); ok {
		return userAgent
	}
	return ""
}

// GetClientIP retrieves client IP from context
func GetClientIP(ctx context.Context) string {
	if clientIP, ok := ctx.Value(clientIPKey).(string); ok {
		return clientIP
	}
	return ""
}

// GetRequestURI retrieves request URI from context
func GetRequestURI(ctx context.Context) string {
	if uri, ok := ctx.Value(requestURIKey).(string); ok {
		return uri
	}
	return ""
}

// GetRequestHost retrieves request host from context
func GetRequestHost(ctx context.Context) string {
	if host, ok := ctx.Value(requestHostKey).(string); ok {
		return host
	}
	return ""
}

// GetGatewayMode retrieves gateway mode flag from context
func GetGatewayMode(ctx context.Context) bool {
	if mode, ok := ctx.Value(gatewayModeKey).(bool); ok {
		return mode
	}
	return false
}

// GetQueryParams retrieves the original query params from context.
func GetQueryParams(ctx context.Context) map[string]string {
	params, ok := ctx.Value(queryParamsKey).(map[string]string)
	if !ok || len(params) == 0 {
		return map[string]string{}
	}

	cloned := make(map[string]string, len(params))
	for key, value := range params {
		cloned[key] = value
	}
	return cloned
}
