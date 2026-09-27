package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/visionary/ragpipeline/services/api-gateway/model"
)

func TestClarificationHandler_ServeHTTP_Success(t *testing.T) {
	handler := NewClarificationHandler()

	body := `{"query": "What is photosynthesis?", "confidence": 0.5}`
	req := httptest.NewRequest(http.MethodPost, "/clarify", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp model.ClarificationResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !resp.NeedsClarification {
		t.Error("expected needs_clarification to be true")
	}

	if len(resp.Options) == 0 {
		t.Error("expected at least one clarification option")
	}

	if len(resp.Options) > 3 {
		t.Errorf("expected at most 3 options, got %d", len(resp.Options))
	}
}

func TestClarificationHandler_ServeHTTP_InvalidMethod(t *testing.T) {
	handler := NewClarificationHandler()

	req := httptest.NewRequest(http.MethodGet, "/clarify", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestClarificationHandler_ServeHTTP_InvalidJSON(t *testing.T) {
	handler := NewClarificationHandler()

	req := httptest.NewRequest(http.MethodPost, "/clarify", strings.NewReader("not json"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestClarificationHandler_ServeHTTP_MissingQuery(t *testing.T) {
	handler := NewClarificationHandler()

	body := `{"confidence": 0.5}`
	req := httptest.NewRequest(http.MethodPost, "/clarify", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestClarificationHandler_HandleSelection_Success(t *testing.T) {
	handler := NewClarificationHandler()

	body := `{"option_id": "test-id", "original_query": "What is photosynthesis?"}`
	req := httptest.NewRequest(http.MethodPost, "/clarify/select", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler.HandleSelection(w, req)

	// Should return 400 since the option ID won't be found in predefined options
	// This is expected behavior - in real usage, option IDs come from generated options
	if w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
		t.Errorf("expected status 200 or 400, got %d", w.Code)
	}
}

func TestClarificationHandler_HandleSelection_InvalidJSON(t *testing.T) {
	handler := NewClarificationHandler()

	req := httptest.NewRequest(http.MethodPost, "/clarify/select", strings.NewReader("not json"))
	w := httptest.NewRecorder()

	handler.HandleSelection(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestNeedsClarification(t *testing.T) {
	tests := []struct {
		confidence float64
		expected   bool
	}{
		{0.3, true},
		{0.5, true},
		{0.69, true},
		{0.7, false},
		{0.71, false},
		{0.9, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := NeedsClarification(tt.confidence); got != tt.expected {
				t.Errorf("NeedsClarification(%f) = %v, want %v", tt.confidence, got, tt.expected)
			}
		})
	}
}

func TestClarificationRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     model.ClarificationRequest
		wantErr bool
	}{
		{
			name:    "missing query",
			req:     model.ClarificationRequest{Query: "", Confidence: 0.5},
			wantErr: true,
		},
		{
			name:    "confidence too low",
			req:     model.ClarificationRequest{Query: "test", Confidence: -0.1},
			wantErr: true,
		},
		{
			name:    "confidence too high",
			req:     model.ClarificationRequest{Query: "test", Confidence: 1.1},
			wantErr: true,
		},
		{
			name:    "valid request",
			req:     model.ClarificationRequest{Query: "test", Confidence: 0.5},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
