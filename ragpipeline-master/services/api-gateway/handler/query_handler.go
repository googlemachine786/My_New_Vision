package handler

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/visionary/ragpipeline/services/api-gateway/cache"
	"github.com/visionary/ragpipeline/services/api-gateway/client"
	sessionctx "github.com/visionary/ragpipeline/services/api-gateway/context"
	"github.com/visionary/ragpipeline/services/api-gateway/middleware"
	"github.com/visionary/ragpipeline/pkg/types"
	"github.com/visionary/ragpipeline/services/api-gateway/session"

	"github.com/rs/zerolog/log"
)

// QueryHandlerDeps holds the dependencies for QueryHandler.
type QueryHandlerDeps struct {
	EmbeddingClient          *client.ServiceClient
	VectorSearchClient       *client.ServiceClient
	QueryUnderstandingClient *client.ServiceClient
	LLMClient                *client.ServiceClient
	Cache                    *cache.ResponseCache     // optional
	SessionManager           *session.SessionManager  // optional
	CAGOrchestrator          *cache.CAGOrchestrator   // optional
	FallbackHandler          *FallbackHandler         // Story C: optional fallback handler
	ContextRegistry          *sessionctx.SessionRegistry // Story G/H: session context registry
	NLIClient                *NLIClient               // Story J: NLI client for grounding
}

// QueryHandler handles the main query endpoint
type QueryHandler struct {
	EmbeddingClient          *client.ServiceClient
	VectorSearchClient       *client.ServiceClient
	QueryUnderstandingClient *client.ServiceClient
	LLMClient                *client.ServiceClient
	Cache                    *cache.ResponseCache
	SessionManager           *session.SessionManager
	CAGOrchestrator          *cache.CAGOrchestrator
	FallbackHandler          *FallbackHandler // Story C: fallback handler
	Enricher                 *QueryEnricher   // Story H: query enricher
	NLIClient                *NLIClient       // Story J: NLI client for grounding
}

// NewQueryHandler creates a new query handler with the given dependencies.
func NewQueryHandler(deps QueryHandlerDeps) *QueryHandler {
	enricher := (*QueryEnricher)(nil)
	if deps.ContextRegistry != nil {
		enricher = NewQueryEnricher(deps.ContextRegistry)
	}

	return &QueryHandler{
		EmbeddingClient:          deps.EmbeddingClient,
		VectorSearchClient:       deps.VectorSearchClient,
		QueryUnderstandingClient: deps.QueryUnderstandingClient,
		LLMClient:                deps.LLMClient,
		Cache:                    deps.Cache,
		SessionManager:           deps.SessionManager,
		CAGOrchestrator:          deps.CAGOrchestrator,
		FallbackHandler:          deps.FallbackHandler,
		Enricher:                 enricher,
		NLIClient:                deps.NLIClient,
	}
}

// ServeHTTP handles POST /query
func (h *QueryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "method_not_allowed", Message: "Only POST is allowed"},
		})
		return
	}

	// Parse request
	var req types.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "invalid_request", Message: "Invalid JSON body"},
		})
		return
	}

	if err := req.Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "validation_error", Message: err.Error()},
		})
		return
	}

	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	if req.Stream {
		h.handleStreaming(w, r, &req, userID, requestID)
	} else {
		h.handleNonStreaming(w, r, &req, userID, requestID)
	}
}

