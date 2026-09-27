package prompts

import "fmt"

// NLIPromptType defines the type of NLI validation prompt.
type NLIPromptType string

const (
	// NLITypeClaimVerification verifies if a claim is supported by the context.
	NLITypeClaimVerification NLIPromptType = "claim_verification"
	// NLITypeFactExtraction extracts factual claims from a response.
	NLITypeFactExtraction NLIPromptType = "fact_extraction"
)

// NLIClaimVerificationPrompt generates a prompt to verify if a claim is supported by context.
func NLIClaimVerificationPrompt(claim string, context string) string {
	return fmt.Sprintf(`You are a factual verification assistant. Determine whether the given CLAIM is supported by the provided CONTEXT.

CONTEXT:
%s

CLAIM:
%s

TASK:
Determine if the claim is:
- SUPPORTED: The context clearly supports this claim (exact match or paraphrase)
- CONTRADICTED: The context contradicts this claim
- NOT_ENOUGH_INFO: The context doesn't have enough information to verify

OUTPUT FORMAT:
Return ONLY a valid JSON object:
{
  "verdict": "supported" | "contradicted" | "not_enough_info",
  "confidence": 0.95,
  "reasoning": "brief explanation of why this verdict was reached"
}

Return ONLY the JSON object, no additional text.`, context, claim)
}

// NLIFactExtractionPrompt generates a prompt to extract factual claims from text.
func NLIFactExtractionPrompt(text string) string {
	return fmt.Sprintf(`You are a factual claim extraction engine. Extract individual factual claims from the given text.

TEXT:
%s

TASK:
Break down the text into individual factual statements/claims. Each claim should be:
1. A single, atomic fact statement
2. Self-contained (understandable without the original text)
3. Verifiable independently

OUTPUT FORMAT:
Return ONLY a valid JSON array:
{
  "claims": [
    "claim 1",
    "claim 2"
  ]
}

Return ONLY the JSON object, no additional text.`, text)
}

// BuildContextForValidation formats retrieved chunks for validation context.
func BuildContextForValidation(chunks []string) string {
	if len(chunks) == 0 {
		return "(no context provided)"
	}

	result := ""
	for i, chunk := range chunks {
		result += fmt.Sprintf("[Source %d]\n%s\n\n", i+1, chunk)
	}
	return result
}
