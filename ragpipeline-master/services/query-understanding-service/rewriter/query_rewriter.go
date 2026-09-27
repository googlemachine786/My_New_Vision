package rewriter

import (
	"context"
	"fmt"

	"query-understanding-service/llm"
	"query-understanding-service/prompts"
)

// RewriteRequest is the input for query rewriting.
type RewriteRequest struct {
	Query   string                  `json:"query"`
	History []prompts.HistoryMessage `json:"history,omitempty"`
}

// RewriteResponse is the output from query rewriting.
type RewriteResponse struct {
	RewrittenQuery string   `json:"rewritten_query"`
	Confidence     float64  `json:"confidence"`
	ChangesMade    []string `json:"changes_made,omitempty"`
}

// Validate checks that the rewrite request has the required fields.
func (r *RewriteRequest) Validate() error {
	if r.Query == "" {
		return fmt.Errorf("query is required")
	}
	if len(r.Query) > 2000 {
		return fmt.Errorf("query exceeds maximum length of 2000 characters")
	}
	if len(r.History) > 10 {
		return fmt.Errorf("history exceeds maximum of 10 turns")
	}
	return nil
}

// ServiceOption configures a Service instance.
type ServiceOption interface {
	apply(*serviceOptions)
}

type serviceOptions struct {
	maxTurns      int
	promptVersion string
}

type maxTurnsOption struct {
	MaxTurns int
}

func (o maxTurnsOption) apply(opts *serviceOptions) { opts.maxTurns = o.MaxTurns }

type promptVersionOption struct {
	Version string
}

func (o promptVersionOption) apply(opts *serviceOptions) { opts.promptVersion = o.Version }

// WithMaxTurns sets the maximum number of conversation turns to consider.
func WithMaxTurns(n int) ServiceOption {
	return maxTurnsOption{MaxTurns: n}
}

// WithPromptVersion sets the prompt version to use.
func WithPromptVersion(v string) ServiceOption {
	return promptVersionOption{Version: v}
}

// Service handles query rewriting with conversation context.
type Service struct {
	llmClient    *llm.Client
	maxTurns     int
	promptVersion string
}

// NewService creates a new rewrite service.
// Defaults: MaxTurns=10, PromptVersion="v1".
func NewService(client *llm.Client, opts ...ServiceOption) *Service {
	options := serviceOptions{
		maxTurns:      10,
		promptVersion: "v1",
	}
	for _, o := range opts {
		o.apply(&options)
	}

	return &Service{
		llmClient:    client,
		maxTurns:     options.maxTurns,
		promptVersion: options.promptVersion,
	}
}

// Rewrite rewrites the query incorporating conversation history.
func (s *Service) Rewrite(ctx context.Context, req *RewriteRequest) (*RewriteResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Truncate history to max turns
	history := req.History
	if len(history) > s.maxTurns {
		history = history[len(history)-s.maxTurns:]
	}

	// If no history, return the original query with high confidence
	if len(history) == 0 {
		return &RewriteResponse{
			RewrittenQuery: req.Query,
			Confidence:     1.0,
			ChangesMade:    []string{"no history provided, query returned as-is"},
		}, nil
	}

	// Build the prompt and call Gemini
	prompt := prompts.QueryRewritePrompt(req.Query, history, s.promptVersion)

	var result RewriteResponse
	if err := s.llmClient.GenerateJSON(ctx, prompt, &result); err != nil {
		// Fallback: return original query with reduced confidence
		return s.fallbackResponse(req.Query), nil
	}

	// Validate the response
	if result.RewrittenQuery == "" {
		result.RewrittenQuery = req.Query
		result.Confidence = 0.5
	}

	return &result, nil
}

// fallbackResponse returns a safe fallback when the LLM fails.
func (s *Service) fallbackResponse(query string) *RewriteResponse {
	return &RewriteResponse{
		RewrittenQuery: query,
		Confidence:     0.5,
		ChangesMade:    []string{"LLM unavailable, returning original query"},
	}
}
