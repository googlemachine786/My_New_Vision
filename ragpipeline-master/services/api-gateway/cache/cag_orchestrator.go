// Package cache provides CAG (Cache Augmented Generation) orchestration.
package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// EmbeddingCache is the interface for embedding cache (defined locally to avoid cross-service imports).
type EmbeddingCache interface {
	Get(ctx context.Context, text, taskType string) (interface{}, error)
	Set(ctx context.Context, text, taskType string, embedding interface{}) error
	Stats() map[string]interface{}
	CostSaved() float64
}

// SearchCache defines the interface for search result caching.
// This avoids import cycles with the vector-search-service package.
type SearchCache interface {
	Get(ctx context.Context, embedding []float64, topK int, filters map[string][]string) (interface{}, error)
	Set(ctx context.Context, embedding []float64, topK int, filters map[string][]string, result interface{}) error
}

// CAGOrchestrator coordinates all cache layers in the CAG pipeline
type CAGOrchestrator struct {
	ExactCache     *ResponseCache
	SemanticCache  *SemanticCache
	EmbeddingCache EmbeddingCache // Interface to avoid cross-service imports
	SearchCache    SearchCache // Typed interface, not interface{}
	TemplateCache  *TemplateCache

	// Metrics
	mu      sync.RWMutex
	metrics CAGMetrics
}

// CAGMetrics tracks cache performance across all layers
type CAGMetrics struct {
	// Layer 1: Exact Match
	ExactMatchHits   int64 `json:"exact_match_hits"`
	ExactMatchMisses int64 `json:"exact_match_misses"`

	// Layer 2: Semantic
	SemanticHits   int64   `json:"semantic_hits"`
	SemanticMisses int64   `json:"semantic_misses"`
	AvgSimilarity  float64 `json:"avg_similarity"`

	// Layer 3: Embedding
	EmbeddingCacheHits   int64 `json:"embedding_cache_hits"`
	EmbeddingCacheMisses int64 `json:"embedding_cache_misses"`

	// Layer 4: Search
	SearchCacheHits   int64 `json:"search_cache_hits"`
	SearchCacheMisses int64 `json:"search_cache_misses"`

	// Layer 5: Template
	TemplateCacheHits   int64 `json:"template_cache_hits"`
	TemplateCacheMisses int64 `json:"template_cache_misses"`

	// Overall
	TotalCostSaved        float64 `json:"total_cost_saved"`
	LLMCallsAvoided       int64   `json:"llm_calls_avoided"`
	EmbeddingCallsAvoided int64   `json:"embedding_calls_avoided"`
}

// CAGResponse represents a response from the CAG pipeline
type CAGResponse struct {
	Content   string   `json:"content"`
	Sources   []Source `json:"sources"`
	CacheHit  string   `json:"cache_hit,omitempty"` // "exact", "semantic", "template", or ""
	Latency   time.Duration `json:"latency_ms"`
}

// NewCAGOrchestrator creates a new CAG orchestrator
func NewCAGOrchestrator(
	exactCache *ResponseCache,
	semanticCache *SemanticCache,
	templateCache *TemplateCache,
) *CAGOrchestrator {
	return &CAGOrchestrator{
		ExactCache:    exactCache,
		SemanticCache: semanticCache,
		TemplateCache: templateCache,
	}
}

// SetEmbeddingCache sets the embedding cache (from embedding service)
func (o *CAGOrchestrator) SetEmbeddingCache(cache EmbeddingCache) {
	o.EmbeddingCache = cache
}

// SetSearchCache sets the search cache (from vector search service)
func (o *CAGOrchestrator) SetSearchCache(cache SearchCache) {
	o.SearchCache = cache
}

