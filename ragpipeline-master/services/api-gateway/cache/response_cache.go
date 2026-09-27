package cache

import (
	"container/list"
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
	CacheKeyPrefix = "rag:cache:"
	DefaultTTL     = time.Hour
	MaxEntries     = 1000
)

// ResponseCache provides Redis-backed response caching with LRU eviction
type ResponseCache struct {
	redis    *redis.Client
	ttl      time.Duration
	maxSize  int

	// Local LRU tracking for eviction using container/list for O(1) operations
	mu   sync.RWMutex
	lru  *list.List        // list of string keys, front=oldest, back=newest
	lruE map[string]*list.Element // map key -> list element

	// Metrics
	muMetrics sync.RWMutex
	hits      int64
	misses    int64
}

// NewResponseCache creates a new response cache.
// Defaults: TTL=1h, MaxSize=1000.
func NewResponseCache(redisClient *redis.Client, opts ...CacheOption) *ResponseCache {
	options := cacheOptions{
		ttl:     DefaultTTL,
		maxSize: MaxEntries,
	}
	for _, o := range opts {
		o.apply(&options)
	}

	return &ResponseCache{
		redis:   redisClient,
		ttl:     options.ttl,
		maxSize: options.maxSize,
		lru:     list.New(),
		lruE:    make(map[string]*list.Element, options.maxSize),
	}
}

// CachedResponse represents a cached response
type CachedResponse struct {
	Content   string    `json:"content"`
	Sources   []Source  `json:"sources"`
	CreatedAt time.Time `json:"created_at"`
}

// Source represents a retrieved chunk in cache
type Source struct {
	ParentID string  `json:"parent_id"`
	Content  string  `json:"content"`
	Score    float64 `json:"score"`
}

// Get retrieves a cached response by query parameters
func (c *ResponseCache) Get(ctx context.Context, query, sessionContext, grade, subject string) (*CachedResponse, error) {
	key := c.buildKey(query, sessionContext, grade, subject)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		c.recordMiss()
		return nil, nil
	}
	if err != nil {
		c.recordMiss()
		log.Debug().Err(err).Str("cache_key", key).Msg("Cache get error")
		return nil, nil
	}

	var cached CachedResponse
	if err := json.Unmarshal(data, &cached); err != nil {
		c.recordMiss()
		log.Warn().Err(err).Str("cache_key", key).Msg("Cache unmarshal error")
		return nil, nil
	}

	c.recordHit()
	c.touchLRU(key)

	log.Debug().Str("cache_key", key).Msg("Cache hit")
	return &cached, nil
}

// Set stores a response in the cache
func (c *ResponseCache) Set(ctx context.Context, query, sessionContext, grade, subject string, response *CachedResponse) error {
	key := c.buildKey(query, sessionContext, grade, subject)

	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal cached response: %w", err)
	}

	// Check if we need to evict
	c.mu.RLock()
	needsEviction := c.lru.Len() >= c.maxSize
	c.mu.RUnlock()

	if needsEviction {
		c.evictOldest(ctx)
	}

	pipe := c.redis.Pipeline()
	pipe.Set(ctx, key, data, c.ttl)
	pipe.SAdd(ctx, "rag:cache:keys", key)
	_, err = pipe.Exec(ctx)

	if err != nil {
		log.Warn().Err(err).Str("cache_key", key).Msg("Cache set error")
		return fmt.Errorf("failed to cache response: %w", err)
	}

	c.addLRU(key)
	log.Debug().Str("cache_key", key).Msg("Cache set")
	return nil
}

// Invalidate removes a specific cache entry
func (c *ResponseCache) Invalidate(ctx context.Context, query, sessionContext, grade, subject string) error {
	key := c.buildKey(query, sessionContext, grade, subject)
	return c.redis.Del(ctx, key).Err()
}

// Clear removes all cache entries
func (c *ResponseCache) Clear(ctx context.Context) error {
	keys, err := c.redis.SMembers(ctx, "rag:cache:keys").Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		if err := c.redis.Del(ctx, keys...).Err(); err != nil {
			return err
		}
	}

	c.redis.Del(ctx, "rag:cache:keys")

	c.mu.Lock()
	c.lru = list.New()
	c.lruE = make(map[string]*list.Element)
	c.mu.Unlock()

	return nil
}

// Stats returns cache statistics
func (c *ResponseCache) Stats() map[string]interface{} {
	c.muMetrics.RLock()
	hits := c.hits
	misses := c.misses
	c.muMetrics.RUnlock()

	total := hits + misses
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}

	c.mu.RLock()
	size := c.lru.Len()
	c.mu.RUnlock()

	return map[string]interface{}{
		"hits":      hits,
		"misses":    misses,
		"hit_rate":  hitRate,
		"size":      size,
		"max_size":  c.maxSize,
		"ttl":       c.ttl.String(),
	}
}

// buildKey creates a cache key from request parameters
func (c *ResponseCache) buildKey(query, sessionContext, grade, subject string) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s|%s|%s|%s", query, sessionContext, grade, subject)))
	hash := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("%s%s", CacheKeyPrefix, hash)
}

// recordHit increments cache hit counter
func (c *ResponseCache) recordHit() {
	c.muMetrics.Lock()
	c.hits++
	c.muMetrics.Unlock()
}

// recordMiss increments cache miss counter
func (c *ResponseCache) recordMiss() {
	c.muMetrics.Lock()
	c.misses++
	c.muMetrics.Unlock()
}

// LRU tracking methods using container/list for O(1) operations
func (c *ResponseCache) touchLRU(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.lruE[key]; exists {
		c.lru.MoveToBack(elem)
	}
}

func (c *ResponseCache) addLRU(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem := c.lru.PushBack(key)
	c.lruE[key] = elem
}

func (c *ResponseCache) evictOldest(ctx context.Context) {
	c.mu.Lock()
	if c.lru.Len() == 0 {
		c.mu.Unlock()
		return
	}

	// Remove oldest (front of list)
	elem := c.lru.Front()
	oldest := elem.Value.(string)
	c.lru.Remove(elem)
	delete(c.lruE, oldest)
	c.mu.Unlock()

	c.redis.Del(ctx, oldest)
	log.Debug().Str("evicted_key", oldest).Msg("Evicted cache entry")
}
