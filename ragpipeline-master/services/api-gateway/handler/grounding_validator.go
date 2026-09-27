// Package handler provides NLI-based response grounding validation.
//
// Instead of hand-rolled keyword overlap similarity, this delegates to the
// NLI microservice which uses CrossEncoder models for proper natural language
// inference. spaCy handles sentence boundary detection correctly (abbreviations,
// decimals, etc.) and the NLI model produces entailment scores for each claim.
package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

const (
	defaultNLIServiceURL = "http://localhost:8084"
	defaultNLITimeout    = 10 * time.Second
	defaultBatchSize     = 50  // max claims per batch request
	groundingWarnPercent = 0.3 // warn if >30% claims ungrounded
)

// NLIClient wraps the HTTP client to the NLI microservice.
type NLIClient struct {
	baseURL string
	client  *http.Client
}

// NewNLIClient creates a client for the NLI microservice.
func NewNLIClient(baseURL string) *NLIClient {
	if baseURL == "" {
		baseURL = defaultNLIServiceURL
	}
	return &NLIClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: defaultNLITimeout},
	}
}

// SplitSentencesRequest is the request for /split-sentences.
type SplitSentencesRequest struct {
	Text string `json:"text"`
}

// SplitSentencesResponse is the response from /split-sentences.
type SplitSentencesResponse struct {
	Sentences []string `json:"sentences"`
}

// BatchValidateRequest is the request for /batch-validate.
type BatchValidateRequest struct {
	Claims        []string `json:"claims"`
	ContextChunks []string `json:"context_chunks"`
}

// ClaimResult is a single claim validation result.
type ClaimResult struct {
	Claim         string  `json:"claim"`
	IsGrounded    bool    `json:"is_grounded"`
	GroundingScore float64 `json:"grounding_score"`
	BestChunkIdx  int     `json:"best_chunk_idx"`
	Label         string  `json:"label"`
}

// BatchValidateResponse is the response from /batch-validate.
type BatchValidateResponse struct {
	Results             []ClaimResult `json:"results"`
	OverallGroundingScore float64     `json:"overall_grounding_score"`
	ProcessingTimeMS    float64       `json:"processing_time_ms"`
}

// SplitSentences calls the NLI service to properly split text into sentences
// using spaCy's sentence boundary detection (handles abbreviations, decimals,
// ellipsis, etc. correctly).
func (c *NLIClient) SplitSentences(ctx context.Context, text string) ([]string, error) {
	reqBody, err := json.Marshal(SplitSentencesRequest{Text: text})
	if err != nil {
		return nil, fmt.Errorf("marshal split-sentences request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/split-sentences", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create split-sentences request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call split-sentences: %w", err)
	}
	// FIX P0: Defer close immediately, not inside a loop
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("split-sentences returned %d: %s", resp.StatusCode, string(body))
	}

	var result SplitSentencesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode split-sentences response: %w", err)
	}

	return result.Sentences, nil
}

// BatchValidateClaims validates multiple claims against context chunks using
// NLI CrossEncoder model. Returns grounding score for each claim.
func (c *NLIClient) BatchValidateClaims(ctx context.Context, claims, contextChunks []string) (*BatchValidateResponse, error) {
	if len(claims) == 0 {
		return &BatchValidateResponse{}, nil
	}

	// Process in batches to avoid overwhelming the service
	var allResults []ClaimResult
	overallScore := 0.0
	totalWeight := 0.0

	for start := 0; start < len(claims); start += defaultBatchSize {
		end := start + defaultBatchSize
		if end > len(claims) {
			end = len(claims)
		}
		batch := claims[start:end]

		reqBody, err := json.Marshal(BatchValidateRequest{
			Claims:        batch,
			ContextChunks: contextChunks,
		})
		if err != nil {
			return nil, fmt.Errorf("marshal batch-validate request: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/batch-validate", bytes.NewReader(reqBody))
		if err != nil {
			return nil, fmt.Errorf("create batch-validate request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			log.Warn().Err(err).Msg("NLI service unavailable, skipping grounding validation")
			// FIX P1: Return permissive results when NLI service is down (fail-open, not fail-closed)
			for _, claim := range batch {
				allResults = append(allResults, ClaimResult{
					Claim:         claim,
					IsGrounded:    true,
					GroundingScore: 1.0,
					Label:         "unknown",
				})
			}
			overallScore = 1.0
			break
		}
		// FIX P0: Close body IMMEDIATELY, not deferred in a loop (FD leak)
		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("read batch-validate response: %w", readErr)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("batch-validate returned %d: %s", resp.StatusCode, string(body))
		}

		var result BatchValidateResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("decode batch-validate response: %w", err)
		}

		allResults = append(allResults, result.Results...)
		weight := float64(len(result.Results))
		overallScore += result.OverallGroundingScore * weight
		totalWeight += weight
	}

	if totalWeight > 0 {
		overallScore /= totalWeight
	}

	return &BatchValidateResponse{
		Results:             allResults,
		OverallGroundingScore: overallScore,
		ProcessingTimeMS:    0,
	}, nil
}