// QueryCAG executes the full CAG query flow through all cache layers
// Returns (response, true) if cache hit, (nil, false) if all caches missed
func (o *CAGOrchestrator) QueryCAG(ctx context.Context, query string, embedding []float64, sessionContext, grade, subject string, filters map[string][]string) (*CAGResponse, bool) {
	startTime := time.Now()

	// Layer 1: Exact match cache
	if o.ExactCache != nil {
		cached, err := o.ExactCache.Get(ctx, query, sessionContext, grade, subject)
		if err == nil && cached != nil {
			o.recordExactHit()
			return &CAGResponse{
				Content:  cached.Content,
				Sources:  cached.Sources,
				CacheHit: "exact",
				Latency:  time.Since(startTime),
			}, true
		}
		o.recordExactMiss()
	}

	// Layer 2: Semantic cache
	if o.SemanticCache != nil && len(embedding) > 0 {
		similar, err := o.SemanticCache.SearchSimilar(ctx, embedding, 0)
		if err == nil && similar != nil {
			o.recordSemanticHit()
			return &CAGResponse{
				Content:  similar.Response,
				Sources:  similar.Sources,
				CacheHit: "semantic",
				Latency:  time.Since(startTime),
			}, true
		}
		o.recordSemanticMiss()
	}

	// Layer 3: Embedding cache (checked in embedding service, just track here)
	if o.EmbeddingCache != nil && len(embedding) > 0 {
		o.recordEmbeddingCheck()
	}

	// Layer 4: Search cache (checked in search service)
	// Tracked separately in vector-search-service

	// Layer 5: Template cache (for common query patterns with known intent)
	if o.TemplateCache != nil {
		// Template cache is checked at the LLM response generation level
		// The query handler should check this after retrieving chunks
		o.CacheTemplateMiss() // Default to miss; will be overridden if template found
	}

	// All caches missed - caller should execute full pipeline
	return nil, false
}

// CacheResult caches all layers after a successful query.
// Uses fire-and-forget goroutines with background context so cache writes
// never block the response or hang on a cancelled request context.
func (o *CAGOrchestrator) CacheResult(ctx context.Context, query string, embedding []float64, sessionContext, grade, subject string, response *CAGResponse) {
	if response == nil {
		return
	}

	// Use a detached background context with a timeout for async cache writes.
	// The caller's ctx may be cancelled as soon as the HTTP response is flushed,
	// which would cause wg.Wait() to hang indefinitely if goroutines block on I/O.
	bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Cache in exact match cache
	if o.ExactCache != nil {
		go func() {
			cachedResp := &CachedResponse{
				Content:   response.Content,
				Sources:   response.Sources,
				CreatedAt: time.Now(),
			}
			if err := o.ExactCache.Set(bgCtx, query, sessionContext, grade, subject, cachedResp); err != nil {
				log.Debug().Err(err).Msg("Failed to cache in exact match layer")
			}
		}()
	}

	// Cache in semantic cache
	if o.SemanticCache != nil && len(embedding) > 0 {
		go func() {
			semanticResp := &CachedSemanticResponse{
				QueryText:      query,
				QueryEmbedding: embedding,
				Response:       response.Content,
				Sources:        response.Sources,
				CreatedAt:      time.Now(),
			}
			if err := o.SemanticCache.Add(bgCtx, embedding, semanticResp); err != nil {
				log.Debug().Err(err).Msg("Failed to cache in semantic layer")
			}
		}()
	}
}

// CacheEmbeddingHit records an embedding cache hit (called from embedding service)
func (o *CAGOrchestrator) CacheEmbeddingHit() {
	o.mu.Lock()
	o.metrics.EmbeddingCacheHits++
	o.metrics.EmbeddingCallsAvoided++
	o.metrics.TotalCostSaved += 0.0001 // ~$0.0001 per embedding
	o.mu.Unlock()
}

// CacheEmbeddingMiss records an embedding cache miss
func (o *CAGOrchestrator) CacheEmbeddingMiss() {
	o.mu.Lock()
	o.metrics.EmbeddingCacheMisses++
	o.mu.Unlock()
}

// CacheSearchHit records a search cache hit
func (o *CAGOrchestrator) CacheSearchHit() {
	o.mu.Lock()
	o.metrics.SearchCacheHits++
	o.mu.Unlock()
}

