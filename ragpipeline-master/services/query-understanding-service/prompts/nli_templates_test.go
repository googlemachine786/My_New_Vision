package prompts

import "testing"

func TestNLIClaimVerificationPrompt(t *testing.T) {
	claim := "Photosynthesis occurs in chloroplasts."
	context := "Photosynthesis is the process by which plants convert sunlight into energy. This process occurs in organelles called chloroplasts."

	prompt := NLIClaimVerificationPrompt(claim, context)
	if prompt == "" {
		t.Fatal("expected non-empty prompt")
	}
}

func TestNLIFactExtractionPrompt(t *testing.T) {
	text := "Water boils at 100 degrees Celsius at sea level. Ice melts at 0 degrees Celsius."

	prompt := NLIFactExtractionPrompt(text)
	if prompt == "" {
		t.Fatal("expected non-empty prompt")
	}
}

func TestBuildContextForValidation(t *testing.T) {
	chunks := []string{
		"First chunk content.",
		"Second chunk content.",
	}

	result := BuildContextForValidation(chunks)
	if result == "" {
		t.Fatal("expected non-empty context")
	}
}

func TestBuildContextForValidation_Empty(t *testing.T) {
	result := BuildContextForValidation(nil)
	if result != "(no context provided)" {
		t.Errorf("expected '(no context provided)', got %q", result)
	}
}
