// Package sanitizer provides input validation and sanitization to prevent prompt injection.
package sanitizer

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	// MaxQueryLength is the maximum allowed query length (1000 chars)
	MaxQueryLength = 1000
	
	// MaxHistoryEntryLength is the max length for each history entry (2000 chars)
	MaxHistoryEntryLength = 2000
	
	// MaxHistoryEntries is the maximum number of history entries
	MaxHistoryEntries = 20
)

// Dangerous patterns that should be rejected from user input
var dangerousPatterns = []*regexp.Regexp{
	// System prompt injection attempts
	regexp.MustCompile(`(?i)ignore\s+previous`),
	regexp.MustCompile(`(?i)disregard\s+(the\s+)?(above|previous)`),
	regexp.MustCompile(`(?i)you\s+are\s+now`),
	regexp.MustCompile(`(?i)system\s*:\s*`),
	regexp.MustCompile(`(?i)<\|system\|>`),
	regexp.MustCompile(`(?i)\[INST\]`),
	
	// Prompt structure manipulation
	regexp.MustCompile(`(?i)new\s+instructions`),
	regexp.MustCompile(`(?i)override\s+(security|rules|instructions)`),
	regexp.MustCompile(`(?i)bypass\s+(safety|restrictions|filters)`),
	regexp.MustCompile(`(?i)jailbreak`),
	
	// Code injection
	regexp.MustCompile(`\{\{.*\}\}`),  // Template injection
	regexp.MustCompile(`\$\{.*\}`),    // Shell variable injection
}

// SanitizeQuery validates and sanitizes a user query.
func SanitizeQuery(query string) error {
	if query == "" {
		return fmt.Errorf("query cannot be empty")
	}

	if utf8.RuneCountInString(query) > MaxQueryLength {
		return fmt.Errorf("query exceeds maximum length of %d characters", MaxQueryLength)
	}

	// Check for dangerous patterns
	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(query) {
			return fmt.Errorf("query contains potentially dangerous pattern: %q", pattern.String())
		}
	}

	return nil
}

// SanitizeHistoryEntry validates a conversation history entry.
func SanitizeHistoryEntry(role, content string) error {
	if role == "" {
		return fmt.Errorf("history entry role cannot be empty")
	}

	if content == "" {
		return fmt.Errorf("history entry content cannot be empty")
	}

	// Validate role
	if role == "system" {
		return fmt.Errorf("users cannot inject system messages")
	}

	validRoles := map[string]struct{}{
		"user":      {},
		"assistant": {},
	}

	if _, ok := validRoles[role]; !ok {
		return fmt.Errorf("invalid history entry role: %q (must be 'user' or 'assistant')", role)
	}

	if utf8.RuneCountInString(content) > MaxHistoryEntryLength {
		return fmt.Errorf("history entry content exceeds maximum length of %d characters", MaxHistoryEntryLength)
	}

	return nil
}

// ValidateHistory validates the entire conversation history.
func ValidateHistory(history []map[string]string) error {
	if len(history) > MaxHistoryEntries {
		return fmt.Errorf("history exceeds maximum of %d entries (got %d)", MaxHistoryEntries, len(history))
	}

	for i, entry := range history {
		role, roleOk := entry["role"]
		content, contentOk := entry["content"]

		if !roleOk || !contentOk {
			return fmt.Errorf("history entry %d missing 'role' or 'content' field", i)
		}

		if err := SanitizeHistoryEntry(role, content); err != nil {
			return fmt.Errorf("invalid history entry %d: %w", i, err)
		}
	}

	return nil
}

// SanitizeForPrompt escapes user input to make it safe for inclusion in prompts.
// This prevents users from breaking out of the intended prompt structure.
func SanitizeForPrompt(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")
	
	// Remove control characters except newlines and tabs
	var result strings.Builder
	for _, r := range input {
		if r == '\n' || r == '\t' || (r >= 32 && r != 127) {
			result.WriteRune(r)
		}
	}
	
	return strings.TrimSpace(result.String())
}

// ValidateConfidence checks if a confidence score is valid (between 0 and 1).
func ValidateConfidence(confidence float64) error {
	if confidence < 0 || confidence > 1 {
		return fmt.Errorf("confidence score must be between 0 and 1, got %f", confidence)
	}
	return nil
}

// ValidatePromptVersion checks if a prompt version string is valid.
func ValidatePromptVersion(version string) error {
	if version == "" {
		return fmt.Errorf("prompt version cannot be empty")
	}

	// Must match pattern like "v1", "v2.1", "2024-01-15"
	validPattern := regexp.MustCompile(`^v\d+(\.\d+)?$|^\d{4}-\d{2}-\d{2}$`)
	if !validPattern.MatchString(version) {
		return fmt.Errorf("invalid prompt version format: %q (expected 'v1', 'v2.1', or '2024-01-15')", version)
	}

	return nil
}
