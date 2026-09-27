# Enterprise Production Hardening Plan - v2.0

**Project:** Visionary RAG Pipeline
**Current Production Readiness:** 34%
**Target Production Readiness:** 95%+
**Timeline:** 20 weeks (team of 4-6 engineers)
**Priority:** P0 (Security) → P1 (Reliability) → P2 (Scalability)

---

## Executive Summary

The deep enterprise audit identified **127 gaps** across 8 categories. This plan prioritizes fixing **18 CRITICAL (P0)** and **34 HIGH (P1)** issues before any production deployment.

**DO NOT DEPLOY TO PRODUCTION** until Phase 0 (Security Foundation) and Phase 1 (Reliability Foundation) are complete.

---

## Phase 0: Security Foundation (Weeks 1-4) 🔴 CRITICAL

**Goal:** Address all P0 security vulnerabilities
**Team:** 2 backend engineers + 1 security engineer
**Success Criteria:** 85%+ security score, 0 critical vulnerabilities

### Week 1: Authentication & Authorization

#### Task 0.1.1: JWT Middleware Implementation
- **File:** `orchestrator/middleware/auth.go` (NEW)
- **Effort:** 2 days
- **Acceptance Criteria:**
  - All endpoints require valid JWT
  - Invalid tokens return 401 with error code
  - Claims extracted to context (user_id, taxonomy_id, role, grade)
  - Unit tests with 100% coverage

```go
// Implementation checklist:
// [ ] JWT parsing with golang-jwt/jwt/v5
// [ ] Bearer token extraction
// [ ] Signature validation
// [ ] Expiration check
// [ ] Claims extraction
// [ ] Context propagation
// [ ] Error handling (missing header, invalid token, expired)
// [ ] Unit tests
```

#### Task 0.1.2: Grade-Level Isolation
- **File:** `orchestrator/handler/rag_handler.go` (MODIFY)
- **Effort:** 1 day
- **Acceptance Criteria:**
  - Queries filtered by taxonomy_id from JWT
  - Grade 6 students cannot access Grade 8 content
  - Integration tests verify isolation

```sql
-- Query must include taxonomy_id filter
SELECT content, metadata FROM chunks
WHERE taxonomy_id = $1  -- From JWT claims
  AND metadata->>'grade' = $2  -- Enforce grade level
```

#### Task 0.1.3: OAuth2 Integration (Optional for Enterprise SSO)
- **File:** `orchestrator/middleware/oauth.go` (NEW)
- **Effort:** 2 days
- **Acceptance Criteria:**
  - Google Workspace SSO integration
  - Microsoft Entra ID integration
  - Session management with refresh tokens

### Week 2: Encryption & Data Security

#### Task 0.2.1: Encryption at Rest
- **File:** `terraform/main.tf` (MODIFY)
- **Effort:** 2 days
- **Acceptance Criteria:**
  - AlloyDB encryption with CMEK
  - KMS key rotation configured (90 days)
  - Terraform apply successful

```hcl
resource "google_alloydb_cluster" "visionary" {
  encryption_config {
    kms_key_name = google_kms_crypto_key.alloydb_key.id
  }
}
```

#### Task 0.2.2: TLS Enforcement
- **File:** `terraform/main.tf` (MODIFY)
- **Effort:** 1 day
- **Acceptance Criteria:**
  - `ssl = on` database flag
  - `require_secure_transport = on`
  - Connection fails without TLS

#### Task 0.2.3: Field-Level Encryption for PII
- **File:** `orchestrator/security/encryption.py` (NEW)
- **Effort:** 2 days
- **Acceptance Criteria:**
  - PII fields encrypted before insert
  - Decryption on read authorized
  - Key management via GCP Secret Manager

```python
class FieldEncryptor:
    def encrypt_pii(self, plaintext: str) -> str:
        # Fernet encryption
    def decrypt_pii(self, ciphertext: str) -> str:
        # Decryption with auth check
```

### Week 3: API Security

#### Task 0.3.1: Input Validation & Sanitization
- **File:** `orchestrator/middleware/security.go` (NEW)
- **Effort:** 2 days
- **Acceptance Criteria:**
  - SQL injection prevention (parameterized queries only)
  - XSS prevention (HTML entity encoding)
  - Request size limits (10MB max)
  - Content-Type validation

#### Task 0.3.2: Security Headers
- **File:** `orchestrator/middleware/security.go` (MODIFY)
- **Effort:** 1 day
- **Acceptance Criteria:**
  - X-Content-Type-Options: nosniff
  - X-Frame-Options: DENY
  - Strict-Transport-Security: max-age=31536000
  - Content-Security-Policy: default-src 'self'

#### Task 0.3.3: Rate Limiting Enhancement
- **File:** `orchestrator/middleware/rate_limiter.go` (MODIFY)
- **Effort:** 2 days
- **Acceptance Criteria:**
  - Per-user rate limits (not just IP)
  - Tiered limits (free vs premium)
  - Rate limit headers (X-RateLimit-*)
  - Graceful degradation when limit exceeded

