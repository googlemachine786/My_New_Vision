package handler

import (
	"testing"

	"github.com/visionary/ragpipeline/pkg/types"
)

func TestReexplainSessionKey(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		queryID   string
		want      string
	}{
		{
			name:      "with session ID",
			sessionID: "session123",
			queryID:   "query456",
			want:      "rag:reexplain:session:session123:query456",
		},
		{
			name:      "without session ID",
			sessionID: "",
			queryID:   "query456",
			want:      "rag:reexplain:session:query456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reexplainSessionKey(tt.sessionID, tt.queryID)
			if got != tt.want {
				t.Errorf("reexplainSessionKey(%q, %q) = %q, want %q", tt.sessionID, tt.queryID, got, tt.want)
			}
		})
	}
}

func TestReexplainHandler_validateRequest(t *testing.T) {
	handler := &ReexplainHandler{}

	tests := []struct {
		name    string
		req     types.ReexplainRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: types.ReexplainRequest{
				QueryID:         "q1",
				OriginalQuery:   "What is photosynthesis?",
				PreviousResponse: "Photosynthesis is...",
				SessionID:       "s1",
			},
			wantErr: false,
		},
		{
			name: "missing query_id",
			req: types.ReexplainRequest{
				QueryID:         "",
				OriginalQuery:   "What is photosynthesis?",
				PreviousResponse: "Photosynthesis is...",
			},
			wantErr: true,
			errMsg:  "query_id is required",
		},
		{
			name: "missing original_query",
			req: types.ReexplainRequest{
				QueryID:         "q1",
				OriginalQuery:   "",
				PreviousResponse: "Photosynthesis is...",
			},
			wantErr: true,
			errMsg:  "original_query is required",
		},
		{
			name: "missing previous_response",
			req: types.ReexplainRequest{
				QueryID:         "q1",
				OriginalQuery:   "What is photosynthesis?",
				PreviousResponse: "",
			},
			wantErr: true,
			errMsg:  "previous_response is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateRequest(&tt.req)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateRequest() expected error, got nil")
				} else if err.Error() != tt.errMsg {
					t.Errorf("validateRequest() error = %q, want %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateRequest() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestMaxReexplainAttempts(t *testing.T) {
	if MaxReexplainAttempts != 3 {
		t.Errorf("MaxReexplainAttempts = %d, want 3", MaxReexplainAttempts)
	}
}

func TestReexplainTTL(t *testing.T) {
	if ReexplainTTL <= 0 {
		t.Errorf("ReexplainTTL = %v, should be positive", ReexplainTTL)
	}
}
