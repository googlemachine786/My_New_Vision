package handler

import (
	"testing"
	"time"
)

func TestNewNLIClient_DefaultURL(t *testing.T) {
	client := NewNLIClient("")
	if client.baseURL != "http://localhost:8084" {
		t.Errorf("expected default URL http://localhost:8084, got %s", client.baseURL)
	}
}

func TestNewNLIClient_CustomURL(t *testing.T) {
	client := NewNLIClient("http://nli-service:8084")
	if client.baseURL != "http://nli-service:8084" {
		t.Errorf("expected custom URL, got %s", client.baseURL)
	}
}

func TestNLIClient_Timeout(t *testing.T) {
	client := NewNLIClient("http://localhost:9999") // non-existent
	client.client.Timeout = 100 * time.Millisecond

	// Should fail gracefully with timeout
	_, err := client.SplitSentences(nil, "test")
	if err == nil {
		t.Error("expected error for unreachable service")
	}
}

func TestBatchValidateClaims_EmptyClaims(t *testing.T) {
	client := NewNLIClient("http://localhost:9999")
	result, err := client.BatchValidateClaims(nil, []string{}, []string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result for empty claims")
	}
}
