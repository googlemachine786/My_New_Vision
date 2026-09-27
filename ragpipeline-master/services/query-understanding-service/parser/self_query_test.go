package parser

import (
	"testing"
)

func TestParseRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     ParseRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid request",
			req:     ParseRequest{Query: "Find me class 10 science chapters"},
			wantErr: false,
		},
		{
			name:    "empty query",
			req:     ParseRequest{Query: ""},
			wantErr: true,
			errMsg:  "query is required",
		},
		{
			name:    "query at max length",
			req:     ParseRequest{Query: makeString(2000)},
			wantErr: false,
		},
		{
			name:    "query exceeds max length",
			req:     ParseRequest{Query: makeString(2001)},
			wantErr: true,
			errMsg:  "query exceeds maximum length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseRequest.Validate() = nil, want error containing %q", tt.errMsg)
					return
				}
				if !containsStr(err.Error(), tt.errMsg) {
					t.Errorf("ParseRequest.Validate() error = %q, want containing %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ParseRequest.Validate() = %v, want nil", err)
				}
			}
		})
	}
}

func TestNormalizeGrade(t *testing.T) {
	tests := []struct {
		name  string
		grade string
		want  string
	}{
		{"numeric grade", "10", "10"},
		{"class prefix", "class 10", "10"},
		{"grade prefix", "grade 9", "9"},
		{"suffix th", "12th", "12"},
		{"mixed case", "Class 8", "8"},
		{"whitespace", "  7  ", "7"},
		{"no number", "advanced", "advanced"},
		{"empty string", "", ""},
		{"multiple numbers", "class 9 to 10", "910"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeGrade(tt.grade)
			if got != tt.want {
				t.Errorf("normalizeGrade(%q) = %q, want %q", tt.grade, got, tt.want)
			}
		})
	}
}

func TestNormalizeBoard(t *testing.T) {
	tests := []struct {
		name  string
		board string
		want  string
	}{
		{"cbse lowercase", "cbse", "CBSE"},
		{"cbse mixed case", "CBSE", "CBSE"},
		{"cbse full name", "Central Board of Secondary Education (CBSE)", "CBSE"},
		{"icse lowercase", "icse", "ICSE"},
		{"icse full", "ICSE", "ICSE"},
		{"ib lowercase", "ib", "IB"},
		{"ib full", "IB", "IB"},
		{"unknown board", "state board", "STATE BOARD"},
		{"whitespace", "  cbse  ", "CBSE"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeBoard(tt.board)
			if got != tt.want {
				t.Errorf("normalizeBoard(%q) = %q, want %q", tt.board, got, tt.want)
			}
		})
	}
}

func TestRemoveEmpty(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "no empty strings",
			input: []string{"a", "b", "c"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "with empty strings",
			input: []string{"a", "", "b", "", "c"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "whitespace only",
			input: []string{"a", "  ", "b", "\t", "c"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "all empty",
			input: []string{"", "", ""},
			want:  []string{},
		},
		{
			name:  "empty slice",
			input: []string{},
			want:  []string{},
		},
		{
			name:  "nil slice",
			input: nil,
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removeEmpty(tt.input)

			if len(got) != len(tt.want) {
				t.Errorf("removeEmpty() returned %d items, want %d", len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("removeEmpty()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestFilterSet_EmptyByDefault(t *testing.T) {
	filters := FilterSet{}

	if filters.Grade != nil {
		t.Errorf("Grade = %v, want nil", filters.Grade)
	}
	if filters.Subject != nil {
		t.Errorf("Subject = %v, want nil", filters.Subject)
	}
	if filters.Chapter != nil {
		t.Errorf("Chapter = %v, want nil", filters.Chapter)
	}
	if filters.Topic != nil {
		t.Errorf("Topic = %v, want nil", filters.Topic)
	}
	if filters.ContentType != nil {
		t.Errorf("ContentType = %v, want nil", filters.ContentType)
	}
	if filters.Board != nil {
		t.Errorf("Board = %v, want nil", filters.Board)
	}
	if filters.Language != nil {
		t.Errorf("Language = %v, want nil", filters.Language)
	}
}

func TestParseResponse_Structure(t *testing.T) {
	resp := ParseResponse{
		Filters: FilterSet{
			Grade:   []string{"10"},
			Subject: []string{"science"},
		},
		Confidence:  0.9,
		Explanation: "Found grade and subject filters",
	}

	if len(resp.Filters.Grade) != 1 || resp.Filters.Grade[0] != "10" {
		t.Errorf("Filters.Grade = %v, want [10]", resp.Filters.Grade)
	}
	if len(resp.Filters.Subject) != 1 || resp.Filters.Subject[0] != "science" {
		t.Errorf("Filters.Subject = %v, want [science]", resp.Filters.Subject)
	}
	if resp.Confidence != 0.9 {
		t.Errorf("Confidence = %v, want 0.9", resp.Confidence)
	}
	if resp.Explanation == "" {
		t.Error("Explanation is empty")
	}
}

func TestFallbackResponse(t *testing.T) {
	svc := &Service{}
	resp := svc.fallbackResponse("test query")

	if resp == nil {
		t.Fatal("fallbackResponse returned nil")
	}
	if resp.Confidence != 0.3 {
		t.Errorf("Confidence = %v, want 0.3", resp.Confidence)
	}
	if resp.Explanation == "" {
		t.Error("Explanation is empty")
	}
	if resp.Filters.Grade != nil {
		t.Errorf("Filters.Grade = %v, want nil", resp.Filters.Grade)
	}
}

func makeString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
