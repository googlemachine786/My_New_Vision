// Package llm provides the Gemini LLM client via Vertex AI aiplatform SDK.
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/cenkalti/backoff/v4"
	"golang.org/x/time/rate"
	"google.golang.org/protobuf/types/known/structpb"

	"query-understanding-service/config"
)

// Client wraps the Vertex AI aiplatform PredictionClient for Gemini LLM calls.
type Client struct {
	predictor   *aiplatform.PredictionClient
	projectID   string
	location    string
	modelName   string
	config      config.GeminiConfig
	mu          sync.Mutex
	rateLimiter *RateLimiter
	circuitBreaker *CircuitBreaker
}

// NewClient creates a new Gemini client using Vertex AI aiplatform SDK.
func NewClient(ctx context.Context, cfg config.GeminiConfig) (*Client, error) {
	if cfg.Model == "" {
		return nil, fmt.Errorf("model name is required")
	}
	if cfg.ProjectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if cfg.Location == "" {
		return nil, fmt.Errorf("location is required")
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 30 * time.Second
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}

	predictor, err := aiplatform.NewPredictionClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vertex AI prediction client: %w", err)
	}

	rateLimiter := NewRateLimiter(cfg.RateLimitPerSecond, cfg.RateLimitBurst)
	circuitBreaker := NewCircuitBreaker(5, 60*time.Second)

	return &Client{
		predictor:      predictor,
		projectID:      cfg.ProjectID,
		location:       cfg.Location,
		modelName:      cfg.Model,
		config:         cfg,
		rateLimiter:    rateLimiter,
		circuitBreaker: circuitBreaker,
	}, nil
}

// GenerateJSON sends a prompt to Gemini and parses the JSON response.
func (c *Client) GenerateJSON(ctx context.Context, prompt string, target interface{}) error {
	return c.generateWithRetry(ctx, prompt, target, GenerationConfig{})
}

// GenerateText sends a prompt and returns the raw text response.
func (c *Client) GenerateText(ctx context.Context, prompt string) (string, error) {
	return c.generateTextWithRetry(ctx, prompt)
}

// GenerateJSONWithConfig allows overriding generation parameters.
func (c *Client) GenerateJSONWithConfig(ctx context.Context, prompt string, target interface{}, genCfg GenerationConfig) error {
	return c.generateWithRetry(ctx, prompt, target, genCfg)
}

// generateWithRetry handles retry logic for JSON generation.
func (c *Client) generateWithRetry(ctx context.Context, prompt string, target interface{}, genCfg GenerationConfig) error {
	err := c.executeWithCircuitBreaker(ctx, func(ctx context.Context) error {
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limit exceeded: %w", err)
		}

		var lastErr error
		bo := backoff.NewExponentialBackOff()
		bo.InitialInterval = 200 * time.Millisecond
		bo.MaxInterval = 5 * time.Second
		bo.MaxElapsedTime = c.config.RequestTimeout
		bo.Multiplier = 2.0

		err := backoff.Retry(func() error {
			select {
			case <-ctx.Done():
				return backoff.Permanent(ctx.Err())
			default:
			}

			if err := c.rateLimiter.Wait(ctx); err != nil {
				return backoff.Permanent(fmt.Errorf("rate limit exceeded: %w", err))
			}

			timeoutCtx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
			defer cancel()

			respText, err := c.doPredict(timeoutCtx, prompt, genCfg)
			if err != nil {
				lastErr = err
				if isRetryableError(err) {
					return err
				}
				return backoff.Permanent(err)
			}

			text := stripMarkdownJSON(respText)
			if err := json.Unmarshal([]byte(text), target); err != nil {
				lastErr = err
				return backoff.Permanent(fmt.Errorf("parse JSON response: %w", err))
			}

			return nil
		}, backoff.WithContext(bo, ctx))

		if err != nil {
			return fmt.Errorf("gemini request failed after %d attempts: %w", c.config.MaxRetries, lastErr)
		}

		return nil
	})

	return err
}

// generateTextWithRetry handles retry logic for text generation.
func (c *Client) generateTextWithRetry(ctx context.Context, prompt string) (string, error) {
	var result string

	err := c.executeWithCircuitBreaker(ctx, func(ctx context.Context) error {
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limit exceeded: %w", err)
		}

		var lastErr error
		bo := backoff.NewExponentialBackOff()
		bo.InitialInterval = 200 * time.Millisecond
		bo.MaxInterval = 5 * time.Second
		bo.MaxElapsedTime = c.config.RequestTimeout
		bo.Multiplier = 2.0

		err := backoff.Retry(func() error {
			select {
			case <-ctx.Done():
				return backoff.Permanent(ctx.Err())
			default:
			}

			timeoutCtx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
			defer cancel()

			respText, err := c.doPredict(timeoutCtx, prompt, GenerationConfig{})
			if err != nil {
				lastErr = err
				if isRetryableError(err) {
					return err
				}
				return backoff.Permanent(err)
			}

			result = stripMarkdownJSON(respText)
			return nil
		}, backoff.WithContext(bo, ctx))

		if err != nil {
			return fmt.Errorf("gemini request failed: %w", lastErr)
		}

		return nil
	})

	return result, err
}

