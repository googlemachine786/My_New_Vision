package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSMiddleware handles Cross-Origin Resource Sharing headers
type CORSMiddleware struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// NewCORSMiddleware creates a new CORS middleware with sensible defaults
// FIX P1: Default to restrictive origins; AllowCredentials with wildcard is a security risk
func NewCORSMiddleware() *CORSMiddleware {
	return &CORSMiddleware{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173"}, // Dev origins only
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "X-CSRF-Token"},
		ExposedHeaders:   []string{"X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// Handler is the HTTP middleware that adds CORS headers
func (m *CORSMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			m.handlePreflight(w, r, origin)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Add CORS headers to actual requests
		m.setCORSHeaders(w, r, origin)
		next.ServeHTTP(w, r)
	})
}

func (m *CORSMiddleware) handlePreflight(w http.ResponseWriter, r *http.Request, origin string) {
	m.setCORSHeaders(w, r, origin)
}

func (m *CORSMiddleware) setCORSHeaders(w http.ResponseWriter, r *http.Request, origin string) {
	// Allow-Origin
	if len(m.AllowedOrigins) == 1 && m.AllowedOrigins[0] == "*" {
		if !m.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
	} else {
		for _, allowedOrigin := range m.AllowedOrigins {
			if allowedOrigin == origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}
	}

	// Allow-Methods
	if len(m.AllowedMethods) > 0 {
		w.Header().Set("Access-Control-Allow-Methods", joinStrings(m.AllowedMethods, ", "))
	}

	// Allow-Headers
	if len(m.AllowedHeaders) > 0 {
		w.Header().Set("Access-Control-Allow-Headers", joinStrings(m.AllowedHeaders, ", "))
	}

	// Expose-Headers
	if len(m.ExposedHeaders) > 0 {
		w.Header().Set("Access-Control-Expose-Headers", joinStrings(m.ExposedHeaders, ", "))
	}

	// Allow-Credentials
	if m.AllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	// Max-Age
	if m.MaxAge > 0 {
		w.Header().Set("Access-Control-Max-Age", stringInt(m.MaxAge))
	}
}

func joinStrings(strs []string, sep string) string {
	return strings.Join(strs, sep)
}

func stringInt(i int) string {
	return strconv.Itoa(i)
}
