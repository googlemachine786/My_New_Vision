package classifier

import (
	"context"
	"fmt"
	"strings"
	"time"

	"query-understanding-service/llm"
	"query-understanding-service/middleware"
	"query-understanding-service/prompts"
)

// IntentType represents a query intent category.
type IntentType string

// Story F: 4 required categories
const (
	IntentConcept     IntentType = "concept"      // what is, explain, define
	IntentNumerical   IntentType = "numerical"    // solve, calculate, find value
	IntentFactCheck   IntentType = "fact_check"   // true/false, verify, is it correct
	IntentStrategy    IntentType = "strategy"     // how to, steps, process
)

// Legacy intent types (maintained for backward compatibility)
const (
	IntentInformationSeeking IntentType = "information_seeking"
	IntentProblemSolving     IntentType = "problem_solving"
	IntentContentRetrieval   IntentType = "content_retrieval"
	IntentClarification      IntentType = "clarification"
	IntentComparison         IntentType = "comparison"
	IntentDefinition         IntentType = "definition"
	IntentProcedural         IntentType = "procedural"
	IntentConversational     IntentType = "conversational"
	IntentExamPrep           IntentType = "exam_prep"
)

// ValidIntents returns all recognized intent types (Story F: 4 categories).
func ValidIntents() []IntentType {
	return []IntentType{
		IntentConcept,
		IntentNumerical,
		IntentFactCheck,
		IntentStrategy,
	}
}

// LegacyValidIntents returns the original 9 intent types for backward compatibility.
func LegacyValidIntents() []IntentType {
	return []IntentType{
		IntentInformationSeeking,
		IntentProblemSolving,
		IntentContentRetrieval,
		IntentClarification,
		IntentComparison,
		IntentDefinition,
		IntentProcedural,
		IntentConversational,
		IntentExamPrep,
	}
}

// IsValidIntent checks if an intent string is a recognized type (checks both 4 and 9 category sets).
func IsValidIntent(s string) bool {
	t := IntentType(s)
	// Check 4 required categories first
	for _, valid := range ValidIntents() {
		if t == valid {
			return true
		}
	}
	// Check legacy intents for backward compatibility
	for _, valid := range LegacyValidIntents() {
		if t == valid {
			return true
		}
	}
	return false
}

// ClassifyRequest is the input for intent classification.
type ClassifyRequest struct {
	Query string `json:"query"`
}

// ClassifyResponse is the output from intent classification.
type ClassifyResponse struct {
	Intent        IntentType `json:"intent"`
	Confidence    float64    `json:"confidence"`
	SecondaryIntent *IntentType `json:"secondary_intent,omitempty"`
	Reasoning     string     `json:"reasoning"`
}

// Validate checks that the classify request has the required fields.
func (r *ClassifyRequest) Validate() error {
	if r.Query == "" {
		return fmt.Errorf("query is required")
	}
	if len(r.Query) > 2000 {
		return fmt.Errorf("query exceeds maximum length of 2000 characters")
	}
	return nil
}

// Service handles intent classification.
type Service struct {
	llmClient     *llm.Client
	promptVersion string
}

// NewService creates a new intent classification service.
func NewService(client *llm.Client, promptVersion string) *Service {
	return &Service{
		llmClient:     client,
		promptVersion: promptVersion,
	}
}

// Classify determines the intent of a user query.
func (s *Service) Classify(ctx context.Context, req *ClassifyRequest) (*ClassifyResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	start := time.Now()
	prompt := prompts.IntentClassificationPrompt(req.Query, s.promptVersion)

	var result ClassifyResponse
	if err := s.llmClient.GenerateJSON(ctx, prompt, &result); err != nil {
		return s.fallbackResponse(req.Query), nil
	}

	// Record LLM latency
	middleware.RecordLLMLatency("intent_classification", time.Since(start).Seconds())

	// Validate and normalize intent
	if result.Intent == "" {
		result.Intent = IntentInformationSeeking
		result.Confidence = 0.3
	}

	if !IsValidIntent(string(result.Intent)) {
		// Map common variants
		result.Intent = mapIntent(string(result.Intent))
	}

	// Record confidence metric
	middleware.RecordIntentConfidence(string(result.Intent), result.Confidence)

	return &result, nil
}

// mapIntent maps common intent string variants to canonical types.
func mapIntent(s string) IntentType {
	s = normalizeIntentString(s)
	intentMap := map[string]IntentType{
		// Story F: 4 required categories
		"concept":       IntentConcept,
		"numerical":     IntentNumerical,
		"fact_check":    IntentFactCheck,
		"factcheck":     IntentFactCheck,
		"strategy":      IntentStrategy,

		// Legacy mappings - map to 4 categories
		"information_seeking": IntentConcept,
		"information":         IntentConcept,
		"question":            IntentConcept,
		"explanation":         IntentConcept,
		"info":                IntentConcept,
		"define":              IntentConcept,
		"definition":          IntentConcept,
		"meaning":             IntentConcept,
		"what_is":             IntentConcept,
		"explain":             IntentConcept,
		"clarification":       IntentConcept,
		"clarify":             IntentConcept,
		"elaborate":           IntentConcept,

		"problem_solving":     IntentNumerical,
		"problem":             IntentNumerical,
		"solve":               IntentNumerical,
		"calculation":         IntentNumerical,
		"calculate":           IntentNumerical,
		"math":                IntentNumerical,
		"find_value":          IntentNumerical,
		"exam_prep":           IntentNumerical,
		"exam":                IntentNumerical,
		"test":                IntentNumerical,

		"content_retrieval":   IntentFactCheck,
		"retrieval":           IntentFactCheck,
		"verify":              IntentFactCheck,
		"true_false":          IntentFactCheck,
		"is_correct":          IntentFactCheck,
		"comparison":          IntentFactCheck,
		"compare":             IntentFactCheck,
		"difference":          IntentFactCheck,

		"procedural":          IntentStrategy,
		"procedure":           IntentStrategy,
		"how_to":              IntentStrategy,
		"howto":               IntentStrategy,
		"instructions":        IntentStrategy,
		"steps":               IntentStrategy,
		"process":             IntentStrategy,
		"method":              IntentStrategy,

		"conversational":      IntentConcept, // Default to Concept for conversational
		"conversation":        IntentConcept,
		"greeting":            IntentConcept,
		"chitchat":            IntentConcept,
		"chit_chat":           IntentConcept,
		"social":              IntentConcept,
	}

	if intent, ok := intentMap[s]; ok {
		return intent
	}

	// Default fallback
	return IntentConcept
}

// normalizeIntentString normalizes an intent string for lookup.
func normalizeIntentString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	// Replace spaces and hyphens with underscores
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '-' {
			result[i] = '_'
		} else {
			result[i] = s[i]
		}
	}
	return string(result)
}

// fallbackResponse returns a safe fallback when the LLM fails.
func (s *Service) fallbackResponse(query string) *ClassifyResponse {
	intent := IntentInformationSeeking
	return &ClassifyResponse{
		Intent:     intent,
		Confidence: 0.3,
		Reasoning:  "LLM unavailable, defaulting to information_seeking",
	}
}
