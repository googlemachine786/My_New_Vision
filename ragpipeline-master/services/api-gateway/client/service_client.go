package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/visionary/ragpipeline/pkg/circuitbreaker"
)

const (
	RequestIDHeader = "X-Request-ID"
)

// ServiceCaller defines the interface for calling downstream services
type ServiceCaller interface {
	Do(ctx context.Context, req Request) (*http.Response, error)
	Get(ctx context.Context, path string) (*http.Response, error)
	PostJSON(ctx context.Context, path string, body interface{}, result interface{}) error
	PostJSONRaw(ctx context.Context, path string, body interface{}) ([]byte, error)
	Stream(ctx context.Context, path string, body interface{}) (*http.Response, error)
	Name() string
	Stats() map[string]interface{}
}

// Compile-time interface satisfaction check
var _ ServiceCaller = (*ServiceClient)(nil)

// ServiceClient is an HTTP client for calling downstream services
type ServiceClient struct {
	httpClient *http.Client
	baseURL    string
	name       string

	CircuitBreaker *circuitbreaker.CircuitBreaker
	metrics        *clientMetrics

	DefaultTimeout time.Duration
	Tracer         trace.Tracer

	// Retry configuration
	MaxRetries int
	RetryCodes map[int]bool // HTTP status codes that should trigger retry
}

// clientMetrics tracks service client metrics
type clientMetrics struct {
	mu           sync.Mutex
	RequestCount int64
	ErrorCount   int64
	TotalLatency time.Duration
}

func (m *clientMetrics) RecordRequest(latency time.Duration, isError bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RequestCount++
	if isError {
		m.ErrorCount++
	}
	m.TotalLatency += latency
}

func (m *clientMetrics) Stats() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	avgLatency := time.Duration(0)
	if m.RequestCount > 0 {
		avgLatency = m.TotalLatency / time.Duration(m.RequestCount)
	}

	return map[string]interface{}{
		"request_count":   m.RequestCount,
		"error_count":     m.ErrorCount,
		"average_latency": avgLatency.String(),
	}
}

// NewServiceClient creates a new service client
func NewServiceClient(name, baseURL string, timeout time.Duration) *ServiceClient {
	return &ServiceClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL:        baseURL,
		name:           name,
		CircuitBreaker: circuitbreaker.New(),
		metrics:        &clientMetrics{},
		DefaultTimeout: timeout,
		Tracer:         otel.Tracer("api-gateway"),
		MaxRetries:     2,
		RetryCodes: map[int]bool{
			http.StatusTooManyRequests:      true,
			http.StatusInternalServerError:  true,
			http.StatusBadGateway:           true,
			http.StatusServiceUnavailable:   true,
			http.StatusGatewayTimeout:       true,
		},
	}
}

// Request represents an outgoing request
type Request struct {
	Method string
	Path   string
	Body   interface{}
	Header http.Header
}

// Do executes an HTTP request with circuit breaking, retries, and tracing
func (sc *ServiceClient) Do(ctx context.Context, req Request) (*http.Response, error) {
	// Check circuit breaker
	if !sc.CircuitBreaker.AllowRequest() {
		return nil, fmt.Errorf("circuit breaker open for %s", sc.name)
	}

	start := time.Now()
	defer func() {
		sc.metrics.RecordRequest(time.Since(start), false)
	}()

	url := sc.baseURL + req.Path

	// FIX P0: Marshal body once, then create fresh reader for each retry
	var jsonBody []byte
	if req.Body != nil {
		var err error
		jsonBody, err = json.Marshal(req.Body)
		if err != nil {
			sc.CircuitBreaker.RecordFailure()
			sc.metrics.RecordRequest(time.Since(start), true)
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	// FIX P0: Build request inside retry operation so bodyReader is fresh each attempt
	buildRequest := func() (*http.Request, error) {
		var bodyReader io.Reader
		if jsonBody != nil {
			bodyReader = bytes.NewReader(jsonBody) // Fresh reader each attempt
		}
		httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, bodyReader)
		if err != nil {
			return nil, err
		}
		if req.Header != nil {
			httpReq.Header = req.Header.Clone()
		}
		httpReq.Header.Set("Content-Type", "application/json")
		if httpReq.Header.Get(RequestIDHeader) == "" {
			if requestID, ok := ctx.Value(RequestIDHeader).(string); ok && requestID != "" {
				httpReq.Header.Set(RequestIDHeader, requestID)
			}
		}
		return httpReq, nil
	}

	// Create span for tracing
	ctx, span := sc.Tracer.Start(ctx, fmt.Sprintf("%s %s", sc.name, req.Path))
	defer span.End()

	// Execute with retries
	var resp *http.Response

	operation := func() error {
		httpReq, err := buildRequest()
		if err != nil {
			return backoff.Permanent(err)
		}
		resp, err = sc.httpClient.Do(httpReq)
		if err != nil {
			return backoff.Permanent(err)
		}

		// Check if response indicates a retryable error
		if sc.RetryCodes[resp.StatusCode] {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return fmt.Errorf("retryable status %d: %s", resp.StatusCode, string(body))
		}

		return nil
	}

	b := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), uint64(sc.MaxRetries))
	b = backoff.WithContext(b, ctx)

	if err := backoff.Retry(operation, b); err != nil {
		sc.CircuitBreaker.RecordFailure()
		sc.metrics.RecordRequest(time.Since(start), true)
		span.RecordError(err)
		return nil, fmt.Errorf("request to %s failed: %w", sc.name, err)
	}

	sc.CircuitBreaker.RecordSuccess()
	return resp, nil
}

