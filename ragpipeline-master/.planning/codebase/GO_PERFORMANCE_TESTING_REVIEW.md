# Go Performance & Testing Review

**Date:** 2026-04-03
**Scope:** 71 Go files across 4 microservices + shared libs + orchestrator
**Focus:** Performance patterns, testing best practices, benchmarking, linting, race detection

---

## Critical Issues (Must Fix)

### C1. String concatenation in loops (performance degradation)

**Files affected:**

| File | Line | Pattern | Impact |
|------|------|---------|--------|
| `services/api-gateway/handler/query_handler.go` | 219 | `answerContent += src.Content + "\n"` | O(n²) in loop |
| `services/api-gateway/handler/query_handler.go` | 568 | `content += src.Content + "\n"` | O(n²) in loop |
| `orchestrator/handler/rag_handler.go` | 268 | `fullAnswer.WriteString(token)` (correct) BUT `buildContext` at line 264 uses `fmt.Sprintf` in loop | Suboptimal |
| `services/embedding-service/main.go` | 225-240 | Multiple `requestID.(string)` type assertions in logging middleware | Repeated unsafe casts |
| `services/vector-search-service/main.go` | 175-195 | Same repeated type assertion pattern | Repeated unsafe casts |

**Fix for query_handler.go lines 219, 568:**
```go
// BEFORE (O(n²)):
content := ""
for _, src := range sources {
    content += src.Content + "\n"
}

// AFTER (O(n)):
var sb strings.Builder
for _, src := range sources {
    sb.WriteString(src.Content)
    sb.WriteByte('\n')
}
content := sb.String()
```

### C2. Repeated `[]byte(string)` conversions in hot paths

**Files affected:**

| File | Line | Pattern |
|------|------|---------|
| `services/vector-search-service/cache/search_cache.go` | 175-195 | `h.Write([]byte(fmt.Sprintf(...)))` inside loop |
| `services/api-gateway/cache/cag_orchestrator.go` | 310-320 | `h.Write([]byte(...))` repeated in loop |
| `services/api-gateway/cache/template_cache.go` | 183-187 | `h.Write([]byte(hash))` in loop |
| `services/api-gateway/cache/response_cache.go` | 175 | `h.Write([]byte(fmt.Sprintf(...)))` |

**Fix for search_cache.go:**
```go
// BEFORE:
for _, v := range embedding {
    h.Write([]byte(fmt.Sprintf("%.6f", v)))
}

// AFTER: use strconv.AppendFloat to avoid allocation
buf := make([]byte, 0, 32)
for _, v := range embedding {
    buf = strconv.AppendFloat(buf[:0], v, 'f', 6, 64)
    h.Write(buf)
}
```

### C3. No `.golangci.yml` configuration — linting not enforced

**Status:** No `.golangci.yml` or `.golangci.yaml` found anywhere in the project.

**Impact:** No automated static analysis, meaning missed `errcheck` violations, unused imports, style issues, and potential security issues (gosec) go undetected.

**Fix:** Create `.golangci.yml`:
```yaml
linters:
  enable:
    - errcheck
    - govet
    - staticcheck
    - goimports
    - revive
    - gosec
    - ineffassign
    - misspell
    - unconvert

run:
  timeout: 5m

linters-settings:
  govet:
    check-shadowing: true
  revive:
    rules:
      - name: unused-parameter
      - name: blank-imports
      - name: context-as-argument
```

### C4. Integration test uses `testify` assertions instead of `cmp.Diff`

**File:** `tests/integration_test.go`

**Issue:** Imports `github.com/stretchr/testify/assert` and `github.com/stretchr/testify/require` — these are third-party assertion libraries. The Go testing philosophy prefers `cmp.Diff` from `github.com/google/go-cmp` for structural comparisons.

**Impact:** Inconsistent testing patterns across the codebase. The unit tests in `orchestrator/` use standard library patterns (good), but integration tests use testify.

---

## High Priority (Should Fix)

### H1. Missing capacity hints for known-size slices/maps

**Files affected:**

| File | Line | Issue |
|------|------|-------|
| `services/vector-search-service/search/rrf.go` | 62 | `entries := make(map[uuid.UUID]*rrfEntry)` — no hint, but size is `len(denseResults) + len(sparseResults)` |
| `services/api-gateway/handler/query_handler.go` | 157 | `filters := make(map[string][]string)` — should be `make(map[string][]string, 2)` (only grade + subject) |
| `services/vector-search-service/db/postgres.go` | 286 | `strings.Join(conditions, " AND ")` — `conditions` slice built without capacity hint |
| `shared/go/analytics/evaluator.go` | 126 | `recallAccum[k] = make([]float64, 0, len(qaPairs))` — **this is correct** ✅ |

