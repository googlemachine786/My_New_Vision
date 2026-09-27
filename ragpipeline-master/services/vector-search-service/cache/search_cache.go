// Package cache provides Redis-backed search results cache to avoid redundant database queries.
package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	SearchCachePrefix  = "rag:search_cache:"
	SearchCacheTTL     = 30 * time.Minute
	SearchCacheMaxSize = 3000
)

// SearchCache caches vector search results
type SearchCache struct {
	redis   *redis.Client
	ttl     time.Duration
	maxSize int

	// Metrics
	mu      sync.RWMutex
	hits    int64
	misses  int64
}

// NewSearchCache creates a new search results cache.
// Defaults: TTL=30m, MaxSize=3000.
func NewSearchCache(redisClient *redis.Client, opts ...CacheOption) *SearchCache {
	options := cacheOptions{
		ttl:     SearchCacheTTL,
		maxSize: SearchCacheMaxSize,
	}
	for _, o := range opts {
		o.apply(&options)
	}

	return &SearchCache{
		redis:   redisClient,
		ttl:     options.ttl,
		maxSize: options.maxSize,
	}
}

// CachedSearchResult represents cached search results
type CachedSearchResult struct {
	ChunkIDs   []string             `json:"chunk_ids"`
	Scores     []float64            `json:"scores"`
	Chunks     []SearchResultChunk  `json:"chunks"`
	TopK       int                  `json:"top_k"`
	Filters    string               `json:"filters"`
	CreatedAt  time.Time            `json:"created_at"`
}

// SearchResultChunk represents a single chunk in cached search results
type SearchResultChunk struct {
	ParentID   string                 `json:"parent_id"`
	Content    string                 `json:"content"`
	Score      float64                `json:"score"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Get retrieves cached search results by embedding hash
func (c *SearchCache) Get(ctx context.Context, embedding []float64, topK int, filters map[string][]string) (*CachedSearchResult, error) {
	key := c.buildKey(embedding, topK, filters)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		c.recordMiss()
		return nil, nil
	}
	if err != nil {
		c.recordMiss()
		log.Debug().Err(err).Str("cache_key", key).Msg("Search cache get error")
		return nil, nil
	}

	var cached CachedSearchResult
	if err := json.Unmarshal(data, &cached); err != nil {
		c.recordMiss()
		log.Warn().Err(err).Str("cache_key", key).Msg("Search cache unmarshal error")
		return nil, nil
	}

	c.recordHit()
	log.Debug().Str("cache_key", key).Int("results", len(cached.ChunkIDs)).Msg("Search cache hit")
	return &cached, nil
}

// Set stores search results in the cache
func (c *SearchCache) Set(ctx context.Context, embedding []float64, topK int, filters map[string][]string, result *CachedSearchResult) error {
	key := c.buildKey(embedding, topK, filters)

	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal search result: %w", err)
	}

	// Check cache size and evict if needed
	currentSize, err := c.redis.DBSize(ctx).Result()
	if err == nil && int(currentSize) > c.maxSize {
		c.evictOldest(ctx)
	}

	pipe := c.redis.Pipeline()
	pipe.Set(ctx, key, data, c.ttl)
	pipe.SAdd(ctx, "rag:search_cache:keys", key)
	_, err = pipe.Exec(ctx)

	if err != nil {
		log.Warn().Err(err).Str("cache_key", key).Msg("Search cache set error")
		return fmt.Errorf("failed to cache search result: %w", err)
	}

	log.Debug().Str("cache_key", key).Int("results", len(result.ChunkIDs)).Msg("Search result cached")
	return nil
}

// Invalidate removes a specific search result from cache
func (c *SearchCache) Invalidate(ctx context.Context, embedding []float64, topK int, filters map[string][]string) error {
	key := c.buildKey(embedding, topK, filters)
	return c.redis.Del(ctx, key).Err()
}

// Clear removes all search results from cache
func (c *SearchCache) Clear(ctx context.Context) error {
	keys, err := c.redis.SMembers(ctx, "rag:search_cache:keys").Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		if err := c.redis.Del(ctx, keys...).Err(); err != nil {
			return err
		}
	}

	c.redis.Del(ctx, "rag:search_cache:keys")
	log.Info().Msg("Search cache cleared")
	return nil
}

// Stats returns cache statistics
func (c *SearchCache) Stats() map[string]interface{} {
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

// buildKey creates a cache key from embedding, top_k, and filters
func (c *SearchCache) buildKey(embedding []float64, topK int, filters map[string][]string) string {
	h := sha256.New()

	// Hash embedding using strconv.AppendFloat for efficiency
	buf := make([]byte, 0, 20) // capacity hint for formatted float
	for _, v := range embedding {
		buf = strconv.AppendFloat(buf[:0], v, 'f', 6, 64)
		h.Write(buf)
	}

	// Hash top_k
	h.Write([]byte(fmt.Sprintf("|%d", topK)))

	// Hash filters (sorted for consistency)
	if len(filters) > 0 {
		// Sort filter keys for deterministic output
		keys := make([]string, 0, len(filters))
		for k := range filters {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			h.Write([]byte(fmt.Sprintf("|%s=", k)))
			for _, v := range filters[k] {
				h.Write([]byte(v))
			}
		}
	}

	hash := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("%s%s", SearchCachePrefix, hash)
}

// recordHit increments cache hit counter
func (c *SearchCache) recordHit() {
	c.mu.Lock()
	c.hits++
	c.mu.Unlock()
}

// recordMiss increments cache miss counter
func (c *SearchCache) recordMiss() {
	c.mu.Lock()
	c.misses++
	c.mu.Unlock()
}

// evictOldest removes the oldest cache entry
func (c *SearchCache) evictOldest(ctx context.Context) {
	keys, err := c.redis.SMembers(ctx, "rag:search_cache:keys").Result()
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
		c.redis.SRem(ctx, "rag:search_cache:keys", interfaceKeys...)
		log.Debug().Int("evicted", toDelete).Msg("Search cache evicted oldest entries")
	}
}