// GroundingResult is the result of grounding validation for a response.
type GroundingResult struct {
	GroundingScore    float64  `json:"grounding_score"`    // 0.0–1.0
	TotalClaims       int      `json:"total_claims"`
	GroundedClaims    int      `json:"grounded_claims"`
	UngroundedClaims  int      `json:"ungrounded_claims"`
	UngroundedDetails []string `json:"ungrounded_details,omitempty"`
	Warning           string   `json:"grounding_warning,omitempty"`
	LatencyMS         int64    `json:"latency_ms"`
}

// groundingHistogram is a package-level Prometheus histogram for grounding scores.
var (
	groundingHist    prometheus.Histogram
	groundingHistOnce sync.Once
)

// GetGroundingHistogram returns or creates the grounding score histogram.
// Uses sync.Once to prevent panic from duplicate MustRegister calls.
func GetGroundingHistogram(reg prometheus.Registerer) prometheus.Histogram {
	groundingHistOnce.Do(func() {
		groundingHist = prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "rag_grounding_score",
			Help:    "Grounding score distribution for LLM responses (0.0 = ungrounded, 1.0 = fully grounded)",
			Buckets: prometheus.LinearBuckets(0.0, 0.1, 11),
		})

		// FIX P0: Use Register instead of MustRegister to prevent panic on duplicate
		if reg != nil {
			// Attempt to register; ignore "already registered" errors
			if err := reg.Register(groundingHist); err != nil {
				// Check if it's an AlreadyRegisteredError — that's fine
				if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
					log.Warn().Err(err).Msg("Failed to register grounding histogram")
				}
			}
		}
	})

	return groundingHist
}

// NewGroundingValidator creates a grounding validator with configurable thresholds.
// groundingThreshold: minimum score for a claim to be considered grounded (default 0.6)
// warnPercent: maximum percentage of ungrounded claims before warning (default 0.3)
func NewGroundingValidator(groundingThreshold, warnPercent float64) *GroundingValidator {
	return &GroundingValidator{
		groundingThreshold: groundingThreshold,
		warnPercent:        warnPercent,
	}
}

// GroundingValidator validates response grounding.
type GroundingValidator struct {
	groundingThreshold float64
	warnPercent        float64
}

// Validate runs grounding validation using the NLIClient.
func (v *GroundingValidator) Validate(ctx context.Context, responseText string, contextChunks []string, nliClient *NLIClient, metricsReg prometheus.Registerer) (*GroundingResult, error) {
	start := time.Now()

	sentences, err := nliClient.SplitSentences(ctx, responseText)
	if err != nil {
		return nil, fmt.Errorf("split sentences: %w", err)
	}

	var claims []string
	for _, sent := range sentences {
		if len(sent) >= 15 {
			claims = append(claims, sent)
		}
	}

	if len(claims) == 0 {
		return &GroundingResult{GroundingScore: 1.0}, nil
	}

	validation, err := nliClient.BatchValidateClaims(ctx, claims, contextChunks)
	if err != nil {
		return nil, fmt.Errorf("batch validate claims: %w", err)
	}

	groundedCount := 0
	var ungroundedDetails []string
	for _, r := range validation.Results {
		if r.IsGrounded {
			groundedCount++
		} else {
			ungroundedDetails = append(ungroundedDetails,
				fmt.Sprintf("%s (score: %.2f, label: %s)", r.Claim, r.GroundingScore, r.Label))
		}
	}

	total := len(validation.Results)
	groundingScore := validation.OverallGroundingScore
	ungroundedCount := total - groundedCount

	var warning string
	if total > 0 && float64(ungroundedCount)/float64(total) > v.warnPercent {
		warning = fmt.Sprintf("This response may contain unverified information. %d of %d claims could not be grounded.", ungroundedCount, total)
	}

	// Record metrics safely
	if metricsReg != nil {
		GetGroundingHistogram(metricsReg).Observe(groundingScore)
	}

	log.Info().
		Float64("grounding_score", groundingScore).
		Int("total_claims", total).
		Int("grounded", groundedCount).
		Int("ungrounded", ungroundedCount).
		Dur("latency_ms", time.Since(start)).
		Msg("Grounding validation complete")

	return &GroundingResult{
		GroundingScore:    groundingScore,
		TotalClaims:       total,
		GroundedClaims:    groundedCount,
		UngroundedClaims:  ungroundedCount,
		UngroundedDetails: ungroundedDetails,
		Warning:           warning,
		LatencyMS:         time.Since(start).Milliseconds(),
	}, nil
}
