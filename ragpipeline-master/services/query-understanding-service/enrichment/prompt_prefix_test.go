package enrichment

import "testing"

func TestPromptPrefixBuilder_FullContext(t *testing.T) {
	builder := NewPromptPrefixBuilder()
	result := builder.
		Board("CBSE").
		Grade("10").
		Language("English").
		CurrentSubject("Physics").
		CurrentChapter("Optics").
		Build()

	if result == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestPromptPrefixBuilder_MinimalContext(t *testing.T) {
	builder := NewPromptPrefixBuilder()
	result := builder.Grade("8").Build()

	if result == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestPromptPrefixBuilder_WithErrors(t *testing.T) {
	errors := []PracticeError{
		{Topic: "physics", ErrorMsg: "unit conversion", Count: 2},
	}

	builder := NewPromptPrefixBuilder()
	result := builder.
		Grade("9").
		PracticeErrors(errors).
		Build()

	if result == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestPromptPrefixBuilder_BuildWithQuery(t *testing.T) {
	builder := NewPromptPrefixBuilder()
	result := builder.
		Grade("10").
		BuildWithQuery("What is force?")

	if result == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestPromptPrefixBuilder_Empty(t *testing.T) {
	builder := NewPromptPrefixBuilder()
	result := builder.Build()

	if result != "" {
		t.Errorf("expected empty result for no context, got %q", result)
	}
}

func TestPromptPrefixBuilder_BuildWithQuery_Empty(t *testing.T) {
	builder := NewPromptPrefixBuilder()
	result := builder.BuildWithQuery("What is force?")

	if result != "What is force?" {
		t.Errorf("expected just the query for empty context, got %q", result)
	}
}

func TestItoS(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{-1, "-1"},
		{100, "100"},
	}

	for _, tt := range tests {
		result := itos(tt.input)
		if result != tt.expected {
			t.Errorf("itos(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