### Week 4: Audit Logging & Secrets

#### Task 0.4.1: Audit Logging Implementation
- **File:** `orchestrator/observability/audit_logger.go` (NEW)
- **Effort:** 3 days
- **Acceptance Criteria:**
  - All data access logged (who, what, when)
  - Log integrity (HMAC signing)
  - Tamper detection
  - 7-year retention policy

```go
type AuditLogger struct {
    // Logs every query with user context
}

func (al *AuditLogger) LogQuery(userID, query, resultCount string) {
    // Write to immutable audit log table
}
```

#### Task 0.4.2: Secrets Management Migration
- **File:** `orchestrator/config/secrets.go` (NEW)
- **Effort:** 2 days
- **Acceptance Criteria:**
  - Migrate from env vars to GCP Secret Manager
  - Secret versioning support
  - Automatic secret rotation
  - Access logging

---

## Phase 1: Reliability Foundation (Weeks 5-8) 🔴 CRITICAL

**Goal:** Build resilient, fault-tolerant system
**Team:** 3 backend engineers
**Success Criteria:** 85%+ reliability score, circuit breakers on all external calls

### Week 5: Error Handling & Retry Logic

#### Task 1.1.1: Retry with Exponential Backoff
- **File:** `orchestrator/reliability/retry.go` (NEW)
- **Effort:** 2 days
- **Acceptance Criteria:**
  - Retry transient failures (5xx, timeout)
  - Exponential backoff (100ms, 200ms, 400ms, 800ms, 1600ms)
  - Max 5 retries
  - Jitter to prevent thundering herd

```go
func RetryWithBackoff(fn func() error, maxRetries int) error {
    // Implement exponential backoff with jitter
}
```

#### Task 1.1.2: Error Classification
- **File:** `orchestrator/errors/errors.go` (NEW)
- **Effort:** 1 day
- **Acceptance Criteria:**
  - Error taxonomy (Transient, Permanent, Expected)
  - Proper HTTP status mapping
  - User-friendly error messages

### Week 6: Circuit Breakers

#### Task 1.2.1: Circuit Breaker Implementation
- **File:** `orchestrator/reliability/circuit_breaker.go` (NEW)
- **Effort:** 3 days
- **Acceptance Criteria:**
  - Circuit breaker pattern (closed, open, half-open)
  - Configurable threshold (5 failures)
  - Configurable timeout (30 seconds)
  - Metrics exported (circuit_state, failures_count)

```go
type CircuitBreaker struct {
    state     string  // closed, open, half-open
    failures  int
    threshold int
    timeout   time.Duration
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    // Implement circuit breaker logic
}
```

#### Task 1.2.2: Circuit Breaker Integration
- **Files:** All external call sites (MODIFY)
- **Effort:** 2 days
- **Acceptance Criteria:**
  - Vertex AI calls protected
  - Database calls protected
  - Ollama calls protected
  - Tests verify circuit opens on failures

### Week 7: Health Checks & Monitoring

#### Task 1.3.1: Deep Health Checks
- **File:** `orchestrator/handler/health_handler.go` (NEW)
- **Effort:** 2 days
- **Acceptance Criteria:**
  - Liveness probe (is process running?)
  - Readiness probe (are dependencies healthy?)
  - Startup probe (is initialization complete?)
  - Dependency health (DB, Redis, Vertex AI)

```go
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
    // Check DB connection
    // Check Redis connection
    // Check Vertex AI connectivity
    // Return 200 only if ALL healthy
}
```

#### Task 1.3.2: Custom Metrics
- **File:** `orchestrator/observability/metrics.go` (NEW)
- **Effort:** 3 days
- **Acceptance Criteria:**
  - Request latency histogram
  - Error rate counter
  - Cache hit rate gauge
  - Token usage counter
  - Cloud Monitoring integration

### Week 8: Disaster Recovery

#### Task 1.4.1: Automated Backups
- **File:** `orchestrator/disaster_recovery/backup.go` (NEW)
- **Effort:** 2 days
- **Acceptance Criteria:**
  - Daily automated backups
  - Backup verification (restore test)
  - Backup retention (30 days)
  - Alert on backup failure

#### Task 1.4.2: DR Runbook
- **File:** `docs/runbooks/disaster_recovery.md` (NEW)
- **Effort:** 2 days
- **Acceptance Criteria:**
  - Step-by-step recovery procedure
  - RTO target: 4 hours
  - RPO target: 1 hour
  - Contact list
  - Tested quarterly

---

## Phase 2: Observability (Weeks 9-12) 🟡 HIGH PRIORITY

**Goal:** Full visibility into system behavior
**Team:** 2 backend engineers + 1 DevOps engineer
**Success Criteria:** Distributed tracing on all requests, actionable dashboards

### Week 9-10: Distributed Tracing
- Implement OpenTelemetry
- Trace all requests end-to-end
- Trace context propagation
- Cloud Trace integration