// doPredict calls the Vertex AI aiplatform Predict endpoint for Gemini.
func (c *Client) doPredict(ctx context.Context, prompt string, genCfg GenerationConfig) (string, error) {
	endpoint := fmt.Sprintf("projects/%s/locations/%s/publishers/google/models/%s",
		c.projectID, c.location, c.modelName)

	// Build parameters for generation config
	temperature := c.config.Temperature
	maxTokens := int32(c.config.MaxTokens)
	topP := c.config.TopP
	topK := int32(c.config.TopK)

	if genCfg.Temperature != 0 {
		temperature = genCfg.Temperature
	}
	if genCfg.MaxTokens > 0 {
		maxTokens = int32(genCfg.MaxTokens)
	}
	if genCfg.TopP != 0 {
		topP = genCfg.TopP
	}
	if genCfg.TopK > 0 {
		topK = int32(genCfg.TopK)
	}

	params, _ := structpb.NewStruct(map[string]interface{}{
		"temperature":      temperature,
		"maxOutputTokens":  maxTokens,
		"topP":             topP,
		"topK":             topK,
	})

	instance, _ := structpb.NewStruct(map[string]interface{}{
		"prompt": prompt,
	})

	req := &aiplatformpb.PredictRequest{
		Endpoint:   endpoint,
		Instances:  []*structpb.Value{structpb.NewStructValue(instance)},
		Parameters: structpb.NewStructValue(params),
	}

	resp, err := c.predictor.Predict(ctx, req)
	if err != nil {
		return "", fmt.Errorf("aiplatform predict: %w", err)
	}

	if resp == nil || len(resp.Predictions) == 0 {
		return "", fmt.Errorf("aiplatform returned empty response")
	}

	// Extract text from prediction response
	pred := resp.Predictions[0]
	pbStruct := pred.GetStructValue()
	if pbStruct == nil {
		return "", fmt.Errorf("aiplatform returned non-struct prediction")
	}

	contentField := pbStruct.GetFields()["content"]
	if contentField == nil {
		// Try "text" field as fallback
		contentField = pbStruct.GetFields()["text"]
	}
	if contentField == nil {
		return "", fmt.Errorf("aiplatform response missing 'content' or 'text' field")
	}

	return contentField.GetStringValue(), nil
}

// Close releases the aiplatform prediction client.
func (c *Client) Close() error {
	if c.predictor != nil {
		return c.predictor.Close()
	}
	return nil
}

// executeWithCircuitBreaker wraps execution with circuit breaker pattern.
func (c *Client) executeWithCircuitBreaker(ctx context.Context, fn func(context.Context) error) error {
	if c.circuitBreaker.IsOpen() {
		return fmt.Errorf("circuit breaker is open: LLM service temporarily unavailable")
	}

	err := fn(ctx)

	if err != nil {
		c.circuitBreaker.RecordFailure()
	} else {
		c.circuitBreaker.RecordSuccess()
	}

	return err
}

// isRetryableError determines if an error should be retried.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	retryableCodes := []string{
		"RESOURCE_EXHAUSTED", "UNAVAILABLE", "DEADLINE_EXCEEDED", "INTERNAL",
		"429", "500", "503", "transient", "rate limit",
	}

	for _, code := range retryableCodes {
		if strings.Contains(errStr, code) {
			return true
		}
	}
	return false
}

// stripMarkdownJSON removes markdown code fence markers from JSON text.
func stripMarkdownJSON(text string) string {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```json") {
		text = strings.TrimPrefix(text, "```json")
	} else if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```")
	}
	text = strings.TrimSuffix(text, "```")
	return strings.TrimSpace(text)
}

// truncate limits a string to n characters.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// GenerationConfig holds generation parameters.
type GenerationConfig struct {
	Temperature  float32
	MaxTokens    int
	TopP         float32
	TopK         int
	ResponseMIME string
}

// RateLimiter wraps golang.org/x/time/rate for leak-free rate limiting.
type RateLimiter struct {
	limiter *rate.Limiter
}

// NewRateLimiter creates a rate limiter with no goroutine leaks.
func NewRateLimiter(ratePerSec int, burst int) *RateLimiter {
	limiter := rate.NewLimiter(rate.Every(time.Second/time.Duration(ratePerSec)), burst)
	return &RateLimiter{limiter: limiter}
}

// Wait blocks until a token is available or context is cancelled.
func (rl *RateLimiter) Wait(ctx context.Context) error {
	return rl.limiter.Wait(ctx)
}

// CircuitBreaker implements the circuit breaker pattern.
type CircuitBreaker struct {
	mu               sync.Mutex
	failures         int
	successes        int
	state            string
	failureThreshold int
	resetTimeout     time.Duration
	lastFailureAt    time.Time
}

// NewCircuitBreaker creates a circuit breaker.
func NewCircuitBreaker(failureThreshold int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            "closed",
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
	}
}

// IsOpen returns true if circuit breaker is open (tripped).
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == "open" {
		if time.Since(cb.lastFailureAt) > cb.resetTimeout {
			cb.state = "half-open"
			return false
		}
		return true
	}
	return false
}

// RecordSuccess records a successful request.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.successes++
	cb.failures = 0

	if cb.state == "half-open" {
		cb.state = "closed"
		cb.successes = 0
	}
}

// RecordFailure records a failed request.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailureAt = time.Now()

	if cb.failures >= cb.failureThreshold {
		cb.state = "open"
	}
}

// GetState returns the current state.
func (cb *CircuitBreaker) GetState() string {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
