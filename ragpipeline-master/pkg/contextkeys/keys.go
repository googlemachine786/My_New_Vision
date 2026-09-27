// Package contextkeys provides typed context key helpers.
package contextkeys

import "context"

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	RequestIDKey contextKey = "request_id"
	AuthTokenKey contextKey = "auth_token"
)

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// UserIDFromContext extracts user ID from context
func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// RequestIDFromContext extracts request ID from context
func RequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(RequestIDKey).(string)
	return requestID, ok
}

// WithAuthToken adds auth token to context
func WithAuthToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, AuthTokenKey, token)
}

// AuthTokenFromContext extracts auth token from context
func AuthTokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(AuthTokenKey).(string)
	return token, ok
}
