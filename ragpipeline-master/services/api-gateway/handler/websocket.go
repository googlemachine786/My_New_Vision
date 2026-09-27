package handler

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/visionary/ragpipeline/services/api-gateway/cache"
	"github.com/visionary/ragpipeline/services/api-gateway/client"
	"github.com/visionary/ragpipeline/services/api-gateway/middleware"
	"github.com/visionary/ragpipeline/pkg/types"
	"github.com/visionary/ragpipeline/services/api-gateway/session"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// WebSocket message types
const (
	MsgTypeQuery = "query"
	MsgTypeChunk = "chunk"
	MsgTypeDone  = "done"
	MsgTypeError = "error"
	MsgTypePing  = "ping"
	MsgTypePong  = "pong"
	MsgTypeClose = "close"
)

// WSMessage represents a WebSocket message frame
type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"data,omitempty"`
}

// WSQueryPayload represents the payload for a query message
type WSQueryPayload struct {
	Query     string `json:"query"`
	SessionID string `json:"session_id,omitempty"`
	Grade     string `json:"grade,omitempty"`
	Subject   string `json:"subject,omitempty"`
}

// WSChunkPayload represents the payload for a chunk message
type WSChunkPayload struct {
	Content string `json:"content"`
}

// WSDonePayload represents the payload for a done message
type WSDonePayload struct {
	SessionID   string         `json:"session_id"`
	RequestID   string         `json:"request_id"`
	Sources     []types.Source `json:"sources"`
	TotalTokens int            `json:"total_tokens"`
	DurationMs  int64          `json:"duration_ms"`
}

// WSErrorPayload represents the payload for an error message
type WSErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WebSocketUpgrader configures the WebSocket upgrader
var WebSocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		allowedOrigins := getAllowedWSOrigins()
		if len(allowedOrigins) == 0 {
			return true
		}
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				return true
			}
		}
		return false
	},
}

// activeConnections tracks the number of active WebSocket connections
var activeConnections int64

// GetActiveConnections returns the current number of active WebSocket connections
func GetActiveConnections() int64 {
	return atomic.LoadInt64(&activeConnections)
}

// allowedWSOrigins caches the allowed origins list
var allowedWSOrigins []string
var allowedWSOriginsOnce sync.Once

// getAllowedWSOrigins returns the list of allowed WebSocket origins from environment
func getAllowedWSOrigins() []string {
	allowedWSOriginsOnce.Do(func() {
		origins := os.Getenv("WS_ALLOWED_ORIGINS")
		if origins == "" {
			allowedWSOrigins = []string{}
			return
		}
		allowedWSOrigins = strings.Split(origins, ",")
		for i, o := range allowedWSOrigins {
			allowedWSOrigins[i] = strings.TrimSpace(o)
		}
	})
	return allowedWSOrigins
}

// WebSocketHandler handles WebSocket connections for low-latency streaming
type WebSocketHandler struct {
	EmbeddingClient          *client.ServiceClient
	VectorSearchClient       *client.ServiceClient
	QueryUnderstandingClient *client.ServiceClient
	LLMClient                *client.ServiceClient
	CAGOrchestrator          *cache.CAGOrchestrator
	SessionManager           *session.SessionManager
	RedisClient              *redis.Client

	pingInterval time.Duration
	pongWait     time.Duration
	writeWait    time.Duration
}

// WebSocketHandlerDeps holds dependencies for the WebSocket handler
type WebSocketHandlerDeps struct {
	EmbeddingClient          *client.ServiceClient
	VectorSearchClient       *client.ServiceClient
	QueryUnderstandingClient *client.ServiceClient
	LLMClient                *client.ServiceClient
	CAGOrchestrator          *cache.CAGOrchestrator
	SessionManager           *session.SessionManager
	RedisClient              *redis.Client
	PingInterval             time.Duration
	PongWait                 time.Duration
	WriteWait                time.Duration
}

// NewWebSocketHandler creates a new WebSocket handler with the given dependencies
func NewWebSocketHandler(deps WebSocketHandlerDeps) *WebSocketHandler {
	pingInterval := deps.PingInterval
	if pingInterval == 0 {
		pingInterval = 30 * time.Second
	}
	pongWait := deps.PongWait
	if pongWait == 0 {
		pongWait = 60 * time.Second
	}
	writeWait := deps.WriteWait
	if writeWait == 0 {
		writeWait = 10 * time.Second
	}

	return &WebSocketHandler{
		EmbeddingClient:          deps.EmbeddingClient,
		VectorSearchClient:       deps.VectorSearchClient,
		QueryUnderstandingClient: deps.QueryUnderstandingClient,
		LLMClient:                deps.LLMClient,
		CAGOrchestrator:          deps.CAGOrchestrator,
		SessionManager:           deps.SessionManager,
		RedisClient:              deps.RedisClient,
		pingInterval:             pingInterval,
		pongWait:                 pongWait,
		writeWait:                writeWait,
	}
}

// ServeHTTP upgrades the HTTP connection to WebSocket and handles the connection lifecycle
func (h *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := WebSocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade to WebSocket")
		return
	}

	atomic.AddInt64(&activeConnections, 1)
	defer atomic.AddInt64(&activeConnections, -1)

	defer func() {
		if err := conn.Close(); err != nil {
			log.Debug().Err(err).Msg("Error closing WebSocket connection")
		}
	}()

	log.Info().
		Str("remote_addr", conn.RemoteAddr().String()).
		Int64("active_connections", GetActiveConnections()).
		Msg("WebSocket connection established")

	h.handleConnection(conn, r)
}

// handleConnection manages the bidirectional WebSocket communication
func (h *WebSocketHandler) handleConnection(conn *websocket.Conn, r *http.Request) {
	conn.SetReadLimit(1 << 16)
	conn.SetReadDeadline(time.Now().Add(h.pongWait))

	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(h.pongWait))
		return nil
	})

	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	go h.startPingLoop(conn)
	h.startReadLoop(conn, userID, requestID)
}

// startPingLoop sends periodic pings to keep the connection alive
func (h *WebSocketHandler) startPingLoop(conn *websocket.Conn) {
	ticker := time.NewTicker(h.pingInterval)
	defer ticker.Stop()

	for range ticker.C {
		conn.SetWriteDeadline(time.Now().Add(h.writeWait))
		if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			log.Debug().Err(err).Msg("Failed to send ping")
			return
		}
	}
}

// startReadLoop reads incoming messages and processes queries
func (h *WebSocketHandler) startReadLoop(conn *websocket.Conn, userID, requestID string) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Warn().Err(err).Msg("WebSocket unexpected close")
			}
			return
		}

		conn.SetReadDeadline(time.Now().Add(h.pongWait))

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			h.sendErrorMessage(conn, "invalid_message", "Failed to parse message")
			continue
		}

		switch wsMsg.Type {
		case MsgTypeQuery:
			h.handleWSQuery(conn, wsMsg, userID, requestID)
		case MsgTypePing:
			h.sendWSMessage(conn, MsgTypePong, map[string]string{"status": "ok"})
		case MsgTypeClose:
			return
		default:
			h.sendErrorMessage(conn, "unknown_message_type", fmt.Sprintf("Unknown message type: %s", wsMsg.Type))
		}
	}
}

// handleWSQuery processes a query received via WebSocket
func (h *WebSocketHandler) handleWSQuery(conn *websocket.Conn, wsMsg WSMessage, userID, requestID string) {
	startTime := time.Now()

	payloadBytes, err := json.Marshal(wsMsg.Payload)
	if err != nil {
		h.sendErrorMessage(conn, "invalid_payload", "Failed to parse query payload")
		return
	}

	var queryPayload WSQueryPayload
	if err := json.Unmarshal(payloadBytes, &queryPayload); err != nil {
		h.sendErrorMessage(conn, "invalid_payload", "Failed to parse query: "+err.Error())
		return
	}

	if queryPayload.Query == "" {
		h.sendErrorMessage(conn, "validation_error", "Query is required")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	req := &types.QueryRequest{
		Query:     queryPayload.Query,
		SessionID: queryPayload.SessionID,
		Grade:     queryPayload.Grade,
		Subject:   queryPayload.Subject,
		Stream:    true,
	}

	sessionID := req.SessionID
	if sessionID == "" && h.SessionManager != nil {
		newSession, err := h.SessionManager.CreateSession(ctx, userID)
		if err != nil {
			h.sendErrorMessage(conn, "session_error", "Failed to create session")
			return
		}
		sessionID = newSession.ID
	}

	// Get session history
	var history []session.ConversationTurn
	if h.SessionManager != nil && sessionID != "" {
		if turns, err := h.SessionManager.GetHistory(ctx, sessionID); err == nil {
			history = turns
		}
	}

	sessionContext := ""
	if len(history) > 0 {
		sessionContext = history[len(history)-1].Query + " " + history[len(history)-1].Response
	}

	rewrittenQuery := req.Query
	if len(history) > 0 {
		rewrittenQuery = h.rewriteQueryForWS(ctx, req.Query, history)
	}

	embedding, err := h.generateEmbeddingForWS(ctx, rewrittenQuery)
	if err != nil {
		h.sendErrorMessage(conn, "embedding_failed", "Failed to generate embedding: "+err.Error())
		return
	}

	// CAG cache check
	if h.CAGOrchestrator != nil {
		filters := make(map[string][]string, 2)
		if req.Grade != "" {
			filters["grade"] = []string{req.Grade}
		}
		if req.Subject != "" {
			filters["subject"] = []string{req.Subject}
		}

		cachedResp, hit := h.CAGOrchestrator.QueryCAG(ctx, req.Query, embedding, sessionContext, req.Grade, req.Subject, filters)
		if hit && cachedResp != nil {
			log.Info().Str("cache_hit", cachedResp.CacheHit).Str("query", req.Query).Msg("CAG cache hit (WebSocket)")

			for _, source := range cachedResp.Sources {
				modelSource := types.Source{
					ParentID: source.ParentID,
					Content:  source.Content,
					Score:    source.Score,
				}
				h.streamContentChunks(conn, modelSource.Content)
			}

			h.sendWSMessage(conn, MsgTypeDone, WSDonePayload{
				SessionID:  sessionID,
				RequestID:  requestID,
				TotalTokens: 0,
				DurationMs: time.Since(startTime).Milliseconds(),
			})
			return
		}
	}

	// Execute full pipeline
	sources, err := h.searchVectorsForWS(ctx, embedding, req.Grade, req.Subject)
	if err != nil {
		h.sendErrorMessage(conn, "search_failed", "Failed to search vectors: "+err.Error())
		return
	}

	totalTokens, err := h.streamLLMResponseWS(ctx, conn, req.Query, sources)
	if err != nil {
		h.sendErrorMessage(conn, "llm_failed", "Failed to stream LLM response: "+err.Error())
		return
	}

	h.sendWSMessage(conn, MsgTypeDone, WSDonePayload{
		SessionID:   sessionID,
		RequestID:   requestID,
		Sources:     sources,
		TotalTokens: totalTokens,
		DurationMs:  time.Since(startTime).Milliseconds(),
	})

	h.cacheResponseForWS(ctx, req.Query, embedding, sessionContext, req.Grade, req.Subject, sources)

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

// streamContentChunks splits content into word-level chunks and streams them
func (h *WebSocketHandler) streamContentChunks(conn *websocket.Conn, content string) {
	words := strings.Fields(content)
	for i := 0; i < len(words); i += 3 {
		end := i + 3
		if end > len(words) {
			end = len(words)
		}
		chunk := strings.Join(words[i:end], " ")
		h.sendWSMessage(conn, MsgTypeChunk, WSChunkPayload{Content: chunk})
	}
}

// streamLLMResponseWS streams the LLM response via WebSocket
func (h *WebSocketHandler) streamLLMResponseWS(ctx context.Context, conn *websocket.Conn, query string, sources []types.Source) (int, error) {
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

	go func() {
		<-ctx.Done()
		resp.Body.Close()
	}()

	scanner := bufio.NewScanner(resp.Body)
	totalTokens := 0
	lastActivity := time.Now()

	for scanner.Scan() {
		if time.Since(lastActivity) > 15*time.Second {
			h.sendWSMessage(conn, MsgTypeChunk, WSChunkPayload{Content: ""})
			lastActivity = time.Now()
		}

		line := scanner.Text()
		lastActivity = time.Now()

		if line == "" {
			continue
		}

		if len(line) > 5 && line[:5] == "data:" {
			var event types.SSEChunkEvent
			if err := json.Unmarshal([]byte(line[5:]), &event); err == nil {
				if event.Content != "" {
					h.streamContentChunks(conn, event.Content)
					totalTokens += len(event.Content) / 4
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		if ctx.Err() != nil {
			return totalTokens, fmt.Errorf("client disconnected")
		}
		return totalTokens, fmt.Errorf("error reading stream: %w", err)
	}

	return totalTokens, nil
}

// sendWSMessage sends a typed message over WebSocket
func (h *WebSocketHandler) sendWSMessage(conn *websocket.Conn, msgType string, payload interface{}) {
	msg := WSMessage{
		Type:    msgType,
		Payload: payload,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal WebSocket message")
		return
	}

	conn.SetWriteDeadline(time.Now().Add(h.writeWait))
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Debug().Err(err).Msg("Failed to write WebSocket message")
	}
}

// sendErrorMessage sends an error message over WebSocket
func (h *WebSocketHandler) sendErrorMessage(conn *websocket.Conn, code, message string) {
	h.sendWSMessage(conn, MsgTypeError, WSErrorPayload{
		Code:    code,
		Message: message,
	})
}

// rewriteQueryForWS calls the Query Understanding Service to rewrite the query
func (h *WebSocketHandler) rewriteQueryForWS(ctx context.Context, query string, history []session.ConversationTurn) string {
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
		log.Warn().Err(err).Msg("Failed to rewrite query (WebSocket), using original")
		return query
	}

	if result.RewrittenQuery != "" {
		return result.RewrittenQuery
	}

	return query
}

// generateEmbeddingForWS calls the Embedding Service
func (h *WebSocketHandler) generateEmbeddingForWS(ctx context.Context, text string) ([]float64, error) {
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

// searchVectorsForWS calls the Vector Search Service
func (h *WebSocketHandler) searchVectorsForWS(ctx context.Context, embedding []float64, grade, subject string) ([]types.Source, error) {
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

// cacheResponseForWS caches the response in all available cache layers
func (h *WebSocketHandler) cacheResponseForWS(ctx context.Context, query string, embedding []float64, sessionContext, grade, subject string, sources []types.Source) {
	var sb strings.Builder
	for _, src := range sources {
		sb.WriteString(src.Content)
		sb.WriteByte('\n')
	}

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

	if h.CAGOrchestrator != nil {
		h.CAGOrchestrator.CacheResult(ctx, query, embedding, sessionContext, grade, subject, response)
	}
}
