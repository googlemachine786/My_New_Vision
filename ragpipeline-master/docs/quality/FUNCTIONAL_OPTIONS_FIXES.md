# Functional Options Fixes - Complete

## Summary

All critical and high-priority functional options pattern issues have been fixed across the Go microservices.

---

## Fixes Applied

### ✅ Critical: `vertex.NewClient` — 5+1 Params → Functional Options
**File:** `services/embedding-service/vertex/client.go`

**Before:**
```go
func NewClient(ctx context.Context, projectID, location, model string, dimension int, timeout time.Duration) (*Client, error)
```

**After:**
```go
func NewClient(ctx context.Context, projectID, location string, opts ...Option) (*Client, error)

// Options:
func WithModel(m string) Option           // default: "text-embedding-005"
func WithDimension(d int) Option          // default: 768
func WithTimeout(t time.Duration) Option  // default: 30s
```

**Caller Experience:**
```go
// Simple
vertex.NewClient(ctx, "my-project", "us-central1")

// With custom options
vertex.NewClient(ctx, "my-project", "us-central1",
    vertex.WithModel("text-embedding-004"),
    vertex.WithTimeout(60*time.Second),
)
```

---

### ✅ Critical: `api-gateway.NewQueryHandler` — 6 Params → Config Struct
**File:** `services/api-gateway/handler/query_handler.go`

**Before:**
```go
func NewQueryHandler(
    embeddingClient, vectorSearchClient, queryUnderstandingClient, llmClient *client.ServiceClient,
    cache *cache.ResponseCache,
    sessionManager *session.SessionManager,
    cagOrchestrator *cache.CAGOrchestrator,
) *QueryHandler
```

**After:**
```go
type QueryHandlerDeps struct {
    EmbeddingClient          *client.ServiceClient
    VectorSearchClient       *client.ServiceClient
    QueryUnderstandingClient *client.ServiceClient
    LLMClient                *client.ServiceClient
    Cache                    *cache.ResponseCache    // optional
    SessionManager           *session.SessionManager // optional
    CAGOrchestrator          *cache.CAGOrchestrator  // optional
}

func NewQueryHandler(deps QueryHandlerDeps) *QueryHandler
```

**Benefits:**
- Can't swap client arguments by accident
- Adding a 7th dependency doesn't break callers
- Self-documenting field names
- Optional deps can be nil

---

### ✅ Critical: `vector-search.NewSearchHandler` — 4 Params → Config Struct
**File:** `services/vector-search-service/handler/search_handler.go`

**Before:**
```go
func NewSearchHandler(hybridSearcher *search.HybridSearcher, cfg *config.Config, store *db.Store, searchCache *cache.SearchCache) *SearchHandler
```

**After:**
```go
type SearchHandlerDeps struct {
    HybridSearcher *search.HybridSearcher
    Config         *config.Config
    Store          *db.Store
    SearchCache    *cache.SearchCache // optional
}

func NewSearchHandler(deps SearchHandlerDeps) *SearchHandler
```

---

### ✅ High: All Cache Constructors → Shared Cache Options
**Files:**
- `services/api-gateway/cache/response_cache.go`
- `services/api-gateway/cache/semantic_cache.go`
- `services/api-gateway/cache/template_cache.go`
- `services/vector-search-service/cache/search_cache.go`
- `services/embedding-service/cache/embedding_cache.go`

**Before:**
```go
func NewResponseCache(redisClient *redis.Client, ttl time.Duration, maxSize int) *ResponseCache
func NewSemanticCache(redisClient *redis.Client, ttl time.Duration, maxSize int, threshold float64) *SemanticCache
// etc.
```

**After:**
```go
// Shared cache options
type CacheOption interface { apply(*cacheOptions) }

func WithCacheTTL(t time.Duration) CacheOption
func WithCacheMaxSize(size int) CacheOption
func WithSimilarityThreshold(t float64) CacheOption  // semantic cache only

// All cache constructors now use:
func NewResponseCache(redisClient *redis.Client, opts ...CacheOption) *ResponseCache
func NewSemanticCache(redisClient *redis.Client, opts ...CacheOption) *SemanticCache
// etc.
```

