package middleware

import "context"

// Context keys for storing user information
type contextKey string

const (
	userIDKey    contextKey = "userID"
	sessionIDKey contextKey = "sessionID"
)

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