### Week 11-12: Dashboards & Alerting
- Operational dashboards (Grafana/Cloud Monitoring)
- SLO dashboards
- Alerting rules (PagerDuty integration)
- Error budget tracking

---

## Phase 3: Scalability (Weeks 13-16) 🟡 MEDIUM PRIORITY

**Goal:** Handle 10x traffic growth
**Team:** 3 backend engineers
**Success Criteria:** Linear scaling to 1000 RPS, <500ms P99 latency

### Week 13-14: Caching
- Redis caching layer
- Query result caching
- Embedding caching
- Cache invalidation strategy

### Week 15-16: Horizontal Scaling
- Connection pooling (PgBouncer)
- Database read replicas
- Load balancing configuration
- Auto-scaling policies

---

## Phase 4: Data Quality & Governance (Weeks 17-20) 🟡 MEDIUM PRIORITY

**Goal:** Ensure data quality and compliance
**Team:** 2 data engineers
**Success Criteria:** Data quality SLAs met, compliance audits passed

### Week 17-18: Data Contracts
- Schema validation
- Data quality checks
- Data lineage tracking

### Week 19-20: Compliance Automation
- GDPR automation (export/delete)
- Consent management
- Data residency enforcement
- Compliance dashboards

---

## Gating Criteria for Production Deployment

**Before ANY production traffic:**

### Phase 0 Gates (Security)
- [ ] All 18 P0 security issues resolved
- [ ] Penetration test passed
- [ ] Security audit completed
- [ ] JWT auth enforced on all endpoints
- [ ] Encryption at rest and in transit enabled
- [ ] Audit logging operational

### Phase 1 Gates (Reliability)
- [ ] All 34 P1 reliability issues resolved
- [ ] Circuit breakers on all external calls
- [ ] Health checks passing
- [ ] Monitoring dashboards operational
- [ ] On-call rotation established
- [ ] DR runbook tested

### Phase 2 Gates (Observability)
- [ ] Distributed tracing on 100% of requests
- [ ] Custom metrics for all business KPIs
- [ ] Alerting rules configured and tested
- [ ] SLO dashboards operational

### Phase 3 Gates (Scalability)
- [ ] Load test passed (1000 RPS)
- [ ] P99 latency < 500ms under load
- [ ] Auto-scaling tested
- [ ] Cache hit rate > 80%

### Phase 4 Gates (Compliance)
- [ ] GDPR compliance verified
- [ ] Data quality SLAs met (>99% accuracy)
- [ ] Compliance audit passed

---

## Resource Requirements

### Team Composition
- **2-3 Senior Backend Engineers** (Go, Python)
- **1 Security Engineer**
- **1 DevOps Engineer**
- **1 Data Engineer**
- **0.5 Product Manager** (prioritization)

### Infrastructure Costs (Monthly)
| Resource | Current | Target | Cost |
|----------|---------|--------|------|
| Cloud Run | $50 | $500 | +$450 |
| AlloyDB | $200 | $800 | +$600 |
| Redis | $0 | $150 | +$150 |
| Cloud Monitoring | $0 | $200 | +$200 |
| Cloud Trace | $0 | $100 | +$100 |
| Secret Manager | $0 | $50 | +$50 |
| **Total** | **$250** | **$1,800** | **+$1,550** |

### Timeline Summary
| Phase | Duration | Team Size | Effort (weeks) |
|-------|----------|-----------|----------------|
| Phase 0: Security | 4 weeks | 3 engineers | 12 |
| Phase 1: Reliability | 4 weeks | 3 engineers | 12 |
| Phase 2: Observability | 4 weeks | 3 engineers | 12 |
| Phase 3: Scalability | 4 weeks | 3 engineers | 12 |
| Phase 4: Compliance | 4 weeks | 2 engineers | 8 |
| **Total** | **20 weeks** | **4-6 engineers** | **56 weeks** |

---

## Risk Mitigation

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Team capacity insufficient | Medium | High | Hire contractors, extend timeline |
| Security vulnerabilities found late | Low | Critical | Early pen test, continuous security scanning |
| Performance regression | Medium | High | Load testing every sprint |
| Scope creep | High | Medium | Strict prioritization, MVP focus |
| Dependency delays (GCP features) | Low | Medium | Fallback options, workarounds |

---

## Next Steps

1. **Immediate (This Week):**
   - Review and approve this plan
   - Assign team members
   - Set up project tracking (Jira/Linear)
   - Schedule kickoff meeting

2. **Week 1:**
   - Start Phase 0 (Security Foundation)
   - Implement JWT middleware
   - Begin encryption at rest setup

3. **Ongoing:**
   - Weekly progress reviews
   - Bi-weekly demos
   - Monthly stakeholder updates

---

**Approval Required:** This plan requires executive approval before proceeding. Production deployment is blocked until Phase 0 and Phase 1 are complete.

**Document Version:** 2.0
**Last Updated:** April 1, 2026
**Next Review:** Weekly (every Monday)
