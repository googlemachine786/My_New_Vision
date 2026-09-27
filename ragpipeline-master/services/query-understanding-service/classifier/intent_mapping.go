// Package classifier provides query intent classification services.
package classifier

import "strings"

// MapLegacyToFourCategory maps a legacy intent to the 4-category system (Story F)
func MapLegacyToFourCategory(legacy IntentType) IntentType {
	legacyMap := map[IntentType]IntentType{
		IntentInformationSeeking: IntentConcept,
		IntentDefinition:         IntentConcept,
		IntentClarification:      IntentConcept,
		IntentConversational:     IntentConcept,

		IntentProblemSolving:   IntentNumerical,
		IntentExamPrep:         IntentNumerical,

		IntentContentRetrieval: IntentFactCheck,
		IntentComparison:       IntentFactCheck,

		IntentProcedural: IntentStrategy,
	}

	if mapped, ok := legacyMap[legacy]; ok {
		return mapped
	}

	// Default fallback
	return IntentConcept
}

// MapKeywordToIntent maps common query keywords to the 4 intent categories (Story F)
func MapKeywordToIntent(query string) IntentType {
	// Normalize query to lowercase (spaces → underscores)
	q := normalizeIntentString(query)

	// Check for "step_by_step" or "process_for" or "method_to" patterns first — strong strategy signals
	// (but NOT "process_of" which is a common concept phrase)
	if strings.Contains(q, "step_by_step") || strings.Contains(q, "process_for") ||
		strings.Contains(q, "method_to") || strings.Contains(q, "method_for") ||
		strings.Contains(q, "how_do_i_") || strings.Contains(q, "how_do_we_") {
		return IntentStrategy
	}

	// Strategy keywords (how_to, steps_to, process_for) — check first as most specific
	strategyKeywords := []string{
		"how_to", "steps_to", "steps_for", "process_for",
		"approach_to", "best_way_to", "procedure_for",
		"guide_me", "walk_me_through", "instructions_for",
		"walk_through",
	}
	for _, kw := range strategyKeywords {
		if containsIntentKeyword(q, kw) {
			return IntentStrategy
		}
	}

	// Fact-check keywords (is_x_y?, true/false, verify, did_x, compare)
	factCheckKeywords := []string{
		"true_or_false", "is_it_true", "is_it_correct",
		"verify", "validate", "confirm", "check_if",
		"fact_check", "is_this_accurate", "did_this_happen",
		"compare", "difference_between", "versus", "vs",
	}
	// Check "did_x" pattern for fact-checking (not "when did" or "where did")
	if strings.HasPrefix(q, "did_") {
		return IntentFactCheck
	}
	// Check "is_x_the_y?" pattern for fact-checking
	if strings.HasPrefix(q, "is_") || strings.HasPrefix(q, "was_") || strings.HasPrefix(q, "are_") {
		return IntentFactCheck
	}
	for _, kw := range factCheckKeywords {
		if containsIntentKeyword(q, kw) {
			return IntentFactCheck
		}
	}

	// Numerical keywords (solve, calculate, %, find value, divide)
	numericalKeywords := []string{
		"calculate", "compute", "find_the_value", "find_the",
		"what_is_the_value", "evaluate", "simplify",
		"derivative", "integral",
		"arithmetic", "algebra",
		"divided_by", "multiply", "subtract", "add_",
	}
	// Check for percentage or numeric patterns
	if strings.Contains(q, "%") || strings.Contains(q, "_of_") && hasNumericPattern(q) {
		return IntentNumerical
	}
	for _, kw := range numericalKeywords {
		if containsIntentKeyword(q, kw) {
			return IntentNumerical
		}
	}
	// "solve" is numerical but only when not "how_to solve"
	if strings.Contains(q, "solve") && !strings.Contains(q, "how_to") {
		return IntentNumerical
	}

	// Concept keywords (what_is, explain, define) — default category
	conceptKeywords := []string{
		"what_is", "what_are", "what_was", "what_were",
		"explain", "describe", "define", "meaning",
		"who_is", "who_was", "where_is", "where_did",
		"why_did", "why_is", "why_are", "how_does",
		"tell_me_about", "understand", "concept",
		"equation", "formula", "math", "geometry",
	}
	for _, kw := range conceptKeywords {
		if containsIntentKeyword(q, kw) {
			return IntentConcept
		}
	}

	// Default to Concept
	return IntentConcept
}

// hasNumericPattern checks if query contains number patterns
func hasNumericPattern(q string) bool {
	for _, r := range q {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

// containsIntentKeyword checks if the query contains a keyword (handling spaces/underscores)
func containsIntentKeyword(query, keyword string) bool {
	// Check with spaces
	if contains(query, keyword) {
		return true
	}
	// Check with underscores
	underscored := replaceSpaces(keyword, '_')
	if contains(query, underscored) {
		return true
	}
	// Check with hyphens
	hyphenated := replaceSpaces(keyword, '-')
	if contains(query, hyphenated) {
		return true
	}
	return false
}

// contains checks if s contains substr (simple string contains)
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// replaceSpaces replaces spaces in s with the given character
func replaceSpaces(s string, repl byte) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' {
			result[i] = repl
		} else {
			result[i] = s[i]
		}
	}
	return string(result)
}

// ClassifyIntentFromKeywords classifies a query's intent based on keyword matching (Story F)
// This is a fast, rule-based classifier that can be used as a fallback or validation.
func ClassifyIntentFromKeywords(query string) IntentType {
	return MapKeywordToIntent(query)
}
