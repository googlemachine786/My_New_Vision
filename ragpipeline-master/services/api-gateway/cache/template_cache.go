// Package cache provides Redis-backed LLM template cache for common query patterns.
package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	TemplateCachePrefix = "rag:template_cache:"
	TemplateCacheTTL    = 12 * time.Hour
	TemplateCacheMaxSize = 500
)

// TemplateCache caches LLM response templates for common query intents
type TemplateCache struct {
	redis   *redis.Client
	ttl     time.Duration
	maxSize int

	// Metrics
	mu      sync.RWMutex
	hits    int64
	misses  int64
}

// NewTemplateCache creates a new template cache.
// Defaults: TTL=12h, MaxSize=500.
func NewTemplateCache(redisClient *redis.Client, opts ...CacheOption) *TemplateCache {
	options := cacheOptions{
		ttl:     TemplateCacheTTL,
		maxSize: TemplateCacheMaxSize,
	}
	for _, o := range opts {
		o.apply(&options)
	}

	return &TemplateCache{
		redis:   redisClient,
		ttl:     options.ttl,
		maxSize: options.maxSize,
	}
}

// CachedTemplateResponse represents a cached template response
type CachedTemplateResponse struct {
	Intent      string   `json:"intent"`
	QueryHash   string   `json:"query_hash"`
	Template    string   `json:"template"`
	Variables   map[string]string `json:"variables"`
	Response    string   `json:"response"`
	Sources     []Source `json:"sources"`
	ChunkHashes []string `json:"chunk_hashes"`
	CreatedAt   time.Time `json:"created_at"`
}

// Get retrieves a cached template response by intent, query, and chunks
func (c *TemplateCache) Get(ctx context.Context, intent, query string, chunkHashes []string) (*CachedTemplateResponse, error) {
	key := c.buildKey(intent, query, chunkHashes)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		c.recordMiss()
		return nil, nil
	}
	if err != nil {
		c.recordMiss()
		log.Debug().Err(err).Str("cache_key", key).Msg("Template cache get error")
		return nil, nil
	}

	var cached CachedTemplateResponse
	if err := json.Unmarshal(data, &cached); err != nil {
		c.recordMiss()
		log.Warn().Err(err).Str("cache_key", key).Msg("Template cache unmarshal error")
		return nil, nil
	}

	c.recordHit()
	log.Debug().Str("cache_key", key).Str("intent", cached.Intent).Msg("Template cache hit")
	return &cached, nil
}

// Set stores a template response in the cache
func (c *TemplateCache) Set(ctx context.Context, intent, query string, chunkHashes []string, response *CachedTemplateResponse) error {
	key := c.buildKey(intent, query, chunkHashes)

	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal template response: %w", err)
	}

	// Check cache size and evict if needed
	currentSize, err := c.redis.DBSize(ctx).Result()
	if err == nil && int(currentSize) > c.maxSize {
		c.evictOldest(ctx)
	}

	pipe := c.redis.Pipeline()
	pipe.Set(ctx, key, data, c.ttl)
	pipe.SAdd(ctx, "rag:template_cache:keys", key)
	_, err = pipe.Exec(ctx)

	if err != nil {
		log.Warn().Err(err).Str("cache_key", key).Msg("Template cache set error")
		return fmt.Errorf("failed to cache template response: %w", err)
	}

	log.Debug().Str("cache_key", key).Str("intent", response.Intent).Msg("Template response cached")
	return nil
}

// Invalidate removes a specific template from cache
func (c *TemplateCache) Invalidate(ctx context.Context, intent, query string, chunkHashes []string) error {
	key := c.buildKey(intent, query, chunkHashes)
	return c.redis.Del(ctx, key).Err()
}

// Clear removes all templates from cache
func (c *TemplateCache) Clear(ctx context.Context) error {
	keys, err := c.redis.SMembers(ctx, "rag:template_cache:keys").Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		if err := c.redis.Del(ctx, keys...).Err(); err != nil {
			return err
		}
	}

	c.redis.Del(ctx, "rag:template_cache:keys")
	log.Info().Msg("Template cache cleared")
	return nil
}

// Stats returns cache statistics
func (c *TemplateCache) Stats() map[string]interface{} {
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

// CostSaved calculates approximate cost saved by template caching
func (c *TemplateCache) CostSaved() float64 {
	// Each hit avoids an LLM call (~$0.0025)
	c.mu.RLock()
	hits := c.hits
	c.mu.RUnlock()

	costPerLLMCall := 0.0025
	return float64(hits) * costPerLLMCall
}

// InterpolateTemplate fills template variables with actual values
func InterpolateTemplate(template string, variables map[string]string) string {
	result := template
	for key, value := range variables {
		placeholder := fmt.Sprintf("{%s}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// buildKey creates a cache key from intent, query, and chunk hashes
func (c *TemplateCache) buildKey(intent, query string, chunkHashes []string) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s|%s", intent, query)))

	for _, hash := range chunkHashes {
		h.Write([]byte(hash))
	}

	hash := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("%s%s", TemplateCachePrefix, hash)
}

// recordHit increments cache hit counter
func (c *TemplateCache) recordHit() {
	c.mu.Lock()
	c.hits++
	c.mu.Unlock()
}

// recordMiss increments cache miss counter
func (c *TemplateCache) recordMiss() {
	c.mu.Lock()
	c.misses++
	c.mu.Unlock()
}

// evictOldest removes the oldest cache entry
func (c *TemplateCache) evictOldest(ctx context.Context) {
	keys, err := c.redis.SMembers(ctx, "rag:template_cache:keys").Result()
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
		c.redis.SRem(ctx, "rag:template_cache:keys", interfaceKeys...)
		log.Debug().Int("evicted", toDelete).Msg("Template cache evicted oldest entries")
	}
}
