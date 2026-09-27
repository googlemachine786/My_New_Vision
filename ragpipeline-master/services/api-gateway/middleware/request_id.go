package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/visionary/ragpipeline/pkg/contextkeys"
)

const (
	RequestIDHeader = "X-Request-ID"
)

// RequestIDMiddleware adds a unique request ID to each request
type RequestIDMiddleware struct{}

// NewRequestIDMiddleware creates a new request ID middleware
func NewRequestIDMiddleware() *RequestIDMiddleware {
	return &RequestIDMiddleware{}
}

// Handler is the HTTP middleware that adds request IDs
func (m *RequestIDMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add to response headers
		w.Header().Set(RequestIDHeader, requestID)

		// Add to context using shared contextkeys package
		ctx := contextkeys.WithRequestID(r.Context(), requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := contextkeys.RequestIDFromContext(ctx); ok {
		return id
	}
	return ""
}
