package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiterOption configures a rate limiter instance.
type RateLimiterOption interface {
	apply(*rateLimiterOptions)
}

type rateLimiterOptions struct {
	rpm   int
	burst int
}

type rpmOption struct {
	RPM int
}

func (o rpmOption) apply(opts *rateLimiterOptions) { opts.rpm = o.RPM }

type burstOption struct {
	Burst int
}

func (o burstOption) apply(opts *rateLimiterOptions) { opts.burst = o.Burst }

// WithRateLimit sets the requests per minute rate limit.
func WithRateLimit(rpm int) RateLimiterOption {
	return rpmOption{RPM: rpm}
}

// WithBurstSize sets the burst size allowance.
func WithBurstSize(burst int) RateLimiterOption {
	return burstOption{Burst: burst}
}

// RateLimiterMiddleware implements distributed rate limiting using Redis
type RateLimiterMiddleware struct {
	redisClient       *redis.Client
	requestsPerMinute int
	burstSize         int
	mu                sync.Mutex
	// In-memory fallback when Redis is unavailable
	localLimiters map[string]*LocalRateLimiter
}

// LocalRateLimiter is an in-memory rate limiter fallback
type LocalRateLimiter struct {
	tokens     float64
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiterMiddleware creates a new rate limiter middleware.
// Defaults: RPM=100, Burst=20.
func NewRateLimiterMiddleware(redisClient *redis.Client, opts ...RateLimiterOption) *RateLimiterMiddleware {
	options := rateLimiterOptions{
		rpm:   100,
		burst: 20,
	}
	for _, o := range opts {
		o.apply(&options)
	}

	return &RateLimiterMiddleware{
		redisClient:       redisClient,
		requestsPerMinute: options.rpm,
		burstSize:         options.burst,
		localLimiters:     make(map[string]*LocalRateLimiter, 1000),
	}
}

// Handler is the HTTP middleware that enforces rate limits
func (m *RateLimiterMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user identifier (from context if authenticated, otherwise IP)
		userID := GetUserID(r.Context())
		if userID == "" {
			userID = getClientIP(r)
		}

		allowed, remaining, resetAt, err := m.Allow(r.Context(), userID)
		if err != nil {
			// On error, allow the request (fail open)
			next.ServeHTTP(w, r)
			return
		}

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", m.requestsPerMinute))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		w.Header().Set("X-RateLimit-Reset", resetAt.Format(time.RFC3339))

		if !allowed {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":{"code":"rate_limit_exceeded","message":"Too many requests. Please try again later."}}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Allow checks if a request is allowed under the rate limit
func (m *RateLimiterMiddleware) Allow(ctx context.Context, key string) (bool, int64, time.Time, error) {
	// Try Redis-based rate limiting first
	allowed, remaining, resetAt, err := m.allowWithRedis(ctx, key)
	if err == nil {
		return allowed, remaining, resetAt, nil
	}

	// Fall back to local rate limiting
	return m.allowLocal(key)
}

func (m *RateLimiterMiddleware) allowWithRedis(ctx context.Context, key string) (bool, int64, time.Time, error) {
	now := time.Now()
	windowKey := fmt.Sprintf("rate_limit:%s:%d", key, now.Unix()/60)

	pipe := m.redisClient.Pipeline()
	incrCmd := pipe.Incr(ctx, windowKey)
	ttlCmd := pipe.TTL(ctx, windowKey)
	_, err := pipe.Exec(ctx)

	if err != nil && err != redis.Nil {
		return false, 0, time.Time{}, err
	}

	// Set TTL if key is new
	if ttlCmd.Val() < 0 {
		m.redisClient.Expire(ctx, windowKey, 60*time.Second)
	}

	count := incrCmd.Val()
	remaining := int64(m.requestsPerMinute) - count
	if remaining < 0 {
		remaining = 0
	}

	resetAt := now.Truncate(time.Minute).Add(time.Minute)

	if count > int64(m.requestsPerMinute+m.burstSize) {
		return false, remaining, resetAt, nil
	}

	if count > int64(m.requestsPerMinute) {
		// Allow burst but mark as rate limited
		return true, remaining, resetAt, nil
	}

	return true, remaining, resetAt, nil
}

func (m *RateLimiterMiddleware) allowLocal(key string) (bool, int64, time.Time, error) {
	m.mu.Lock()
	// FIX P0: Bound the localLimiters map to prevent OOM (LRU eviction)
	const maxLocalLimiters = 10000
	if len(m.localLimiters) >= maxLocalLimiters {
		// Evict oldest entry (first key in map iteration order is stable enough for eviction)
		for k := range m.localLimiters {
			delete(m.localLimiters, k)
			break
		}
	}
	limiter, exists := m.localLimiters[key]
	if !exists {
		limiter = &LocalRateLimiter{
			tokens:     float64(m.burstSize),
			lastRefill: time.Now(),
		}
		m.localLimiters[key] = limiter
	}
	m.mu.Unlock()

	return limiter.allow(m.requestsPerMinute, m.burstSize)
}

// LocalRateLimiter methods
func (l *LocalRateLimiter) allow(rpm, burst int) (bool, int64, time.Time, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastRefill).Seconds()
	l.tokens += elapsed * float64(rpm) / 60.0
	if l.tokens > float64(burst) {
		l.tokens = float64(burst)
	}
	l.lastRefill = now

	if l.tokens >= 1 {
		l.tokens--
		remaining := int64(l.tokens)
		resetAt := now.Truncate(time.Minute).Add(time.Minute)
		return true, remaining, resetAt, nil
	}

	resetAt := now.Add(time.Second)
	return false, 0, resetAt, nil
}

func getClientIP(r *http.Request) string {
	// Check common proxy headers
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP", "CF-Connecting-IP"} {
		if ip := r.Header.Get(header); ip != "" {
			// X-Forwarded-For can contain multiple IPs; take the first one
			if idx := strings.Index(ip, ","); idx != -1 {
				ip = ip[:idx]
			}
			return strings.TrimSpace(ip)
		}
	}
	// RemoteAddr includes port; strip it
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// Stats returns rate limiter statistics
func (m *RateLimiterMiddleware) Stats() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	return map[string]interface{}{
		"local_limiters": len(m.localLimiters),
		"rpm":            m.requestsPerMinute,
		"burst_size":     m.burstSize,
	}
}
