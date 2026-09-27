package model

import "testing"

func TestClarificationRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     ClarificationRequest
		wantErr bool
	}{
		{
			name:    "missing query",
			req:     ClarificationRequest{Query: "", Confidence: 0.5},
			wantErr: true,
		},
		{
			name:    "negative confidence",
			req:     ClarificationRequest{Query: "test", Confidence: -0.1},
			wantErr: true,
		},
		{
			name:    "confidence > 1",
			req:     ClarificationRequest{Query: "test", Confidence: 1.1},
			wantErr: true,
		},
		{
			name:    "valid request",
			req:     ClarificationRequest{Query: "What is photosynthesis?", Confidence: 0.5},
			wantErr: false,
		},
		{
			name:    "confidence = 0",
			req:     ClarificationRequest{Query: "test", Confidence: 0.0},
			wantErr: false,
		},
		{
			name:    "confidence = 1",
			req:     ClarificationRequest{Query: "test", Confidence: 1.0},
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

func TestClarificationSelection_Validate(t *testing.T) {
	tests := []struct {
		name    string
		sel     ClarificationSelection
		wantErr bool
	}{
		{
			name:    "missing option_id",
			sel:     ClarificationSelection{OptionID: "", OriginalQuery: "test"},
			wantErr: true,
		},
		{
			name:    "missing original_query",
			sel:     ClarificationSelection{OptionID: "opt-1", OriginalQuery: ""},
			wantErr: true,
		},
		{
			name:    "valid selection",
			sel:     ClarificationSelection{OptionID: "opt-1", OriginalQuery: "test query"},
			wantErr: false,
		},
		{
			name:    "valid with session_id",
			sel:     ClarificationSelection{OptionID: "opt-1", OriginalQuery: "test", SessionID: "sess-1"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sel.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
