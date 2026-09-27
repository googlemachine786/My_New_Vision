// Package types provides shared type definitions for all RAG pipeline services.
package types

import "fmt"

// QueryRequest represents the incoming query request
type QueryRequest struct {
	Query     string `json:"query"`
	SessionID string `json:"session_id,omitempty"`
	Grade     string `json:"grade,omitempty"`
	Subject   string `json:"subject,omitempty"`
	Stream    bool   `json:"stream"`
}

// Validate checks if the request is valid
func (r *QueryRequest) Validate() error {
	if r.Query == "" {
		return &ValidationError{Field: "query", Message: "query is required"}
	}
	if len(r.Query) > 10000 {
		return &ValidationError{Field: "query", Message: "query exceeds maximum length of 10000 characters"}
	}
	return nil
}

// QueryResponse represents a complete query response (non-streaming)
type QueryResponse struct {
	SessionID       string   `json:"session_id"`
	RequestID       string   `json:"request_id"`
	Answer          string   `json:"answer,omitempty"`
	Sources         []Source `json:"sources,omitempty"`
	TotalTokens     int      `json:"total_tokens,omitempty"`
	ConfidenceScore *float64 `json:"confidence_score,omitempty"`   // Story E: 0.0 to 1.0 confidence
	WasFallback     *bool    `json:"was_fallback,omitempty"`       // Story C: indicates fallback was used
	GroundingScore  *float64 `json:"grounding_score,omitempty"`    // Story J: 0.0 to 1.0 grounding validation
	UngroundedClaims []string `json:"ungrounded_claims,omitempty"` // Story J: claims not supported by context
	GroundingWarning *string  `json:"grounding_warning,omitempty"` // Story J: warning message if grounding failed
}

// Source represents a retrieved chunk
type Source struct {
	ParentID string  `json:"parent_id"`
	Content  string  `json:"content"`
	Score    float64 `json:"score"`
}

// ValidationError represents a request validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error code and message
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// SSEEvent represents a Server-Sent Event
type SSEEvent struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// SSEStartEvent is sent when streaming begins
type SSEStartEvent struct {
	SessionID string `json:"session_id"`
	RequestID string `json:"request_id"`
}

// SSEChunkEvent is sent for each content chunk
type SSEChunkEvent struct {
	Content string `json:"content"`
}

// SSEEndEvent is sent when streaming completes
type SSEEndEvent struct {
	Sources     []Source `json:"sources"`
	TotalTokens int      `json:"total_tokens"`
}

// SSEErrorEvent is sent when an error occurs
type SSEErrorEvent struct {
	Error ErrorDetail `json:"error"`
}

// FeedbackRequest represents user feedback
type FeedbackRequest struct {
	Query        string   `json:"query"`
	ResponseID   string   `json:"response_id,omitempty"`
	SessionID    string   `json:"session_id,omitempty"`
	ThumbsUp     *bool    `json:"thumbs_up,omitempty"`
	Rating       string   `json:"rating,omitempty"`
	Comment      string   `json:"comment,omitempty"`
	FeedbackText string   `json:"feedback_text,omitempty"` // Story A: optional "What was wrong?" input
	Grade        string   `json:"grade,omitempty"`
	Subject      string   `json:"subject,omitempty"`
}

// Validate checks if the feedback request is valid
func (r FeedbackRequest) Validate() error {
	if r.Query == "" {
		return fmt.Errorf("query is required")
	}
	return nil
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Services  map[string]string `json:"services,omitempty"`
	Version   string            `json:"version,omitempty"`
}

// StatsResponse represents service statistics
type StatsResponse struct {
	RequestCount    int64                  `json:"request_count"`
	ErrorCount      int64                  `json:"error_count"`
	AverageLatency  string                 `json:"average_latency"`
	CacheHitRate    float64                `json:"cache_hit_rate"`
	CircuitBreakers map[string]interface{} `json:"circuit_breakers,omitempty"`
	RateLimiting    map[string]interface{} `json:"rate_limiting,omitempty"`
	Uptime          string                 `json:"uptime"`
}

// FeedbackContext stores full query + response + retrieved context with feedback (Story A)
type FeedbackContext struct {
	FeedbackID       string   `json:"feedback_id"`
	Query            string   `json:"query"`
	Response         string   `json:"response"`
	RetrievedContext []string `json:"retrieved_context"`
	Grade            string   `json:"grade,omitempty"`
	Subject          string   `json:"subject,omitempty"`
	Timestamp        string   `json:"timestamp"`
}

// ReexplainRequest represents a request to re-explain a concept differently (Story D)
type ReexplainRequest struct {
	QueryID         string `json:"query_id"`
	OriginalQuery   string `json:"original_query"`
	PreviousResponse string `json:"previous_response"`
	SessionID       string `json:"session_id,omitempty"`
}

// ReexplainResponse represents a re-explain response (Story D)
type ReexplainResponse struct {
	ReexplainedAnswer string `json:"reexplained_answer"`
	StrategyUsed      string `json:"strategy_used"`
	AttemptNumber   int    `json:"attempt_number"`
	MaxAttempts     int    `json:"max_attempts"`
	WasFallback       *bool  `json:"was_fallback,omitempty"`
}

// QualityExportRequest represents a request to export weekly quality dataset (Story B)
type QualityExportRequest struct {
	Week string `json:"week"` // YYYY-WW format
}

// QualityExportItem represents a single item in the quality export (Story B)
type QualityExportItem struct {
	Query            string   `json:"query"`
	Response         string   `json:"response"`
	RetrievedContext []string `json:"retrieved_context"`
	FeedbackReason   string   `json:"feedback_reason"`
	ConfidenceScore  float64  `json:"confidence_score"`
	Timestamp        string   `json:"timestamp"`
}

// ReexplainStrategy represents a re-explain prompt template strategy (Story D)
type ReexplainStrategy string

const (
	StrategyAnalogyBased      ReexplainStrategy = "analogy_based"
	StrategyStepByStep        ReexplainStrategy = "step_by_step"
	StrategyRealWorldExample  ReexplainStrategy = "real_world_example"
	StrategySimplifiedLanguage ReexplainStrategy = "simplified_language"
	StrategyVisualDescription ReexplainStrategy = "visual_description"
)
