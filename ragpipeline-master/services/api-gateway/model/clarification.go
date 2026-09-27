// Package model provides shared types for clarification request/response flows.
package model

import "fmt"

// ClarificationOption represents a single clarification choice presented to the user
type ClarificationOption struct {
	// ID is a unique identifier for this option
	ID string `json:"id"`
	// Label is the human-readable text shown to the user
	Label string `json:"label"`
	// Description provides additional context about what this option means
	Description string `json:"description,omitempty"`
	// Type indicates the kind of clarification (topic, grade, intent, etc.)
	Type string `json:"type"`
}

// ClarificationRequest represents a request for clarification options
type ClarificationRequest struct {
	// Query is the original user query that triggered low confidence
	Query string `json:"query"`
	// Intent is the detected intent (if any)
	Intent string `json:"intent,omitempty"`
	// Confidence is the confidence score of the intent classification
	Confidence float64 `json:"confidence"`
	// Grade is the detected grade (if applicable)
	Grade string `json:"grade,omitempty"`
	// Subject is the detected subject (if applicable)
	Subject string `json:"subject,omitempty"`
}

// Validate checks if the clarification request has required fields
func (r *ClarificationRequest) Validate() error {
	if r.Query == "" {
		return fmt.Errorf("query is required")
	}
	if r.Confidence < 0 || r.Confidence > 1 {
		return fmt.Errorf("confidence must be between 0 and 1")
	}
	return nil
}

// ClarificationResponse represents the response with clarification options
type ClarificationResponse struct {
	// NeedsClarification indicates whether clarification is needed
	NeedsClarification bool `json:"needs_clarification"`
	// Options contains the available clarification choices
	Options []ClarificationOption `json:"options,omitempty"`
	// Message is a human-readable prompt asking for clarification
	Message string `json:"message"`
	// OriginalQuery is the query that triggered the clarification
	OriginalQuery string `json:"original_query"`
	// Intent is the detected intent (may be uncertain)
	Intent string `json:"intent,omitempty"`
	// Confidence is the confidence score that triggered clarification
	Confidence float64 `json:"confidence"`
}

// ClarificationSelection represents the user's selection from clarification options
type ClarificationSelection struct {
	// OptionID is the ID of the selected clarification option
	OptionID string `json:"option_id"`
	// OriginalQuery is the original query that triggered clarification
	OriginalQuery string `json:"original_query"`
	// SessionID is the session identifier for conversation tracking
	SessionID string `json:"session_id,omitempty"`
}

// Validate checks if the clarification selection is valid
func (s *ClarificationSelection) Validate() error {
	if s.OptionID == "" {
		return fmt.Errorf("option_id is required")
	}
	if s.OriginalQuery == "" {
		return fmt.Errorf("original_query is required")
	}
	return nil
}

// ClarificationResult represents the result after the user selects an option
type ClarificationResult struct {
	// EnhancedQuery is the query enhanced with the user's clarification
	EnhancedQuery string `json:"enhanced_query"`
	// SelectedOption is the option the user selected
	SelectedOption ClarificationOption `json:"selected_option"`
	// OriginalQuery is the original query
	OriginalQuery string `json:"original_query"`
}

// Clarification types
const (
	ClarificationTypeTopic   = "topic"
	ClarificationTypeGrade   = "grade"
	ClarificationTypeIntent  = "intent"
)