**Fix for rrf.go:**
```go
// BEFORE:
entries := make(map[uuid.UUID]*rrfEntry)

// AFTER:
entries := make(map[uuid.UUID]*rrfEntry, len(denseResults)+len(sparseResults))
```

### H2. Bubble sort in production code (`shared/go/utils/math.go`)

**File:** `shared/go/utils/math.go`, lines 96-108

```go
// Current: Insertion sort (O(n²))
sorted := make([]float64, len(values))
copy(sorted, values)
for i := 1; i < len(sorted); i++ {
    key := sorted[i]
    j := i - 1
    for j >= 0 && sorted[j] > key {
        sorted[j+1] = sorted[j]
        j--
    }
    sorted[j+1] = key
}
```

**Fix:** Use `slices.Sort` from `cmp` package (Go 1.21+):
```go
import "slices"
sorted := make([]float64, len(values))
copy(sorted, values)
slices.Sort(sorted)
```

### H3. Duplicate sorting logic in multiple cache files

Three cache files implement the same bubble sort for filter key ordering:

| File | Lines |
|------|-------|
| `services/vector-search-service/cache/search_cache.go` | 183-190 |
| `services/api-gateway/cache/cag_orchestrator.go` | 313-319 |
| `services/api-gateway/cache/template_cache.go` | (uses `sort.Strings` via imported `sort` — **correct** ✅) |

**Fix:** Import `sort` and use `sort.Strings(keys)` consistently. The `search_cache.go` and `cag_orchestrator.go` implementations should be replaced.

### H4. Unsafe type assertions without ok-check in middleware

**Files affected:**

| File | Line | Pattern |
|------|------|---------|
| `services/vector-search-service/main.go` | 178 | `requestID.(string)` — panics if type wrong |
| `services/embedding-service/main.go` | 228 | `requestID.(string)` — same issue |

**Fix:**
```go
// BEFORE:
Str("request_id", requestID.(string))

// AFTER:
Str("request_id", requestID.(string))  // Safe because we always set it above
// OR use type assertion with ok-check if external code might call this
```

**Note:** In these specific cases the assertion IS safe because the middleware sets the value as a string just above. However, if the middleware chain changes, this will panic. Consider using a typed context key (see H6).

### H5. `fmt.Sscanf` used for integer parsing instead of `strconv.Atoi`

**File:** `services/vector-search-service/main.go`, line 253

```go
// BEFORE:
var v int
if _, err := fmt.Sscanf(value, "%d", &v); err == nil {
    return v
}

// AFTER:
if v, err := strconv.Atoi(value); err == nil {
    return v
}
```

**Impact:** `fmt.Sscanf` is significantly slower than `strconv.Atoi` (allocates a format parser). This is in an env-var helper so rarely called, but the pattern matters for code quality.

### H6. Untyped context keys risk collision

**Files affected:**

| File | Line |
|------|------|
| `services/vector-search-service/main.go` | 173 |
| `services/embedding-service/main.go` | 223 |
| `services/query-understanding-service/main.go` | 108 |
| `services/api-gateway/middleware/request_id.go` | (if exists) |

All use string `"request_id"` as context key. Any other package using the same string key would collide.

**Fix:**
```go
type contextKey string
const contextKeyRequestID contextKey = "request_id"

ctx := context.WithValue(r.Context(), contextKeyRequestID, requestID)
// Retrieval:
requestID, _ := ctx.Value(contextKeyRequestID).(string)
```

---

## Recommendations

### R1. Add benchmark functions

**Current state:** No `Benchmark*` functions found anywhere in the codebase.

**Suggested benchmarks:**

| Package | Function to Benchmark | Rationale |
|---------|----------------------|-----------|
| `shared/go/utils` | `CosineSimilarity` | Called on every query, 768-dim vectors |
| `shared/go/analytics` | `CalculateRecallAtK`, `CalculateNDCG` | Evaluation pipeline |
| `services/vector-search-service/search` | `ReciprocalRankFuse` | Core search algorithm |
| `pkg/circuitbreaker` | `AllowRequest`, `RecordFailure` | Called on every service call |
| `services/api-gateway/cache` | `buildKey` | Called on every cache operation |
| `orchestrator/llm/gemini` | Rate limiter `Wait()` | Controls throughput |

