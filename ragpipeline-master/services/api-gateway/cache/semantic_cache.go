// Package cache provides Redis-backed semantic caching for query responses.
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"gonum.org/v1/gonum/mat"
)

const (
	SemanticCachePrefix      = "rag:semantic_cache:"
	SemanticCacheTTL         = 6 * time.Hour
	SemanticCacheMaxSize     = 5000
	DefaultSimilarityThreshold = 0.95
)

// SemanticCache caches query responses with embeddings for semantic similarity lookup
type SemanticCache struct {
	redis     *redis.Client
	ttl       time.Duration
	maxSize   int
	threshold float64

	// Metrics
	mu            sync.RWMutex
	hits          int64
	misses        int64
	totalSimilarity float64
	similarityCount  int64
}

// semanticCacheOptions extends cacheOptions with semantic-specific settings.
type semanticCacheOptions struct {
	cacheOptions
	threshold float64
}

// SimilarityThresholdOption configures the similarity threshold for semantic cache.
type SimilarityThresholdOption struct {
	Threshold float64
}

func (o SimilarityThresholdOption) apply(opts *cacheOptions) {
	if so, ok := interface{}(opts).(*semanticCacheOptions); ok {
		so.threshold = o.Threshold
	}
}

// WithSimilarityThreshold sets the minimum cosine similarity for cache hits.
func WithSimilarityThreshold(t float64) CacheOption {
	return SimilarityThresholdOption{Threshold: t}
}

// NewSemanticCache creates a new semantic cache.
// Defaults: TTL=6h, MaxSize=5000, Threshold=0.95.
func NewSemanticCache(redisClient *redis.Client, opts ...CacheOption) *SemanticCache {
	options := semanticCacheOptions{
		cacheOptions: cacheOptions{
			ttl:     SemanticCacheTTL,
			maxSize: SemanticCacheMaxSize,
		},
		threshold: DefaultSimilarityThreshold,
	}
	for _, o := range opts {
		o.apply(&options.cacheOptions)
	}
	// Apply threshold-specific option if set via SimilarityThresholdOption
	for _, o := range opts {
		if so, ok := o.(SimilarityThresholdOption); ok {
			if so.Threshold > 0 && so.Threshold <= 1 {
				options.threshold = so.Threshold
			}
		}
	}

	return &SemanticCache{
		redis:     redisClient,
		ttl:       options.ttl,
		maxSize:   options.maxSize,
		threshold: options.threshold,
	}
}

// CachedSemanticResponse represents a cached response with semantic embedding
type CachedSemanticResponse struct {
	QueryText      string    `json:"query_text"`
	QueryEmbedding []float64 `json:"query_embedding"`
	Response       string    `json:"response"`
	Sources        []Source  `json:"sources"`
	CreatedAt      time.Time `json:"created_at"`
}

// SearchSimilar finds a semantically similar cached response
func (c *SemanticCache) SearchSimilar(ctx context.Context, embedding []float64, threshold float64) (*CachedSemanticResponse, error) {
	if threshold == 0 {
		threshold = c.threshold
	}

	// Get all cached embeddings
	keys, err := c.redis.SMembers(ctx, SemanticCachePrefix+"keys").Result()
	if err != nil {
		log.Debug().Err(err).Msg("Semantic cache SMembers error")
		return nil, nil
	}

	if len(keys) == 0 {
		c.recordMiss()
		return nil, nil
	}

	// Use MGET pipeline to fetch all values in one round-trip
	pipe := c.redis.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}
	_, _ = pipe.Exec(ctx)

	var bestMatch *CachedSemanticResponse
	var bestSimilarity float64

	for i, cmd := range cmds {
		data, err := cmd.Bytes()
		if err != nil {
			continue
		}

		var cached CachedSemanticResponse
		if err := json.Unmarshal(data, &cached); err != nil {
			continue
		}

		// Calculate cosine similarity
		similarity := cosineSimilarity(embedding, cached.QueryEmbedding)

		if similarity > threshold && similarity > bestSimilarity {
			bestSimilarity = similarity
			bestMatch = &cached
		}

		_ = i // suppress unused warning
	}

	if bestMatch != nil {
		c.recordHit()
		c.recordSimilarity(bestSimilarity)
		log.Debug().
			Float64("similarity", bestSimilarity).
			Str("query_text", bestMatch.QueryText).
			Msg("Semantic cache hit")
		return bestMatch, nil
	}

	c.recordMiss()
	return nil, nil
}

