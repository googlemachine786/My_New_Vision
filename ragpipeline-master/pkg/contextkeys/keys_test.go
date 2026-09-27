package contextkeys

import (
	"context"
	"testing"
)

func TestWithUserID_And_UserIDFromContext(t *testing.T) {
	tests := []struct {
		name   string
		userID string
	}{
		{"valid user", "user-123"},
		{"empty user", ""},
		{"unicode user", "用户-456"},
		{"long user ID", string(make([]byte, 1000))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithUserID(context.Background(), tt.userID)
			got, ok := UserIDFromContext(ctx)

			if !ok {
				t.Errorf("UserIDFromContext() ok = false, want true")
			}
			if got != tt.userID {
				t.Errorf("UserIDFromContext() = %q, want %q", got, tt.userID)
			}
		})
	}
}

func TestUserIDFromContext_EmptyContext(t *testing.T) {
	got, ok := UserIDFromContext(context.Background())
	if ok {
		t.Errorf("UserIDFromContext(empty) ok = true, want false")
	}
	if got != "" {
		t.Errorf("UserIDFromContext(empty) = %q, want empty string", got)
	}
}

func TestUserIDFromContext_WrongType(t *testing.T) {
	// Simulate wrong type being stored
	type wrongKey string
	ctx := context.WithValue(context.Background(), wrongKey("user_id"), 12345)

	got, ok := UserIDFromContext(ctx)
	if ok {
		t.Errorf("UserIDFromContext(wrong type) ok = true, want false")
	}
	if got != "" {
		t.Errorf("UserIDFromContext(wrong type) = %q, want empty string", got)
	}
}

func TestWithRequestID_And_RequestIDFromContext(t *testing.T) {
	tests := []struct {
		name      string
		requestID string
	}{
		{"valid request", "req-abc-123"},
		{"empty request", ""},
		{"uuid format", "550e8400-e29b-41d4-a716-446655440000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithRequestID(context.Background(), tt.requestID)
			got, ok := RequestIDFromContext(ctx)

			if !ok {
				t.Errorf("RequestIDFromContext() ok = false, want true")
			}
			if got != tt.requestID {
				t.Errorf("RequestIDFromContext() = %q, want %q", got, tt.requestID)
			}
		})
	}
}

func TestRequestIDFromContext_EmptyContext(t *testing.T) {
	got, ok := RequestIDFromContext(context.Background())
	if ok {
		t.Errorf("RequestIDFromContext(empty) ok = true, want false")
	}
	if got != "" {
		t.Errorf("RequestIDFromContext(empty) = %q, want empty string", got)
	}
}

func TestWithAuthToken_And_AuthTokenFromContext(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{"valid token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test"},
		{"empty token", ""},
		{"bearer token", "Bearer eyJhbGciOiJIUzI1NiJ9.signature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithAuthToken(context.Background(), tt.token)
			got, ok := AuthTokenFromContext(ctx)

			if !ok {
				t.Errorf("AuthTokenFromContext() ok = false, want true")
			}
			if got != tt.token {
				t.Errorf("AuthTokenFromContext() = %q, want %q", got, tt.token)
			}
		})
	}
}

func TestAuthTokenFromContext_EmptyContext(t *testing.T) {
	got, ok := AuthTokenFromContext(context.Background())
	if ok {
		t.Errorf("AuthTokenFromContext(empty) ok = true, want false")
	}
	if got != "" {
		t.Errorf("AuthTokenFromContext(empty) = %q, want empty string", got)
	}
}

func TestContextKeys_DoNotCollide(t *testing.T) {
	// Verify that all three context keys are distinct
	ctx := context.Background()
	ctx = WithUserID(ctx, "user-1")
	ctx = WithRequestID(ctx, "req-1")
	ctx = WithAuthToken(ctx, "token-1")

	userID, userOK := UserIDFromContext(ctx)
	requestID, reqOK := RequestIDFromContext(ctx)
	authToken, authOK := AuthTokenFromContext(ctx)

	if !userOK || userID != "user-1" {
		t.Errorf("UserID = %q (ok=%v), want %q (ok=true)", userID, userOK, "user-1")
	}
	if !reqOK || requestID != "req-1" {
		t.Errorf("RequestID = %q (ok=%v), want %q (ok=true)", requestID, reqOK, "req-1")
	}
	if !authOK || authToken != "token-1" {
		t.Errorf("AuthToken = %q (ok=%v), want %q (ok=true)", authToken, authOK, "token-1")
	}
}

func TestContextKeyChaining(t *testing.T) {
	ctx := context.Background()

	// Chain all three
	ctx = WithUserID(ctx, "user-123")
	ctx = WithRequestID(ctx, "req-456")
	ctx = WithAuthToken(ctx, "token-789")

	// Verify all are accessible
	userID, _ := UserIDFromContext(ctx)
	if userID != "user-123" {
		t.Errorf("UserID = %q, want %q", userID, "user-123")
	}

	requestID, _ := RequestIDFromContext(ctx)
	if requestID != "req-456" {
		t.Errorf("RequestID = %q, want %q", requestID, "req-456")
	}

	authToken, _ := AuthTokenFromContext(ctx)
	if authToken != "token-789" {
		t.Errorf("AuthToken = %q, want %q", authToken, "token-789")
	}
}

func TestContextKeyOverwriting(t *testing.T) {
	ctx := WithUserID(context.Background(), "user-old")
	ctx = WithUserID(ctx, "user-new")

	got, ok := UserIDFromContext(ctx)
	if !ok {
		t.Fatal("UserIDFromContext() ok = false, want true")
	}
	if got != "user-new" {
		t.Errorf("UserIDFromContext() = %q, want %q", got, "user-new")
	}
}
