# Critical Fixes - Top 10 Must-Fix Issues Before Production

**Project:** Visionary RAG Pipeline
**Audit Date:** April 1, 2026
**Production Readiness:** 34%
**Target:** 95%+ for production deployment

---

## Executive Summary

This document identifies the **10 most critical issues** that must be fixed before any production deployment. These issues represent existential risks to security, reliability, and business continuity.

**Do not deploy to production until all 10 issues are resolved.**

---

## Issue #1: No Authentication & Authorization

### Risk Level: CRITICAL 🔴

### Issue Description
The API has **no authentication middleware**. Anyone with the endpoint URL can:
- Query any student's data
- Access any grade level's content
- Exhaust API quotas
- Potentially inject malicious queries

### Current State
```go
// orchestrator/handler/rag_handler.go
func (h *RAGHandler) extractClaims(r *http.Request) (*JWTClaims, error) {
    // Implementation exists but NEVER CALLED
    // No middleware protects the /query endpoint
}

// orchestrator/middleware/rate_limiter.go
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
    // Only rate limits - does NOT authenticate
    // Uses IP address as fallback, not user identity
}
```

### Risk If Not Fixed
- **Data breach:** Unauthorized access to student data
- **Compliance violation:** GDPR/FERPA violations
- **Resource exhaustion:** API abuse without accountability
- **Grade contamination:** Grade 6 students accessing Grade 8 content

### Implementation Plan