func (h *QueryHandler) handleStreaming(w http.ResponseWriter, r *http.Request, req *types.QueryRequest, userID, requestID string) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeSSEError(w, flusher, "streaming not supported")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	// Send start event
	sessionID := req.SessionID
	if sessionID == "" {
		// Create new session
		if h.SessionManager != nil {
			newSession, err := h.SessionManager.CreateSession(ctx, userID)
			if err != nil {
				writeSSEError(w, flusher, "failed to create session")
				return
			}
			sessionID = newSession.ID
		}
	}

	startEvent := types.SSEStartEvent{
		SessionID: sessionID,
		RequestID: requestID,
	}
	writeSSEEvent(w, flusher, "start", startEvent)

	// Get session history
	var history []session.ConversationTurn
	if h.SessionManager != nil && sessionID != "" {
		if turns, err := h.SessionManager.GetHistory(ctx, sessionID); err == nil {
			history = turns
		}
	}

	// Build session context for cache key
	sessionContext := ""
	if len(history) > 0 {
		sessionContext = history[len(history)-1].Query + " " + history[len(history)-1].Response
	}

	// Rewrite query if history exists
	rewrittenQuery := req.Query
	if len(history) > 0 {
		rewrittenQuery = h.rewriteQuery(ctx, req.Query, history)
	}

	// Generate embedding (needed for semantic cache and search)
	embedding, err := h.generateEmbedding(ctx, rewrittenQuery)
	if err != nil {
		writeSSEError(w, flusher, "failed to generate embedding: "+err.Error())
		return
	}

	// CAG: Check all cache layers before executing full pipeline
	if h.CAGOrchestrator != nil {
		filters := make(map[string][]string, 2) // grade + subject
		if req.Grade != "" {
			filters["grade"] = []string{req.Grade}
		}
		if req.Subject != "" {
			filters["subject"] = []string{req.Subject}
		}

		cachedResp, hit := h.CAGOrchestrator.QueryCAG(ctx, req.Query, embedding, sessionContext, req.Grade, req.Subject, filters)
		if hit && cachedResp != nil {
			log.Info().Str("cache_hit", cachedResp.CacheHit).Str("query", req.Query).Msg("CAG cache hit")

			// Convert cache.Source to types.Source
			modelSources := make([]types.Source, len(cachedResp.Sources))
			for i, src := range cachedResp.Sources {
				modelSources[i] = types.Source{
					ParentID: src.ParentID,
					Content:  src.Content,
					Score:    src.Score,
				}
			}

			// Return cached response
			for _, source := range modelSources {
				writeSSEEvent(w, flusher, "chunk", types.SSEChunkEvent{Content: source.Content})
			}
			writeSSEEvent(w, flusher, "end", types.SSEEndEvent{
				Sources:     modelSources,
				TotalTokens: 0,
			})
			return
		}
	}

	// Fall back to existing cache check (legacy)
	if h.Cache != nil {
		cached, err := h.Cache.Get(ctx, req.Query, sessionContext, req.Grade, req.Subject)
		if err == nil && cached != nil {
			// Convert cache.Source to types.Source
			modelSources := make([]types.Source, len(cached.Sources))
			for i, src := range cached.Sources {
				modelSources[i] = types.Source{
					ParentID: src.ParentID,
					Content:  src.Content,
					Score:    src.Score,
				}
			}

			// Return cached response
			for _, source := range modelSources {
				writeSSEEvent(w, flusher, "chunk", types.SSEChunkEvent{Content: source.Content})
			}
			writeSSEEvent(w, flusher, "end", types.SSEEndEvent{
				Sources:     modelSources,
				TotalTokens: 0,
			})
			return
		}
	}

	// All caches missed - execute full pipeline

	// Step 1: Call Vector Search Service to find relevant chunks
	sources, err := h.searchVectors(ctx, embedding, req.Grade, req.Subject)
	if err != nil {
		writeSSEError(w, flusher, "failed to search vectors: "+err.Error())
		return
	}

	// Story C: Check if search results are empty and trigger fallback
	if ShouldUseFallback(sources) && h.FallbackHandler != nil {
		RecordFallbackMetric()
		log.Info().Str("query", req.Query).Msg("Vector search returned zero results, triggering fallback")

		answer, totalTokens, err := h.FallbackHandler.GenerateFallbackResponse(ctx, req.Query, req.Grade, req.Subject)
		if err != nil {
			writeSSEError(w, flusher, "failed to generate fallback response: "+err.Error())
			return
		}

		// Stream the fallback answer
		chunkSize := 50
		for i := 0; i < len(answer); i += chunkSize {
			end := i + chunkSize
			if end > len(answer) {
				end = len(answer)
			}
			writeSSEEvent(w, flusher, "chunk", types.SSEChunkEvent{Content: answer[i:end]})
			totalTokens += chunkSize / 4
		}

		// Calculate confidence score for fallback (typically lower)
		confidence := calculateConfidenceScore(sources, 0.0)
		if confidence < 0.6 {
			middleware.RecordLowConfidenceResponse("fallback")
		}

		// Send end event with fallback flag
		writeSSEEvent(w, flusher, "end", types.SSEEndEvent{
			Sources:     sources,
			TotalTokens: totalTokens,
		})
		return
	}

	// Story H: Enrich query with user profile and session context
	enrichedQuery := h.enrichQuery(ctx, req.Query, sessionID, userID, req.Grade)

	// Step 2: Stream response from LLM Service
	totalTokens, err := h.streamLLMResponse(ctx, w, flusher, enrichedQuery.EnrichedPrompt, sources)
	if err != nil {
		writeSSEError(w, flusher, "failed to stream LLM response: "+err.Error())
		return
	}

	// Story E: Calculate and log confidence score
	confidence := calculateConfidenceScore(sources, 0.0)
	if confidence < 0.6 {
		middleware.RecordLowConfidenceResponse("rag")
	}

	// Send end event
	writeSSEEvent(w, flusher, "end", types.SSEEndEvent{
		Sources:     sources,
		TotalTokens: totalTokens,
	})

	// Cache the response in all layers
	h.cacheResponse(ctx, req.Query, embedding, sessionContext, req.Grade, req.Subject, sources)

	// Append to session history
	if h.SessionManager != nil && sessionID != "" {
		var sb strings.Builder
		for _, src := range sources {
			sb.WriteString(src.Content)
			sb.WriteByte('\n')
		}

		_ = h.SessionManager.AddTurn(ctx, sessionID, session.ConversationTurn{
			Query:     req.Query,
			Response:  sb.String(),
			Timestamp: time.Now(),
			Grade:     req.Grade,
			Subject:   req.Subject,
		})
	}
}

