# Go Microservices Codebase Audit

**Scope:** `services/`, `pkg/`, `shared/go/`  
**Date:** 2026-04-06  
**Auditor:** Qwen Code (static analysis — Go not installed locally, no `go build`/`go vet` executed)

---

## Executive Summary

The RAG pipeline consists of **4 Go microservices**, a **shared `pkg/`** library, and a **`shared/go/`** utilities module. Overall code quality is good with structured logging, circuit breakers, and reasonable error handling. However, several critical gaps exist: **zero Dockerfiles**, **~60% of packages lack tests**, **import style inconsistencies**, **swallowed errors in handler code**, and **a compilation bug in vector-search-service main.go**.

---

## Priority Matrix

| Priority | Category | Count | Impact |
|----------|----------|-------|--------|
| P0 — Critical | Build failure, missing Dockerfiles | 5 | Cannot deploy or build |
| P1 — High | Missing test coverage, error swallowing | 30+ packages | Silent bugs, no regression safety |
| P2 — Medium | Import inconsistencies, missing go.sum | 6 modules | Confusing diffs, potential CI failures |
| P3 — Low | Documentation gaps, Makefile gaps | 10+ items | Developer experience |

---

## 1. Missing Tests (P1 — High)

### Packages with ZERO test coverage

| Service/Module | Package | Files | Risk |
|----------------|---------|-------|------|
| **api-gateway** | `cache/` | 5 files (`cache_options.go`, `cag_orchestrator.go`, `response_cache.go`, `semantic_cache.go`, `template_cache.go`) | Critical: CAG orchestration logic has no tests |
| **api-gateway** | `handler/` (partial) | `query_handler.go` (~550 lines), `stats_handler.go`, `feedback_handler.go`, `cag_stats_handler.go` | Critical: Core query pipeline untested |
| **api-gateway** | `client/` | `service_client.go` (~280 lines) | High: HTTP client with circuit breaker, retries, streaming |
| **api-gateway** | `session/` | `session_manager.go` | Medium: Session lifecycle untested |
| **api-gateway** | `middleware/` (partial) | `auth.go`, `rate_limit.go`, `recovery.go`, `request_id.go`, `cors.go` | Medium: Only `middleware_test.go` exists; unclear coverage |
| **embedding-service** | `cache/` | `cache_options.go`, `embedding_cache.go` | Medium: Cache logic untested |
| **embedding-service** | `handler/` | `embedding_handler.go` (~250 lines) | High: Embedding API untested |
| **embedding-service** | `vertex/` | `client.go` | High: Vertex AI client wrapper untested |
| **embedding-service** | `middleware/` | `cors.go` | Low |
| **embedding-service** | `config/` | `config.go` | Medium: Config validation untested |
| **embedding-service** | (root) | `main.go` (~320 lines) | Medium: No integration/smoke test |
| **vector-search-service** | `cache/` | `cache_options.go`, `search_cache.go` | Medium: Search cache untested |
| **vector-search-service** | `handler/` | `search_handler.go` (~400 lines) | High: Search endpoints untested |
| **vector-search-service** | `db/` | `postgres.go` (~450 lines) | High: All SQL query builders untested |
| **vector-search-service** | `search/` (partial) | `dense.go`, `sparse.go` | Medium: Only `errors_test.go` and `rrf_test.go` exist |
| **vector-search-service** | `config/` | `config.go` | Medium |
| **vector-search-service** | (root) | `main.go` (~290 lines) | Medium |
| **query-understanding-service** | `handler/` | `query_handler.go` (~230 lines) | High: Rewrite/parse/classify endpoints untested |
| **query-understanding-service** | `llm/` | `gemini_client.go` (~370 lines) | High: Gemini client with retry/circuit breaker untested |
| **query-understanding-service** | `rewriter/` | `query_rewriter.go` | Medium |
| **query-understanding-service** | `parser/` (partial) | `self_query.go` has tests, but coverage likely shallow | Low |
| **query-understanding-service** | `config/` | `config.go` | Medium |
| **query-understanding-service** | (root) | `main.go` (~130 lines) | Medium |
| **pkg/** | `httperrors/` | `errors.go` | Low: Simple helpers |
| **pkg/** | `types/` | `types.go` | Medium: Validation logic (`QueryRequest.Validate()`, `FeedbackRequest.Validate()`) untested |
| **shared/go/** | `config/` | `ab_test.go`, `rag_config.go` | Medium |
| **shared/go/** | `analytics/` | `evaluator.go` (has `evaluator_test.go`) | Low |
| **shared/go/** | `utils/` | `math.go` (has `math_test.go`) | Low |
| **shared/go/** | `errors/`, `middleware/`, `models/` | **Empty directories** | Info: Placeholder dirs with no code |

**Packages WITH tests (8 total):**
- `api-gateway/config/config_test.go`
- `api-gateway/handler/health_handler_test.go`
- `api-gateway/middleware/middleware_test.go`
- `query-understanding-service/classifier/intent_test.go`
- `query-understanding-service/parser/self_query_test.go`
- `query-understanding-service/sanitizer/sanitizer_test.go`
- `vector-search-service/search/errors_test.go`
- `vector-search-service/search/rrf_test.go`
- `pkg/circuitbreaker/circuitbreaker_test.go`
- `pkg/contextkeys/keys_test.go`

**Estimated overall test coverage: ~15-20%** (only small utility packages tested; core handlers, clients, and DB layers entirely untested)

---

## 2. Missing Dockerfiles (P0 — Critical)

**None of the 4 services has a Dockerfile.**

| Service | Expected Path | Status |
|---------|---------------|--------|
| api-gateway | `services/api-gateway/Dockerfile` | **MISSING** |
| embedding-service | `services/embedding-service/Dockerfile` | **MISSING** |
| vector-search-service | `services/vector-search-service/Dockerfile` | **MISSING** |
| query-understanding-service | `services/query-understanding-service/Dockerfile` | **MISSING** |

The Makefile references `docker-compose build` and `docker-compose up`, but `docker-compose.yml` likely expects Dockerfiles that don't exist. This is a **blocker for containerized deployment**.

---

## 3. Missing go.sum Files (P2 — Medium)

| Module | go.mod | go.sum | Status |
|--------|--------|--------|--------|
| `services/api-gateway/` | ✅ | ✅ | OK |
| `services/embedding-service/` | ✅ | ✅ | OK |
| `services/vector-search-service/` | ✅ | ✅ | OK |
| `services/query-understanding-service/` | ✅ | ✅ | OK |
| `pkg/` | ✅ | **MISSING** | `go.mod` exists with no deps, but `go.sum` is absent — should still have an empty go.sum for tooling consistency |
| `shared/go/` | ✅ | **MISSING** | Same as pkg/ — needs `go.sum` even if empty |

---

## 4. Import Inconsistencies (P2 — Medium)

### 4.1 Mixed import styles — query-understanding-service

**File:** `services/query-understanding-service/main.go` (line 13-19)  
**File:** `services/query-understanding-service/handler/query_handler.go` (line 12-16)  
**File:** `services/query-understanding-service/llm/gemini_client.go` (line 14)

Uses **relative/non-standard module path**:
```go
import (
    "query-understanding-service/classifier"
    "query-understanding-service/config"
    "query-understanding-service/handler"
    ...
)
```

**All other services** use the canonical module path:
```go
import (
    "github.com/visionary/ragpipeline/services/embedding-service/cache"
    "github.com/visionary/ragpipeline/services/vector-search-service/config"
)
```

The `go.mod` for this service declares `module query-understanding-service` (line 1), while all other modules use `github.com/visionary/ragpipeline/services/<name>`. This creates:
- Inconsistent import styles across the codebase
- Potential import confusion for developers
- Inability to import query-understanding-service types from other services

### 4.2 Import grouping inconsistency

**File:** `services/api-gateway/main.go`  
Standard library imports are mixed with third-party imports without clear grouping:
```go
import (
    "context"
    "net/http"
    "os"
    // ... stdlib
    "github.com/gorilla/mux"   // third party — no blank line separator
)
```

**File:** `services/embedding-service/main.go` has proper grouping with blank lines between stdlib, third-party, and internal imports.

**Recommendation:** Enforce `goimports` (already configured in `.golangci.yml`) or add a `gci` linter for deterministic import ordering.

### 4.3 String-keyed context values (anti-pattern)

**File:** `services/embedding-service/main.go` (line 240):
```go
ctx := context.WithValue(r.Context(), "request_id", requestID)
```
Uses a raw string key `"request_id"` instead of the typed `contextkeys.WithRequestID()` used elsewhere.

**File:** `services/vector-search-service/main.go` uses the correct `contextkeys.WithRequestID()`.  
**File:** `services/query-understanding-service/main.go` (line 113) also uses raw string `"request_id"`.

---

## 5. Error Handling Gaps (P1 — High)

### 5.1 Swallowed errors in query_handler.go

**File:** `services/api-gateway/handler/query_handler.go`

| Line area | Issue | Severity |
|-----------|-------|----------|
| `handleNonStreaming` — session history fetch | `if turns, err := h.SessionManager.GetHistory(ctx, req.SessionID); err == nil {` — error silently ignored | Medium |
| `handleNonStreaming` — session AddTurn | `_ = h.SessionManager.AddTurn(...)` — error discarded with blank identifier | Medium |
| `handleStreaming` — session history | Same pattern — `err == nil` guard swallows error | Medium |
| `handleStreaming` — session AddTurn | `_ = h.SessionManager.AddTurn(...)` — error discarded | Medium |
| `rewriteQuery` — client PostJSON | `log.Warn().Err(err).Msg(...)` then returns original query — acceptable but should log requestID | Low |
| `cacheResponse` — legacy cache Set | `_ = h.Cache.Set(...)` — cache write failure silently ignored | Low (fire-and-forget is intentional) |

### 5.2 Response body close after read

**File:** `services/api-gateway/client/service_client.go` — `PostJSONRaw` (line ~220)
```go
respBody, err := io.ReadAll(resp.Body)
```
`resp.Body` is deferred-close in caller, but if `io.ReadAll` succeeds and status code check fails, `respBody` is re-read in error message — correct but the defer is in the calling method, not here. **No bug, but worth noting the pattern is fragile.**

### 5.3 `w.Write()` return value ignored

**File:** `services/api-gateway/main.go` (line ~210):
```go
w.Write([]byte(`{"version":"` + Version + `",...}`))
```
The return value of `w.Write` (bytes written, error) is ignored. Same pattern in:
- `services/api-gateway/middleware/auth.go` — `writeAuthError`
- `services/embedding-service/main.go` — `versionHandler`
- `services/vector-search-service/main.go` — `versionHandler`

### 5.4 fmt.Errorf vs errors.New

Throughout the codebase, `fmt.Errorf("simple message")` is used where `errors.New("simple message")` would be more appropriate. Examples:
- `services/api-gateway/client/service_client.go`: `fmt.Errorf("circuit breaker open for %s", sc.name)` — correct (has format verb)
- `pkg/types/types.go`: `fmt.Errorf("query is required")` — should be `errors.New("query is required")`
- `services/query-understanding-service/llm/gemini_client.go`: `fmt.Errorf("model name is required")` — should be `errors.New`

This is a style concern (already flagged by `revive` in `.golangci.yml` via `errorf` rule).

### 5.5 Vector-search-service main.go — undefined variable (P0 — Compilation Bug)

**File:** `services/vector-search-service/main.go` (line ~195):
```go
log.Info().
    Str("method", r.Method).
    Str("path", r.URL.Path).
    Str("request_id", reqID).   // <-- BUG: `reqID` is undefined; should be `requestID`
    Int64("latency_ms", time.Since(start).Milliseconds()).
    Msg("Request completed")
```
This will **fail `go build`**. The variable is named `requestID` on line ~183 but referenced as `reqID` on line ~195.

### 5.6 Embedding-service middleware defined in main.go

**File:** `services/embedding-service/main.go` defines `requestIDMiddleware`, `loggingMiddleware`, `recoveryMiddleware`, and `corsMiddleware` as functions in `main` package, but **also** has a `middleware/` directory with `cors.go`. The main.go functions shadow or duplicate the middleware package. These middleware functions are never used from the `middleware/` package — they're all inline in `main.go`.

---

## 6. Missing Makefile Targets (P3 — Low)

**File:** `Makefile`

### Missing targets:
| Target | Purpose | Priority |
|--------|---------|----------|
| `go-mod-tidy` | Run `go mod tidy` in all services | Medium |
| `go-mod-verify` | Run `go mod verify` | Low |
| `build-%` | Build a single service (e.g., `make build-api-gateway`) | Medium |
| `test-%` | Test a single service | Medium |
| `test-pkg` | Test the shared `pkg/` module | Medium |
| `test-shared` | Test the `shared/go/` module | Low |
| `lint-fix` | Auto-fix lint issues (`golangci-lint run --fix`) | Low |
| `imports` | Run `goimports` | Low |
| `docker-build-%` | Build Docker image for a single service | Medium |
| `gen-mocks` | Generate mock interfaces for testing | Low |
| `tidy` | Shortcut for `go-mod-tidy` | Low |
| `check-all` | Run `test` + `lint` + `vet` in one command | Medium |

### Broken targets:
- **`build`** target references `./cmd/server/main.go` but **no service has a `cmd/server/` directory** — all `main.go` files are at the service root. This target will fail.
- **`run`** target similarly references `cmd/server/main.go` — will fail.

---

## 7. Missing .env.example Vars (P3 — Low)

### api-gateway `.env.example`
- Missing `LLM_SERVICE_URL` in the CAG configuration section (referenced in config)
- `CAG_ENABLED` is documented but not parsed in `config.Load()` — dead variable

### embedding-service `.env.example`
- Complete and well-documented. ✅

### vector-search-service `.env.example`
- Complete and well-documented. ✅

### query-understanding-service `.env.example`
- Missing `GEMINI_RATE_LIMIT` and `GEMINI_RATE_BURST` variables (used in config defaults but not documented in `.env.example`)

---

## 8. Documentation Gaps (P3 — Low)

| Path | Issue |
|------|-------|
| `services/api-gateway/README.md` | Exists ✅ |
| `services/embedding-service/README.md` | Exists ✅ |
| `services/vector-search-service/README.md` | Exists ✅ |
| `services/query-understanding-service/README.md` | Exists ✅ |
| `shared/go/README.md` | Exists ✅ |
| `pkg/` | **No README** — documents purpose of shared types |
| `pkg/httperrors/` | No package doc comment |
| `pkg/types/` | Has package doc comment ✅ |
| `pkg/circuitbreaker/` | Has package doc comment ✅ |
| `pkg/contextkeys/` | Has package doc comment ✅ |
| `services/query-understanding-service/CODE_REVIEW.md` | Exists (non-standard — typically in `.planning/` or similar) |
| `shared/go/errors/`, `shared/go/middleware/`, `shared/go/models/` | **Empty directories** — dead placeholders that should be removed |

---

## 9. Build Verification (P0 — Critical)

**Go compiler is not installed on this machine**, so `go build` could not be executed. However, static analysis reveals:

### Known compilation failures:
1. **`services/vector-search-service/main.go`** — `reqID` undefined (should be `requestID`) in `loggingMiddleware` — **will not compile**.
2. **`Makefile` `build` target** — references `./cmd/server/main.go` but entry points are at `services/*/main.go` — **make build will fail**.

### Likely compilation success (pending Go toolchain):
- `services/api-gateway/` — imports look correct, no obvious syntax errors
- `services/embedding-service/` — imports correct
- `services/query-understanding-service/` — imports use non-standard module path but match its `go.mod`

---

## 10. Dependency Freshness (P2 — Medium)

### Version inconsistencies across services:

| Dependency | api-gateway | embedding-service | query-understanding | vector-search |
|------------|-------------|-------------------|---------------------|---------------|
| `cenkalti/backoff/v4` | v4.3.0 | v4.3.0 | v4.3.0 | (not used) |
| `google/uuid` | v1.6.0 | v1.6.0 | (not used) | v1.6.0 |
| `gorilla/mux` | v1.8.1 | v1.8.1 | (not used) | v1.8.1 |
| `joho/godotenv` | v1.5.1 | v1.5.1 | v1.5.1 | v1.5.1 |
| `redis/go-redis/v9` | v9.5.0 | v9.5.0 | (not used) | v9.5.0 |
| `rs/zerolog` | v1.31.0 | v1.31.0 | (not used) | v1.31.0 |
| `otel` | v1.35.0 | v1.42.0 | v1.35.0 | (not used) |
| `otel/trace` | v1.35.0 | v1.42.0 | v1.35.0 | (not used) |
| `cloud.google.com/go/vertexai` | (not used) | v0.16.0 | v0.14.0 | (not used) |
| `golang.org/x/sys` | v0.33.0 | v0.42.0 | v0.33.0 | v0.33.0 |
| `google.golang.org/grpc` | (not used) | v1.79.3 | v1.72.2 | (not used) |
| `google.golang.org/protobuf` | (not used) | v1.36.11 | v1.36.6 | (not used) |

**OTel version drift:** embedding-service uses v1.42.0 while api-gateway and query-understanding use v1.35.0. This could cause subtle incompatibilities if shared OTel types cross service boundaries.

**Go version mismatch:**
- `embedding-service`: `go 1.25.0` (future version — Go 1.25 doesn't exist yet as of April 2026)
- `api-gateway`, `vector-search`, `query-understanding`: `go 1.23.0`
- `pkg/`, `shared/go/`: `go 1.24.5`

The `go 1.25.0` directive in embedding-service will fail with any current Go toolchain.

### `go mod tidy` status:
Cannot verify without Go toolchain. Recommend running:
```bash
for svc in services/*/; do (cd "$svc" && go mod tidy); done
(cd pkg && go mod tidy)
(cd shared/go && go mod tidy)
```

---

## Prioritized Improvement List

### P0 — Critical (fix immediately)

1. **Fix `reqID` → `requestID` typo** in `services/vector-search-service/main.go` line ~195  
   *Blocks compilation.*

2. **Create Dockerfiles for all 4 services**  
   `services/api-gateway/Dockerfile`, `services/embedding-service/Dockerfile`, `services/vector-search-service/Dockerfile`, `services/query-understanding-service/Dockerfile`  
   *Blocks containerized deployment.*

3. **Fix Makefile `build` and `run` targets** — change `./cmd/server/main.go` to `./main.go`  
   *`make build` and `make run` are broken.*

4. **Fix `go 1.25.0` in `services/embedding-service/go.mod`** — change to `go 1.24` or current stable  
   *Will fail any Go toolchain build.*

### P1 — High (fix within sprint)

5. **Add tests for `api-gateway/handler/query_handler.go`** (~550 lines, zero coverage)  
   *Core query pipeline — streaming, non-streaming, CAG orchestration, caching.*

6. **Add tests for `api-gateway/client/service_client.go`** (~280 lines)  
   *HTTP client with circuit breaker, retries, streaming — critical infrastructure.*

7. **Add tests for `api-gateway/cache/cag_orchestrator.go`**  
   *CAG multi-layer cache orchestration logic.*

8. **Add tests for `vector-search-service/db/postgres.go`** (~450 lines)  
   *SQL query builders for dense and sparse search — use testcontainers or pgx mocks.*

9. **Add tests for `vector-search-service/handler/search_handler.go`** (~400 lines)  
   *Search, dense, sparse, health, stats endpoints.*

10. **Add tests for `embedding-service/handler/embedding_handler.go`** (~250 lines)  
    *Embedding API, batch processing, caching.*

11. **Add tests for `query-understanding-service/llm/gemini_client.go`** (~370 lines)  
    *Gemini client with retry, rate limiter, circuit breaker.*

12. **Fix swallowed errors in `api-gateway/handler/query_handler.go`**  
    - Log errors for `SessionManager.GetHistory` failures instead of silent `err == nil` guard
    - Log or metric errors for `SessionManager.AddTurn` instead of `_ =`
    - Same pattern in both `handleStreaming` and `handleNonStreaming`

13. **Fix string-keyed context values** in `embedding-service/main.go` and `query-understanding-service/main.go`  
    - Use `contextkeys.WithRequestID()` instead of `context.WithValue(ctx, "request_id", ...)`

14. **Add tests for `pkg/types/`** — `QueryRequest.Validate()` and `FeedbackRequest.Validate()` are untested

### P2 — Medium (fix within release)

15. **Standardize module path for query-understanding-service**  
    Change `go.mod` from `module query-understanding-service` to `module github.com/visionary/ragpipeline/services/query-understanding-service` and update all imports.

16. **Add `go.sum` to `pkg/` and `shared/go/`**  
    Even if empty, for tooling consistency.

17. **Align OTel versions** across services (v1.35.0 vs v1.42.0)

18. **Align `golang.org/x/*` dependency versions** across services  
    Notably `golang.org/x/sys` (v0.33.0 vs v0.42.0), `golang.org/x/crypto`, `golang.org/x/net`.

19. **Enforce import grouping** via `goimports` or `gci` linter in CI

20. **Replace `fmt.Errorf("static message")` with `errors.New("static message")`**  
    Already flagged by `.golangci.yml` `errorf` rule — run `golangci-lint run --fix`.

21. **Run `go mod tidy` across all modules** and verify no unused deps

### P3 — Low (backlog)

22. **Add missing Makefile targets:** `go-mod-tidy`, `build-%`, `test-%`, `test-pkg`, `check-all`, `lint-fix`

23. **Remove empty directories:** `shared/go/errors/`, `shared/go/middleware/`, `shared/go/models/`

24. **Add `pkg/README.md`** documenting shared library purpose and usage

25. **Document missing `.env.example` vars** for query-understanding-service (`GEMINI_RATE_LIMIT`, `GEMINI_RATE_BURST`)

26. **Add package doc comment** to `pkg/httperrors/`

27. **Move `CODE_REVIEW.md`** to `.planning/` or integrate into PR review process

28. **Extract middleware from `embedding-service/main.go`** into `middleware/` package (currently duplicates `middleware/cors.go`)

29. **Add `w.Write()` error handling** in version handlers across services

30. **Add dead code analysis** — `CAG_ENABLED` in api-gateway `.env.example` but not parsed in config

---

## Scoring Summary

| Dimension | Score | Notes |
|-----------|-------|-------|
| Test Coverage | 3/10 | ~15-20% coverage; core handlers untested |
| Deployability | 2/10 | No Dockerfiles, broken Makefile targets |
| Dependency Hygiene | 5/10 | Version drift, invalid Go version in one module |
| Error Handling | 6/10 | Generally good; some swallowed errors, one compilation bug |
| Import Consistency | 6/10 | One service uses non-standard module path, string context keys |
| Documentation | 7/10 | Service READMEs exist; pkg/ undocumented |
| Build Readiness | 4/10 | Known compilation bug, broken Makefile, Go not installed locally |
| **Overall** | **4.7/10** | Functional but needs test coverage and deployment artifacts |