**Example:**
```go
func BenchmarkCosineSimilarity(b *testing.B) {
    a := make([]float64, 768)
    b_vec := make([]float64, 768)
    for i := range a {
        a[i] = float64(i) / 1000
        b_vec[i] = float64(i+1) / 1000
    }
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        utils.CosineSimilarity(a, b_vec)
    }
}
```

### R2. Add `goleak` for goroutine leak detection

**Current state:** No `TestMain` with `goleak.VerifyTestMain(m)` found.

The integration test `tests/integration_test.go` has a `TestMain` but only prints messages — no leak verification.

**Fix:**
```go
import "go.uber.org/goleak"

func TestMain(m *testing.M) {
    goleak.VerifyTestMain(m)
}
```

### R3. Run tests with `-race` flag

**Current state:** No evidence of race detector usage in Makefile, CI, or documentation.

**Command to add:**
```bash
go test -race ./...
```

**Particularly important for:**
- `pkg/circuitbreaker/circuitbreaker.go` — heavy mutex usage
- `services/api-gateway/cache/cag_orchestrator.go` — concurrent goroutines in `CacheResult()`
- All cache implementations — mutex-protected counters

### R4. Convert table-driven tests to use `cmp.Diff`

**Current test files using manual comparisons:**

| File | Current Pattern | Suggested |
|------|----------------|-----------|
| `orchestrator/prompts/rag_prompts_test.go` | `strings.Contains` checks | `cmp.Diff` for template output |
| `orchestrator/parser/rag_parser_test.go` | Manual field comparison | `cmp.Diff` with struct comparison |
| `orchestrator/llm/gemini/client_test.go` | Manual `got != want` | `cmp.Diff` |

**Example transformation:**
```go
// BEFORE:
if got.Answer != tt.wantAnswer {
    t.Errorf("ParseRAGResponse().Answer = %q, want %q", got.Answer, tt.wantAnswer)
}

// AFTER:
if diff := cmp.Diff(tt.want, got); diff != "" {
    t.Errorf("ParseRAGResponse() mismatch (-want +got):\n%s", diff)
}
```

### R5. Add test helpers with `t.Helper()`

**Current state:** Test helper functions in `tests/integration_test.go` (`createTestConfig`, `createMockEmbedding`) do NOT call `t.Helper()`.

**Fix:**
```go
func createTestConfig(t *testing.T) *config.Config {
    t.Helper()
    return &config.Config{...}
}
```

### R6. `context.WithValue` on `context.Background()` in main.go

**File:** `services/vector-search-service/main.go`, line 44

```go
ctx := context.Background()
```

This context is used for startup checks but then never propagated to request handlers. Each request creates its own context, which is correct. However, the `ctx` variable shadows any parent context and is used for Redis ping checks — this is fine for startup but worth noting.

### R7. `cacheResponse` in query_handler.go launches unbounded goroutines

**File:** `services/api-gateway/cache/cag_orchestrator.go`, lines 135-165

The `CacheResult` method spawns goroutines for each cache layer but has no semaphore or worker pool. Under high load, this could create many goroutines.

**Recommendation:** Add a `sync.WaitGroup` with a bounded semaphore or use a worker pool for cache writes.

### R8. Streaming: no `-benchmem` or profiling setup

The `Makefile` and project have no `go test -bench=. -benchmem` targets. Add to Makefile:
```makefile
.PHONY: bench
bench:
	go test -bench=. -benchmem -count=5 ./...

.PHONY: profile
profile:
	go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ./...
```

### R9. `InterpolateTemplate` uses manual string replacement

**File:** `services/api-gateway/cache/template_cache.go`, lines 170-178

```go
func replaceAll(s, old, new string) string {
    result := ""
    lastIdx := 0
    for i := 0; i <= len(s)-len(old); i++ {
        if s[i:i+len(old)] == old {
            result += s[lastIdx:i] + new  // O(n²) string concat!
            lastIdx = i + len(old)
        }
    }
    result += s[lastIdx:]
    return result
}
```

**Fix:** Use `strings.ReplaceAll` — it uses `strings.Builder` internally:
```go
func InterpolateTemplate(template string, variables map[string]string) string {
    result := template
    for key, value := range variables {
        placeholder := fmt.Sprintf("{%s}", key)
        result = strings.ReplaceAll(result, placeholder, value)
    }
    return result
}
```

