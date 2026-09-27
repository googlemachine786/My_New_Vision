package parser

import (
	"context"
	"fmt"
	"strings"

	"query-understanding-service/llm"
	"query-understanding-service/prompts"
)

// FilterSet holds extracted metadata filters from a query.
type FilterSet struct {
	Grade       []string `json:"grade"`
	Subject     []string `json:"subject"`
	Chapter     []string `json:"chapter"`
	Topic       []string `json:"topic"`
	ContentType []string `json:"content_type"`
	Board       []string `json:"board"`
	Language    []string `json:"language"`
}

// ParseRequest is the input for self-query parsing.
type ParseRequest struct {
	Query string `json:"query"`
}

// ParseResponse is the output from self-query parsing.
type ParseResponse struct {
	Filters     FilterSet `json:"filters"`
	Confidence  float64   `json:"confidence"`
	Explanation string    `json:"explanation,omitempty"`
}

// Validate checks that the parse request has the required fields.
func (r *ParseRequest) Validate() error {
	if r.Query == "" {
		return fmt.Errorf("query is required")
	}
	if len(r.Query) > 2000 {
		return fmt.Errorf("query exceeds maximum length of 2000 characters")
	}
	return nil
}

// Service handles self-query parsing (metadata filter extraction).
type Service struct {
	llmClient     *llm.Client
	promptVersion string
}

// NewService creates a new parse service.
func NewService(client *llm.Client, promptVersion string) *Service {
	return &Service{
		llmClient:     client,
		promptVersion: promptVersion,
	}
}

// Parse extracts metadata filters from a query.
func (s *Service) Parse(ctx context.Context, req *ParseRequest) (*ParseResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	prompt := prompts.SelfQueryParsePrompt(req.Query, s.promptVersion)

	var result ParseResponse
	if err := s.llmClient.GenerateJSON(ctx, prompt, &result); err != nil {
		return s.fallbackResponse(req.Query), nil
	}

	// Normalize extracted filters
	s.normalizeFilters(&result.Filters)

	return &result, nil
}

// normalizeFilters applies normalization rules to extracted filters.
func (s *Service) normalizeFilters(filters *FilterSet) {
	// Normalize subject: lowercase and trim
	for i, subj := range filters.Subject {
		filters.Subject[i] = strings.ToLower(strings.TrimSpace(subj))
	}

	// Normalize content_type: lowercase and trim
	for i, ct := range filters.ContentType {
		filters.ContentType[i] = strings.ToLower(strings.TrimSpace(ct))
	}

	// Normalize grade: extract just the number
	for i, grade := range filters.Grade {
		filters.Grade[i] = normalizeGrade(grade)
	}

	// Normalize board: uppercase standard boards
	for i, board := range filters.Board {
		filters.Board[i] = normalizeBoard(board)
	}

	// Normalize language: lowercase
	for i, lang := range filters.Language {
		filters.Language[i] = strings.ToLower(strings.TrimSpace(lang))
	}

	// Remove empty strings from all slices
	filters.Subject = removeEmpty(filters.Subject)
	filters.Grade = removeEmpty(filters.Grade)
	filters.Chapter = removeEmpty(filters.Chapter)
	filters.Topic = removeEmpty(filters.Topic)
	filters.ContentType = removeEmpty(filters.ContentType)
	filters.Board = removeEmpty(filters.Board)
	filters.Language = removeEmpty(filters.Language)
}

// normalizeGrade extracts the numeric portion from grade strings.
// "class 10" -> "10", "grade 9" -> "9", "12th" -> "12"
func normalizeGrade(grade string) string {
	grade = strings.TrimSpace(grade)

	// Try to find a number in the grade string
	for _, ch := range grade {
		if ch >= '0' && ch <= '9' {
			// Extract all consecutive digits
			var num []rune
			for _, c := range grade {
				if c >= '0' && c <= '9' {
					num = append(num, c)
				}
			}
			if len(num) > 0 {
				return string(num)
			}
		}
	}
	return grade
}

// normalizeBoard standardizes education board names.
func normalizeBoard(board string) string {
	board = strings.TrimSpace(board)
	upper := strings.ToUpper(board)

	// Map common variants
	switch {
	case strings.Contains(upper, "CBSE"):
		return "CBSE"
	case strings.Contains(upper, "ICSE"):
		return "ICSE"
	case strings.Contains(upper, "IB"):
		return "IB"
	default:
		return strings.ToUpper(board)
	}
}

// removeEmpty removes empty strings from a slice.
func removeEmpty(items []string) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			result = append(result, item)
		}
	}
	return result
}

// fallbackResponse returns a safe fallback when the LLM fails.
func (s *Service) fallbackResponse(query string) *ParseResponse {
	return &ParseResponse{
		Filters:    FilterSet{},
		Confidence: 0.3,
		Explanation: "LLM unavailable, no filters extracted",
	}
}
