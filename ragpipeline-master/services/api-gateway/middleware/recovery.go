package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/rs/zerolog/log"
)

// RecoveryMiddleware recovers from panics and returns 500
type RecoveryMiddleware struct{}

// NewRecoveryMiddleware creates a new recovery middleware
func NewRecoveryMiddleware() *RecoveryMiddleware {
	return &RecoveryMiddleware{}
}

// Handler is the HTTP middleware that recovers from panics
func (m *RecoveryMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				log.Error().
					Interface("panic", err).
					Str("stack", string(debug.Stack())).
					Str("request_id", GetRequestID(r.Context())).
					Msg("Recovered from panic")

				// Return 500 error
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":{"code":"internal_error","message":"An internal error occurred. Please try again later."}}`))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