func (h *QueryHandler) handleNonStreaming(w http.ResponseWriter, r *http.Request, req *types.QueryRequest, userID, requestID string) {
	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	// Get session history
	var history []session.ConversationTurn
	if h.SessionManager != nil && req.SessionID != "" {
		if turns, err := h.SessionManager.GetHistory(ctx, req.SessionID); err == nil {
			history = turns
		}
	}

	// Build session context for cache key
	sessionContext := ""
	if len(history) > 0 {
		sessionContext = history[len(history)-1].Query + " " + history[len(history)-1].Response
	}

	// Rewrite query if history exists
	rewrittenQuery := req.Query
	if len(history) > 0 {
		rewrittenQuery = h.rewriteQuery(ctx, req.Query, history)
	}

	// Generate embedding (needed for semantic cache and search)
	embedding, err := h.generateEmbedding(ctx, rewrittenQuery)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "embedding_failed", Message: "Failed to generate embedding: " + err.Error()},
		})
		return
	}

	// CAG: Check all cache layers before executing full pipeline
	if h.CAGOrchestrator != nil {
		filters := make(map[string][]string, 2) // grade + subject
		if req.Grade != "" {
			filters["grade"] = []string{req.Grade}
		}
		if req.Subject != "" {
			filters["subject"] = []string{req.Subject}
		}

		cachedResp, hit := h.CAGOrchestrator.QueryCAG(ctx, req.Query, embedding, sessionContext, req.Grade, req.Subject, filters)
		if hit && cachedResp != nil {
			log.Info().Str("cache_hit", cachedResp.CacheHit).Str("query", req.Query).Msg("CAG cache hit")

			// Convert cache.Source to types.Source
			modelSources := make([]types.Source, len(cachedResp.Sources))
			for i, src := range cachedResp.Sources {
				modelSources[i] = types.Source{
					ParentID: src.ParentID,
					Content:  src.Content,
					Score:    src.Score,
				}
			}

			writeJSON(w, http.StatusOK, types.QueryResponse{
				SessionID: req.SessionID,
				RequestID: requestID,
				Sources:   modelSources,
			})
			return
		}
	}

	// Fall back to existing cache check (legacy)
	if h.Cache != nil {
		cached, err := h.Cache.Get(ctx, req.Query, sessionContext, req.Grade, req.Subject)
		if err == nil && cached != nil {
			// Convert cache.Source to types.Source
			modelSources := make([]types.Source, len(cached.Sources))
			for i, src := range cached.Sources {
				modelSources[i] = types.Source{
					ParentID: src.ParentID,
					Content:  src.Content,
					Score:    src.Score,
				}
			}

			writeJSON(w, http.StatusOK, types.QueryResponse{
				SessionID: req.SessionID,
				RequestID: requestID,
				Sources:   modelSources,
			})
			return
		}
	}

	// All caches missed - execute full pipeline

	// Search vectors
	sources, err := h.searchVectors(ctx, embedding, req.Grade, req.Subject)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "search_failed", Message: "Failed to search vectors: " + err.Error()},
		})
		return
	}

	// Story C: Check if search results are empty and trigger fallback
	if ShouldUseFallback(sources) && h.FallbackHandler != nil {
		RecordFallbackMetric()
		log.Info().Str("query", req.Query).Msg("Vector search returned zero results, triggering fallback")

		answer, totalTokens, err := h.FallbackHandler.GenerateFallbackResponse(ctx, req.Query, req.Grade, req.Subject)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, types.ErrorResponse{
				Error: types.ErrorDetail{Code: "fallback_failed", Message: "Failed to generate fallback response: " + err.Error()},
			})
			return
		}

		// Story E: Calculate confidence score for fallback
		confidence := calculateConfidenceScore(sources, 0.0)
		if confidence < 0.6 {
			middleware.RecordLowConfidenceResponse("fallback")
		}

		// Cache response in all layers
		h.cacheResponse(ctx, req.Query, embedding, sessionContext, req.Grade, req.Subject, sources)

		// Append to session history
		if h.SessionManager != nil && req.SessionID != "" {
			_ = h.SessionManager.AddTurn(ctx, req.SessionID, session.ConversationTurn{
				Query:     req.Query,
				Response:  answer,
				Timestamp: time.Now(),
				Grade:     req.Grade,
				Subject:   req.Subject,
			})
		}

		wasFallback := true
		writeJSON(w, http.StatusOK, types.QueryResponse{
			SessionID:       req.SessionID,
			RequestID:       requestID,
			Answer:          answer,
			Sources:         sources,
			TotalTokens:     totalTokens,
			ConfidenceScore: &confidence,
			WasFallback:     &wasFallback,
		})
		return
	}

	// Story H: Enrich query with user profile and session context
	enrichedQuery := h.enrichQuery(ctx, req.Query, req.SessionID, userID, req.Grade)

	// Get LLM response (non-streaming)
	answer, totalTokens, err := h.getLLMResponse(ctx, enrichedQuery.EnrichedPrompt, sources)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "llm_failed", Message: "Failed to get LLM response: " + err.Error()},
		})
		return
	}

	// Story E: Calculate confidence score
	confidence := calculateConfidenceScore(sources, 0.0)
	if confidence < 0.6 {
		middleware.RecordLowConfidenceResponse("rag")
	}

	// Story J: Run grounding validation via NLI service
	var groundingScore float64
	var groundingWarning string
	if h.NLIClient != nil && len(sources) > 0 {
		contextChunks := make([]string, len(sources))
		for i, s := range sources {
			contextChunks[i] = s.Content
		}
		validation, _ := h.NLIClient.BatchValidateClaims(ctx, []string{answer}, contextChunks)
		if validation != nil {
			groundingScore = validation.OverallGroundingScore
			if groundingScore < 0.6 {
				groundingWarning = "This response may contain unverified information."
			}
		}
	}

	// Cache response in all layers
	h.cacheResponse(ctx, req.Query, embedding, sessionContext, req.Grade, req.Subject, sources)

	// Append to session history
	if h.SessionManager != nil && req.SessionID != "" {
		_ = h.SessionManager.AddTurn(ctx, req.SessionID, session.ConversationTurn{
			Query:     req.Query,
			Response:  answer,
			Timestamp: time.Now(),
			Grade:     req.Grade,
			Subject:   req.Subject,
		})
	}

	// Build response with optional grounding data
	resp := types.QueryResponse{
		SessionID:       req.SessionID,
		RequestID:       requestID,
		Answer:          answer,
		Sources:         sources,
		TotalTokens:     totalTokens,
		ConfidenceScore: &confidence,
	}

	if groundingScore > 0 {
		resp.GroundingScore = &groundingScore
		if groundingWarning != "" {
			warning := groundingWarning
			resp.GroundingWarning = &warning
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

// rewriteQuery calls the Query Understanding Service to rewrite the query
func (h *QueryHandler) rewriteQuery(ctx context.Context, query string, history []session.ConversationTurn) string {
	if h.QueryUnderstandingClient == nil {
		return query
	}

	type rewriteRequest struct {
		Query   string                   `json:"query"`
		History []session.ConversationTurn `json:"history"`
	}

	type rewriteResponse struct {
		RewrittenQuery string `json:"rewritten_query"`
	}

	var result rewriteResponse
	err := h.QueryUnderstandingClient.PostJSON(ctx, "/rewrite", rewriteRequest{
		Query:   query,
		History: history,
	}, &result)

	if err != nil {
		log.Warn().Err(err).Msg("Failed to rewrite query, using original query")
		return query
	}

	if result.RewrittenQuery != "" {
		return result.RewrittenQuery
	}

	return query
}

// generateEmbedding calls the Embedding Service
func (h *QueryHandler) generateEmbedding(ctx context.Context, text string) ([]float64, error) {
	if h.EmbeddingClient == nil {
		return nil, fmt.Errorf("embedding service not configured")
	}

	type embeddingRequest struct {
		Text string `json:"text"`
	}

	type embeddingResponse struct {
		Embedding []float64 `json:"embedding"`
	}

	var result embeddingResponse
	err := h.EmbeddingClient.PostJSON(ctx, "/embed", embeddingRequest{Text: text}, &result)
	if err != nil {
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}

	return result.Embedding, nil
}

// searchVectors calls the Vector Search Service
func (h *QueryHandler) searchVectors(ctx context.Context, embedding []float64, grade, subject string) ([]types.Source, error) {
	if h.VectorSearchClient == nil {
		return nil, fmt.Errorf("vector search service not configured")
	}

	type searchRequest struct {
		Embedding []float64 `json:"embedding"`
		Grade     string    `json:"grade,omitempty"`
		Subject   string    `json:"subject,omitempty"`
		TopK      int       `json:"top_k"`
	}

	type searchResponse struct {
		Results []types.Source `json:"results"`
	}

	var result searchResponse
	err := h.VectorSearchClient.PostJSON(ctx, "/search", searchRequest{
		Embedding: embedding,
		Grade:     grade,
		Subject:   subject,
		TopK:      5,
	}, &result)
	if err != nil {
		return nil, fmt.Errorf("vector search request failed: %w", err)
	}

	return result.Results, nil
}

// streamLLMResponse streams the LLM response via SSE
func (h *QueryHandler) streamLLMResponse(ctx context.Context, w http.ResponseWriter, flusher http.Flusher, query string, sources []types.Source) (int, error) {
	if h.LLMClient == nil {
		return 0, fmt.Errorf("LLM service not configured")
	}

	type llmRequest struct {
		Query   string         `json:"query"`
		Sources []types.Source `json:"sources"`
		Stream  bool           `json:"stream"`
	}

	resp, err := h.LLMClient.Stream(ctx, "/generate", llmRequest{
		Query:   query,
		Sources: sources,
		Stream:  true,
	})
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Launch goroutine to close response body on client disconnect.
	// This unblocks scanner.Scan() when the client disconnects,
	// avoiding the busy-poll pattern.
	go func() {
		<-ctx.Done()
		resp.Body.Close()
	}()

	// Read and forward SSE events
	scanner := bufio.NewScanner(resp.Body)
	totalTokens := 0
	lastActivity := time.Now()

	for scanner.Scan() {
		// Send heartbeat if no activity for 15 seconds
		if time.Since(lastActivity) > 15*time.Second {
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
			lastActivity = time.Now()
		}

		line := scanner.Text()
		lastActivity = time.Now()

		if line == "" {
			continue
		}

		// Forward SSE event
		fmt.Fprintln(w, line)
		flusher.Flush()

		// Try to parse as data event to count tokens
		if len(line) > 5 && line[:5] == "data:" {
			var event types.SSEChunkEvent
			if err := json.Unmarshal([]byte(line[5:]), &event); err == nil {
				// Count approximate tokens (words)
				totalTokens += len(event.Content) / 4
			}
		}
	}

	// Check why scanner.Scan() returned false
	if err := scanner.Err(); err != nil {
		// Context cancellation means client disconnected
		if ctx.Err() != nil {
			return totalTokens, fmt.Errorf("client disconnected")
		}
		return totalTokens, fmt.Errorf("error reading stream: %w", err)
	}
	return totalTokens, nil
}

// getLLMResponse gets a non-streaming LLM response
func (h *QueryHandler) getLLMResponse(ctx context.Context, query string, sources []types.Source) (string, int, error) {
	if h.LLMClient == nil {
		return "", 0, fmt.Errorf("LLM service not configured")
	}

	type llmRequest struct {
		Query   string         `json:"query"`
		Sources []types.Source `json:"sources"`
		Stream  bool           `json:"stream"`
	}

	type llmResponse struct {
		Answer      string `json:"answer"`
		TotalTokens int    `json:"total_tokens"`
	}

	var result llmResponse
	body, err := h.LLMClient.PostJSONRaw(ctx, "/generate", llmRequest{
		Query:   query,
		Sources: sources,
		Stream:  false,
	})
	if err != nil {
		return "", 0, err
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", 0, fmt.Errorf("failed to decode LLM response: %w", err)
	}

	return result.Answer, result.TotalTokens, nil
}

// SSE helper functions
func writeSSEEvent(w http.ResponseWriter, flusher http.Flusher, event string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		writeSSEError(w, flusher, "failed to marshal event data")
		return
	}

	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, string(jsonData))
	if flusher != nil {
		flusher.Flush()
	}
}

