// Package middleware provides shared HTTP middleware functions.
package middleware

import (
	"fmt"
	"net/http"
	"strings"
)

// CORSConfig holds CORS configuration options.
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int // Cache preflight requests (seconds)
}

// DefaultCORSConfig returns a sensible default CORS configuration.
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-Request-ID"},
		ExposedHeaders: []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge: 86400, // 24 hours
	}
}

// CORS returns a middleware handler that adds CORS headers to responses.
func CORS(cfg *CORSConfig) func(http.Handler) http.Handler {
	if cfg == nil {
		cfg = DefaultCORSConfig()
	}

	allowedOrigins := make(map[string]bool, len(cfg.AllowedOrigins))
	for _, origin := range cfg.AllowedOrigins {
		allowedOrigins[origin] = true
	}

	allowedMethods := make(map[string]bool, len(cfg.AllowedMethods))
	for _, method := range cfg.AllowedMethods {
		allowedMethods[strings.ToUpper(method)] = true
	}

	allowedHeaders := make(map[string]bool, len(cfg.AllowedHeaders))
	for _, header := range cfg.AllowedHeaders {
		allowedHeaders[strings.ToLower(header)] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			if len(cfg.AllowedOrigins) == 0 || allowedOrigins["*"] || allowedOrigins[origin] {
				if origin != "" && !allowedOrigins["*"] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				} else if allowedOrigins["*"] {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				}

				if cfg.AllowCredentials && !allowedOrigins["*"] {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
			w.Header().Set("Access-Control-Expose-Headers", strings.Join(cfg.ExposedHeaders, ", "))
			w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", cfg.MaxAge))

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
