// Package vertex provides the Vertex AI embedding client with retry and observability.
package vertex

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/cenkalti/backoff/v4"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"
)

// TaskType represents the type of embedding task.
type TaskType string

const (
	// TaskTypeDocument is for document embeddings (indexing).
	TaskTypeDocument TaskType = "RETRIEVAL_DOCUMENT"
	// TaskTypeQuery is for query embeddings (search).
	TaskTypeQuery TaskType = "RETRIEVAL_QUERY"
)

// EmbedResponse represents a single embedding response.
type EmbedResponse struct {
	Embedding  []float64
	Dimensions int
}

// Client wraps the Vertex AI embedding client with retry and observability.
type Client struct {
	predictor *aiplatform.PredictionClient
	projectID string
	location  string
	model     string
	dimension int
	timeout   time.Duration
	tracer    trace.Tracer

	// Statistics
	requestCount  atomic.Int64
	errorCount    atomic.Int64
	totalLatency  atomic.Int64 // milliseconds
	lastRequestAt atomic.Int64 // unix timestamp
}

// NewClient creates a new Vertex AI embedding client.
func NewClient(ctx context.Context, projectID, location, model string, dimension int, timeout time.Duration) (*Client, error) {
	if projectID == "" {
		return nil, fmt.Errorf("projectID is required")
	}
	if location == "" {
		return nil, fmt.Errorf("location is required")
	}

	predictor, err := aiplatform.NewPredictionClient(ctx, option.WithEndpoint(fmt.Sprintf("%s-aiplatform.googleapis.com:443", location)))
	if err != nil {
		return nil, fmt.Errorf("failed to create Vertex AI prediction client: %w", err)
	}

	tracer := otel.Tracer("embedding-service.vertex")

	log.Info().
		Str("project", projectID).
		Str("location", location).
		Str("model", model).
		Int("dimension", dimension).
		Msg("Vertex AI client initialized")

	return &Client{
		predictor: predictor,
		projectID: projectID,
		location:  location,
		model:     model,
		dimension: dimension,
		timeout:   timeout,
		tracer:    tracer,
	}, nil
}

// Close releases resources held by the client.
func (c *Client) Close() error {
	return c.predictor.Close()
}

// EmbedText generates an embedding for a single text.
func (c *Client) EmbedText(ctx context.Context, text string, taskType TaskType) (*EmbedResponse, error) {
	ctx, span := c.tracer.Start(ctx, "vertex.embed_text",
		trace.WithAttributes(
			attribute.String("task_type", string(taskType)),
			attribute.Int("text_length", len(text)),
			attribute.String("model", c.model),
		),
	)
	defer span.End()

	startTime := time.Now()

	// Use retry logic for transient failures
	result, err := c.embedWithRetry(ctx, text, taskType)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		c.recordError(time.Since(startTime))
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	span.SetAttributes(attribute.Int("embedding_dimension", result.Dimensions))
	c.recordSuccess(time.Since(startTime))

	return result, nil
}

// EmbedBatch generates embeddings for multiple texts.
func (c *Client) EmbedBatch(ctx context.Context, texts []string, taskType TaskType) ([]*EmbedResponse, error) {
	ctx, span := c.tracer.Start(ctx, "vertex.embed_batch",
		trace.WithAttributes(
			attribute.String("task_type", string(taskType)),
			attribute.Int("batch_size", len(texts)),
			attribute.String("model", c.model),
		),
	)
	defer span.End()

	startTime := time.Now()

	results := make([]*EmbedResponse, len(texts))
	var firstErr error

	for i, text := range texts {
		emb, err := c.EmbedText(ctx, text, taskType)
		if err != nil && firstErr == nil {
			firstErr = fmt.Errorf("text %d: %w", i, err)
		}
		results[i] = emb
	}

	if firstErr != nil {
		span.SetStatus(codes.Error, firstErr.Error())
		c.recordError(time.Since(startTime))
		return nil, firstErr
	}

	span.SetStatus(codes.Ok, "")
	span.SetAttributes(attribute.Int("results_count", len(results)))
	c.recordSuccess(time.Since(startTime))

	return results, nil
}

// embedWithRetry executes embedding with exponential backoff retry.
func (c *Client) embedWithRetry(ctx context.Context, text string, taskType TaskType) (*EmbedResponse, error) {
	var result *EmbedResponse

	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 100 * time.Millisecond
	bo.MaxInterval = 5 * time.Second
	bo.MaxElapsedTime = c.timeout
	bo.Multiplier = 2.0

	attempt := 0
	err := backoff.Retry(func() error {
		attempt++
		select {
		case <-ctx.Done():
			return backoff.Permanent(ctx.Err())
		default:
		}

		resp, err := c.doEmbed(ctx, text, taskType)
		if err != nil {
			if isRetryable(err) {
				log.Warn().
					Err(err).
					Int("attempt", attempt).
					Str("task_type", string(taskType)).
					Msg("Vertex AI embedding request failed, retrying")
				return err
			}
			return backoff.Permanent(err)
		}

		result = resp
		return nil
	}, backoff.WithContext(bo, ctx))

	if err != nil {
		return nil, fmt.Errorf("embedding failed after %d attempts: %w", attempt, err)
	}

	return result, nil
}