func writeSSEError(w http.ResponseWriter, flusher http.Flusher, message string) {
	data := types.SSEErrorEvent{
		Error: types.ErrorDetail{
			Code:    "internal_error",
			Message: message,
		},
	}
	writeSSEEvent(w, flusher, "error", data)
}

// cacheResponse caches the response in all available cache layers concurrently
func (h *QueryHandler) cacheResponse(ctx context.Context, query string, embedding []float64, sessionContext, grade, subject string, sources []types.Source) {
	// Build content from sources using strings.Builder for efficiency
	var sb strings.Builder
	for _, src := range sources {
		sb.WriteString(src.Content)
		sb.WriteByte('\n')
	}

	// Convert types.Source to cache.Source
	cacheSources := make([]cache.Source, len(sources))
	for i, src := range sources {
		cacheSources[i] = cache.Source{
			ParentID: src.ParentID,
			Content:  src.Content,
			Score:    src.Score,
		}
	}

	response := &cache.CAGResponse{
		Content: sb.String(),
		Sources: cacheSources,
	}

	// Use CAG orchestrator if available (caches in all layers concurrently)
	if h.CAGOrchestrator != nil {
		h.CAGOrchestrator.CacheResult(ctx, query, embedding, sessionContext, grade, subject, response)
	} else if h.Cache != nil {
		// Fall back to legacy cache
		cachedResp := &cache.CachedResponse{
			Content:   sb.String(),
			Sources:   cacheSources,
			CreatedAt: time.Now(),
		}
		_ = h.Cache.Set(ctx, query, sessionContext, grade, subject, cachedResp)
	}
}