**Step 1: Create JWT middleware** (2 days)
```go
// File: orchestrator/middleware/auth.go
package middleware

import (
    "context"
    "net/http"
    "strings"
    
    "github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
    UserID      string `json:"sub"`
    TaxonomyID  int    `json:"taxonomy_id"`
    Role        string `json:"role"`  // "student", "teacher", "admin"
    Grade       int    `json:"grade"`
    jwt.RegisteredClaims
}

type AuthMiddleware struct {
    secretKey []byte
}

func NewAuthMiddleware(secretKey string) *AuthMiddleware {
    return &AuthMiddleware{secretKey: []byte(secretKey)}
}

func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract token from Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, `{"error": {"code": "MISSING_AUTH", "message": "Authorization header required"}}`, http.StatusUnauthorized)
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        
        // Parse and validate token
        token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
            return m.secretKey, nil
        })

        if err != nil || !token.Valid {
            http.Error(w, `{"error": {"code": "INVALID_TOKEN", "message": "Invalid or expired token"}}`, http.StatusUnauthorized)
            return
        }

        // Extract claims and add to context
        claims, ok := token.Claims.(*JWTClaims)
        if !ok {
            http.Error(w, `{"error": {"code": "INVALID_CLAIMS", "message": "Invalid token claims"}}`, http.StatusUnauthorized)
            return
        }

        // Add claims to context
        ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
        ctx = context.WithValue(ctx, "taxonomy_id", claims.TaxonomyID)
        ctx = context.WithValue(ctx, "role", claims.Role)
        ctx = context.WithValue(ctx, "grade", claims.Grade)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

**Step 2: Apply middleware to router** (1 day)
```go
// File: orchestrator/cmd/server/main.go
func main() {
    // ... existing setup ...
    
    // Create auth middleware
    authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret)
    
    // Create router
    ragHandler := handler.NewRAGHandler(cfg)
    router := handler.NewRouter(ragHandler)
    
    // Apply middleware chain
    chained := securityMiddleware(
        authMiddleware.Middleware(
            rateLimiter.Middleware(router),
        ),
    )
    
    // Start server with chained middleware
    http.ListenAndServe(":8080", chained)
}
```

**Step 3: Enforce taxonomy_id in queries** (1 day)
```go
// File: orchestrator/handler/rag_handler.go
func (h *RAGHandler) Query(w http.ResponseWriter, r *http.Request) {
    // Extract taxonomy_id from context (set by auth middleware)
    taxonomyID, ok := r.Context().Value("taxonomy_id").(int)
    if !ok || taxonomyID == 0 {
        h.writeError(w, http.StatusForbidden, "MISSING_TAXONOMY", 
            fmt.Errorf("taxonomy_id not found in token"))
        return
    }

    // Use taxonomy_id in retrieval query
    results, err := h.retriever.Search(ctx, query, taxonomyID)
    // ...
}
```

**Step 4: Write tests** (1 day)
```go
// File: orchestrator/middleware/auth_test.go
func TestAuthMiddleware(t *testing.T) {
    tests := []struct {
        name           string
        authHeader     string
        expectedStatus int
    }{
        {"Missing header", "", http.StatusUnauthorized},
        {"Invalid token", "Bearer invalid", http.StatusUnauthorized},
        {"Valid token", "Bearer <valid_jwt>", http.StatusOK},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Verification
- [ ] All endpoints require valid JWT
- [ ] Invalid tokens return 401
- [ ] Taxonomy ID enforced in queries
- [ ] Tests pass with 100% coverage

---

## Issue #2: No Encryption at Rest or In Transit

### Risk Level: CRITICAL 🔴

### Issue Description
- **Database files are unencrypted** - anyone with disk access can read all data
- **TLS not enforced** - data in transit can be intercepted
- **PII stored in plaintext** - emails, names visible in database

### Current State
```hcl
# terraform/main.tf - No encryption configuration
resource "google_alloydb_cluster" "visionary" {
  cluster_id = "visionary-rag-cluster"
  # No encryption_config block
  # No TLS enforcement
}
```

```python
# orchestrator/compliance/gdpr.py - PII detected but not encrypted
class PIIDetector:
    def redact(self, text: str):
        # Detects PII but database still stores plaintext
        pass
```

### Risk If Not Fixed
- **Data breach:** Stolen database = all data exposed
- **Man-in-the-middle:** Interception of queries/responses
- **Compliance failure:** GDPR requires encryption
- **Insider threat:** DB admins can read all data

### Implementation Plan

**Step 1: Enable encryption at rest** (2 days)
```hcl
# File: terraform/main.tf
resource "google_alloydb_cluster" "visionary" {
  cluster_id = "visionary-rag-cluster"
  location   = var.region
  
  # Enable encryption
  encryption_config {
    kms_key_name = google_kms_crypto_key.alloydb_key.id
  }
  
  # ... rest of config
}

resource "google_kms_crypto_key" "alloydb_key" {
  name     = "alloydb-encryption-key"
  key_ring = google_kms_key_ring.key_ring.id
  
  rotation_period = "7776000s"  # 90 days
  
  version_template {
    algorithm = "GOOGLE_SYMMETRIC_ENCRYPTION"
  }
}
```

**Step 2: Enforce TLS** (1 day)
```hcl
# File: terraform/main.tf
resource "google_alloydb_instance" "primary" {
  # ... existing config
  
  database_flags = {
    "ssl" = "on"
    "require_secure_transport" = "on"
  }
}
```

**Step 3: Field-level encryption for PII** (3 days)
```python
# File: orchestrator/security/encryption.py
from cryptography.fernet import Fernet
import base64
import os

class FieldEncryptor:
    def __init__(self, key: bytes = None):
        if key is None:
            key = os.getenv('ENCRYPTION_KEY')
            if key is None:
                raise ValueError("ENCRYPTION_KEY required")
        self.cipher = Fernet(base64.urlsafe_b64encode(key))
    
    def encrypt(self, plaintext: str) -> str:
        """Encrypt PII field."""
        return self.cipher.encrypt(plaintext.encode()).decode()
    
    def decrypt(self, ciphertext: str) -> str:
        """Decrypt PII field."""
        return self.cipher.decrypt(ciphertext.encode()).decode()

# Usage in writer
async def insert_user_data(conn, user_id: str, email: str, encryptor: FieldEncryptor):
    encrypted_email = encryptor.encrypt(email)
    await conn.execute(
        "INSERT INTO users (user_id, email) VALUES ($1, $2)",
        user_id, encrypted_email
    )
```

**Step 4: TLS configuration for clients** (1 day)
```go
// File: orchestrator/db/alloydb.go
func NewAlloyDBPool(dsn string) (*pgxpool.Pool, error) {
    config, err := pgxpool.ParseConfig(dsn)
    if err != nil {
        return nil, err
    }
    
    // Require SSL
    config.ConnConfig.TLSConfig = &tls.Config{
        MinVersion: tls.VersionTLS13,
    }
    
    return pgxpool.NewWithConfig(context.Background(), config)
}
```

### Verification
- [ ] Database encrypted with KMS key
- [ ] TLS 1.3 enforced
- [ ] PII fields encrypted
- [ ] Connection requires SSL

---

## Issue #3: No Circuit Breakers or Retry Logic

### Risk Level: CRITICAL 🔴

### Issue Description
When Vertex AI, AlloyDB, or Redis experience transient failures:
- **Requests fail immediately** - no retry
- **Cascading failures** - one failure brings down entire system
- **No graceful degradation** - all-or-nothing behavior

### Current State
```go
// orchestrator/handler/rag_handler.go
func (h *RAGHandler) Query(w http.ResponseWriter, r *http.Request) {
    // Direct call - no retry on transient failure
    embedding, err := h.vertex.EmbedQuery(ctx, query)
    if err != nil {
        h.writeError(w, http.StatusInternalServerError, "EMBEDDING_FAILED", err)
        return  // Fails immediately
    }
    
    // Direct call - no circuit breaker
    results, err := h.retriever.Search(ctx, embedding)
    if err != nil {
        h.writeError(w, http.StatusInternalServerError, "RETRIEVAL_FAILED", err)
        return  // Fails immediately
    }
}
```

### Risk If Not Fixed
- **Cascading failures:** One slow dependency takes down entire system
- **Poor user experience:** Transient errors appear as permanent failures
- **Resource exhaustion:** Retries without backoff overwhelm services
- **No graceful degradation:** System doesn't adapt to partial outages

### Implementation Plan

**Step 1: Implement circuit breaker** (2 days)
```go
// File: orchestrator/reliability/circuit_breaker.go
package reliability

import (
    "sync"
    "time"
)

type CircuitState int

const (
    StateClosed CircuitState = iota
    StateOpen
    StateHalfOpen
)

type CircuitBreaker struct {
    mu           sync.RWMutex
    failures     int
    threshold    int
    timeout      time.Duration
    state        CircuitState
    lastFailure  time.Time
    successes    int
    halfOpenReqs int
}

func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        threshold: threshold,
        timeout:   timeout,
        state:     StateClosed,
    }
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    cb.mu.Lock()
    
    // Check if circuit is open
    if cb.state == StateOpen {
        if time.Since(cb.lastFailure) > cb.timeout {
            cb.state = StateHalfOpen
            cb.halfOpenReqs = 0
        } else {
            cb.mu.Unlock()
            return fmt.Errorf("circuit breaker open")
        }
    }
    
    if cb.state == StateHalfOpen {
        cb.halfOpenReqs++
    }
    
    cb.mu.Unlock()
    
    // Execute function
    err := fn()
    
    cb.mu.Lock()
    if err != nil {
        cb.failures++
        cb.lastFailure = time.Now()
        
        if cb.failures >= cb.threshold {
            cb.state = StateOpen
        }
    } else {
        cb.failures = 0
        cb.successes++
        
        if cb.state == StateHalfOpen {
            if cb.halfOpenReqs >= 3 {
                cb.state = StateClosed
            }
        }
    }
    cb.mu.Unlock()
    
    return err
}

func (cb *CircuitBreaker) State() CircuitState {
    cb.mu.RLock()
    defer cb.mu.RUnlock()
    return cb.state
}
```

**Step 2: Implement retry with backoff** (1 day)
```go
// File: orchestrator/reliability/retry.go
package reliability

import (
    "time"
    "math"
)

type RetryConfig struct {
    MaxRetries   int
    InitialDelay time.Duration
    MaxDelay     time.Duration
    Multiplier   float64
}

func DefaultRetryConfig() *RetryConfig {
    return &RetryConfig{
        MaxRetries:   3,
        InitialDelay: 100 * time.Millisecond,
        MaxDelay:     10 * time.Second,
        Multiplier:   2.0,
    }
}

func RetryWithBackoff(fn func() error, config *RetryConfig) error {
    if config == nil {
        config = DefaultRetryConfig()
    }
    
    var lastErr error
    delay := config.InitialDelay
    
    for attempt := 0; attempt <= config.MaxRetries; attempt++ {
        lastErr = fn()
        if lastErr == nil {
            return nil
        }
        
        if attempt < config.MaxRetries {
            time.Sleep(delay)
            delay = time.Duration(math.Min(
                float64(delay) * config.Multiplier,
                float64(config.MaxDelay),
            ))
        }
    }
    
    return lastErr
}
```

**Step 3: Apply to Vertex AI calls** (1 day)
```go
// File: orchestrator/handler/rag_handler.go
func (h *RAGHandler) Query(w http.ResponseWriter, r *http.Request) {
    var embedding []float32
    
    // Wrap with circuit breaker and retry
    cb := h.circuitBreakers["vertex"]
    err := cb.Execute(func() error {
        return RetryWithBackoff(func() error {
            var err error
            embedding, err = h.vertex.EmbedQuery(ctx, query)
            return err
        }, DefaultRetryConfig())
    })
    
    if err != nil {
        // Check if circuit is open - return graceful error
        if cb.State() == StateOpen {
            h.writeError(w, http.StatusServiceUnavailable, "SERVICE_DEGRADED", 
                fmt.Errorf("Vertex AI temporarily unavailable"))
            return
        }
        h.writeError(w, http.StatusInternalServerError, "EMBEDDING_FAILED", err)
        return
    }
    
    // Continue with retrieval...
}
```

**Step 4: Add graceful degradation** (1 day)
```go
// File: orchestrator/handler/rag_handler.go
func (h *RAGHandler) Query(w http.ResponseWriter, r *http.Request) {
    // Try primary (Vertex AI)
    embedding, err := h.tryVertexEmbedding(ctx, query)
    
    if err != nil {
        // Fallback to Ollama (local)
        embedding, err = h.tryOllamaEmbedding(ctx, query)
        if err != nil {
            // All embedders failed - return cached response if available
            cached, ok := h.cache.Get(query)
            if ok {
                h.writeResponse(w, cached)
                return
            }
            h.writeError(w, http.StatusServiceUnavailable, "ALL_EMBEDDERS_FAILED", err)
            return
        }
    }
    
    // Continue with retrieval...
}
```

### Verification
- [ ] Circuit breakers on all external calls
- [ ] Retry with exponential backoff
- [ ] Graceful degradation on failure
- [ ] Tests simulate failures

---

## Issue #4: No Distributed Tracing

### Risk Level: CRITICAL 🔴

### Issue Description
When a query is slow, there's **no way to determine which stage is the bottleneck**:
- Embedding generation?
- Database retrieval?
- RRF fusion?
- LLM generation?

### Current State
```go
// orchestrator/observability/ directory - COMPLETELY EMPTY
// No OpenTelemetry imports anywhere
// No trace context propagation
// No span creation
```

### Risk If Not Fixed
- **Blind debugging:** Cannot identify slow stages
- **SLA violations:** Cannot prove TTFT compliance
- **Performance regression:** Undetected degradation
- **Customer complaints:** Cannot diagnose issues

### Implementation Plan

**Step 1: Add OpenTelemetry SDK** (2 days)
```go
// File: orchestrator/observability/tracer.go
package observability

import (
    "context"
    "time"
    
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func InitTracer(serviceName string) (*sdktrace.TracerProvider, error) {
    ctx := context.Background()
    
    // Create exporter (Cloud Trace)
    exporter, err := otlptracehttp.New(ctx)
    if err != nil {
        return nil, err
    }
    
    // Create resource
    res, err := resource.New(ctx,
        resource.WithAttributes(
            semconv.ServiceName(serviceName),
        ),
    )
    if err != nil {
        return nil, err
    }
    
    // Create tracer provider
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(res),
        sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.1))), // 10% sampling
    )
    
    // Set global tracer
    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))
    
    return tp, nil
}
```

**Step 2: Instrument query handler** (2 days)
```go
// File: orchestrator/handler/rag_handler.go
import "go.opentelemetry.io/otel"

func (h *RAGHandler) Query(w http.ResponseWriter, r *http.Request) {
    ctx, span := otel.Tracer("visionary").Start(r.Context(), "rag.query")
    defer span.End()
    
    // Add attributes
    span.SetAttributes(
        attribute.String("query", truncate(query, 100)),
        attribute.String("session_id", sessionID),
    )
    
    // Instrument embedding stage
    var embedding []float32
    embedCtx, embedSpan := otel.Tracer("visionary").Start(ctx, "vertex.embed")
    err := RetryWithBackoff(func() error {
        var err error
        embedding, err = h.vertex.EmbedQuery(embedCtx, query)
        return err
    }, DefaultRetryConfig())
    embedSpan.End()
    
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "embedding failed")
        h.writeError(w, http.StatusInternalServerError, "EMBEDDING_FAILED", err)
        return
    }
    
    // Instrument retrieval stage
    retCtx, retSpan := otel.Tracer("visionary").Start(ctx, "alloydb.retrieve")
    results, err := h.retriever.Search(retCtx, embedding, taxonomyID)
    retSpan.End()
    
    // ... continue for all stages
}
```

**Step 3: Propagate trace context** (1 day)
```go
// File: orchestrator/middleware/tracing.go
func TracingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract trace context from headers
        ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
        
        // Add request ID for correlation
        requestID := generateRequestID()
        ctx = context.WithValue(ctx, "request_id", requestID)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

**Step 4: Add custom metrics** (1 day)
```go
// File: orchestrator/observability/metrics.go
import "go.opentelemetry.io/otel/metric"

var (
    queryLatencyMs metric.Float64Histogram
    retrievalCount metric.Int64Counter
)

func InitMetrics() error {
    meter := otel.Meter("visionary")
    
    var err error
    queryLatencyMs, err = meter.Float64Histogram(
        "rag.query.latency_ms",
        metric.WithDescription("Query latency in milliseconds"),
        metric.WithUnit("ms"),
    )
    if err != nil {
        return err
    }
    
    return nil
}

func RecordQueryLatency(latencyMs float64) {
    queryLatencyMs.Record(context.Background(), latencyMs)
}
```

### Verification
- [ ] All requests traced
- [ ] Each stage has spans
- [ ] Trace context propagated
- [ ] Traces visible in Cloud Trace

---

## Issue #5: No Input Validation or Sanitization

### Risk Level: CRITICAL 🔴

### Issue Description
User queries are **passed directly to the LLM and database** without validation:
- SQL injection possible through crafted queries
- XSS attacks through query text
- Prompt injection attacks
- Resource exhaustion via large payloads

### Current State
```go
// orchestrator/handler/rag_handler.go
type QueryRequest struct {
    Query     string `json:"query"`
    SessionID string `json:"session_id"`
}

func (h *RAGHandler) Query(w http.ResponseWriter, r *http.Request) {
    var req QueryRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // NO VALIDATION - query used directly
    embedding, _ := h.vertex.EmbedQuery(ctx, req.Query)
    results, _ := h.retriever.Search(ctx, embedding)
    
    // NO SANITIZATION - passed to LLM
    prompt := fmt.Sprintf("Context: %s\nQuestion: %s", results, req.Query)
}
```

### Risk If Not Fixed
- **SQL injection:** Malicious queries access unauthorized data
- **XSS attacks:** Script injection in responses
- **Prompt injection:** Manipulate LLM behavior
- **DoS attacks:** Large payloads exhaust resources

### Implementation Plan

**Step 1: Add input validation** (2 days)
```go
// File: orchestrator/validation/validator.go
package validation

import (
    "regexp"
    "unicode/utf8"
)

var (
    // Allow letters, numbers, spaces, and basic punctuation
    validQueryPattern = regexp.MustCompile(`^[\p{L}\p{N}\s.,!?;:'"()-]+$`)
    maxQueryLength    = 1000
    minQueryLength    = 3
)

type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return e.Field + ": " + e.Message
}

func ValidateQuery(query string) error {
    // Check length
    if utf8.RuneCountInString(query) < minQueryLength {
        return &ValidationError{"query", "query too short"}
    }
    if utf8.RuneCountInString(query) > maxQueryLength {
        return &ValidationError{"query", "query too long"}
    }
    
    // Check for valid characters
    if !validQueryPattern.MatchString(query) {
        return &ValidationError{"query", "query contains invalid characters"}
    }
    
    // Check for SQL injection patterns
    sqlPatterns := []string{
        `(?i)union\s+select`,
        `(?i)drop\s+table`,
        `(?i)delete\s+from`,
        `(?i)insert\s+into`,
        `(?i)--`,
        `(?i);\s*`,
    }
    for _, pattern := range sqlPatterns {
        if matched, _ := regexp.MatchString(pattern, query); matched {
            return &ValidationError{"query", "potential SQL injection detected"}
        }
    }
    
    return nil
}
```

**Step 2: Add request size limits** (1 day)
```go
// File: orchestrator/middleware/security.go
func SecurityMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Limit request body to 1MB
        r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
        
        // Validate content type
        contentType := r.Header.Get("Content-Type")
        if contentType != "application/json" {
            http.Error(w, `{"error": "Content-Type must be application/json"}`, http.StatusUnsupportedMediaType)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

**Step 3: Sanitize output** (1 day)
```go
// File: orchestrator/validation/sanitizer.go
package validation

import (
    "html"
    "strings"
)

func SanitizeOutput(text string) string {
    // Escape HTML to prevent XSS
    text = html.EscapeString(text)
    
    // Remove potential script tags
    text = strings.ReplaceAll(text, "<script>", "&lt;script&gt;")
    text = strings.ReplaceAll(text, "</script>", "&lt;/script&gt;")
    
    return text
}
```

**Step 4: Apply validation in handler** (1 day)
```go
// File: orchestrator/handler/rag_handler.go
func (h *RAGHandler) Query(w http.ResponseWriter, r *http.Request) {
    var req QueryRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.writeError(w, http.StatusBadRequest, "INVALID_JSON", err)
        return
    }
    
    // Validate input
    if err := validation.ValidateQuery(req.Query); err != nil {
        h.writeError(w, http.StatusBadRequest, "INVALID_QUERY", err)
        return
    }
    
    // Sanitize before using
    query := validation.SanitizeInput(req.Query)
    
    // Continue with sanitized query...
}
```

### Verification
- [ ] All inputs validated
- [ ] SQL injection blocked
- [ ] XSS prevented
- [ ] Request size limited

---

## Issue #6: No Health Check Depth or Readiness Probes

### Risk Level: HIGH 🟠

### Issue Description
The `/health` endpoint returns "healthy" even when:
- Database connections are exhausted
- Redis is unreachable
- Vertex AI is down

Traffic continues to be sent to unhealthy instances.

### Current State
```go
// orchestrator/handler/rag_handler.go
func (h *RAGHandler) HealthHandler(w http.ResponseWriter, r *http.Request) {
    status := HealthStatus{
        Status:    "healthy",
        Database:  "unknown",  // Never actually checked
        Redis:     "unknown",
        VertexAI:  "unknown",
    }
    
    // Checks exist but don't affect status properly
    if err := h.alloydb.HealthCheck(r.Context()); err != nil {
        status.Database = "unhealthy"
        // But status remains "healthy"
    }
}
```

### Risk If Not Fixed
- **Traffic to dead instances:** K8s sends traffic to unhealthy pods
- **Cascading failures:** Unhealthy instances cause more failures
- **Slow degradation:** System slowly dies without alerting
- **No startup gating:** Traffic before dependencies ready

### Implementation Plan

**Step 1: Implement deep health checks** (2 days)
```go
// File: orchestrator/handler/health.go
func (h *RAGHandler) DeepHealthCheck(ctx context.Context) error {
    // Check database with timeout
    dbCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()
    
    if err := h.alloydb.Ping(dbCtx); err != nil {
        return fmt.Errorf("database unhealthy: %w", err)
    }
    
    // Check Redis with timeout
    redisCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
    defer cancel()
    
    if err := h.redis.Ping(redisCtx).Err(); err != nil {
        return fmt.Errorf("redis unhealthy: %w", err)
    }
    
    // Check Vertex AI with timeout
    vertexCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()
    
    if err := h.vertex.HealthCheck(vertexCtx); err != nil {
        return fmt.Errorf("vertex AI unhealthy: %w", err)
    }
    
    return nil
}
```

**Step 2: Separate liveness and readiness** (1 day)
```go
// File: orchestrator/handler/health.go
func (h *RAGHandler) LivenessHandler(w http.ResponseWriter, r *http.Request) {
    // Liveness = is process running?
    // Always return healthy unless deadlocked
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
}

func (h *RAGHandler) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
    // Readiness = can we serve traffic?
    if err := h.DeepHealthCheck(r.Context()); err != nil {
        http.Error(w, `{"status": "not_ready", "error": "`+err.Error()+`"}`, http.StatusServiceUnavailable)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}
```

**Step 3: Update Terraform probes** (1 day)
```hcl
# File: terraform/main.tf
resource "google_cloud_run_v2_service" "orchestrator" {
  # ...
  
  template {
    containers {
      # Liveness probe - is process alive?
      liveness_probe {
        http_get {
          path = "/health/live"
          port = 8080
        }
        initial_delay_seconds = 5
        timeout_seconds      = 5
        period_seconds       = 10
        failure_threshold    = 3
      }
      
      # Readiness probe - can serve traffic?
      readiness_probe {
        http_get {
          path = "/health/ready"
          port = 8080
        }
        initial_delay_seconds = 3
        timeout_seconds      = 3
        period_seconds       = 5
        failure_threshold    = 2
      }
      
      # Startup probe - slow start handling
      startup_probe {
        http_get {
          path = "/health/ready"
          port = 8080
        }
        initial_delay_seconds = 0
        timeout_seconds      = 5
        period_seconds       = 10
        failure_threshold    = 30  # Allow 5 minutes to start
      }
    }
  }
}
```

### Verification
- [ ] Liveness endpoint exists
- [ ] Readiness checks all dependencies
- [ ] Startup probe configured
- [ ] Unhealthy instances removed from load balancer

---

## Issue #7: No Secrets Management (Hardcoded in Environment)

### Risk Level: HIGH 🟠

### Issue Description
Secrets are stored in **environment variables**:
- Visible in process listings (`ps aux`)
- Leaked in logs and error messages
- No rotation mechanism
- No access auditing

### Current State
```go
// orchestrator/config/config.go
type Config struct {
    JWTSecret     string `env:"JWT_SECRET"`      // Plain text in env
    AlloyDBDSN    string `env:"ALLOYDB_DSN"`     // Contains password
    RedisAddr     string `env:"REDIS_ADDR"`
}

// Loaded from environment - visible to anyone with process access
```

### Risk If Not Fixed
- **Secret leakage:** Process listings expose secrets
- **No rotation:** Compromised secrets valid forever
- **No auditing:** Cannot track secret access
- **Compliance failure:** SOC2 requires secret management

### Implementation Plan

**Step 1: Integrate GCP Secret Manager** (2 days)
```go
// File: orchestrator/secrets/manager.go
package secrets

import (
    "context"
    "fmt"
    
    secretmanager "cloud.google.com/go/secretmanager/apiv1"
    "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

type SecretManager struct {
    client *secretmanager.Client
    projectID string
}

func NewSecretManager(ctx context.Context, projectID string) (*SecretManager, error) {
    client, err := secretmanager.NewClient(ctx)
    if err != nil {
        return nil, err
    }
    
    return &SecretManager{
        client: client,
        projectID: projectID,
    }, nil
}

func (sm *SecretManager) GetSecret(ctx context.Context, secretID string) (string, error) {
    req := &secretmanagerpb.AccessSecretVersionRequest{
        Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", sm.projectID, secretID),
    }
    
    resp, err := sm.client.AccessSecretVersion(ctx, req)
    if err != nil {
        return "", err
    }
    
    return string(resp.Payload.Data), nil
}

func (sm *SecretManager) Close() error {
    return sm.client.Close()
}
```

**Step 2: Load config from Secret Manager** (1 day)
```go
// File: orchestrator/config/config.go
func LoadConfig(ctx context.Context) (*Config, error) {
    sm, err := secrets.NewSecretManager(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
    if err != nil {
        return nil, err
    }
    defer sm.Close()
    
    // Load secrets from Secret Manager
    jwtSecret, err := sm.GetSecret(ctx, "jwt-secret")
    if err != nil {
        return nil, fmt.Errorf("failed to load jwt-secret: %w", err)
    }
    
    alloyDBDSN, err := sm.GetSecret(ctx, "alloydb-dsn")
    if err != nil {
        return nil, fmt.Errorf("failed to load alloydb-dsn: %w", err)
    }
    
    return &Config{
        JWTSecret:   jwtSecret,
        AlloyDBDSN:  alloyDBDSN,
        RedisAddr:   os.Getenv("REDIS_ADDR"),  // Non-secret can stay in env
    }, nil
}
```

**Step 3: Configure secret rotation** (1 day)
```go
// File: orchestrator/secrets/rotation.go
func (sm *SecretManager) RotateSecret(ctx context.Context, secretID string) error {
    // Generate new secret value
    newSecret := generateSecureSecret()
    
    // Add new version
    req := &secretmanagerpb.AddSecretVersionRequest{
        Parent: fmt.Sprintf("projects/%s/secrets/%s", sm.projectID, secretID),
        Payload: &secretmanagerpb.SecretPayload{
            Data: []byte(newSecret),
        },
    }
    
    _, err := sm.client.AddSecretVersion(ctx, req)
    return err
}

// Schedule rotation every 90 days
func StartRotationScheduler(ctx context.Context, sm *SecretManager) {
    ticker := time.NewTicker(90 * 24 * time.Hour)
    go func() {
        for range ticker.C {
            sm.RotateSecret(ctx, "jwt-secret")
            sm.RotateSecret(ctx, "alloydb-dsn")
        }
    }()
}
```

### Verification
- [ ] All secrets in Secret Manager
- [ ] No secrets in environment
- [ ] Rotation configured
- [ ] Access logging enabled

---

## Issue #8: No Automated Testing (Low Coverage)

### Risk Level: HIGH 🟠

### Issue Description
Current test coverage is **~40%** with:
- No E2E tests
- No integration tests for critical paths
- No load testing
- No security testing

### Current State
```
tests/
├── integration_test.go (basic tests only)
└── (no E2E tests)

ingestion-go/tests/
├── parser_test.go
├── chunker_test.go
└── keywords_test.go

# Python tests: Minimal coverage
# Go tests: ~60% coverage
# Overall: ~40% coverage
```

### Risk If Not Fixed
- **Regression bugs:** Breaking changes undetected
- **Production incidents:** Bugs found by users
- **Slow development:** Manual testing required
- **Security vulnerabilities:** No security tests

### Implementation Plan

**Step 1: Add unit test coverage requirement** (1 week)
```yaml
# File: .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'
      
      - name: Run Go tests with coverage
        run: |
          go test ./... -race -coverprofile=coverage.out -covermode=atomic
          go tool cover -func=coverage.out
          
          # Fail if coverage < 80%
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
          if (( $(echo "$coverage < 80" | bc -l) )); then
            echo "Coverage $coverage% is below 80% threshold"
            exit 1
          fi
      
      - name: Set up Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.11'
      
      - name: Run Python tests with coverage
        run: |
          pip install -r requirements.txt
          pytest --cov=orchestrator --cov=ingestion --cov-report=xml --cov-fail-under=80
```

**Step 2: Add E2E tests** (1 week)
```go
// File: tests/e2e/query_test.go
package e2e

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestE2E_CompleteQueryFlow(t *testing.T) {
    // Setup test server
    server := httptest.NewServer(createApp())
    defer server.Close()
    
    // Create query request
    req := QueryRequest{
        Query:     "What is photosynthesis?",
        SessionID: "test-session",
    }
    
    body, _ := json.Marshal(req)
    resp, err := http.Post(server.URL+"/query", "application/json", bytes.NewReader(body))
    
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    // Parse response
    var result QueryResponse
    json.NewDecoder(resp.Body).Decode(&result)
    
    // Validate response structure
    assert.NotEmpty(t, result.Answer)
    assert.Greater(t, len(result.Sources), 0)
    assert.Greater(t, result.LatencyMs, 0.0)
}
```

**Step 3: Add load tests** (1 week)
```javascript
// File: eval/load_test.js (k6)
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
    stages: [
        { duration: '2m', target: 100 },   // Ramp to 100 users
        { duration: '5m', target: 100 },   // Stay at 100
        { duration: '2m', target: 500 },   // Ramp to 500
        { duration: '5m', target: 500 },   // Stay at 500
        { duration: '2m', target: 1000 },  // Ramp to 1000
        { duration: '10m', target: 1000 }, // Stay at 1000
        { duration: '2m', target: 0 },     // Ramp down
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'],  // 95% under 500ms
        errors: ['rate<0.01'],             // <1% errors
    },
};

export default function() {
    const payload = JSON.stringify({
        query: 'What is photosynthesis?',
        session_id: 'load-test-session',
    });
    
    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer test-token',
        },
    };
    
    const res = http.post('http://localhost:8080/query', payload, params);
    
    const success = check(res, {
        'status is 200': (r) => r.status === 200,
        'has answer': (r) => JSON.parse(r.body).answer !== '',
    });
    
    errorRate.add(!success);
    
    sleep(1);
}
```

### Verification
- [ ] 80%+ test coverage
- [ ] E2E tests pass
- [ ] Load tests pass (1000 users, p95<500ms)
- [ ] Security tests pass

---

## Issue #9: No API Documentation or Versioning

### Risk Level: HIGH 🟠

### Issue Description
The API has:
- **No OpenAPI specification**
- **No versioning** (`/v1/`, `/v2/`)
- **No documentation** for consumers
- **No deprecation policy**

### Current State
```
# No OpenAPI spec
# No API versioning in routes
# No documentation beyond README
```

### Risk If Not Fixed
- **Breaking changes:** Clients break without warning
- **Integration delays:** Consumers guess API behavior
- **Support burden:** Constant questions about API
- **Version lock-in:** Cannot evolve API

### Implementation Plan

**Step 1: Create OpenAPI specification** (2 days)
```yaml
# File: api/openapi.yaml
openapi: 3.0.3
info:
  title: Visionary RAG API
  description: Enterprise RAG API for CBSE Science education
  version: 1.0.0
  contact:
    email: api-support@visionary.edu

servers:
  - url: https://api.visionary.edu/v1
    description: Production server

paths:
  /query:
    post:
      summary: Ask a question and get an AI-generated answer
      operationId: postQuery
      tags:
        - RAG
      security:
        - BearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/QueryRequest'
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/QueryResponse'
        '400':
          description: Invalid request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        '401':
          description: Unauthorized
        '429':
          description: Rate limited

components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
  
  schemas:
    QueryRequest:
      type: object
      required:
        - query
      properties:
        query:
          type: string
          minLength: 3
          maxLength: 1000
          description: The question to ask
        session_id:
          type: string
          description: Session identifier for conversation history
    
    QueryResponse:
      type: object
      properties:
        answer:
          type: string
          description: AI-generated answer
        sources:
          type: array
          items:
            $ref: '#/components/schemas/Source'
        latency_ms:
          type: number
          description: Total latency in milliseconds
    
    Error:
      type: object
      properties:
        error:
          type: object
          properties:
            code:
              type: string
            message:
              type: string
```

**Step 2: Add API versioning** (1 day)
```go
// File: orchestrator/cmd/server/main.go
func main() {
    r := mux.NewRouter()
    
    // Version 1 API
    v1 := r.PathPrefix("/v1").Subrouter()
    v1.Use(middleware.AuthMiddleware)
    v1.HandleFunc("/query", handler.Query).Methods("POST")
    v1.HandleFunc("/health", handler.HealthHandler).Methods("GET")
    
    // Version 2 API (future)
    // v2 := r.PathPrefix("/v2").Subrouter()
    
    // Root redirects to latest version
    r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.Redirect(w, r, "/v1", http.StatusMovedPermanently)
    })
}
```

**Step 3: Generate documentation** (1 day)
```bash
# File: scripts/generate-docs.sh
#!/bin/bash

# Generate Swagger UI
docker run -d -p 8080:8080 -e SWAGGER_JSON=/api/openapi.yaml -v $(pwd)/api:/api swaggerapi/swagger-ui

# Generate SDKs
openapi-generator generate -i api/openapi.yaml -g go -o sdk/go
openapi-generator generate -i api/openapi.yaml -g python -o sdk/python
openapi-generator generate -i api/openapi.yaml -g typescript-axios -o sdk/typescript
```

### Verification
- [ ] OpenAPI spec complete
- [ ] API versioned (/v1/)
- [ ] Swagger UI accessible
- [ ] SDKs generated

---

## Issue #10: No Cost Tracking or Management

### Risk Level: HIGH 🟠

### Issue Description
There is **no visibility into costs**:
- Token usage untracked
- Cost per query unknown
- No budget alerts
- No cost optimization

### Current State
```go
// No token tracking
// No cost calculation
// No budget monitoring
```

### Risk If Not Fixed
- **Cost overruns:** Unexpected bills
- **No optimization:** Wasteful spending continues
- **No attribution:** Cannot charge back to users
- **Budget violations:** Exceed allocated budget

### Implementation Plan

**Step 1: Track token usage** (2 days)
```go
// File: orchestrator/observability/cost.go
package observability

import (
    "context"
    "sync/atomic"
)

type TokenTracker struct {
    promptTokens  int64
    completionTokens int64
}

func (t *TokenTracker) RecordTokens(ctx context.Context, prompt, completion int) {
    atomic.AddInt64(&t.promptTokens, int64(prompt))
    atomic.AddInt64(&t.completionTokens, int64(completion))
    
    // Record metric
    tokenUsage.Record(ctx, int64(prompt+completion),
        attribute.String("type", "llm"),
        attribute.String("model", "gemini-2.0-flash"),
    )
}

func (t *TokenTracker) GetTotalTokens() (prompt, completion int64) {
    return atomic.LoadInt64(&t.promptTokens), atomic.LoadInt64(&t.completionTokens)
}
```

**Step 2: Calculate cost per query** (1 day)
```go
// File: orchestrator/observability/cost.go
var geminiPricing = map[string]float64{
    "input":  0.00000025,  // $0.25 per million tokens
    "output": 0.00000075,  // $0.75 per million tokens
}

func CalculateQueryCost(promptTokens, completionTokens int) float64 {
    inputCost := float64(promptTokens) * geminiPricing["input"] / 1_000_000
    outputCost := float64(completionTokens) * geminiPricing["output"] / 1_000_000
    return inputCost + outputCost
}

// Add to response
type QueryResponse struct {
    Answer      string  `json:"answer"`
    Sources     []Source `json:"sources"`
    LatencyMs   float64 `json:"latency_ms"`
    CostUSD     float64 `json:"cost_usd,omitempty"`  // Add cost
}
```

**Step 3: Add budget alerts** (1 day)
```go
// File: orchestrator/observability/budget.go
type BudgetMonitor struct {
    dailyBudget   float64
    monthlyBudget float64
    currentSpend  float64
}

func (b *BudgetMonitor) CheckBudget(ctx context.Context, cost float64) error {
    b.currentSpend += cost
    
    // Check daily budget
    if b.currentSpend > b.dailyBudget {
        // Send alert
        sendAlert("Daily budget exceeded", b.currentSpend, b.dailyBudget)
        return fmt.Errorf("daily budget exceeded")
    }
    
    return nil
}

func sendAlert(message string, current, limit float64) {
    // Send to Slack/PagerDuty
    // Implementation depends on alerting system
}
```

**Step 4: Create cost dashboard** (1 day)
```go
// File: orchestrator/observability/dashboard.go
type CostDashboard struct {
    TodayCost      float64 `json:"today_cost"`
    MonthToDate    float64 `json:"month_to_date"`
    AvgCostPerQuery float64 `json:"avg_cost_per_query"`
    ProjectedMonthly float64 `json:"projected_monthly"`
    BudgetRemaining float64 `json:"budget_remaining"`
}

func GetCostDashboard() (*CostDashboard, error) {
    // Query metrics database
    // Return dashboard data
}
```

### Verification
- [ ] Token usage tracked
- [ ] Cost per query calculated
- [ ] Budget alerts configured
- [ ] Cost dashboard available

---

## Summary: Critical Fixes Timeline

| Week | Issues | Estimated Effort |
|------|--------|------------------|
| 1-2 | #1 Auth, #2 Encryption | 2 engineers × 2 weeks |
| 3-4 | #3 Circuit Breakers, #4 Tracing | 2 engineers × 2 weeks |
| 5-6 | #5 Input Validation, #6 Health Checks | 2 engineers × 2 weeks |
| 7-8 | #7 Secrets, #8 Testing | 2 engineers × 2 weeks |
| 9-10 | #9 API Docs, #10 Cost Tracking | 2 engineers × 2 weeks |

**Total:** 10 weeks with 2 engineers (or 5 weeks with 4 engineers)

**Do not deploy to production until all 10 issues are resolved.**