// doEmbed performs the actual embedding API call using the stable aiplatform SDK.
func (c *Client) doEmbed(ctx context.Context, text string, taskType TaskType) (*EmbedResponse, error) {
	// Build the endpoint name: projects/{project}/locations/{location}/publishers/google/models/{model}
	endpoint := fmt.Sprintf("projects/%s/locations/%s/publishers/google/models/%s",
		c.projectID, c.location, c.model)

	// Build instance using structpb.Value
	instance, err := structpb.NewStruct(map[string]interface{}{
		"content":   text,
		"task_type": string(taskType),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build instance: %w", err)
	}

	instanceVal := structpb.NewStructValue(instance)

	// Build parameters
	params, _ := structpb.NewStruct(map[string]interface{}{
		"autoTruncate": true,
	})
	paramsVal := structpb.NewStructValue(params)

	// Create predict request
	req := &aiplatformpb.PredictRequest{
		Endpoint:  endpoint,
		Instances: []*structpb.Value{instanceVal},
		Parameters: paramsVal,
	}

	// Call predict
	resp, err := c.predictor.Predict(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("vertex predict: %w", err)
	}

	// Extract embedding from response
	if len(resp.Predictions) == 0 {
		return nil, fmt.Errorf("vertex returned empty predictions")
	}

	// The embedding values are in the first prediction's "embeddings.values" field
	prediction := resp.Predictions[0]
	embedding, err := extractEmbeddingValues(prediction)
	if err != nil {
		return nil, fmt.Errorf("failed to extract embedding: %w", err)
	}

	return &EmbedResponse{
		Embedding:  embedding,
		Dimensions: len(embedding),
	}, nil
}

// extractEmbeddingValues extracts float64 values from a prediction structpb.Value.
func extractEmbeddingValues(prediction *structpb.Value) ([]float64, error) {
	// Navigate the response structure: predictions[].embeddings.values
	structVal := prediction.GetStructValue()
	if structVal == nil {
		return nil, fmt.Errorf("prediction is not a struct")
	}

	embeddingsField, ok := structVal.Fields["embeddings"]
	if !ok {
		return nil, fmt.Errorf("no 'embeddings' field in prediction")
	}

	embeddingsStruct := embeddingsField.GetStructValue()
	if embeddingsStruct == nil {
		return nil, fmt.Errorf("'embeddings' is not a struct")
	}

	valuesField, ok := embeddingsStruct.Fields["values"]
	if !ok {
		return nil, fmt.Errorf("no 'values' field in embeddings")
	}

	valuesList := valuesField.GetListValue()
	if valuesList == nil {
		return nil, fmt.Errorf("'values' is not a list")
	}

	result := make([]float64, len(valuesList.Values))
	for i, v := range valuesList.Values {
		result[i] = v.GetNumberValue()
	}

	return result, nil
}

// HealthCheck verifies Vertex AI connectivity.
func (c *Client) HealthCheck(ctx context.Context) error {
	ctx, span := c.tracer.Start(ctx, "vertex.health_check")
	defer span.End()

	// Attempt a minimal embedding to verify connectivity
	_, err := c.EmbedText(ctx, "health check", TaskTypeQuery)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("health check failed: %w", err)
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// Stats returns client statistics.
func (c *Client) Stats() map[string]interface{} {
	reqCount := c.requestCount.Load()
	errCount := c.errorCount.Load()
	totalLat := c.totalLatency.Load()

	avgLatency := time.Duration(0)
	if reqCount > 0 {
		avgLatency = time.Duration(totalLat/reqCount) * time.Millisecond
	}

	return map[string]interface{}{
		"project":          c.projectID,
		"location":         c.location,
		"model":            c.model,
		"request_count":    reqCount,
		"error_count":      errCount,
		"average_latency":  avgLatency.String(),
		"last_request_at":  time.Unix(c.lastRequestAt.Load(), 0).Format(time.RFC3339),
	}
}

// recordSuccess records a successful embedding request.
func (c *Client) recordSuccess(latency time.Duration) {
	c.requestCount.Add(1)
	c.totalLatency.Add(latency.Milliseconds())
	c.lastRequestAt.Store(time.Now().Unix())
}

// recordError records a failed embedding request.
func (c *Client) recordError(latency time.Duration) {
	c.requestCount.Add(1)
	c.errorCount.Add(1)
	c.totalLatency.Add(latency.Milliseconds())
	c.lastRequestAt.Store(time.Now().Unix())
}

// isRetryable checks if an error is transient and should be retried.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unavailable") ||
		strings.Contains(msg, "deadline exceeded") ||
		strings.Contains(msg, "resource exhausted") ||
		strings.Contains(msg, "internal") ||
		strings.Contains(msg, "503") ||
		strings.Contains(msg, "429")
}