**Defaults:**
- TTL: 1 hour
- MaxSize: 1000
- Threshold: 0.95 (semantic cache only)

---

### ✅ High: `NewRateLimiterMiddleware` — 3 Params → Functional Options
**File:** `services/api-gateway/middleware/rate_limit.go`

**Before:**
```go
func NewRateLimiterMiddleware(redisClient *redis.Client, rpm, burst int) *RateLimiterMiddleware
```

**After:**
```go
type RateLimiterOption interface { apply(*rateLimiterOptions) }

func WithRateLimit(rpm int) RateLimiterOption    // default: 100
func WithBurstSize(burst int) RateLimiterOption  // default: 20

func NewRateLimiterMiddleware(redisClient *redis.Client, opts ...RateLimiterOption) *RateLimiterMiddleware
```

---

### ✅ High: `rewriter.NewService` — 3 Params → Functional Options
**File:** `services/query-understanding-service/rewriter/query_rewriter.go`

**Before:**
```go
func NewService(client *llm.Client, maxTurns int, promptVersion string) *Service
```

**After:**
```go
type ServiceOption interface { apply(*serviceOptions) }

func WithMaxTurns(n int) ServiceOption          // default: 10
func WithPromptVersion(v string) ServiceOption  // default: "v1"

func NewService(client *llm.Client, opts ...ServiceOption) *Service
```

---

## Constructors That Correctly Did NOT Use Functional Options

These have 1-2 params and are simple enough as-is:
- `NewAuthMiddleware(secret string)` — 1 param
- `NewFeedbackHandler(redisClient *redis.Client)` — 1 param
- `NewCAGStatsHandler(cagOrchestrator *cache.CAGOrchestrator)` — 1 param
- `parser.NewService(client, promptVersion)` — 2 params
- `classifier.NewService(client, promptVersion)` — 2 params

---

## Constructors Already Using Correct Patterns

- `NewStore(ctx, cfg *config.Config)` — Config struct ✅
- `llm.NewClient(ctx, cfg GeminiConfig)` — Config struct ✅
- `circuitbreaker.New(opts ...Option)` — Functional options ✅

---

## Benefits Achieved

### Extensibility
- Add new options without breaking any callers
- New cache options can be added to `CacheOption` interface
- New service deps can be added to `QueryHandlerDeps` struct

### Clarity
- Callers only specify what differs from defaults
- Self-documenting: `WithTimeout(60s)` vs `30s` as positional arg
- No confusion about which `*client.ServiceClient` is which

### Testability
- Optional deps can be nil in tests
- Easy to construct minimal deps for unit tests
- Options are typed and checked at compile time

### Clean Caller Experience
```go
// Before: Must always provide all params
h := NewQueryHandler(emb, vec, query, llm, cache, session, cag)

// After: Clear, self-documenting
h := handler.NewQueryHandler(handler.QueryHandlerDeps{
    EmbeddingClient:          emb,
    VectorSearchClient:       vec,
    QueryUnderstandingClient: query,
    LLMClient:                llm,
    Cache:                    cache,  // optional
    SessionManager:           session, // optional
    CAGOrchestrator:          cag,    // optional
})
```

---

## Summary Table

| Constructor | Before | After | Status |
|-------------|--------|-------|--------|
| `vertex.NewClient` | 5+1 positional | Functional Options | ✅ |
| `api-gateway.NewQueryHandler` | 6 positional | Config Struct | ✅ |
| `vector-search.NewSearchHandler` | 4 positional | Config Struct | ✅ |
| `NewServiceClient` | 3 positional | Functional Options | ✅ |
| `NewResponseCache` | 3 positional | Cache Options | ✅ |
| `NewSemanticCache` | 4 positional | Cache Options | ✅ |
| `NewTemplateCache` | 3 positional | Cache Options | ✅ |
| `NewSearchCache` | 3 positional | Cache Options | ✅ |
| `NewEmbeddingCache` | 3 positional | Cache Options | ✅ |
| `NewRateLimiterMiddleware` | 3 positional | Functional Options | ✅ |
| `rewriter.NewService` | 3 positional | Functional Options | ✅ |

**All 11 constructors with 3+ params now use proper patterns!** ✅