// Get performs a GET request
func (sc *ServiceClient) Get(ctx context.Context, path string) (*http.Response, error) {
	return sc.Do(ctx, Request{
		Method: http.MethodGet,
		Path:   path,
	})
}

// PostJSON performs a POST request with JSON body and returns decoded JSON response
func (sc *ServiceClient) PostJSON(ctx context.Context, path string, body interface{}, result interface{}) error {
	resp, err := sc.Do(ctx, Request{
		Method: http.MethodPost,
		Path:   path,
		Body:   body,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d from %s: %s", resp.StatusCode, sc.name, string(body))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response from %s: %w", sc.name, err)
		}
	}

	return nil
}

// PostJSONRaw performs a POST request and returns the raw response body
func (sc *ServiceClient) PostJSONRaw(ctx context.Context, path string, body interface{}) ([]byte, error) {
	resp, err := sc.Do(ctx, Request{
		Method: http.MethodPost,
		Path:   path,
		Body:   body,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %s: %w", sc.name, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status %d from %s: %s", resp.StatusCode, sc.name, string(respBody))
	}

	return respBody, nil
}

// Stream performs a POST request and returns the response for streaming
func (sc *ServiceClient) Stream(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	// Check circuit breaker
	if !sc.CircuitBreaker.AllowRequest() {
		return nil, fmt.Errorf("circuit breaker open for %s", sc.name)
	}

	start := time.Now()
	defer func() {
		sc.metrics.RecordRequest(time.Since(start), false)
	}()

	url := sc.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			sc.CircuitBreaker.RecordFailure()
			sc.metrics.RecordRequest(time.Since(start), true)
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bodyReader)
	if err != nil {
		sc.CircuitBreaker.RecordFailure()
		sc.metrics.RecordRequest(time.Since(start), true)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	// Propagate request ID
	if requestID, ok := ctx.Value(RequestIDHeader).(string); ok && requestID != "" {
		httpReq.Header.Set(RequestIDHeader, requestID)
	}

	// Create span for tracing
	ctx, span := sc.Tracer.Start(ctx, fmt.Sprintf("%s %s (stream)", sc.name, path))
	defer span.End()

	// No retries for streaming requests
	resp, err := sc.httpClient.Do(httpReq)
	if err != nil {
		sc.CircuitBreaker.RecordFailure()
		sc.metrics.RecordRequest(time.Since(start), true)
		span.RecordError(err)
		return nil, fmt.Errorf("stream request to %s failed: %w", sc.name, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		sc.CircuitBreaker.RecordFailure()
		sc.metrics.RecordRequest(time.Since(start), true)
		return nil, fmt.Errorf("unexpected status %d from %s: %s", resp.StatusCode, sc.name, string(body))
	}

	sc.CircuitBreaker.RecordSuccess()
	return resp, nil
}

// Name returns the service client name
func (sc *ServiceClient) Name() string {
	return sc.name
}

// Stats returns client statistics
func (sc *ServiceClient) Stats() map[string]interface{} {
	return map[string]interface{}{
		"name":            sc.name,
		"base_url":        sc.baseURL,
		"metrics":         sc.metrics.Stats(),
		"circuit_breaker": sc.CircuitBreaker.Stats(),
	}
}