### R10. `fmt.Fprintf` for JSON response in version handlers

**Files affected:**

| File | Line |
|------|------|
| `services/api-gateway/main.go` | 239 |
| `services/vector-search-service/main.go` | 130 |
| `services/embedding-service/main.go` | 173 |

```go
fmt.Fprintf(w, `{"version":"%s",...}`, Version, GitCommit, BuildTime)
```

**Impact:** If version strings contain special characters (like `"` or `\`), this produces invalid JSON. Use `json.Marshal` or `json.NewEncoder(w).Encode()`.

---

## Positive Notes

### P1. Excellent use of table-driven tests with subtests

**Files:**
- `orchestrator/prompts/rag_prompts_test.go` — 6 table-driven tests, all with subtests ✅
- `orchestrator/parser/rag_parser_test.go` — 5 table-driven tests ✅
- `orchestrator/llm/gemini/client_test.go` — 3 table-driven tests ✅

All follow the pattern:
```go
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) { ... })
}
```

### P2. Circuit breaker implements proper mutex patterns

**File:** `pkg/circuitbreaker/circuitbreaker.go`

- Uses `sync.Mutex` with `defer cb.mu.Unlock()` consistently ✅
- Functional options pattern for configuration ✅
- Compile-time interface check: `var _ ServiceCaller = (*ServiceClient)(nil)` ✅
- Clean state machine (Closed → Open → HalfOpen) ✅

### P3. Good error handling with wrapping

**Files:**
- `services/api-gateway/client/service_client.go` — uses `%w` for error wrapping ✅
- `services/api-gateway/handler/query_handler.go` — proper error propagation ✅
- `pkg/circuitbreaker/circuitbreaker.go` — sentinel errors via State type ✅

### P4. Capacity hints used correctly in analytics

**File:** `shared/go/analytics/evaluator.go`

```go
recallAccum[k] = make([]float64, 0, len(qaPairs))  // ✅
mrrAccum := make([]float64, 0, len(qaPairs))       // ✅
ndcgAccum := make([]float64, 0, len(qaPairs))      // ✅
```

### P5. Proper use of `strings.Builder` in some places

**File:** `orchestrator/handler/rag_handler.go`, line 264
```go
var fullAnswer strings.Builder  // ✅ Correct for streaming token accumulation
```

**File:** `services/query-understanding-service/sanitizer/sanitizer.go`, line 111
```go
var result strings.Builder  // ✅ Correct for character-by-character filtering
```

### P6. Graceful shutdown implemented everywhere

All four service `main.go` files implement proper graceful shutdown with:
- `signal.Notify` for SIGINT/SIGTERM ✅
- `context.WithTimeout` for shutdown deadline ✅
- `server.Shutdown(ctx)` for draining connections ✅

### P7. Redis pipelining for batch operations

**Files:**
- `services/api-gateway/cache/response_cache.go` — uses `redis.Pipeline()` ✅
- `services/vector-search-service/cache/search_cache.go` — uses `redis.Pipeline()` ✅

### P8. Performance tests in integration suite

**File:** `tests/integration_test.go`, lines 357-392

Includes `TestPerformance_Chunking` and `TestPerformance_KeywordExtraction` with timing assertions — good practice, though these should be converted to proper `Benchmark*` functions.

---

## Summary Statistics

| Category | Count | Severity |
|----------|-------|----------|
| Critical Issues | 4 | Must Fix |
| High Priority | 6 | Should Fix |
| Recommendations | 10 | Nice to Have |
| Positive Notes | 8 | Keep It Up |

**Test coverage gap:** 71 Go files, only 5 test files (~7% test-to-code ratio). Most service handlers, cache layers, and middleware have zero test coverage.

**Key action items (priority order):**
1. Create `.golangci.yml` and fix reported issues
2. Replace `+=` string concatenation in loops with `strings.Builder`
3. Replace `[]byte(fmt.Sprintf())` with `strconv.AppendFloat` in cache key builders
4. Add `sort.Strings()` to replace manual bubble sort in 2 files
5. Add benchmark functions for hot paths (cosine similarity, RRF fusion, circuit breaker)
6. Add `goleak` and run `go test -race ./...`
7. Expand test coverage for handlers, middleware, and cache layers