// Add stores a response with its embedding for future semantic lookups
func (c *SemanticCache) Add(ctx context.Context, embedding []float64, response *CachedSemanticResponse) error {
	// Check if we need to evict
	c.mu.RLock()
	currentKeys, err := c.redis.SMembers(ctx, SemanticCachePrefix+"keys").Result()
	c.mu.RUnlock()

	if err == nil && len(currentKeys) >= c.maxSize {
		c.evictOldest(ctx)
	}

	key := fmt.Sprintf("%s%x", SemanticCachePrefix, time.Now().UnixNano())

	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal semantic response: %w", err)
	}

	pipe := c.redis.Pipeline()
	pipe.Set(ctx, key, data, c.ttl)
	pipe.SAdd(ctx, SemanticCachePrefix+"keys", key)
	_, err = pipe.Exec(ctx)

	if err != nil {
		log.Warn().Err(err).Str("cache_key", key).Msg("Semantic cache set error")
		return fmt.Errorf("failed to cache semantic response: %w", err)
	}

	log.Debug().
		Str("cache_key", key).
		Str("query_text", response.QueryText).
		Int("embedding_dim", len(embedding)).
		Msg("Semantic response cached")
	return nil
}

// Invalidate removes a specific entry from cache
func (c *SemanticCache) Invalidate(ctx context.Context, key string) error {
	pipe := c.redis.Pipeline()
	pipe.Del(ctx, key)
	pipe.SRem(ctx, SemanticCachePrefix+"keys", key)
	_, err := pipe.Exec(ctx)
	return err
}

// Clear removes all entries from cache
func (c *SemanticCache) Clear(ctx context.Context) error {
	keys, err := c.redis.SMembers(ctx, SemanticCachePrefix+"keys").Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		if err := c.redis.Del(ctx, keys...).Err(); err != nil {
			return err
		}
	}

	c.redis.Del(ctx, SemanticCachePrefix+"keys")
	log.Info().Msg("Semantic cache cleared")
	return nil
}

// Stats returns cache statistics
func (c *SemanticCache) Stats() map[string]interface{} {
	c.mu.RLock()
	hits := c.hits
	misses := c.misses
	avgSimilarity := float64(0)
	if c.similarityCount > 0 {
		avgSimilarity = c.totalSimilarity / float64(c.similarityCount)
	}
	c.mu.RUnlock()

	total := hits + misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}

	return map[string]interface{}{
		"hits":            hits,
		"misses":          misses,
		"hit_rate":        hitRate,
		"avg_similarity":  avgSimilarity,
		"max_size":        c.maxSize,
		"ttl":             c.ttl.String(),
		"threshold":       c.threshold,
	}
}

// CostSaved calculates approximate cost saved by semantic caching
func (c *SemanticCache) CostSaved() float64 {
	// Each hit avoids: embedding ($0.0001) + search ($0.001) + LLM ($0.0025)
	c.mu.RLock()
	hits := c.hits
	c.mu.RUnlock()

	costPerHit := 0.0036 // ~$0.0036 per avoided pipeline run
	return float64(hits) * costPerHit
}

// cosineSimilarity calculates cosine similarity between two vectors
// using gonum's optimized BLAS-level operations.
func cosineSimilarity(a, b []float64) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}

	va := mat.NewVecDense(len(a), a)
	vb := mat.NewVecDense(len(b), b)

	dot := mat.Dot(va, vb)
	normA := mat.Norm(va, 2)
	normB := mat.Norm(vb, 2)

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (normA * normB)
}

// recordHit increments cache hit counter
func (c *SemanticCache) recordHit() {
	c.mu.Lock()
	c.hits++
	c.mu.Unlock()
}

// recordMiss increments cache miss counter
func (c *SemanticCache) recordMiss() {
	c.mu.Lock()
	c.misses++
	c.mu.Unlock()
}

// recordSimilarity records a similarity score for averaging
func (c *SemanticCache) recordSimilarity(similarity float64) {
	c.mu.Lock()
	c.totalSimilarity += similarity
	c.similarityCount++
	c.mu.Unlock()
}

// evictOldest removes the oldest cache entry
func (c *SemanticCache) evictOldest(ctx context.Context) {
	keys, err := c.redis.SMembers(ctx, SemanticCachePrefix+"keys").Result()
	if err != nil || len(keys) == 0 {
		return
	}

	// Delete oldest 10% of cache
	toDelete := len(keys) / 10
	if toDelete == 0 {
		toDelete = 1
	}

	if toDelete < len(keys) {
		c.redis.Del(ctx, keys[:toDelete]...)
		// Convert []string to []interface{} for SRem
		interfaceKeys := make([]interface{}, len(keys[:toDelete]))
		for i, k := range keys[:toDelete] {
			interfaceKeys[i] = k
		}
		c.redis.SRem(ctx, SemanticCachePrefix+"keys", interfaceKeys...)
		log.Debug().Int("evicted", toDelete).Msg("Semantic cache evicted oldest entries")
	}
}
