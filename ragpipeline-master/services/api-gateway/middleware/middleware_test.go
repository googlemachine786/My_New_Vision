package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/visionary/ragpipeline/pkg/contextkeys"
)

// ============================================================================
// Request ID Middleware Tests
// ============================================================================

func TestRequestIDMiddleware_GeneratesID(t *testing.T) {
	middleware := NewRequestIDMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request ID was added to context
		requestID := GetRequestID(r.Context())
		if requestID == "" {
			t.Error("GetRequestID() returned empty string")
		}
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.Handler(handler).ServeHTTP(w, req)

	// Verify X-Request-ID header was set
	if w.Header().Get("X-Request-ID") == "" {
		t.Error("X-Request-ID header not set")
	}
}

func TestRequestIDMiddleware_PreservesExistingID(t *testing.T) {
	middleware := NewRequestIDMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		if requestID != "existing-id-123" {
			t.Errorf("GetRequestID() = %q, want %q", requestID, "existing-id-123")
		}
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "existing-id-123")
	w := httptest.NewRecorder()

	middleware.Handler(handler).ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") != "existing-id-123" {
		t.Errorf("X-Request-ID = %q, want %q", w.Header().Get("X-Request-ID"), "existing-id-123")
	}
}

func TestRequestIDMiddleware_UniqueIDs(t *testing.T) {
	middleware := NewRequestIDMiddleware()

	// Make multiple requests and verify unique IDs
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ids[GetRequestID(r.Context())] = true
		})

		req := httptest.NewRequest("GET", "/test", nil)
		middleware.Handler(handler).ServeHTTP(httptest.NewRecorder(), req)
	}

	if len(ids) != 100 {
		t.Errorf("Expected 100 unique IDs, got %d", len(ids))
	}
}

func TestGetRequestID_EmptyContext(t *testing.T) {
	got := GetRequestID(context.Background())
	if got != "" {
		t.Errorf("GetRequestID(empty context) = %q, want empty string", got)
	}
}

func TestGetRequestID_WithContextValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextkeys.RequestIDKey, "test-id-456")
	got := GetRequestID(ctx)
	if got != "test-id-456" {
		t.Errorf("GetRequestID() = %q, want %q", got, "test-id-456")
	}
}

// ============================================================================
// Auth Middleware Tests
// ============================================================================

func TestAuthMiddleware_SkipsHealthEndpoints(t *testing.T) {
	middleware := NewAuthMiddleware("test-secret")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name string
		path string
	}{
		{"health", "/health"},
		{"version", "/version"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			middleware.Handler(handler).ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Auth middleware should skip %s, got status %d", tt.path, w.Code)
			}
		})
	}
}

func TestAuthMiddleware_RejectsMissingAuth(t *testing.T) {
	middleware := NewAuthMiddleware("test-secret")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called without auth")
	})

	req := httptest.NewRequest("GET", "/query", nil)
	w := httptest.NewRecorder()

	middleware.Handler(handler).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Auth middleware should reject missing auth, got status %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "missing authorization header") {
		t.Errorf("Response should contain 'missing authorization header', got: %s", body)
	}
}

func TestAuthMiddleware_RejectsInvalidFormat(t *testing.T) {
	middleware := NewAuthMiddleware("test-secret")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called with invalid auth format")
	})

	tests := []struct {
		name  string
		auth  string
	}{
		{"no bearer", "Basic token123"},
		{"empty", ""},
		{"only bearer", "Bearer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/query", nil)
			if tt.auth != "" {
				req.Header.Set("Authorization", tt.auth)
			}
			w := httptest.NewRecorder()

			middleware.Handler(handler).ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("Auth middleware should reject invalid auth format, got status %d", w.Code)
			}
		})
	}
}

// ============================================================================
// Auth Security Tests (Rate Limiting, Secret Validation, Expiration)
// ============================================================================

func TestValidateSecret_RejectsShortSecrets(t *testing.T) {
	shortSecrets := []string{
		"short",
		"this-is-too-short",
		"1234567890123456789012345678901", // 31 chars
	}

	for _, secret := range shortSecrets {
		err := ValidateSecret(secret)
		if err == nil {
			t.Errorf("ValidateSecret(%q) should reject secrets < 32 chars", secret)
		}
		if _, ok := err.(*SecretValidationError); !ok {
			t.Errorf("ValidateSecret(%q) should return SecretValidationError, got %T", secret, err)
		}
	}
}

func TestValidateSecret_AcceptsLongSecrets(t *testing.T) {
	validSecrets := []string{
		"12345678901234567890123456789012",                  // 32 chars
		"this-is-a-much-longer-secret-that-is-definitely-safe", // 52 chars
	}

	for _, secret := range validSecrets {
		err := ValidateSecret(secret)
		if err != nil {
			t.Errorf("ValidateSecret(%q) should accept secrets >= 32 chars, got: %v", secret, err)
		}
	}
}

func TestAuthMiddleware_RejectsTokenWithoutExpiration(t *testing.T) {
	// Create a token without expiration
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "test-user",
		// No "exp" claim
	})
	tokenString, _ := token.SignedString([]byte("this-is-a-very-long-test-secret-for-security"))

	middleware := NewAuthMiddleware("this-is-a-very-long-test-secret-for-security")

	req := httptest.NewRequest("GET", "/query", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()

	middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for token without expiration")
	})).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Should reject token without expiration, got status %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if errMsg, _ := resp["error"].(map[string]interface{})["message"].(string); !strings.Contains(errMsg, "expiration") {
		t.Errorf("Error should mention missing expiration, got: %s", errMsg)
	}
}

