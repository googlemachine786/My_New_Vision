package middleware

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
	"github.com/visionary/ragpipeline/pkg/contextkeys"
)

const (
	// authRateLimitWindow is the time window for auth rate limiting (1 minute)
	authRateLimitWindow = 1 * time.Minute
	// authRateLimitMax is the max auth attempts per IP per window
	authRateLimitMax = 20
	// minSecretLength is the minimum acceptable JWT secret length
	minSecretLength = 32
)

// authRateLimiter tracks auth attempts per IP to prevent brute force
type authRateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
}

var globalAuthRateLimiter = &authRateLimiter{
	attempts: make(map[string][]time.Time),
}

// allow checks if an IP is within the rate limit for auth attempts
func (rl *authRateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-authRateLimitWindow)

	// Filter out expired attempts
	var recent []time.Time
	for _, t := range rl.attempts[ip] {
		if t.After(windowStart) {
			recent = append(recent, t)
		}
	}

	if len(recent) >= authRateLimitMax {
		rl.attempts[ip] = recent
		return false
	}

	rl.attempts[ip] = append(recent, now)
	return true
}

// ValidateSecret checks that the JWT secret meets minimum security requirements
func ValidateSecret(secret string) error {
	if len(secret) < minSecretLength {
		return &SecretValidationError{
			Message: "JWT_SECRET must be at least 32 characters",
		}
	}
	return nil
}

// SecretValidationError is returned when the JWT secret is too weak
type SecretValidationError struct {
	Message string
}

func (e *SecretValidationError) Error() string {
	return e.Message
}

// AuthMiddleware validates JWT bearer tokens with rate limiting
type AuthMiddleware struct {
	Secret string
	// SkipPaths contains paths that don't require authentication
	SkipPaths map[string]bool
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(secret string) *AuthMiddleware {
	if err := ValidateSecret(secret); err != nil {
		log.Warn().Err(err).Msg("Weak JWT secret detected - use a strong secret in production")
	}

	return &AuthMiddleware{
		Secret: secret,
		SkipPaths: map[string]bool{
			"/health":  true,
			"/version": true,
		},
	}
}

// Handler is the HTTP middleware that validates JWT tokens
func (m *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for specified paths
		if m.SkipPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		// Rate limit auth attempts to prevent brute force
		clientIP := getClientIP(r)
		if !globalAuthRateLimiter.allow(clientIP) {
			writeRateLimitError(w)
			return
		}

		// Extract Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeAuthError(w, "missing authorization header")
			return
		}

		// Check Bearer scheme
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			writeAuthError(w, "invalid authorization header format")
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method - only allow HMAC (HS256, HS384, HS512)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(m.Secret), nil
		})

		if err != nil {
			log.Debug().Str("ip", clientIP).Err(err).Msg("JWT validation failed")
			writeAuthError(w, "invalid or expired token")
			return
		}

		if !token.Valid {
			writeAuthError(w, "invalid token")
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeAuthError(w, "invalid token claims")
			return
		}

		// Validate expiration claim exists (tokens must have exp)
		if _, hasExp := claims["exp"]; !hasExp {
			writeAuthError(w, "token missing expiration claim")
			return
		}

		// Extract user_id from claims (try common claim names)
		var userID string
		for _, claim := range []string{"user_id", "sub", "uid"} {
			if val, ok := claims[claim].(string); ok && val != "" {
				userID = val
				break
			}
		}

		if userID == "" {
			writeAuthError(w, "user_id not found in token")
			return
		}

		// Add user_id and auth_token to request context using shared contextkeys
		ctx := contextkeys.WithUserID(r.Context(), userID)
		ctx = contextkeys.WithAuthToken(ctx, tokenString)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID extracts user ID from request context
func GetUserID(ctx context.Context) string {
	if userID, ok := contextkeys.UserIDFromContext(ctx); ok {
		return userID
	}
	return ""
}

// GetAuthToken extracts the auth token from request context
func GetAuthToken(ctx context.Context) string {
	if token, ok := contextkeys.AuthTokenFromContext(ctx); ok {
		return token
	}
	return ""
}

func writeAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":{"code":"unauthorized","message":"` + message + `"}}`))
}

func writeRateLimitError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", "60")
	w.WriteHeader(http.StatusTooManyRequests)
	w.Write([]byte(`{"error":{"code":"rate_limited","message":"too many auth attempts, please try again later"}}`))
}