// CacheSearchMiss records a search cache miss
func (o *CAGOrchestrator) CacheSearchMiss() {
	o.mu.Lock()
	o.metrics.SearchCacheMisses++
	o.mu.Unlock()
}

// CacheTemplateHit records a template cache hit
func (o *CAGOrchestrator) CacheTemplateHit() {
	o.mu.Lock()
	o.metrics.TemplateCacheHits++
	o.metrics.LLMCallsAvoided++
	o.metrics.TotalCostSaved += 0.0025
	o.mu.Unlock()
}

// CacheTemplateMiss records a template cache miss
func (o *CAGOrchestrator) CacheTemplateMiss() {
	o.mu.Lock()
	o.metrics.TemplateCacheMisses++
	o.mu.Unlock()
}

// GetMetrics returns current cache metrics
func (o *CAGOrchestrator) GetMetrics() CAGMetrics {
	o.mu.RLock()
	defer o.mu.RUnlock()

	m := o.metrics

	// Aggregate from individual caches
	if o.ExactCache != nil {
		stats := o.ExactCache.Stats()
		if hits, ok := stats["hits"].(int64); ok {
			m.ExactMatchHits = hits
		}
		if misses, ok := stats["misses"].(int64); ok {
			m.ExactMatchMisses = misses
		}
	}

	if o.SemanticCache != nil {
		stats := o.SemanticCache.Stats()
		if hits, ok := stats["hits"].(int64); ok {
			m.SemanticHits = hits
		}
		if misses, ok := stats["misses"].(int64); ok {
			m.SemanticMisses = misses
		}
		if avgSim, ok := stats["avg_similarity"].(float64); ok {
			m.AvgSimilarity = avgSim
		}
	}

	if o.EmbeddingCache != nil {
		stats := o.EmbeddingCache.Stats()
		if hits, ok := stats["hits"].(int64); ok {
			m.EmbeddingCacheHits = hits
		}
		if misses, ok := stats["misses"].(int64); ok {
			m.EmbeddingCacheMisses = misses
		}
	}

	if o.TemplateCache != nil {
		stats := o.TemplateCache.Stats()
		if hits, ok := stats["hits"].(int64); ok {
			m.TemplateCacheHits = hits
		}
		if misses, ok := stats["misses"].(int64); ok {
			m.TemplateCacheMisses = misses
		}
	}

	// Calculate total cost saved from all caches
	m.TotalCostSaved = 0
	if o.ExactCache != nil {
		m.TotalCostSaved += float64(m.ExactMatchHits) * 0.0036
	}
	if o.SemanticCache != nil {
		m.TotalCostSaved += float64(m.SemanticHits) * 0.0036
	}
	if o.EmbeddingCache != nil {
		m.TotalCostSaved += float64(m.EmbeddingCacheHits) * 0.0001
	}
	if o.TemplateCache != nil {
		m.TotalCostSaved += float64(m.TemplateCacheHits) * 0.0025
	}

	m.LLMCallsAvoided = m.ExactMatchHits + m.SemanticHits + m.TemplateCacheHits
	m.EmbeddingCallsAvoided = m.EmbeddingCacheHits

	return m
}