func TestAuthMiddleware_RateLimiting(t *testing.T) {
	// Reset the global rate limiter for clean test
	globalAuthRateLimiter = &authRateLimiter{
		attempts: make(map[string][]time.Time),
	}

	secret := "this-is-a-very-long-test-secret-for-security"
	middleware := NewAuthMiddleware(secret)

	// Send 20 requests (should all be processed, though rejected for invalid token)
	for i := 0; i < authRateLimitMax; i++ {
		req := httptest.NewRequest("GET", "/query", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.here")
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()

		middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(w, req)
	}

	// 21st request should be rate limited
	req := httptest.NewRequest("GET", "/query", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called when rate limited")
	})).ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Should rate limit after %d attempts, got status %d", authRateLimitMax, w.Code)
	}

	retryAfter := w.Header().Get("Retry-After")
	if retryAfter == "" {
		t.Error("Rate limit response should include Retry-After header")
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if errCode, _ := resp["error"].(map[string]interface{})["code"].(string); errCode != "rate_limited" {
		t.Errorf("Error code should be 'rate_limited', got: %s", errCode)
	}
}

func TestGetUserID_EmptyContext(t *testing.T) {
	got := GetUserID(context.Background())
	if got != "" {
		t.Errorf("GetUserID(empty context) = %q, want empty string", got)
	}
}

func TestGetAuthToken_EmptyContext(t *testing.T) {
	got := GetAuthToken(context.Background())
	if got != "" {
		t.Errorf("GetAuthToken(empty context) = %q, want empty string", got)
	}
}

// ============================================================================
// Recovery Middleware Tests
// ============================================================================

func TestRecoveryMiddleware_RecoversFromPanic(t *testing.T) {
	middleware := NewRecoveryMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "test-req-id")
	w := httptest.NewRecorder()

	// Should not panic
	middleware.Handler(handler).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Recovery middleware should return 500 on panic, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "internal_error") {
		t.Errorf("Response should contain 'internal_error', got: %s", body)
	}
}

func TestRecoveryMiddleware_PassesNormalRequests(t *testing.T) {
	middleware := NewRecoveryMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.Handler(handler).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Recovery middleware should pass normal requests, got %d", w.Code)
	}
}

// ============================================================================
// CORS Middleware Tests
// ============================================================================

func TestCORSMiddleware_SetsHeaders(t *testing.T) {
	middleware := NewCORSMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	middleware.Handler(handler).ServeHTTP(w, req)

	// Verify CORS headers
	headers := w.Header()
	if headers.Get("Access-Control-Allow-Origin") == "" {
		t.Error("Missing Access-Control-Allow-Origin header")
	}
	if headers.Get("Access-Control-Allow-Methods") == "" {
		t.Error("Missing Access-Control-Allow-Methods header")
	}
	if headers.Get("Access-Control-Allow-Headers") == "" {
		t.Error("Missing Access-Control-Allow-Headers header")
	}
}

func TestCORSMiddleware_PreflightRequest(t *testing.T) {
	middleware := NewCORSMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for preflight")
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	w := httptest.NewRecorder()

	middleware.Handler(handler).ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Preflight should return 204, got %d", w.Code)
	}
}

// ============================================================================
// Local Rate Limiter Tests
// ============================================================================

func TestLocalRateLimiter_AllowsInitialRequests(t *testing.T) {
	limiter := &LocalRateLimiter{
		tokens:     10,
		lastRefill: time.Now(),
	}

	// First requests should be allowed
	for i := 0; i < 10; i++ {
		allowed, _, _, err := limiter.allow(60, 10)
		if err != nil || !allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}
}

func TestLocalRateLimiter_ExhaustsTokens(t *testing.T) {
	limiter := &LocalRateLimiter{
		tokens:     2,
		lastRefill: time.Now(),
	}

	// Exhaust tokens
	_, _, _, _ = limiter.allow(60, 2)
	_, _, _, _ = limiter.allow(60, 2)

	// Next request should be denied (no refill time elapsed)
	allowed, _, _, _ := limiter.allow(60, 2)
	if allowed {
		t.Error("Request should be denied after tokens exhausted")
	}
}

func TestLocalRateLimiter_RefillsTokens(t *testing.T) {
	limiter := &LocalRateLimiter{
		tokens:     0,
		lastRefill: time.Now().Add(-2 * time.Minute), // 2 minutes ago
	}

	// After 2 minutes at 60rpm, should have refilled 120 tokens (capped at burst=1)
	allowed, _, _, err := limiter.allow(60, 1)
	if err != nil || !allowed {
		t.Error("Request should be allowed after token refill")
	}
}

// ============================================================================
// Helper Tests
// ============================================================================

func TestGetClientIP_FromRemoteAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	got := getClientIP(req)
	if got != "192.168.1.1" {
		t.Errorf("getClientIP() = %q, want %q", got, "192.168.1.1")
	}
}

func TestGetClientIP_FromXForwardedFor(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1, 192.168.1.1")
	req.RemoteAddr = "127.0.0.1:12345"

	got := getClientIP(req)
	if got != "10.0.0.1" {
		t.Errorf("getClientIP() from X-Forwarded-For = %q, want %q", got, "10.0.0.1")
	}
}

func TestGetClientIP_FromXRealIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Real-IP", "10.0.0.2")
	req.RemoteAddr = "127.0.0.1:12345"

	got := getClientIP(req)
	if got != "10.0.0.2" {
		t.Errorf("getClientIP() from X-Real-IP = %q, want %q", got, "10.0.0.2")
	}
}
