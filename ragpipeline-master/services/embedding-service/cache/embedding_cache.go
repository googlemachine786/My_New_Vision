// Package cache provides Redis-backed embedding cache to avoid redundant Vertex AI calls.
package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	EmbeddingCachePrefix = "rag:embedding_cache:"
	EmbeddingCacheTTL    = 24 * time.Hour
	EmbeddingCacheMaxSize = 10000
)

// EmbeddingCache caches Vertex AI embedding results
type EmbeddingCache struct {
	redis   *redis.Client
	ttl     time.Duration
	maxSize int

	// Metrics
	mu      sync.RWMutex
	hits    int64
	misses  int64
}

// NewEmbeddingCache creates a new embedding cache.
// Defaults: TTL=24h, MaxSize=10000.
func NewEmbeddingCache(redisClient *redis.Client, opts ...CacheOption) *EmbeddingCache {
	options := cacheOptions{
		ttl:     EmbeddingCacheTTL,
		maxSize: EmbeddingCacheMaxSize,
	}
	for _, o := range opts {
		o.apply(&options)
	}

	return &EmbeddingCache{
		redis:   redisClient,
		ttl:     options.ttl,
		maxSize: options.maxSize,
	}
}

// CachedEmbedding represents a cached embedding result
type CachedEmbedding struct {
	Embedding  []float64 `json:"embedding"`
	Dimensions int       `json:"dimensions"`
	TaskType   string    `json:"task_type"`
	Text       string    `json:"text"`
	CreatedAt  time.Time `json:"created_at"`
}

// Get retrieves a cached embedding by text and task type
func (c *EmbeddingCache) Get(ctx context.Context, text, taskType string) (*CachedEmbedding, error) {
	key := c.buildKey(text, taskType)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		c.recordMiss()
		return nil, nil
	}
	if err != nil {
		c.recordMiss()
		log.Debug().Err(err).Str("cache_key", key).Msg("Embedding cache get error")
		return nil, nil
	}

	var cached CachedEmbedding
	if err := json.Unmarshal(data, &cached); err != nil {
		c.recordMiss()
		log.Warn().Err(err).Str("cache_key", key).Msg("Embedding cache unmarshal error")
		return nil, nil
	}

	c.recordHit()
	log.Debug().Str("cache_key", key).Msg("Embedding cache hit")
	return &cached, nil
}

// Set stores an embedding in the cache
func (c *EmbeddingCache) Set(ctx context.Context, text, taskType string, embedding *CachedEmbedding) error {
	key := c.buildKey(text, taskType)

	data, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}

	// Check cache size and evict if needed
	currentSize, err := c.redis.DBSize(ctx).Result()
	if err == nil && int(currentSize) > c.maxSize {
		c.evictOldest(ctx)
	}

	pipe := c.redis.Pipeline()
	pipe.Set(ctx, key, data, c.ttl)
	pipe.SAdd(ctx, "rag:embedding_cache:keys", key)
	_, err = pipe.Exec(ctx)

	if err != nil {
		log.Warn().Err(err).Str("cache_key", key).Msg("Embedding cache set error")
		return fmt.Errorf("failed to cache embedding: %w", err)
	}

	log.Debug().Str("cache_key", key).Int("dimensions", embedding.Dimensions).Msg("Embedding cached")
	return nil
}

// Invalidate removes a specific embedding from cache
func (c *EmbeddingCache) Invalidate(ctx context.Context, text, taskType string) error {
	key := c.buildKey(text, taskType)
	return c.redis.Del(ctx, key).Err()
}

// Clear removes all embeddings from cache
func (c *EmbeddingCache) Clear(ctx context.Context) error {
	keys, err := c.redis.SMembers(ctx, "rag:embedding_cache:keys").Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		if err := c.redis.Del(ctx, keys...).Err(); err != nil {
			return err
		}
	}

	c.redis.Del(ctx, "rag:embedding_cache:keys")
	log.Info().Msg("Embedding cache cleared")
	return nil
}

// Stats returns cache statistics
func (c *EmbeddingCache) Stats() map[string]interface{} {
	c.mu.RLock()
	hits := c.hits
	misses := c.misses
	c.mu.RUnlock()

	total := hits + misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}

	return map[string]interface{}{
		"hits":      hits,
		"misses":    misses,
		"hit_rate":  hitRate,
		"max_size":  c.maxSize,
		"ttl":       c.ttl.String(),
	}
}

// CostSaved calculates approximate cost saved by caching embeddings
func (c *EmbeddingCache) CostSaved() float64 {
	// Vertex AI embedding cost: ~$0.0001 per 1K tokens
	// Average text length: ~100 tokens
	c.mu.RLock()
	hits := c.hits
	c.mu.RUnlock()
	
	costPerEmbedding := 0.0001
	return float64(hits) * costPerEmbedding
}

// buildKey creates a cache key from text and task type
func (c *EmbeddingCache) buildKey(text, taskType string) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s|%s", text, taskType)))
	hash := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("%s%s", EmbeddingCachePrefix, hash)
}

// recordHit increments cache hit counter
func (c *EmbeddingCache) recordHit() {
	c.mu.Lock()
	c.hits++
	c.mu.Unlock()
}

// recordMiss increments cache miss counter
func (c *EmbeddingCache) recordMiss() {
	c.mu.Lock()
	c.misses++
	c.mu.Unlock()
}

// evictOldest removes the oldest cache entry
func (c *EmbeddingCache) evictOldest(ctx context.Context) {
	keys, err := c.redis.SMembers(ctx, "rag:embedding_cache:keys").Result()
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
		c.redis.SRem(ctx, "rag:embedding_cache:keys", interfaceKeys...)
		log.Debug().Int("evicted", toDelete).Msg("Embedding cache evicted oldest entries")
	}
}