// GetSummary returns a human-readable summary of cache performance
func (o *CAGOrchestrator) GetSummary() map[string]interface{} {
	metrics := o.GetMetrics()

	totalHits := metrics.ExactMatchHits + metrics.SemanticHits + metrics.EmbeddingCacheHits + metrics.TemplateCacheHits
	totalMisses := metrics.ExactMatchMisses + metrics.SemanticMisses + metrics.EmbeddingCacheMisses + metrics.TemplateCacheMisses
	total := totalHits + totalMisses
	overallHitRate := float64(0)
	if total > 0 {
		overallHitRate = float64(totalHits) / float64(total) * 100
	}

	return map[string]interface{}{
		"overall_hit_rate":    overallHitRate,
		"total_hits":          totalHits,
		"total_misses":        totalMisses,
		"total_cost_saved":    metrics.TotalCostSaved,
		"llm_calls_avoided":   metrics.LLMCallsAvoided,
		"embedding_calls_avoided": metrics.EmbeddingCallsAvoided,
		"layers": map[string]interface{}{
			"exact_match": map[string]interface{}{
				"hits":    metrics.ExactMatchHits,
				"misses":  metrics.ExactMatchMisses,
				"hit_rate": hitRate(metrics.ExactMatchHits, metrics.ExactMatchMisses),
			},
			"semantic": map[string]interface{}{
				"hits":          metrics.SemanticHits,
				"misses":        metrics.SemanticMisses,
				"hit_rate":      hitRate(metrics.SemanticHits, metrics.SemanticMisses),
				"avg_similarity": metrics.AvgSimilarity,
			},
			"embedding": map[string]interface{}{
				"hits":     metrics.EmbeddingCacheHits,
				"misses":   metrics.EmbeddingCacheMisses,
				"hit_rate": hitRate(metrics.EmbeddingCacheHits, metrics.EmbeddingCacheMisses),
			},
			"template": map[string]interface{}{
				"hits":     metrics.TemplateCacheHits,
				"misses":   metrics.TemplateCacheMisses,
				"hit_rate": hitRate(metrics.TemplateCacheHits, metrics.TemplateCacheMisses),
			},
		},
	}
}

// InvalidateAll clears all cache layers.
// Uses a detached background context to prevent hanging on cancelled contexts.
func (o *CAGOrchestrator) InvalidateAll(ctx context.Context) error {
	bgCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	errors := make(chan error, 4)

	if o.ExactCache != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := o.ExactCache.Clear(bgCtx); err != nil {
				errors <- err
			}
		}()
	}

	if o.SemanticCache != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := o.SemanticCache.Clear(bgCtx); err != nil {
				errors <- err
			}
		}()
	}

	if o.TemplateCache != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := o.TemplateCache.Clear(bgCtx); err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Return first error if any
	for err := range errors {
		log.Warn().Err(err).Msg("Error during cache invalidation")
		return err
	}

	log.Info().Msg("All cache layers invalidated")
	return nil
}

// BuildQueryHash creates a hash of query + filters for template caching
func BuildQueryHash(query string, filters map[string][]string) string {
	h := sha256.New()
	h.Write([]byte(query))

	// Sort filter keys for deterministic output
	keys := make([]string, 0, len(filters))
	for k := range filters {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		h.Write([]byte(k))
		for _, v := range filters[k] {
			h.Write([]byte(v))
		}
	}

	return hex.EncodeToString(h.Sum(nil))
}

// hitRate calculates hit rate percentage
func hitRate(hits, misses int64) float64 {
	total := hits + misses
	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total) * 100
}

// recordExactHit increments exact match hit counter
func (o *CAGOrchestrator) recordExactHit() {
	o.mu.Lock()
	o.metrics.ExactMatchHits++
	o.metrics.LLMCallsAvoided++
	o.mu.Unlock()
}

// recordExactMiss increments exact match miss counter
func (o *CAGOrchestrator) recordExactMiss() {
	o.mu.Lock()
	o.metrics.ExactMatchMisses++
	o.mu.Unlock()
}

// recordSemanticHit increments semantic cache hit counter
func (o *CAGOrchestrator) recordSemanticHit() {
	o.mu.Lock()
	o.metrics.SemanticHits++
	o.metrics.LLMCallsAvoided++
	o.mu.Unlock()
}

// recordSemanticMiss increments semantic cache miss counter
func (o *CAGOrchestrator) recordSemanticMiss() {
	o.mu.Lock()
	o.metrics.SemanticMisses++
	o.mu.Unlock()
}

// recordEmbeddingCheck increments embedding cache check counter
func (o *CAGOrchestrator) recordEmbeddingCheck() {
	// Just tracking that we checked the embedding cache
}