// enrichQuery enriches a query with user profile and session context.
// If the enricher is not configured or enrichment fails, it returns the original query.
func (h *QueryHandler) enrichQuery(ctx context.Context, query, sessionID, userID, grade string) *EnrichedQuery {
	if h.Enricher == nil {
		return &EnrichedQuery{
			OriginalQuery:  query,
			EnrichedPrompt: query,
			Grade:          grade,
		}
	}

	enriched, err := h.Enricher.Enrich(ctx, sessionID, userID, query, grade)
	if err != nil {
		log.Warn().Err(err).Str("query", query).Msg("Query enrichment failed, using original query")
		return &EnrichedQuery{
			OriginalQuery:  query,
			EnrichedPrompt: query,
			Grade:          grade,
		}
	}

	return enriched
}

// calculateConfidenceScore computes a confidence score for the response (Story E)
// It combines retrieval quality (chunk scores) and optional LLM confidence.
// Returns a value between 0.0 and 1.0.
func calculateConfidenceScore(sources []types.Source, llmConfidence float64) float64 {
	if len(sources) == 0 {
		// No sources retrieved - very low confidence
		return 0.2
	}

	// Calculate average retrieval score from sources
	var totalScore float64
	for _, src := range sources {
		totalScore += src.Score
	}
	avgRetrievalScore := totalScore / float64(len(sources))

	// Weight retrieval quality more heavily (70%) than LLM confidence (30%)
	weight := 0.7
	if llmConfidence > 0 {
		return weight*avgRetrievalScore + (1-weight)*llmConfidence
	}

	// If no LLM confidence provided, use retrieval score as base
	// Normalize to 0.0-1.0 range (scores are typically 0.0-1.0 already)
	confidence := avgRetrievalScore
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}
