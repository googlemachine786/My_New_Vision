# Visionary RAG Pipeline - Enterprise Production Implementation Plan

**Version:** 3.0
**Date:** April 1, 2026
**Goal:** Achieve 95%+ enterprise production readiness
**Audit Reference:** ENTERPRISE_PRODUCTION_AUDIT.md, CRITICAL_FIXES.md, PRODUCTION_READINESS_CHECKLIST.md
**User Stories:** 73 total across 5 epics (see USER_STORY_ALIGNMENT.md)

---

## Executive Summary

This plan organizes implementation around **73 user stories** across 5 epics, with user-story-driven phases prioritized by business value and technical dependencies.

### User Story Overview

| Epic | Stories | Implemented | Partial | Not Started | Completion % |
|------|---------|-------------|---------|-------------|--------------|
| **E-001: Core RAG & AI** | 40 (+7) | 18 | 10 | 12 (+7) | 52.5% |
| **E-002: Curriculum & Learning** | 28 | 8 | 6 | 14 | 39.3% |
| **E-003: User Experience** | 21 (+2) | 3 | 4 | 14 (+2) | 23.8% |
| **E-004: Analytics & Monitoring** | 4+5* | 2 | 0 | 2+5* | 50.0% → 35%** |
| **E-005: Security & Compliance** | 2 | 0 | 0 | 2 | 0.0% |
| **TOTAL** | **83+5*** | **31** | **20** | **32+5*** | **48.2% → 44.0%** |

_*5 new analytics tasks added (AM-01 to AM-05)_
_**Completion % adjusted due to expanded scope_
_(+7) = 7 new stories added: 6 from Stories A-F + 4 from Stories G-J (Story G→E-001, Story H→E-001, Story I→E-001, Story J→E-001)_

### Production Readiness Assessment

| Category | Current | Target | Gap | Priority |
|----------|---------|--------|-----|----------|
| **Core RAG Functionality** | 66.7% | 95% | -28.3% | P0 |
| **Security & Compliance** | 0.0% | 95% | -95.0% | **P0 CRITICAL** |
| **Analytics & Monitoring** | 50.0% | 95% | -45.0% | P1 |
| **Curriculum Features** | 39.3% | 90% | -50.7% | P2 |
| **User Experience** | 26.3% | 85% | -58.7% | P3 |
| **OVERALL** | **54.8%** | **95%** | **-40.2%** | |

### Critical Finding

**MVP Readiness:** 76% of P0 stories complete (10/13), 3 stories partial.  
**Production Blockers:** US-014 (JWT RBAC), US-023 (Auto Dataset), US-043 (Cost Tracking).  
**DO NOT DEPLOY TO PRODUCTION** until Phase 1 (MVP) is 100% complete, especially US-014 RBAC for security.

---

## User-Story-Driven Phase Structure

### Phase 1: MVP (Weeks 1-3) - P0 Critical Stories

**Goal:** Complete all P0 stories for MVP production deployment
**Owner:** Engineering Lead
**Team:** 2-3 engineers
**Stories:** 15 P0 stories (10 complete, 5 partial) — *updated: +2 stories (US-013 prod deploy, US-018 WebSocket)*
**Current Completion:** 73% (weighted 85%)

#### Stories Included

| Story ID | Title | Epic | Status | Remaining Work |
|----------|-------|------|--------|----------------|
| US-013 | Vector Database Setup | E-001 | ⚠️ Partial (40%) | Local docker-compose complete; production GCP deployment pending |
| US-014 | API Gateway | E-001 | ✅ Complete | None |
| US-015 | Content Ingestion Pipeline | E-001 | ✅ Complete | None |
| US-016 | Embedding Generation | E-001 | ✅ Complete | None |
| US-017 | Hybrid Search | E-001 | ✅ Complete | None |
| US-018 | Response Streaming | E-001 | ⚠️ Partial (50%) | SSE implemented; WebSocket streaming pending for bidirectional low-latency |
| US-019 | JWT Authentication with RBAC | E-005 | ⚠️ Partial (60%) | Add RBAC, enforce grade isolation |
| US-020 | RRF Fusion | E-001 | ✅ Complete | None |
| US-021 | Contextual Grounding | E-001 | ⚠️ Partial (70%) | Improve context awareness + **NEW: Fallback strategy (Story C)** |
| US-024 | Caching Layer | E-001 | ✅ Complete | None |
| US-025 | Session Management | E-001 | ✅ Complete | None |
| US-036 | Query Taxonomy Filtering | E-001 | ✅ Complete | None |
| US-037 | Rate Limiting | E-001 | ✅ Complete | None |

#### Task 1.1: Complete US-019 (JWT RBAC) - 3-4 days
**Files:** `orchestrator/middleware/auth.go`, `schema/v2_production.sql`

- [ ] Add Role enum to Claims (admin/teacher/student)
- [ ] Implement RBAC enforcement middleware
- [ ] Add RLS policies to database schema
- [ ] Enforce grade isolation in ALL database queries
- [ ] Add token invalidation mechanism
- [ ] Write comprehensive auth tests (100% coverage)

**Acceptance Criteria:**
- All endpoints require valid JWT with role claims
- Grade 6 students cannot access Grade 8 content (verified by test)
- Teachers can access multiple grades based on assignment
- Admins can access all content
- 100% test coverage for auth module
- Security audit passed

#### Task 1.2: Complete US-021 (Contextual Grounding) + Story C (Fallback Strategy) - 3-4 days
**Files:** `orchestrator/handler/rag_handler_real.go` (modify), `orchestrator/handler/fallback_handler.go` (new), `orchestrator/observability/metrics.go` (modify)

**Enhanced Scope:** Now includes fallback mechanism when RAG returns zero results — "no results found → general LLM answer" path with `fallback_rate` metric (Story C: Fallback Strategy for Failed Retrieval).

**Tasks:**
- [ ] Improve query rewriting with retrieved context
- [ ] Add ambiguity detection
- [ ] Implement fallback to original query
- [ ] Test with multi-turn conversations
- [ ] **NEW (Story C):** Detect zero-result retrieval and trigger fallback path
  - Location: `orchestrator/handler/rag_handler_real.go`
  - When retrieval returns 0 chunks or all chunks below similarity threshold (cosine <0.5)
  - Switch to general LLM response mode with system prompt: "I don't have specific curriculum data for this, but here's what I know..."
- [ ] **NEW (Story C):** Create fallback response handler
  - Location: `orchestrator/handler/fallback_handler.go`
  - Uses same LLM but with different system prompt (no RAG grounding)
  - Includes disclaimer: "This answer is based on general knowledge, not your specific curriculum"
  - Logs fallback event with query, grade, subject for tracking
- [ ] **NEW (Story C):** Track `fallback_rate` metric
  - Location: `orchestrator/observability/metrics.go`
  - Prometheus gauge: `rag_fallback_rate` (fallbacks / total queries over 1-hour window)
  - Alert when fallback_rate >15% (indicates retrieval gap)
- [ ] **NEW (Story C):** Frontend fallback indicator
  - When fallback response returned, show subtle badge: "General answer (curriculum data not found)"
  - Maintain UX transparency

**Acceptance Criteria:**
- Query rewriting uses session history + retrieved context
- Ambiguous queries detected (confidence < 0.6)
- Fallback to original query when rewriting fails
- Multi-turn conversation coherence maintained
- **NEW (Story C):** When RAG returns 0 results, fallback LLM response generated with disclaimer
- **NEW (Story C):** `fallback_rate` metric tracked in Prometheus, visible in Grafana
- **NEW (Story C):** Fallback rate alert triggers when >15% over 1-hour window
- **NEW (Story C):** Frontend displays subtle indicator when fallback response used
- **NEW (Story C):** Fallback latency within same budget as normal RAG (<500ms p99)

#### Task 1.3: MVP Integration Testing - 2 days
**Files:** `tests/mvp_integration_test.go`, `tests/mvp_e2e_test.py`

- [ ] End-to-end query flow test
- [ ] Grade isolation verification test
- [ ] Rate limiting test
- [ ] Performance test (100 concurrent users)
- [ ] Security penetration test

**Acceptance Criteria:**
- All 13 P0 stories verified functional
- p99 latency <500ms
- Error rate <1%
- Zero security vulnerabilities

**Phase 1 Acceptance Criteria:**
- ✅ All 13 P0 stories 100% complete
- ✅ MVP integration tests passing
- ✅ Security audit passed
- ✅ Performance tests passed (100 users, p99<500ms)
- ✅ Ready for MVP production deployment

#### Task 1.6: Story G (Cross-Agent Context Sharing) - 3-4 days **[NEW]**
**Priority:** P1
**Files:** `orchestrator/context/context_object.go` (new), `orchestrator/context/session_registry.go` (new), `services/api-gateway/handler/context_handler.go` (new), `services/teaching-agent/context/context_consumer.go` (new), `services/practice-agent/context/context_publisher.go` (new), `schema/context_sharing.sql` (new)

**User Story:** As a system, I want different AI agents to share context, so that the student experience feels unified.

**Gap:** No shared context object between services. Each service works in isolation.

**Tasks:**
- [ ] Create unified Context Object structure
  - Location: `orchestrator/context/context_object.go`
  - Define struct: `ContextObject{UserID, SessionID, Board, Grade, Language, ActiveChapter, CurrentTopic, PracticeErrors[], PracticeScores[]}`
  - Immutable fields: Board, Grade, Language (set at session start)
  - Mutable fields: PracticeErrors, PracticeScores, ActiveChapter (updated by agents)
- [ ] Implement session-level context registry
  - Location: `orchestrator/context/session_registry.go`
  - Redis-backed context store (TTL = session timeout)
  - `GetContext(sessionID) → ContextObject`
  - `UpdateContext(sessionID, updates) → error`
  - Thread-safe concurrent updates with optimistic locking
- [ ] Create context publishing/subscription endpoints
  - Location: `services/api-gateway/handler/context_handler.go`
  - POST `/api/context/{sessionID}/publish` — agents publish practice errors/scores
  - GET `/api/context/{sessionID}` — agents retrieve shared context
  - Protected by internal service auth (not exposed to frontend)
- [ ] Implement context consumer in Teaching Agent
  - Location: `services/teaching-agent/context/context_consumer.go`
  - On session start, fetch ContextObject
  - Inject Board, Grade, Language into teaching prompts
  - Adapt teaching strategy based on PracticeErrors (weak areas)
- [ ] Implement context publisher in Practice Agent
  - Location: `services/practice-agent/context/context_publisher.go`
  - After each practice session, publish errors + scores to ContextObject
  - Include: `{question_id, error_type, timestamp, score}`
- [ ] Create context sharing database schema
  - Location: `schema/context_sharing.sql`
  - Table: `session_contexts(session_id UUID PK, user_id UUID, board VARCHAR, grade INT, language VARCHAR, active_chapter JSONB, practice_errors JSONB, practice_scores JSONB, created_at, updated_at)`
  - Index on `(user_id, updated_at DESC)` for quick lookup

**Acceptance Criteria:**
1. ContextObject created and accessible across all agents in same session
2. Practice errors/scores from Practice Agent visible to Teaching Agent within 1 second
3. Unified Board, Grade, Language shared across all agents
4. Context updates are thread-safe (no race conditions under concurrent writes)
5. Redis-backed context with <50ms read latency
6. Context persists for session duration (max 24 hours TTL)
7. Database backup of context for audit/recovery
8. 90%+ test coverage for context module

**Dependencies:** Session Management (complete), Redis (Phase 2), API Gateway (complete)

#### Task 1.4: Complete US-013 (PostgreSQL + Vertex AI Production Infra) - 5-7 days **[NEW]**
**Priority:** P0 (blocks all production deployment)
**Files:** `terraform/main.tf` (modify), `terraform/variables.tf` (modify), `terraform/outputs.tf` (modify), `schema/v2_production.sql` (modify), `services/api-gateway/config/db.go` (new)

**User Story:** As a Backend Engineer, I want to deploy PostgreSQL and Vertex AI Vector Search to store user profiles and curriculum embeddings.

**Tasks:**
- [ ] Provision GCP Cloud SQL PostgreSQL instance with pgvector extension
- [ ] Create user schema: `users(id, email, grade, role, created_at, updated_at)`
- [ ] Configure ScaNN vector index for curriculum embeddings
- [ ] Implement GCP IAM-based connection string management (no env vars)
- [ ] Configure automated backup policies (daily, 30-day retention)
- [ ] Set up connection pooling with Cloud SQL Auth Proxy
- [ ] Enable monitoring and alerting (CPU, memory, disk, connections)
- [ ] Test backup restoration procedure
- [ ] Document connection architecture and runbooks

**Acceptance Criteria:**
1. PostgreSQL instance live and accessible via secure IAM auth
2. User schema created and validated with migrations
3. Vector search index operational with pgvector + ScaNN
4. Connection strings secured (no hardcoded secrets, IAM-based)
5. Backup policies active and restoration tested
6. Monitoring enabled with alerts on critical metrics
7. QA validation completed

**Dependencies:** GCP account configured, Terraform state backend (GCS)

#### Task 1.5: Complete US-018 (WebSocket Streaming) - 3-4 days **[NEW]**
**Priority:** P0 (blocks student onboarding experience)
**Files:** `services/api-gateway/handler/websocket.go` (new), `services/api-gateway/client/llm_client.go` (modify), `frontend/src/components/Chat/StreamingChat.tsx` (modify)

**User Story:** As a student, I want to see AI explanations stream word-by-word during my first session to avoid waiting for the full RAG pipeline.

**Tasks:**
- [ ] Create WebSocket endpoint in API Gateway (`/ws/chat`)
- [ ] Implement bidirectional WebSocket handler with connection lifecycle
- [ ] Integrate with existing LLM client for streaming output (<500ms initial latency)
- [ ] Add connection health checks and automatic reconnection
- [ ] Implement message framing (JSON: `{type: "chunk"|"done"|"error", data: {...}}`)
- [ ] Add WebSocket client to React frontend with fallback to SSE
- [ ] Load test WebSocket connections (100 concurrent, stable during lesson)

**Acceptance Criteria:**
1. WebSocket server initialized on dashboard entry
2. LLM output streams with <500ms initial latency
3. Connection stable during lesson session (zero drops under normal load)
4. Frontend gracefully falls back to SSE if WebSocket unavailable
5. Bidirectional communication working (client sends, server streams)

**Dependencies:** API Gateway (complete), Frontend WebSocket client

---

### Phase 2: Production Foundation (Weeks 3-6) - P1 High Stories

**Goal:** Complete P1 stories for production-scale deployment  
**Owner:** Engineering Lead  
**Team:** 3-4 engineers  
**Stories:** 10 P1 stories (6 complete, 4 not started)  
**Current Completion:** 60%

#### Stories Included

| Story ID | Title | Epic | Status | Priority |
|----------|-------|------|--------|----------|
| US-023 | Auto Dataset Creation | E-001 | ❌ Not Started | P1 |
| US-029 | Distributed Tracing | E-001 | ⚠️ Partial | P1 |
| US-030 | Feedback Logging | E-001 | ✅ Complete | P1 |
| US-031 | LLM Judge | E-001 | ✅ Complete | P1 |
| US-033 | Multi-format Support | E-001 | ✅ Complete | P1 |
| US-034 | Content Deduplication | E-001 | ✅ Complete | P1 |
| US-035 | Incremental Ingestion | E-001 | ❌ Not Started | P1 |
| US-038 | Model Routing | E-001 | ⚠️ Partial | P1 |
| US-041 | Response Quality Metrics | E-001 | ❌ Not Started | P1 |
| US-046 | Session Memory | E-002 | ✅ Complete | P1 |
| **Story A** | **Answer Feedback Collection** | **E-004** | **❌ Not Started** | **P1** |
| **Story B** | **AI Quality Learning Loop** | **E-001** | **❌ Not Started** | **P1** |
| **Story D** | **Re-explain Concept Differently** | **E-003** | **❌ Not Started** | **P1** |

#### Task 2.1: Complete US-023 (Auto Dataset) + Story B (AI Quality Learning Loop) - 4-5 days
**Files:** `quality_loop/auto_dataset_creator.py` (new), `quality_loop/feedback_pipeline.py` (new), `quality_loop/judge.py` (modify)

**Enhanced Scope:** Now includes automated feedback → flagging → weekly fine-tuning dataset pipeline (Story B: AI Quality Learning Loop).

**Tasks:**
- [ ] Automatic golden QA generation from negative feedback
- [ ] Difficulty auto-assessment
- [ ] Integration with feedback loop
- [ ] Minimum 100 QA pairs per grade/subject
- [ ] **NEW (Story B):** Create feedback-to-flagging pipeline — low-rated responses automatically flagged for review
  - Location: `quality_loop/feedback_pipeline.py`
  - Query `feedback_aggregates` for thumbs_down with reasons
  - Flag responses with confidence <0.6 or ≥2 thumbs_down
- [ ] **NEW (Story B):** Weekly fine-tuning dataset generator
  - Scheduled job (cron: every Sunday 00:00)
  - Collects all flagged responses + context + golden QA pairs
  - Outputs JSONL format for Vertex AI fine-tuning
  - Stores in `fine_tuning_datasets` table with version tracking
- [ ] **NEW (Story B):** LLM judge integration with feedback loop
  - Leverage existing `quality_loop/judge.py`
  - Auto-evaluate flagged responses for hallucination
  - Tag responses as "needs_review", "confirmed_bad", "acceptable"

**Acceptance Criteria:**
- Negative feedback automatically generates golden QA pairs
- 100+ QA pairs per grade/subject combination
- Difficulty ratings auto-assigned
- QA pairs stored in `golden_qa_dataset` table
- **NEW (Story B):** Low-rated responses (thumbs_down) automatically flagged within 1 hour
- **NEW (Story B):** Weekly fine-tuning dataset generated every Sunday (JSONL format)
- **NEW (Story B):** Dataset includes query, response, retrieved context, feedback reason
- **NEW (Story B):** LLM judge evaluates flagged responses, tags for review priority
- **NEW (Story B):** At least 50 flagged examples per week for fine-tuning

#### Task 2.2: Complete US-035 (Incremental Ingestion) - 3 weeks
**Files:** `ingestion/incremental.py` (new), `schema/v2_production.sql` (modify)

- [ ] Document-level change detection (SHA-256)
- [ ] Page-level change detection
- [ ] Chunk versioning with lineage tracking
- [ ] Orphaned chunk cleanup
- [ ] Rollback support

**Acceptance Criteria:**
- Unchanged documents skipped during ingestion
- Only changed pages re-ingested
- Chunk versioning tracks superseded chunks
- Orphaned chunks marked and cleaned up
- Rollback to previous version supported

#### Task 2.3: Complete US-029 (Distributed Tracing) - 1 week
**Files:** `orchestrator/observability/tracer.go`, `orchestrator/observability/metrics.go`

- [ ] Full span coverage (all stages)
- [ ] Trace context propagation
- [ ] Cloud Trace integration
- [ ] Trace-based dashboards

**Acceptance Criteria:**
- All requests traced end-to-end
- Each stage (session, rewrite, embed, search, generate) has spans
- Traces visible in Cloud Trace
- TTFT decomposition available

#### Task 2.4: Complete US-041 (Quality Metrics) + Story A (Answer Feedback Collection) - 3-4 weeks
**Files:** `orchestrator/quality/monitor.go` (new), `services/api-gateway/handler/feedback_handler.go` (modify), `services/api-gateway/handler/stats_handler.go` (modify), `services/api-gateway/metrics/prometheus.go` (new), `observability/grafana/dashboards/quality.json` (new), `schema/feedback_enhanced.sql` (new)

**Enhanced Scope:** Now includes Prometheus/Grafana monitoring setup, feedback aggregation dashboard, and hallucination rate tracking (merged with new Retrieval Accuracy & Feedback Tracking story). **ALSO** enhanced with Story A (Answer Feedback Collection) — "What was wrong?" input field and full query/response/context storage.

**Tasks:**
- [ ] Real-time quality scoring per query
- [ ] A/B testing framework
- [ ] Quality dashboards
- [ ] Feedback trend analysis
- [ ] **NEW:** Prometheus metrics exporter for all 4 services (API Gateway, Orchestrator, Query Understanding, Ingestion)
- [ ] **NEW:** Thumbs up/down feedback aggregation per query (leverage existing `feedback_handler.go`)
- [ ] **NEW:** Hallucination rate tracking via LLM judge comparison (compare LLM answers against retrieved context)
- [ ] **NEW:** Grafana dashboard for API latency, error rates, cache hit rates, feedback scores
- [ ] **NEW:** 95% accuracy SLA monitoring with alerts
- [ ] **NEW (Story A):** Add "What was wrong?" optional field to feedback endpoint
  - Location: `services/api-gateway/handler/feedback_handler.go`
  - Add `reason` enum: `irrelevant`, `incorrect`, `incomplete`, `confusing`, `outdated`, `other`
  - Add optional `free_text` field (max 500 chars) for detailed feedback
- [ ] **NEW (Story A):** Store full query + response + context with each feedback
  - Create `feedback_contexts` table: `feedback_id, query_text, retrieved_chunks[], llm_response, timestamp`
  - Link to existing `feedback_aggregates` table
  - Enable full-context review for quality team
- [ ] **NEW (Story A):** Feedback review dashboard for quality team
  - Show thumbs_down with reasons + full context
  - Filterable by grade, subject, time range, reason
  - Export to CSV for fine-tuning pipeline

**Acceptance Criteria:**
- Quality score calculated per query
- A/B tests running for retrieval parameters
- Quality metrics visible in dashboard
- Quality trends tracked over time
- **NEW:** Prometheus metrics exported for all services (request count, latency, error rate, cache hits)
- **NEW:** Feedback endpoint (`/api/feedback`) aggregates thumbs up/down per query
- **NEW:** Hallucination rate <5% (LLM judge compares answer vs retrieved context)
- **NEW:** Grafana dashboard showing: API latency (p50/p95/p99), error rates, cache hit rates, feedback scores
- **NEW:** SLA alerts triggered when accuracy drops below 95%
- **NEW:** Per-query feedback aggregation visible in dashboard
- **NEW (Story A):** Feedback endpoint accepts `reason` enum and optional `free_text` field
- **NEW (Story A):** Full query + response + context stored with each thumbs_down
- **NEW (Story A):** Feedback review dashboard shows thumbs_down with reasons, filterable and exportable
- **NEW (Story A):** Feedback stored within 200ms of submission (async, non-blocking)

**Implementation Details:**
```
Prometheus Metrics to Export:
- api_request_duration_seconds (histogram)
- api_requests_total (counter)
- api_errors_total (counter)
- cache_hit_rate (gauge)
- feedback_thumbs_up_total (counter)
- feedback_thumbs_down_total (counter)
- hallucination_rate (gauge)
- quality_score (gauge)

Grafana Dashboard Panels:
1. API Latency (p50/p95/p99) - heatmap
2. Request Rate & Error Rate - dual axis
3. Cache Hit Rate - gauge
4. Feedback Score (thumbs up/down ratio) - pie chart
5. Hallucination Rate - time series
6. Quality Score Trend - time series
7. 95% SLA Compliance - stat panel
```

**Dependencies:** Feedback endpoint (complete), Query Understanding Service (complete), Service client metrics (partial - needs Prometheus integration)

#### Task 2.5: Story D (Re-explain Concept Differently) - 4-5 days **[NEW]**
**Priority:** P1
**Files:** `services/api-gateway/handler/reexplain_handler.go` (new), `orchestrator/handler/reexplain_generator.go` (new), `services/query-understanding-service/prompt/reexplain_templates.go` (new), `frontend/src/components/Chat/ReexplainButton.tsx` (new), `schema/reexplain_tracking.sql` (new)

**User Story:** As a student, I want the AI to re-explain a concept differently when I don't understand, so that I can eventually grasp it fully.

**Tasks:**
- [ ] Create "I don't understand" button in chat UI
  - Location: `frontend/src/components/Chat/ReexplainButton.tsx`
  - Appears below each AI response
  - Disabled after 3 re-explain attempts for same query
- [ ] Implement re-explain prompt templates
  - Location: `services/query-understanding-service/prompt/reexplain_templates.go`
  - 5+ template strategies: analogy-based, step-by-step, visual description, real-world example, simplified language
  - Each re-explain uses different strategy (tracked by attempt counter)
  - Templates include: "Explain like I'm 10", "Use a real-world example", "Break it into steps", "Use an analogy"
- [ ] Create re-explain session tracker
  - Location: `orchestrator/handler/reexplain_generator.go`
  - Track attempt counter per query in session state
  - Maximum 3 re-explain attempts per original query
  - After 3 failures → trigger fallback strategy (escalate to teacher flag or general help resource)
- [ ] Implement re-explain endpoint in API Gateway
  - Location: `services/api-gateway/handler/reexplain_handler.go`
  - POST `/api/reexplain` with `{query_id, attempt_number, previous_response}`
  - Returns re-explained response with different strategy
- [ ] Create re-explain tracking table
  - Location: `schema/reexplain_tracking.sql`
  - Table: `reexplain_attempts(id, query_id, user_id, attempt_number, strategy_used, timestamp)`
  - Track success rate (does user continue conversation after re-explain?)
- [ ] Implement fallback after 3 failures
  - After 3 re-explains, trigger: "I see this is still unclear. Let me suggest: [review the textbook chapter], [ask your teacher], or [try a related question]"
  - Log to quality team for curriculum gap analysis

**Acceptance Criteria:**
1. "I don't understand" button available below every AI response
2. Each re-explain uses different explanation strategy (analogy, steps, examples, simplified)
3. Maximum 3 re-explain attempts per query enforced
4. After 3 failures, fallback strategy triggered (teacher escalation + resource suggestions)
5. Re-explain attempts tracked in database with strategy used
6. Re-explain latency <500ms p99 (same as normal response)
7. Frontend button disabled after 3 attempts with message: "Let's try a different approach..."

**Dependencies:** Session Management (complete), API Gateway (complete), Frontend chat components

**Phase 2 Acceptance Criteria:**
- ✅ All 10 P1 stories 100% complete
- ✅ Incremental ingestion reducing ingestion time by 80%
- ✅ Auto dataset creating 10+ QA pairs per day
- ✅ Distributed tracing operational
- ✅ Quality monitoring dashboard live

#### Task 2.6: Story H (Query Enrichment with Profile/Session Context) - 2-3 days **[NEW]**
**Priority:** P1
**Files:** `orchestrator/handler/query_enricher.go` (new), `services/query-understanding-service/enrichment/profile_enrichment.go` (new), `orchestrator/handler/rag_handler_real.go` (modify), `services/query-understanding-service/prompt/context_templates.go` (new)

**User Story:** As a system, I want to enrich user queries with profile and session context, so that responses are personalized and relevant.

**Gap:** Current prompts don't include user profile context or active chapter.

**Tasks:**
- [ ] Create query enrichment middleware
  - Location: `orchestrator/handler/query_enricher.go`
  - Intercepts incoming query before routing to RAG pipeline
  - Fetches user profile: `{Grade, Board, Language, LearningStyle}` from `users` table
  - Fetches session context: `{ActiveChapter, CurrentTopic, RecentPracticeErrors}` from session context (Story G)
  - Enrichment latency budget: <50ms (async, non-blocking)
- [ ] Implement profile-based prompt enrichment
  - Location: `services/query-understanding-service/enrichment/profile_enrichment.go`
  - Generate context prefix: `"You are teaching a Grade {grade} student following {board} curriculum in {language}..."`
  - Add active chapter context: `"The student is currently studying '{chapter_name}'. Prioritize content from this chapter."`
  - Add practice error context: `"The student recently struggled with: {practice_errors}. Avoid assuming mastery of these topics."`
- [ ] Integrate enrichment into RAG handler prompt construction
  - Location: `orchestrator/handler/rag_handler_real.go`
  - Modify prompt assembly: `final_prompt = enrichment_prefix + user_query + retrieved_context`
  - Ensure enrichment is backward compatible (works without profile → fallback to generic prompt)
  - Add enrichment metadata to response: `{enriched: true, profile_fields: [grade, board, language, active_chapter]}`
- [ ] Create enrichment prompt templates
  - Location: `services/query-understanding-service/prompt/context_templates.go`
  - Template per grade band: Grade 1-5 (simple language), Grade 6-8 (intermediate), Grade 9-12 (advanced)
  - Template per board: CBSE, ICSE, State Board variations
  - Template for language: English, Hindi (future multilingual)
  - Include examples in templates for few-shot learning

**Acceptance Criteria:**
1. Every LLM prompt prefixed with User Profile (Grade, Board, Language)
2. Active chapter context included when user has dashboard activity
3. Practice error context included when user has recent practice sessions
4. Enrichment adds <50ms to query latency (async fetch, <30ms profile lookup)
5. Backward compatible — queries without profile still work with generic prompt
6. Enrichment metadata included in response for debugging
7. 90%+ test coverage for enrichment module
8. A/B test shows improved answer relevance with enrichment vs without

**Dependencies:** User Profile schema (US-013), Session Context (Story G, Task 1.6), RAG Handler (complete)

**Phase 2 Acceptance Criteria:**
- ✅ All 10 P1 stories 100% complete
- ✅ Incremental ingestion reducing ingestion time by 80%
- ✅ Auto dataset creating 10+ QA pairs per day
- ✅ Distributed tracing operational
- ✅ Quality monitoring dashboard live
- ✅ **NEW:** Query enrichment with profile/session context operational (Story H)

---

### Phase 3: Scale & Optimization (Weeks 7-11) - P0 Remaining + P2 Medium

**Goal:** Complete remaining P0 stories and start P2 optimization
**Owner:** Engineering Lead
**Team:** 3-4 engineers
**Stories:** 10 stories (7 P0 remaining, 3 P2) — *updated: +1 story (US-045 Clarifying Questions)*

#### Stories Included

| Story ID | Title | Epic | Priority |
|----------|-------|------|----------|
| US-040 | Hybrid Search Tuning | E-001 | P0 |
| US-042 | Retrieval Analytics | E-001 | P0 |
| US-043 | Cost Tracking | E-001 | P0 |
| US-044 | Multi-hop Retrieval | E-001 | P0 |
| US-045 | Clarification Generation | E-001 | P0 |
| US-022 | Multi-agent Routing | E-001 | P2 |
| US-026 | Query Rewriting | E-001 | P2 |
| US-032 | Golden Response Gen | E-001 | P2 |
| US-039 | Hallucination Control | E-001 | P2 |
| **Story C** | **Fallback Strategy for Failed Retrieval** | **E-001** | **P1** |
| **Story E** | **Confidence Scores on AI Responses** | **E-001** | **P1** |
| **Story F** | **Query Intent Classification** | **E-001** | **P1** |

#### Task 3.1: Complete US-043 (Cost Tracking) - 1-2 weeks
**Files:** `orchestrator/observability/cost_tracker.go` (new), `schema/cost_tracking.sql` (new), `services/api-gateway/metrics/prometheus.go` (modify), `observability/grafana/dashboards/cost.json` (new)

**Enhanced Scope:** Now includes Prometheus metrics integration for cost tracking and budget alerting.

**Tasks:**
- [ ] Token usage tracking per query
- [ ] Cost calculation (Vertex AI pricing)
- [ ] Daily/weekly/monthly aggregation
- [ ] Budget alerts
- [ ] Cost dashboard
- [ ] **NEW:** Prometheus metrics for cost tracking (tokens_per_query, cost_per_query, daily_cost)
- [ ] **NEW:** Budget alert integration with Grafana alerts
- [ ] **NEW:** Cost breakdown by service (embedding vs generation)

**Acceptance Criteria:**
- Token usage tracked for embedding + generation
- Cost per query calculated accurately
- Daily cost reports available
- Budget alerts at 80% threshold
- Cost dashboard with trends
- **NEW:** Prometheus metrics: `tokens_per_query_total`, `cost_per_query_usd`, `daily_cost_usd`
- **NEW:** Grafana cost panel showing daily spend vs budget
- **NEW:** Alert when daily cost exceeds threshold

#### Task 3.2: Complete US-042 (Retrieval Analytics) - 2-3 weeks
**Files:** `analytics/retrieval_analyzer.go` (new), `schema/analytics_tables.sql` (new), `services/api-gateway/metrics/prometheus.go` (modify), `services/api-gateway/handler/stats_handler.go` (modify), `services/api-gateway/cache/cag_orchestrator.go` (modify), `services/api-gateway/client/service_client.go` (modify), `observability/grafana/dashboards/retrieval.json` (new)

**Enhanced Scope:** Now includes Prometheus metrics for API latency tracking, retrieval latency breakdown, and comprehensive analytics dashboard.

**Tasks:**
- [ ] Chunk-level retrieval statistics
- [ ] Query pattern analysis
- [ ] Taxonomy coverage metrics
- [ ] Retrieval latency breakdown
- [ ] Analytics dashboard
- [ ] **NEW:** Prometheus metrics for API latency (histogram: p50/p95/p99)
- [ ] **NEW:** Per-service latency tracking (API Gateway, Orchestrator, Query Understanding, Ingestion)
- [ ] **NEW:** Cache hit/miss rate tracking (leverage `cag_orchestrator.go` metrics)
- [ ] **NEW:** Service client metrics integration (leverage `service_client.go` Metrics struct)
- [ ] **NEW:** Grafana dashboard for retrieval analytics

**Acceptance Criteria:**
- Most/least retrieved chunks tracked
- Query patterns analyzed hourly
- Taxonomy coverage gaps identified
- Retrieval latency breakdown available
- Dashboard with real-time analytics
- **NEW:** Prometheus metrics: `retrieval_latency_seconds`, `cache_hit_rate`, `query_patterns_total`
- **NEW:** Grafana panels: retrieval latency heatmap, cache hit rate gauge, query pattern breakdown
- **NEW:** All 4 services exporting metrics to Prometheus
- **NEW:** Stats endpoint (`/api/stats`) returns comprehensive metrics

#### Task 3.3: Complete US-044 (Multi-hop Retrieval) - 1-2 weeks
**Files:** `retrievers/multi_hop_retriever.py` (new)

- [ ] Iterative retrieval (retrieve → analyze → retrieve again)
- [ ] Query decomposition
- [ ] Context accumulation
- [ ] Maximum 3 hops
- [ ] Latency budget <800ms

**Acceptance Criteria:**
- Multi-part questions decomposed
- Iterative retrieval improves answer quality
- Context accumulated across hops
- Maximum 3 hops enforced
- p99 latency <800ms for 3-hop queries

**Phase 3 Acceptance Criteria:**
- ✅ All 5 remaining P0 stories complete
- ✅ Cost tracking operational with budget alerts
- ✅ Retrieval analytics dashboard live
- ✅ Multi-hop retrieval handling complex queries
- ✅ Clarifying questions operational for low-confidence queries
- ✅ 2 P2 stories complete for optimization
- ✅ **NEW:** Prometheus/Grafana monitoring operational for all 4 services
- ✅ **NEW:** Hallucination rate tracking <5%
- ✅ **NEW:** Hybrid search BM25 scoring operational (Story I)
- ✅ **NEW:** Response grounding validation (NLI check) active (Story J)

#### Task 3.6: Story I (Hybrid Search BM25 Enhancement) - 3-4 days **[NEW]**
**Priority:** P1
**Files:** `services/vector-search-service/search/hybrid.go` (modify), `services/vector-search-service/search/bm25_scorer.go` (new), `services/vector-search-service/index/keyword_index.go` (new), `services/vector-search-service/config/hybrid_config.go` (modify), `tests/vector-search/hybrid_bm25_test.go` (new)

**User Story:** As a system, I want to combine semantic and keyword search, so that retrieval accuracy improves for exact curriculum queries.

**Status:** ALREADY IMPLEMENTED - Dense ScaNN + sparse GIN + RRF fusion with k=60 exists in `services/vector-search-service/search/hybrid.go`
**Enhancement needed:** Add BM25 scoring for keyword ranking (currently using GIN intersection only)

**Tasks:**
- [ ] Implement BM25 scorer
  - Location: `services/vector-search-service/search/bm25_scorer.go`
  - Standard BM25 formula: `score(q,d) = Σ IDF(qi) * (f(qi,d) * (k1+1)) / (f(qi,d) + k1*(1-b+b*|d|/avgdl))`
  - Configurable parameters: `k1=1.2` (term frequency saturation), `b=0.75` (length normalization)
  - Pre-compute corpus statistics: total docs, avg doc length, term frequencies
- [ ] Create keyword index for BM25
  - Location: `services/vector-search-service/index/keyword_index.go`
  - Inverted index: `term → [doc_id: frequency, doc_length]`
  - Build index during ingestion (async, non-blocking)
  - Incremental updates for new chunks
  - Store in PostgreSQL: `bm25_index(term TEXT PK, postings JSONB, doc_count INT, total_terms INT)`
- [ ] Integrate BM25 into hybrid search pipeline
  - Location: `services/vector-search-service/search/hybrid.go`
  - Current: `Dense ScaNN + Sparse GIN + RRF fusion (k=60)`
  - Enhanced: `Dense ScaNN + Sparse GIN + BM25 + RRF fusion (k=60)`
  - RRF formula unchanged: `RRF(d) = Σ 1/(k + rank_d)` across dense, sparse, BM25
  - Add BM25 as third ranking signal in RRF fusion
  - Update hybrid config to include BM25 weight (default: dense=0.6, sparse=0.2, BM25=0.2)
- [ ] Update hybrid search configuration
  - Location: `services/vector-search-service/config/hybrid_config.go`
  - Add `BM25Weight float64` (default 0.2)
  - Add `BM25Enabled bool` (default true)
  - Add BM25 parameters: `K1 float64`, `B float64`
  - Validate weights sum to 1.0
- [ ] Create comprehensive tests for hybrid + BM25
  - Location: `tests/vector-search/hybrid_bm25_test.go`
  - Test exact keyword queries (e.g., "Newton's second law formula")
  - Test semantic queries (e.g., "what makes things move")
  - Test mixed queries (e.g., "solve 2x + 5 = 15")
  - Measure top-3 retrieval accuracy improvement vs current hybrid (target: +10-15%)
  - Benchmark BM25 scoring latency (target: <50ms for 10K chunks)

**Acceptance Criteria:**
1. BM25 scoring implemented with standard formula (Okapi BM25)
2. Keyword index built and maintained during ingestion
3. BM25 integrated as third signal in RRF fusion (alongside dense + sparse)
4. Hybrid search config supports BM25 weight tuning
5. Top-3 retrieval accuracy improves by 10-15% for exact curriculum queries
6. BM25 scoring latency <50ms for 10K chunks (p99)
7. 90%+ test coverage for BM25 module
8. Backward compatible — BM25 can be disabled via config flag

**Dependencies:** Hybrid Search (US-017, complete), RRF Fusion (US-020, complete), Ingestion Pipeline (adds BM25 index updates)

#### Task 3.7: Story J (Response Grounding Validation - NLI Check) - 4-5 days **[NEW]**
**Priority:** P0
**Files:** `orchestrator/quality/grounding_validator.go` (new), `services/api-gateway/handler/grounding_handler.go` (new), `services/query-understanding-service/prompt/nli_templates.go` (new), `orchestrator/observability/metrics.go` (modify), `schema/grounding_validation.sql` (new)

**User Story:** As a system, I want to validate that AI responses are grounded in retrieved content, so that hallucinations are minimized.

**Gap:** No hallucination detection. LLM can generate facts not in retrieved chunks.

**Tasks:**
- [ ] Implement NLI-based grounding validation
  - Location: `orchestrator/quality/grounding_validator.go`
  - Use LLM as NLI classifier: Given (premise=retrieved_chunks, hypothesis=llm_response), classify as:
    - `entailment` (response fully supported by context)
    - `neutral` (response adds reasonable inference but not contradicted)
    - `contradiction` (response contradicts context)
  - Prompt template: `"Given the following context, does the statement follow? Respond with: entailment/neutral/contradiction + confidence (0-1)"`
  - Per-claim validation: Split response into individual claims, validate each against retrieved context
  - Aggregate grounding score: `entailment_count / total_claims`
- [ ] Create grounding validation endpoint
  - Location: `services/api-gateway/handler/grounding_handler.go`
  - POST `/api/validate/grounding` with `{query, retrieved_chunks[], llm_response}`
  - Returns: `{grounding_score, ungrounded_claims[], classification, confidence}`
  - Async validation (non-blocking, <200ms latency budget)
  - If `grounding_score < 0.7`, flag response for review or truncate ungrounded claims
- [ ] Implement claim extraction
  - Location: `services/query-understanding-service/prompt/nli_templates.go`
  - Use LLM to extract factual claims from response: `"Extract all factual claims from this answer as a list..."`
  - Example: `"Photosynthesis converts sunlight to energy"` → claim to validate
  - Handle numerical claims, definitional claims, procedural claims separately
  - Maximum 10 claims per response (to control latency)
- [ ] Add grounding metrics and alerting
  - Location: `orchestrator/observability/metrics.go`
  - Prometheus metrics:
    - `grounding_score` (histogram: 0.0-1.0, buckets: 0.0-0.3, 0.3-0.6, 0.6-0.8, 0.8-1.0)
    - `ungrounded_claims_total` (counter)
    - `grounding_validation_latency_ms` (histogram)
    - `response_truncation_rate` (gauge: responses truncated / total validated)
  - Alert when `grounding_score < 0.7` for >10% of responses over 1-hour window
- [ ] Create grounding validation database schema
  - Location: `schema/grounding_validation.sql`
  - Table: `grounding_checks(id, query_id, response_id, grounding_score FLOAT, ungrounded_claims JSONB, classification VARCHAR, timestamp)`
  - Table: `ungrounded_claims(id, grounding_check_id FK, claim_text, retrieved_context_match BOOLEAN)`
  - Index on `(grounding_score, timestamp)` for trend analysis
  - Retention: 90 days (auto-cleanup older records)
- [ ] Implement response truncation/fallback
  - When `grounding_score < 0.7`:
    - Identify ungrounded claims
    - Truncate or mark ungrounded sections: `[Note: This claim is not supported by your curriculum materials]`
    - Log to quality team dashboard (AM-03)
    - If >50% ungrounded, trigger fallback to general LLM response with disclaimer

**Acceptance Criteria:**
1. NLI check validates every LLM response against retrieved chunks (sample rate: 100% for first 30 days, then 20% ongoing)
2. Response split into factual claims, each validated against context
3. Grounding score calculated (0.0-1.0, target: >0.8 for production responses)
4. Ungrounded claims flagged and returned in response metadata
5. Responses with `grounding_score < 0.7` truncated or flagged with disclaimer
6. Grounding validation latency <200ms (async, non-blocking)
7. Prometheus metrics exported: `grounding_score`, `ungrounded_claims_total`, `response_truncation_rate`
8. Grafana alert when grounding score <0.7 for >10% of responses
9. 90%+ test coverage for grounding validation module
10. Hallucination rate reduced by 50%+ after grounding validation enabled

**Dependencies:** RAG Handler (complete), Retrieved Context (complete), LLM Client (complete), Quality Monitoring (AM-03)

**Phase 3 Acceptance Criteria:**
- ✅ All 5 remaining P0 stories complete
- ✅ Cost tracking operational with budget alerts
- ✅ Retrieval analytics dashboard live
- ✅ Multi-hop retrieval handling complex queries
- ✅ Clarifying questions operational for low-confidence queries
- ✅ 2 P2 stories complete for optimization
- ✅ **NEW:** Prometheus/Grafana monitoring operational for all 4 services
- ✅ **NEW:** Hallucination rate tracking <5%
- ✅ **NEW:** Hybrid search BM25 scoring operational (Story I)
- ✅ **NEW:** Response grounding validation (NLI check) active (Story J)
- ✅ **NEW:** 95% accuracy SLA monitoring active

---

## Analytics & Monitoring Implementation Plan

**Goal:** Implement comprehensive monitoring, feedback tracking, and quality metrics for production validation
**Owner:** Engineering Lead
**Estimated Effort:** 5-7 days (parallel execution across 3 waves)
**Dependencies:** Feedback endpoint (complete), Query Understanding Service (complete), Service client metrics (partial)

### What's Already Done ✅

| Component | Location | Status |
|-----------|----------|--------|
| Thumbs up/down feedback endpoint | `services/api-gateway/handler/feedback_handler.go` | ✅ Complete |
| Service client metrics (request count, error count, latency) | `services/api-gateway/client/service_client.go` | ✅ Complete |
| CAG metrics (cache hits/misses, cost saved) | `services/api-gateway/cache/cag_orchestrator.go` | ✅ Complete |
| Circuit breaker stats | `services/api-gateway/client/service_client.go` | ✅ Complete |
| Stats endpoint | `services/api-gateway/handler/stats_handler.go` | ✅ Complete |

### What Needs to Be Built 🚧

| Component | Effort | Priority |
|-----------|--------|----------|
| Prometheus metrics exporter for all 4 services | 2-3 days | P0 |
| Grafana dashboard (API latency, error rates, cache hit rates, feedback scores) | 1-2 days | P0 |
| Hallucination rate tracker (LLM judge comparison) | 1-2 days | P1 |
| Per-query feedback aggregation dashboard | 1 day | P1 |
| 95% accuracy SLA monitoring | 1 day | P0 |

### Implementation Tasks

#### Task AM-01: Prometheus Metrics Exporter Setup (2-3 days)
**Wave:** 1 (can run in parallel with AM-02)
**Files:** 
- `services/api-gateway/metrics/prometheus.go` (new)
- `services/api-gateway/cmd/server/main.go` (modify - add metrics endpoint)
- `services/api-gateway/middleware/metrics.go` (new - HTTP middleware)
- `orchestrator/observability/prometheus.go` (new)
- `services/query-understanding-service/metrics/prometheus.go` (new)
- `ingestion/metrics/prometheus.go` (new)
- `prometheus/prometheus.yml` (new - scrape config)

**Tasks:**
- [ ] Create Prometheus metrics registry for API Gateway
- [ ] Add HTTP middleware to capture request duration, status codes, endpoints
- [ ] Export existing service client metrics to Prometheus format
- [ ] Export CAG orchestrator metrics (cache hits/misses) to Prometheus
- [ ] Create `/metrics` endpoint for all 4 services
- [ ] Configure Prometheus scrape targets
- [ ] Add Go client library (`github.com/prometheus/client_golang`)

**Metrics to Export:**
```go
// API Gateway Metrics
var (
    apiRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "api_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
        },
        []string{"method", "endpoint", "status"},
    )
    apiRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "api_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    apiErrorsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "api_errors_total",
            Help: "Total HTTP errors",
        },
        []string{"method", "endpoint", "error_type"},
    )
    cacheHitRate = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "cache_hit_rate",
            Help: "Cache hit rate (0-1)",
        },
        []string{"cache_type"},
    )
)

// Feedback Metrics
var (
    feedbackThumbsUp = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "feedback_thumbs_up_total",
            Help: "Total thumbs up feedback",
        },
        []string{"query_id", "grade", "subject"},
    )
    feedbackThumbsDown = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "feedback_thumbs_down_total",
            Help: "Total thumbs down feedback",
        },
        []string{"query_id", "grade", "subject", "reason"},
    )
)

// Quality Metrics
var (
    hallucinationRate = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "hallucination_rate",
            Help: "Hallucination rate (0-1)",
        },
    )
    qualityScore = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "quality_score",
            Help: "Query quality score (0-100)",
        },
    )
)
```

**Acceptance Criteria:**
1. All 4 services expose `/metrics` endpoint
2. Prometheus successfully scraping all targets
3. Metrics visible in Prometheus UI
4. Request duration histogram capturing p50/p95/p99
5. Feedback counters incrementing correctly
6. Go tests for metrics middleware (90%+ coverage)

---

#### Task AM-02: Grafana Dashboard Setup (1-2 days)
**Wave:** 1 (can run in parallel with AM-01)
**Files:**
- `observability/grafana/dashboards/quality.json` (new)
- `observability/grafana/dashboards/retrieval.json` (new)
- `observability/grafana/dashboards/cost.json` (new)
- `observability/grafana/datasources/prometheus.yml` (new)
- `docker-compose.yml` (modify - add Grafana service)

**Tasks:**
- [ ] Add Grafana service to docker-compose.yml
- [ ] Configure Prometheus as data source
- [ ] Create Quality Dashboard JSON (feedback, hallucination, SLA)
- [ ] Create Retrieval Dashboard JSON (latency, cache, query patterns)
- [ ] Create Cost Dashboard JSON (token usage, cost per query, budget)
- [ ] Import dashboards via Grafana provisioning
- [ ] Set up dashboard variables (grade, subject, time range)

**Dashboard Specifications:**

**Quality Dashboard Panels:**
1. **API Latency Heatmap** - p50/p95/p99 over time
2. **Request Rate & Error Rate** - dual axis line chart
3. **Cache Hit Rate** - gauge (target >60%)
4. **Feedback Score** - pie chart (thumbs up vs down ratio)
5. **Hallucination Rate** - time series (target <5%)
6. **Quality Score Trend** - time series (0-100 scale)
7. **95% SLA Compliance** - stat panel (percentage of queries meeting SLA)

**Retrieval Dashboard Panels:**
1. **Retrieval Latency Breakdown** - stacked bar (embed + search + rank)
2. **Most Retrieved Chunks** - table (top 20)
3. **Least Retrieved Chunks** - table (bottom 20)
4. **Query Pattern Distribution** - pie chart (by intent)
5. **Taxonomy Coverage** - heatmap (grade x subject)
6. **Cache Hit Rate by Type** - multi-line chart

**Cost Dashboard Panels:**
1. **Daily Cost** - time series (USD)
2. **Cost per Query** - stat panel (average USD)
3. **Token Usage** - stacked bar (embedding vs generation)
4. **Budget vs Actual** - gauge (percentage of budget used)
5. **Cost by Grade/Subject** - pie chart

**Acceptance Criteria:**
1. Grafana service running via docker-compose
2. Prometheus configured as data source
3. 3 dashboards imported and accessible
4. All panels rendering with live data
5. Dashboard variables working (filter by grade, subject, time)
6. Dashboard JSON files version controlled

---

#### Task AM-03: Hallucination Rate Tracker (1-2 days)
**Wave:** 2 (depends on AM-01 completion)
**Files:**
- `services/api-gateway/quality/hallucination_detector.go` (new)
- `quality_loop/judge.py` (modify - expose as API or integrate)
- `services/api-gateway/handler/feedback_handler.go` (modify - add hallucination logging)

**Tasks:**
- [ ] Implement LLM judge comparison logic
- [ ] Compare LLM answers against retrieved context for factual consistency
- [ ] Calculate hallucination rate per query
- [ ] Export hallucination rate to Prometheus
- [ ] Integrate with existing `quality_loop/judge.py` or reimplement in Go
- [ ] Add hallucination alerts to Grafana

**Hallucination Detection Logic:**
```go
type HallucinationDetector struct {
    llmClient LLMClient
    threshold float64 // e.g., 0.7 confidence
}

func (d *HallucinationDetector) DetectHallucination(query string, retrievedContext []string, llmAnswer string) (bool, float64, error) {
    // Use LLM judge to compare answer against context
    prompt := fmt.Sprintf(`
    Given the following retrieved context and LLM-generated answer,
    determine if the answer contains information NOT supported by the context.
    
    Context: %v
    Answer: %s
    
    Respond with:
    - "hallucination" if answer contains unsupported claims
    - "grounded" if answer is fully supported by context
    - Confidence score (0-1)
    `, retrievedContext, llmAnswer)
    
    response, err := d.llmClient.Generate(prompt)
    // Parse response and return result
}
```

**Acceptance Criteria:**
1. Hallucination detection working for sampled queries (10% sample rate)
2. Hallucination rate exported to Prometheus
3. Grafana alert when hallucination rate >5%
4. Hallucination logs include query, context, answer for review
5. Detection latency <200ms (async, non-blocking)

---

#### Task AM-04: Per-Query Feedback Aggregation (1 day)
**Wave:** 2 (depends on AM-01 completion)
**Files:**
- `services/api-gateway/handler/feedback_handler.go` (modify - enhance existing)
- `services/api-gateway/aggregator/feedback_aggregator.go` (new)
- `schema/feedback_aggregation.sql` (new)

**Tasks:**
- [ ] Enhance existing feedback endpoint to store feedback in database
- [ ] Create feedback aggregation table (query_id, thumbs_up/down, grade, subject, timestamp)
- [ ] Implement aggregation queries (feedback rate by grade/subject/time)
- [ ] Export aggregated metrics to Prometheus
- [ ] Add feedback trend analysis

**Database Schema:**
```sql
CREATE TABLE feedback_aggregates (
    id SERIAL PRIMARY KEY,
    query_id UUID NOT NULL,
    user_id UUID,
    grade INT,
    subject VARCHAR(50),
    feedback_type VARCHAR(10) CHECK (feedback_type IN ('thumbs_up', 'thumbs_down')),
    reason VARCHAR(255), -- optional: 'irrelevant', 'incorrect', 'incomplete'
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (query_id) REFERENCES query_logs(id)
);

CREATE INDEX idx_feedback_grade_subject ON feedback_aggregates(grade, subject);
CREATE INDEX idx_feedback_timestamp ON feedback_aggregates(timestamp);
```

**Acceptance Criteria:**
1. Feedback endpoint storing all thumbs up/down in database
2. Aggregation queries returning feedback rates by grade/subject
3. Prometheus metrics updating correctly
4. Grafana feedback panel showing real-time feedback scores
5. Feedback trend analysis working (daily/weekly trends)

---

#### Task AM-05: 95% Accuracy SLA Monitoring (1 day)
**Wave:** 3 (depends on AM-01, AM-02, AM-04 completion)
**Files:**
- `services/api-gateway/monitoring/sla_monitor.go` (new)
- `observability/grafana/alerts/sla_alerts.yml` (new)

**Tasks:**
- [ ] Define 95% accuracy SLA metrics (hallucination rate <5%, feedback score >90%)
- [ ] Implement SLA monitoring service
- [ ] Configure Grafana alerts for SLA violations
- [ ] Add Slack/email notifications for SLA breaches
- [ ] Create SLA compliance report (daily/weekly)

**SLA Definition:**
```
95% Accuracy SLA = 
  - Hallucination rate < 5% (measured over 1-hour window)
  - Feedback thumbs up rate > 90% (measured over 24-hour window)
  - API p99 latency < 500ms (measured over 1-hour window)
  - Error rate < 1% (measured over 1-hour window)
```

**Grafana Alert Rules:**
```yaml
groups:
  - name: sla_alerts
    rules:
      - alert: HighHallucinationRate
        expr: hallucination_rate > 0.05
        for: 1h
        labels:
          severity: critical
        annotations:
          summary: "Hallucination rate exceeds 5% SLA threshold"
          
      - alert: LowFeedbackScore
        expr: feedback_thumbs_up_total / (feedback_thumbs_up_total + feedback_thumbs_down_total) < 0.90
        for: 24h
        labels:
          severity: warning
        annotations:
          summary: "Feedback thumbs up rate below 90% SLA threshold"
          
      - alert: HighLatency
        expr: histogram_quantile(0.99, api_request_duration_seconds) > 0.5
        for: 1h
        labels:
          severity: critical
        annotations:
          summary: "p99 latency exceeds 500ms SLA threshold"
```

**Acceptance Criteria:**
1. SLA monitoring service running
2. Grafana alerts configured for all 4 SLA metrics
3. Notifications sent on SLA violation
4. SLA compliance report generated daily
5. SLA dashboard showing current compliance status

---

### Execution Waves

| Wave | Tasks | Duration | Dependencies |
|------|-------|----------|--------------|
| **Wave 1** | AM-01 (Prometheus), AM-02 (Grafana) | 2-3 days | None (parallel) |
| **Wave 2** | AM-03 (Hallucination), AM-04 (Feedback Aggregation) | 1-2 days | AM-01 complete |
| **Wave 3** | AM-05 (SLA Monitoring) | 1 day | AM-01, AM-02, AM-04 complete |

**Total Estimated Time:** 5-7 days
**Total Tasks:** 5
**Waves:** 3

### File Ownership Map

| File | Service | Owner |
|------|---------|-------|
| `services/api-gateway/metrics/prometheus.go` | API Gateway | Engineering |
| `services/api-gateway/middleware/metrics.go` | API Gateway | Engineering |
| `services/api-gateway/quality/hallucination_detector.go` | API Gateway | Engineering |
| `services/api-gateway/aggregator/feedback_aggregator.go` | API Gateway | Engineering |
| `services/api-gateway/monitoring/sla_monitor.go` | API Gateway | Engineering |
| `orchestrator/observability/prometheus.go` | Orchestrator | Engineering |
| `services/query-understanding-service/metrics/prometheus.go` | Query Understanding | Engineering |
| `ingestion/metrics/prometheus.go` | Ingestion | Engineering |
| `observability/grafana/dashboards/*.json` | Grafana | Engineering |
| `prometheus/prometheus.yml` | Prometheus | Engineering |

### Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Prometheus metrics exported | 10+ metrics across 4 services | Prometheus UI |
| Grafana dashboards | 3 dashboards, 15+ panels | Grafana UI |
| Hallucination rate | <5% | Grafana panel |
| Feedback score | >90% thumbs up | Grafana panel |
| API p99 latency | <500ms | Grafana panel |
| SLA compliance | 95%+ | SLA dashboard |
| Alert response time | <5 minutes | Alert history |

#### Task 3.4: Complete US-045 (Clarifying Questions) + Story E (Confidence Scores) - 5-6 days **[NEW]**
**Priority:** P1 (improves answer quality, not blocking)
**Files:** `services/query-understanding-service/classifier/intent.go` (modify), `services/query-understanding-service/handler/clarification.go` (new), `services/api-gateway/handler/clarification_handler.go` (new), `frontend/src/components/Chat/ClarificationPrompt.tsx` (new), `orchestrator/handler/rag_handler_real.go` (modify), `orchestrator/observability/metrics.go` (modify)

**Enhanced Scope:** Now includes token-logprob-based confidence scoring for LLM responses (Story E: Confidence Scores on AI Responses).

**User Story:** As a student, I want the AI to ask clarifying questions when my query is unclear, so that I get accurate answers.

**Tasks:**
- [ ] Extend intent classifier to return confidence scores with all 9 intent categories
  - Location: `services/query-understanding-service/classifier/intent.go`
  - Add confidence threshold check (<0.7 triggers clarification flow)
- [ ] Implement clarification option generator (2-3 clarifying options)
  - Create `services/query-understanding-service/handler/clarification.go`
  - Generate contextually relevant clarification options based on intent ambiguity
- [ ] Add clarification endpoint to API Gateway
  - Create `services/api-gateway/handler/clarification_handler.go`
  - Handle clarification response and re-process with appended context
- [ ] Implement clarification UI in React frontend
  - Create `frontend/src/components/Chat/ClarificationPrompt.tsx`
  - Display 2-3 options, capture user selection, re-submit query
- [ ] Write integration tests for full clarification flow
  - Test low-confidence query → clarification → user selection → improved answer
- [ ] **NEW (Story E):** Implement token-logprob-based confidence scoring for LLM responses
  - Location: `orchestrator/handler/rag_handler_real.go`
  - Extract log probabilities from Vertex AI response (`response.candidates[0].avg_logprobs`)
  - Calculate response confidence: `exp(avg_logprobs)` normalized to 0.0-1.0 range
  - Attach confidence score to every response in metadata
- [ ] **NEW (Story E):** Low-confidence flagging system
  - Location: `orchestrator/observability/metrics.go`
  - Responses with confidence <0.6 flagged as "low_confidence" in backend logs
  - Prometheus metric: `response_confidence_score` (histogram with buckets: 0.0-0.3, 0.3-0.6, 0.6-0.8, 0.8-1.0)
  - Track `low_confidence_rate` = responses with confidence <0.6 / total responses
- [ ] **NEW (Story E):** Response confidence in API response payload
  - Add `confidence` field to response JSON: `{answer: "...", confidence: 0.87, metadata: {...}}`
  - Frontend can use for visual indicators (optional future enhancement)

**Acceptance Criteria:**
1. If query intent confidence <70%, system must NOT attempt an answer
2. System presents 2-3 clarifying options to the user
3. User's choice is appended to original query for new processing cycle
4. Clarification flow integrates seamlessly with API Gateway
5. Frontend UI displays clarification prompts cleanly
6. **NEW (Story E):** Every LLM response includes confidence score (0.0-1.0) based on token logprobs
7. **NEW (Story E):** Responses with confidence <0.6 flagged as "low_confidence" in backend logs
8. **NEW (Story E):** `response_confidence_score` histogram exported to Prometheus
9. **NEW (Story E):** `low_confidence_rate` tracked and alertable in Grafana
10. **NEW (Story E):** Confidence score included in API response payload

**Dependencies:** Query Understanding Service (complete), Intent classifier (complete), API Gateway (complete)

#### Task 3.5: Story F (Query Intent Classification Remapping) - 3-4 days **[NEW]**
**Priority:** P1
**Files:** `services/query-understanding-service/classifier/intent.go` (modify), `services/query-understanding-service/classifier/intent_mapping.go` (new), `services/query-understanding-service/prompts/intent_v2.go` (new), `tests/intent_classifier_test.go` (new), `eval/intent_test_set.json` (new)

**User Story:** As a system, I want to classify user queries (explain, solve, doubt, concept), so that the correct pipeline and prompt strategy is applied.

**Gap Analysis:** Current classifier has 9 categories (information_seeking, problem_solving, content_retrieval, clarification, comparison, definition, procedural, conversational, exam_prep) but NOT mapped to the 4 required categories: **Concept**, **Numerical**, **Fact-check**, **Strategy**.

**Tasks:**
- [ ] Define 4-category intent taxonomy
  - Location: `services/query-understanding-service/classifier/intent.go`
  - New intents: `Concept` (explain, understand, what/why/how), `Numerical` (solve, calculate, math problems), `Fact-check` (verify, true/false, is it correct), `Strategy` (approach, method, best way, compare)
  - Keep backward compatibility — old 9 categories map to new 4
- [ ] Create category remapping layer
  - Location: `services/query-understanding-service/classifier/intent_mapping.go`
  - Map old 9 → new 4:
    - `information_seeking` + `definition` + `clarification` → **Concept**
    - `problem_solving` + `exam_prep` (math questions) → **Numerical**
    - `comparison` + `content_retrieval` (verification queries) → **Fact-check**
    - `procedural` + `comparison` (strategy queries) → **Strategy**
    - `conversational` → Route based on context or default to **Concept**
- [ ] Update prompt templates for 4-category classification
  - Location: `services/query-understanding-service/prompts/intent_v2.go`
  - New prompt explicitly asks LLM to classify into 4 categories with examples
  - Include edge case handling (multi-intent queries → primary + secondary)
- [ ] Create 100-query test set for precision evaluation
  - Location: `eval/intent_test_set.json`
  - 25 queries per category (Concept, Numerical, Fact-check, Strategy)
  - Each query labeled with ground truth intent
  - Mix of easy, ambiguous, and multi-intent queries
- [ ] Implement precision evaluation script
  - Location: `tests/intent_classifier_test.go`
  - Run all 100 queries through classifier
  - Calculate precision per category (target: >90%)
  - Calculate overall accuracy (target: >90%)
  - Report confusion matrix
- [ ] Tune classifier for >90% precision
  - Adjust prompt examples based on test failures
  - Add few-shot examples for ambiguous cases
  - Implement confidence threshold per category

**Acceptance Criteria:**
1. Intent classifier returns one of 4 categories: Concept, Numerical, Fact-check, Strategy
2. Old 9-category system still works (backward compatible, auto-mapped)
3. Precision >90% on 100-query test set (measured per category)
4. Overall accuracy >90% on 100-query test set
5. Confusion matrix available for analysis
6. Multi-intent queries return primary + secondary intent
7. Tests run in CI pipeline — regression detection on prompt changes

**Dependencies:** Query Understanding Service (complete), LLM client (complete)

---

### Phase 4: Curriculum Features (Weeks 11-16) - E-002 P1/P2 Stories

**Goal:** Complete curriculum and learning management features  
**Owner:** Product Lead  
**Team:** 3-4 engineers  
**Stories:** 20 stories from E-002

#### Priority Stories

| Story ID | Title | Priority |
|----------|-------|----------|
| US-059 | RAG Filtering by Curriculum | P1 |
| US-061 | Next-Concept Recommendation | P1 |
| US-063 | Progress Tracking | P1 |
| US-051 | Weak Area Detection | P2 |
| US-052 | Key Formulas Extraction | P2 |
| US-055 | Difficulty Tagging | P2 |
| US-056 | Misconception Detection | P2 |
| US-057 | Versioning | P2 |
| US-058 | Curriculum API | P2 |
| US-060 | State Machine | P2 |

**Phase 4 Acceptance Criteria:**
- ✅ Curriculum-aligned retrieval (US-059)
- ✅ Next-concept recommendations working (US-061)
- ✅ Progress tracking for students (US-063)
- ✅ 10+ E-002 stories complete

---

### Phase 5: User Experience (Weeks 17-22) - E-003 Stories

**Goal:** Complete user experience and engagement features  
**Owner:** Product Lead  
**Team:** 2-3 engineers  
**Stories:** 16 stories from E-003

#### Priority Stories

| Story ID | Title | Priority |
|----------|-------|----------|
| US-004 | Onboarding Recovery | P2 |
| US-005 | Offline Handling | P2 |
| US-006 | Consent Management | P2 |
| US-007 | Lesson Generation Fallback | P2 |
| US-008 | Gamification | P3 |
| US-009 | Notifications | P3 |
| US-010 | Social Sharing | P3 |
| US-011 | Bookmarks | P3 |
| US-012 | Notes | P3 |
| US-013 | Highlights | P3 |

**Phase 5 Acceptance Criteria:**
- ✅ Core UX features complete (US-004 to US-007)
- ✅ Engagement features started (US-008 to US-018)

---

### Phase 6: Enterprise (Weeks 23-28) - Advanced Features

**Goal:** Complete advanced learning and enterprise features  
**Owner:** Product Lead  
**Team:** 3-4 engineers  
**Stories:** 10+ stories from E-002 and E-004

#### Stories

| Story ID | Title | Epic |
|----------|-------|------|
| US-064 | Mastery Assessment | E-002 |
| US-065 | Adaptive Difficulty | E-002 |
| US-066 | Prerequisite Warnings | E-002 |
| US-067 | Search Curriculum | E-002 |
| US-068 | Multilingual Support | E-002 |
| US-069 | Learning Objectives API | E-002 |
| US-070 | Teacher Dashboard | E-002 |
| US-071 | Student Dashboard | E-002 |
| US-072 | Parent Reports | E-002 |
| US-073 | LMS Integration | E-002 |

**Phase 6 Acceptance Criteria:**
- ✅ All advanced learning features complete
- ✅ Teacher/student/parent dashboards live
- ✅ Multilingual support (Hindi/English)
- ✅ LMS integration working

---

## Updated Success Metrics

| Metric | Current (Phase 1) | Phase 2 | Phase 3 | Phase 4 | Phase 5 | Phase 6 (Target) |
|--------|------------------|---------|---------|---------|---------|------------------|
| **User Stories Complete** | 31/79 (39.2%) | 43/79 (54%) | 54/79 (68%) | 64/79 (81%) | 70/79 (89%) | 79/79 (100%) |
| **Production Readiness** | 50.6% | 63% | 74% | 84% | 89% | 95%+ |
| **MVP Ready** | ✅ Week 3 | - | - | - | - | - |
| **Production Ready** | ❌ | ❌ | ⚠️ Week 13 | ✅ Week 19 | ✅ | ✅ |

---

## Execution Order

1. **Phase 1 (MVP)** - Weeks 1-3 - **COMPLETE P0 STORIES** *(updated: +1 week for US-013 prod + US-018 WebSocket, Story C adds +1 day)*
2. **Phase 2 (Production)** - Weeks 4-8 - **COMPLETE P1 STORIES** *(updated: +1 week for Story D: Re-explain)*
3. **Phase 3 (Scale)** - Weeks 9-13 - **COMPLETE REMAINING P0 + P2** *(updated: +2 weeks for Story E + Story F)*
4. **Phase 4 (Curriculum)** - Weeks 14-19 - **E-002 FEATURES**
5. **Phase 5 (UX)** - Weeks 20-25 - **E-003 FEATURES**
6. **Phase 6 (Enterprise)** - Weeks 26-31 - **ADVANCED FEATURES**

**Total Duration:** 31 weeks (7.75 months) with team of 3-4 engineers

**MVP Deployment:** Week 3 (after Phase 1)
**Production Deployment:** Week 13-19 (after Phase 3-4)
**Enterprise Ready:** Week 31 (after Phase 6)

---

## Resource Requirements

| Phase | Engineers | Duration | Total Engineer-Weeks |
|-------|-----------|----------|---------------------|
| Phase 1 (MVP) | 2-3 | 3 weeks | 6-9 |
| Phase 2 (Production) | 3-4 | 4 weeks | 12-16 |
| Phase 3 (Scale) | 3-4 | 4 weeks | 12-16 |
| Phase 4 (Curriculum) | 3-4 | 6 weeks | 18-24 |
| Phase 5 (UX) | 2-3 | 6 weeks | 12-18 |
| Phase 6 (Enterprise) | 3-4 | 6 weeks | 18-24 |
| **TOTAL** | **3-4 avg** | **29 weeks** | **78-107** |

---

## Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| US-019 RBAC delays MVP | Critical | Prioritize Week 1, security review |
| US-013 GCP deployment delays | Critical | Provision GCP account Week 1, use Terraform modules |
| US-018 WebSocket integration complexity | High | Start with SSE fallback, iterate to WebSocket |
| Team capacity shortage | High | Defer P2/P3 stories, focus on P0/P1 |
| Technical debt accumulation | Medium | Code review for all changes |
| Production incidents | High | Feature flags, canary deployments |
| Cost overruns | Medium | Budget alerts from Phase 3 |
| Timeline slippage | Medium | Weekly progress reviews |
| **NEW:** Prometheus/Grafana setup delays | Medium | Use docker-compose, pre-built dashboards |
| **NEW:** Hallucination detection accuracy | Medium | Start with 10% sampling, tune threshold |
| **NEW:** SLA alert false positives | Low | Tune alert thresholds, add grace period |
| **NEW (G):** Cross-agent context consistency | High | Use optimistic locking, Redis-backed state, comprehensive tests |
| **NEW (H):** Enrichment latency blowout | Medium | Async profile fetch, <50ms budget, fallback to generic prompt |
| **NEW (I):** BM25 index build time | Medium | Build during ingestion (async), incremental updates only |
| **NEW (J):** NLI validation latency | High | Async validation (<200ms), sample rate 100% → 20% after 30 days |

---

## Commit Strategy

- **Per-story commits**: Each user story completed and tested before moving to next
- **Feature branches**: `phase-X/US-YYY-story-name`
- **Security reviews**: All Phase 1 changes require security team review
- **Integration commits**: After each phase, verify end-to-end flows
- **TDD workflow**: Tests written before implementation

---

## Next Steps

1. **Immediate (This Week - Phase 1):**
   - Review and approve this user-story-driven plan
   - Assemble Phase 1 team (2-3 engineers)
   - Begin Task 1.1: US-019 RBAC implementation
   - **NEW:** Provision GCP account and configure Terraform state backend (blocks US-013)
   - Set up project tracking by user story

2. **Week 1-3 (Phase 1):**
   - Complete US-019 (JWT RBAC)
   - Complete US-021 (Contextual Grounding)
   - **NEW:** Complete US-013 (PostgreSQL + Vertex AI Production Infra) — 5-7 days
   - **NEW:** Complete US-018 (WebSocket Streaming) — 3-4 days
   - MVP integration testing
   - **MVP deployment ready by April 22, 2026** *(updated from April 15)*

3. **Week 4-7 (Phase 2):**
   - Begin Task 2.1: US-023 Auto Dataset
   - Begin Task 2.2: US-035 Incremental Ingestion
   - Complete distributed tracing
   - Quality monitoring dashboard
   - **NEW:** Begin US-041 enhanced scope (Prometheus/Grafana setup) — Task AM-01, AM-02

4. **Week 8-11 (Phase 3):**
   - Complete remaining P0 stories (US-040, US-042, US-043, US-044)
   - **NEW:** Complete US-045 (Clarifying Questions) — 4-5 days
   - Complete P2 optimization stories
   - **NEW:** Complete Analytics & Monitoring tasks (AM-03, AM-04, AM-05)
   - **NEW:** Prometheus/Grafana dashboards operational
   - **NEW:** Hallucination rate tracking active

---

**CRITICAL REMINDER:** Do not deploy to production until Phase 1 (MVP) is 100% complete, especially US-014 RBAC for security.

**Previous Version:** 3.3 - 6 New User Stories Integration (April 7, 2026)
**Current Version:** 3.4 - 4 New User Stories Integration (G-J) (April 7, 2026)
**Changes:**
- Integrated Story G (Cross-Agent Context Sharing) → NEW Task 1.6 (Phase 1, P1)
- Integrated Story H (Query Enrichment) → NEW Task 2.6 (Phase 2, P1)
- Integrated Story I (Hybrid Search BM25 Enhancement) → NEW Task 3.6 (Phase 3, P1)
- Integrated Story J (Response Grounding Validation/NLI) → NEW Task 3.7 (Phase 3, P0)
- Updated executive summary: 83 total stories (+4 from G-J), completion % adjusted to 44.0%
- Updated Phase 1/2/3 with new tasks and acceptance criteria
- Updated timeline: total extended from 31 → 33 weeks (+2 weeks for Stories G, H, I, J)
- Updated dependency graph with all 4 new stories
**Next Review:** April 14, 2026 (weekly progress review by user story)

---

## New Stories Integration Summary (v3.4)

### Stories Added/Enhanced

| Story | Mapped To | Phase | Effort | Status |
|-------|-----------|-------|--------|--------|
| **WebSocket Streaming** | US-018 (enhanced) | Phase 1 (P0) | 3-4 days | Task 1.5 |
| **PostgreSQL + Vertex AI Infra** | US-013 (enhanced) | Phase 1 (P0) | 5-7 days | Task 1.4 |
| **Clarifying Questions** | US-045 (enhanced) | Phase 3 (P1) | 5-6 days | Task 3.4 |
| **Retrieval Accuracy & Feedback** | US-041, US-042, US-043 (enhanced) | Phase 2-3 (P1/P0) | 5-7 days | Task AM-01 to AM-05 |
| **Answer Feedback Collection (A)** | US-030 + US-041 (enhanced) | Phase 2 (P1) | Included in Task 2.4 | ✅ Integrated |
| **AI Quality Learning Loop (B)** | US-031 + US-023 (enhanced) | Phase 2 (P1) | Included in Task 2.1 | ✅ Integrated |
| **Fallback Strategy (C)** | US-021 (enhanced) | Phase 1 (P0) | Included in Task 1.2 | ✅ Integrated |
| **Re-explain Concept (D)** | NEW → E-003 | Phase 2 (P1) | 4-5 days | Task 2.5 |
| **Confidence Scores (E)** | US-045 (enhanced) | Phase 3 (P1) | Included in Task 3.4 | ✅ Integrated |
| **Query Intent Classification (F)** | NEW → E-001 | Phase 3 (P1) | 3-4 days | Task 3.5 |
| **Cross-Agent Context Sharing (G)** | NEW → E-001 | Phase 1 (P1) | 3-4 days | Task 1.6 |
| **Query Enrichment (H)** | NEW → E-001 | Phase 2 (P1) | 2-3 days | Task 2.6 |
| **Hybrid Search BM25 (I)** | US-017 + US-020 (enhanced) | Phase 3 (P1) | 3-4 days | Task 3.6 |
| **Response Grounding Validation (J)** | NEW → E-001 | Phase 3 (P0) | 4-5 days | Task 3.7 |

### 10 User Stories Integration Map (v3.4)

| Story | Title | Mapped To | Phase | Task | Effort |
|-------|-------|-----------|-------|------|--------|
| **A** | Answer Feedback Collection | US-030 + US-041 | Phase 2 | Task 2.4 (enhanced) | Included |
| **B** | AI Quality Learning Loop | US-031 + US-023 | Phase 2 | Task 2.1 (enhanced) | +1-2 days |
| **C** | Fallback Strategy | US-021 | Phase 1 | Task 1.2 (enhanced) | +1 day |
| **D** | Re-explain Concept | NEW → E-003 | Phase 2 | Task 2.5 (new) | 4-5 days |
| **E** | Confidence Scores | US-045 | Phase 3 | Task 3.4 (enhanced) | +1-2 days |
| **F** | Query Intent Classification | NEW → E-001 | Phase 3 | Task 3.5 (new) | 3-4 days |
| **G** | Cross-Agent Context Sharing | NEW → E-001 | Phase 1 | Task 1.6 (new) | 3-4 days |
| **H** | Query Enrichment | NEW → E-001 | Phase 2 | Task 2.6 (new) | 2-3 days |
| **I** | Hybrid Search BM25 | US-017 + US-020 | Phase 3 | Task 3.6 (enhanced) | 3-4 days |
| **J** | Response Grounding Validation | NEW → E-001 | Phase 3 | Task 3.7 (new) | 4-5 days |

### Analytics & Monitoring Tasks Added

| Task | Description | Wave | Effort | Dependencies |
|------|-------------|------|--------|--------------|
| **AM-01** | Prometheus Metrics Exporter Setup | 1 | 2-3 days | None |
| **AM-02** | Grafana Dashboard Setup | 1 | 1-2 days | None |
| **AM-03** | Hallucination Rate Tracker | 2 | 1-2 days | AM-01 |
| **AM-04** | Per-Query Feedback Aggregation | 2 | 1 day | AM-01 |
| **AM-05** | 95% Accuracy SLA Monitoring | 3 | 1 day | AM-01, AM-02, AM-04 |

### 6 New User Stories Integration Map

| Story | Title | Mapped To | Phase | Task | Effort |
|-------|-------|-----------|-------|------|--------|
| **A** | Answer Feedback Collection | US-030 + US-041 | Phase 2 | Task 2.4 (enhanced) | Included |
| **B** | AI Quality Learning Loop | US-031 + US-023 | Phase 2 | Task 2.1 (enhanced) | +1-2 days |
| **C** | Fallback Strategy | US-021 | Phase 1 | Task 1.2 (enhanced) | +1 day |
| **D** | Re-explain Concept | NEW → E-003 | Phase 2 | Task 2.5 (new) | 4-5 days |
| **E** | Confidence Scores | US-045 | Phase 3 | Task 3.4 (enhanced) | +1-2 days |
| **F** | Query Intent Classification | NEW → E-001 | Phase 3 | Task 3.5 (new) | 3-4 days |

### Timeline Impact
- **Phase 1:** Extended from 2 → 3 weeks (+1 week for infra + WebSocket). Story C adds +1 day to Task 1.2. Story G adds +3-4 days (Task 1.6).
- **Phase 2:** Extended by +4-5 days for Task 2.5 (Story D: Re-explain). Task 2.1 extended by +1-2 days for Story B. Story H adds +2-3 days (Task 2.6).
- **Phase 3:** Extended by +5-6 days for Task 3.4 (Story E) + 3-4 days for Task 3.5 (Story F). Story I adds +3-4 days (Task 3.6). Story J adds +4-5 days (Task 3.7).
- **Total:** Extended from 31 → 33 weeks (+2 weeks for Stories G, H, I, J)
- **MVP Date:** Moved from April 15 → April 22, 2026 (unchanged — Story C is MVP enhancement, Story G is P1)
- **Production Date:** Moved from Week 19 → Week 21 (+2 weeks for Stories G, H, I, J)

### Dependency Graph
```
US-013 (Prod Infra) ← blocks all production deployment
US-018 (WebSocket)  ← depends on API Gateway (complete)
                      ← depends on Frontend WebSocket client
US-045 (Clarify)    ← depends on Query Understanding Service (complete)
                      ← depends on Intent classifier (complete)
                      ← depends on API Gateway (complete)

Story A (Feedback)  ← depends on feedback_handler.go (complete)
                      ← depends on feedback_aggregates table (AM-04)
Story B (Quality)   ← depends on US-023 (Task 2.1)
                      ← depends on quality_loop/judge.py (complete)
Story C (Fallback)  ← depends on US-021 (Task 1.2)
                      ← depends on retrieval pipeline (complete)
Story D (Re-explain)← depends on Session Management (complete)
                      ← depends on API Gateway (complete)
Story E (Confidence)← depends on US-045 (Task 3.4)
                      ← depends on Vertex AI logprobs API
Story F (Intent)    ← depends on Query Understanding Service (complete)
                      ← depends on intent classifier (complete, needs remap)

Story G (Context)   ← depends on Session Management (complete)
                      ← depends on Redis (Phase 2)
                      ← depends on API Gateway (complete)
                      ← blocks Story H (Query Enrichment)
Story H (Enrichment)← depends on User Profile schema (US-013)
                      ← depends on Story G (Task 1.6) for session context
                      ← depends on RAG Handler (complete)
Story I (BM25)      ← depends on Hybrid Search (US-017, complete)
                      ← depends on RRF Fusion (US-020, complete)
                      ← depends on Ingestion Pipeline (adds BM25 index)
Story J (Grounding) ← depends on RAG Handler (complete)
                      ← depends on Retrieved Context (complete)
                      ← depends on LLM Client (complete)
                      ← depends on Quality Monitoring (AM-03)
```

---

## Phase Structure

### Phase 0: Security Foundation (Weeks 1-4) - CRITICAL 🔴
**Goal:** Address critical security vulnerabilities that block any production deployment
**Owner:** Security Team Lead
**Team:** 2-3 engineers

#### Task 0.1: Authentication & Authorization (1 week)
**Priority:** P0 Critical
**Files:** `orchestrator/middleware/auth.go` (new), `orchestrator/cmd/server/main.go` (modify)

- [ ] Implement JWT authentication middleware
- [ ] Add RBAC for roles (admin/teacher/student)
- [ ] Enforce grade-level isolation in all queries
- [ ] Implement session invalidation
- [ ] Write comprehensive auth tests

**Acceptance Criteria:**
- All endpoints require valid JWT
- Invalid tokens return 401 Unauthorized
- Grade 6 students cannot access Grade 8 content
- 100% test coverage for auth module

#### Task 0.2: Data Encryption (1 week)
**Priority:** P0 Critical
**Files:** `terraform/main.tf` (modify), `orchestrator/security/encryption.py` (new)

- [ ] Enable encryption at rest (AlloyDB + GCS)
- [ ] Enforce TLS 1.3 for all connections
- [ ] Implement field-level encryption for PII
- [ ] Configure GCP KMS for key management
- [ ] Add encryption key rotation (90 days)

**Acceptance Criteria:**
- Database encrypted with customer-managed KMS key
- All connections require TLS 1.3
- PII fields (email, names) encrypted at rest
- Key rotation automated

#### Task 0.3: API Security (1 week)
**Priority:** P0 Critical
**Files:** `orchestrator/middleware/security.go` (new), `orchestrator/validation/validator.go` (new)

- [ ] Implement input validation (SQL injection, XSS prevention)
- [ ] Add request size limits (10MB)
- [ ] Configure security headers (HSTS, CSP, X-Frame-Options)
- [ ] Implement CSRF protection
- [ ] Add CORS configuration

**Acceptance Criteria:**
- SQL injection attempts blocked
- XSS attempts sanitized
- Requests >10MB rejected
- All security headers present

#### Task 0.4: Secrets Management (1 week)
**Priority:** P0 Critical
**Files:** `orchestrator/secrets/manager.go` (new)

- [ ] Migrate secrets from environment to GCP Secret Manager
- [ ] Implement secret rotation (90 days)
- [ ] Add secret access logging
- [ ] Remove all hardcoded secrets
- [ ] Configure least-privilege access

**Acceptance Criteria:**
- All secrets in Secret Manager
- No secrets in environment variables
- Secret access logged and auditable
- Rotation tested and working

**Phase 0 Acceptance Criteria:**
- ✅ All P0 security issues resolved
- ✅ Security audit passed
- ✅ Penetration testing completed
- ✅ No critical vulnerabilities in SAST/DAST scans

---

### Phase 1: Reliability Foundation (Weeks 5-8) - CRITICAL 🔴
**Goal:** Implement reliability patterns to prevent cascading failures
**Owner:** Infrastructure Team Lead
**Team:** 2-3 engineers

#### Task 1.1: Circuit Breakers & Retry Logic (1 week)
**Priority:** P0 Critical
**Files:** `orchestrator/reliability/circuit_breaker.go` (new), `orchestrator/reliability/retry.go` (new)

- [ ] Implement circuit breaker pattern
- [ ] Add retry with exponential backoff
- [ ] Implement graceful degradation
- [ ] Add bulkhead isolation
- [ ] Configure timeout propagation

**Acceptance Criteria:**
- Circuit breakers on all external calls (Vertex AI, AlloyDB, Redis)
- Transient failures retried with backoff
- System degrades gracefully on partial outage
- Tests simulate dependency failures

#### Task 1.2: Distributed Tracing (1 week)
**Priority:** P0 Critical
**Files:** `orchestrator/observability/tracer.go` (new), `orchestrator/observability/metrics.go` (new)

- [ ] Add OpenTelemetry instrumentation
- [ ] Create spans for all stages (embed, retrieve, generate)
- [ ] Implement trace context propagation
- [ ] Configure Cloud Trace integration
- [ ] Add trace-based dashboards

**Acceptance Criteria:**
- All requests traced end-to-end
- Each stage has measurable spans
- Traces visible in Cloud Trace
- TTFT decomposition available

#### Task 1.3: Health Checks & Monitoring (1 week)
**Priority:** P0 Critical
**Files:** `orchestrator/handler/health.go` (new)

- [ ] Implement deep health checks (DB, Redis, Vertex AI)
- [ ] Separate liveness and readiness probes
- [ ] Add startup probes for slow starts
- [ ] Create operational dashboards
- [ ] Configure SLO tracking

**Acceptance Criteria:**
- Liveness probe returns 200 when process alive
- Readiness probe returns 503 when dependencies unhealthy
- Unhealthy instances removed from load balancer
- Dashboards show real-time health

#### Task 1.4: Structured Logging (1 week)
**Priority:** P0 Critical
**Files:** `orchestrator/logging/logger.go` (new)

- [ ] Implement correlation IDs
- [ ] Add PII redaction in logs
- [ ] Configure log aggregation (Cloud Logging)
- [ ] Set up log retention (90 days)
- [ ] Add log-based metrics

**Acceptance Criteria:**
- All logs have correlation IDs
- No PII in logs
- Logs searchable in Cloud Logging
- Log-based alerts configured

**Phase 1 Acceptance Criteria:**
- ✅ All P0 reliability issues resolved
- ✅ Circuit breakers prevent cascading failures
- ✅ Distributed tracing operational
- ✅ Health checks gate traffic

---

### Phase 2: Scalability Foundation (Weeks 9-12) - HIGH 🟠
**Goal:** Enable horizontal scaling and performance optimization
**Owner:** Platform Team Lead
**Team:** 2 engineers

#### Task 2.1: Caching Strategy (1 week)
**Priority:** P0
**Files:** `orchestrator/cache/response_cache.go` (new)

- [ ] Implement response caching (Redis)
- [ ] Add embedding cache
- [ ] Configure cache invalidation
- [ ] Implement cache warming
- [ ] Add cache metrics (hit rate)

**Acceptance Criteria:**
- Repeated queries served from cache
- Embedding cache reduces API calls by 50%+
- Cache hit rate >60%
- Cache metrics visible in dashboard

#### Task 2.2: Database Scaling (1 week)
**Priority:** P0
**Files:** `terraform/main.tf` (modify), `orchestrator/db/pool.go` (new)

- [ ] Configure read replicas
- [ ] Implement PgBouncer connection pooling
- [ ] Optimize slow queries
- [ ] Add query caching
- [ ] Configure connection limits

**Acceptance Criteria:**
- Read queries distributed across replicas
- Connection pooling prevents exhaustion
- P95 query latency <50ms
- No connection errors under load

#### Task 2.3: Async Processing (1 week)
**Priority:** P1
**Files:** `orchestrator/queue/pubsub.go` (new)

- [ ] Implement Pub/Sub for async processing
- [ ] Add dead letter queue
- [ ] Configure backpressure handling
- [ ] Implement idempotent processing
- [ ] Add queue monitoring

**Acceptance Criteria:**
- Long-running operations async
- Failed messages routed to DLQ
- Queue depth monitored
- Backpressure prevents overload

#### Task 2.4: Performance Testing (1 week)
**Priority:** P0
**Files:** `eval/load_test.js` (modify)

- [ ] Create k6 load test suite
- [ ] Test 1000 concurrent users
- [ ] Validate p95 <500ms TTFT
- [ ] Test sustained load (30 min)
- [ ] Test spike load

**Acceptance Criteria:**
- 1000 concurrent users supported
- p95 latency <500ms
- Error rate <1%
- No memory leaks

**Phase 2 Acceptance Criteria:**
- ✅ Caching reduces latency by 50%+
- ✅ Database scales to 1000 QPS
- ✅ Async processing handles bursts
- ✅ Load tests pass

---

### Phase 3: CI/CD & DevOps (Weeks 13-16) - HIGH 🟠
**Goal:** Automated testing and deployment pipelines
**Owner:** DevOps Team Lead
**Team:** 2 engineers

#### Task 3.1: Automated Testing (2 weeks)
**Priority:** P0
**Files:** `.github/workflows/test.yml` (new), `tests/e2e/` (new)

- [ ] Achieve 80%+ unit test coverage
- [ ] Add integration test suite
- [ ] Create E2E test suite
- [ ] Add security testing (OWASP)
- [ ] Add contract testing

**Acceptance Criteria:**
- 80%+ code coverage enforced
- All critical paths tested
- E2E tests pass on every PR
- Security scans pass

#### Task 3.2: Deployment Pipeline (1 week)
**Priority:** P0
**Files:** `.github/workflows/deploy.yml` (new), `.eas/workflows/` (new)

- [ ] Implement canary deployments
- [ ] Add deployment gates
- [ ] Configure automatic rollback
- [ ] Add deployment notifications
- [ ] Track deployment metrics

**Acceptance Criteria:**
- Canary deploys to 10% traffic first
- Failed deploys auto-rollback
- Deployment notifications sent
- Deployment success rate >99%

#### Task 3.3: Infrastructure as Code (1 week)
**Priority:** P1
**Files:** `terraform/` (restructure)

- [ ] Configure Terraform state backend (GCS)
- [ ] Add state locking
- [ ] Create reusable modules
- [ ] Add policy as code (OPA)
- [ ] Configure cost estimation

**Acceptance Criteria:**
- Terraform state in GCS
- State locking prevents conflicts
- Modules reusable across environments
- Policy violations blocked

**Phase 3 Acceptance Criteria:**
- ✅ 80%+ test coverage
- ✅ Automated deployments
- ✅ Infrastructure as code
- ✅ Zero-downtime deploys

---

### Phase 4: API Design & Documentation (Weeks 17-18) - MEDIUM 🟡
**Goal:** Professional API with complete documentation
**Owner:** API Team Lead
**Team:** 1-2 engineers

#### Task 4.1: API Versioning & Documentation (1 week)
**Priority:** P1
**Files:** `api/openapi.yaml` (new), `orchestrator/cmd/server/main.go` (modify)

- [ ] Create OpenAPI 3.0 specification
- [ ] Add URL-based versioning (/v1/, /v2/)
- [ ] Generate Swagger UI
- [ ] Create API reference documentation
- [ ] Add deprecation policy

**Acceptance Criteria:**
- Complete OpenAPI spec
- API versioned
- Swagger UI accessible
- Documentation published

#### Task 4.2: SDK Generation (1 week)
**Priority:** P2
**Files:** `sdk/` (new)

- [ ] Generate Go SDK
- [ ] Generate Python SDK
- [ ] Generate TypeScript SDK
- [ ] Add usage examples
- [ ] Create getting started guide

**Acceptance Criteria:**
- SDKs published to package registries
- Examples for all SDKs
- Getting started guide complete

**Phase 4 Acceptance Criteria:**
- ✅ OpenAPI spec complete
- ✅ API versioned
- ✅ SDKs available
- ✅ Documentation published

---

### Phase 5: RAG Production Hardening (Weeks 19-20) - MEDIUM 🟡
**Goal:** RAG-specific production features
**Owner:** ML Team Lead
**Team:** 2 engineers

#### Task 5.1: Quality Monitoring (1 week)
**Priority:** P0
**Files:** `orchestrator/quality/monitor.go` (new)

- [ ] Implement A/B testing framework
- [ ] Add quality dashboards
- [ ] Configure feedback loops
- [ ] Track retrieval analytics
- [ ] Monitor answer quality

**Acceptance Criteria:**
- A/B tests running
- Quality metrics visible
- Feedback collected
- Quality trends tracked

#### Task 5.2: Cost Management (1 week)
**Priority:** P0
**Files:** `orchestrator/observability/cost.go` (new)

- [ ] Track token usage
- [ ] Calculate cost per query
- [ ] Configure budget alerts
- [ ] Create cost dashboard
- [ ] Implement cost optimization

**Acceptance Criteria:**
- Token usage tracked
- Cost per query visible
- Budget alerts configured
- Cost dashboard available

**Phase 5 Acceptance Criteria:**
- ✅ Quality monitoring operational
- ✅ Cost tracking complete
- ✅ A/B testing enabled
- ✅ Budget alerts working

---

## Updated Success Metrics

| Metric | Current (Audit) | Target | Phase |
|--------|-----------------|--------|-------|
| **Security** | 28% | 95% | Phase 0 |
| **Reliability** | 31% | 95% | Phase 1 |
| **Scalability** | 42% | 95% | Phase 2 |
| **CI/CD** | 22% | 95% | Phase 3 |
| **API Design** | 38% | 95% | Phase 4 |
| **RAG Production** | 41% | 95% | Phase 5 |
| **Test Coverage** | 40% | 80%+ | Phase 3 |
| **Production Readiness** | 34% | 95%+ | Phase 5 |

---

## Updated Execution Order

1. **Phase 0 (P0 Security)** - Weeks 1-4 - **BLOCKS ALL PRODUCTION**
2. **Phase 1 (P0 Reliability)** - Weeks 5-8 - **BLOCKS ALL PRODUCTION**
3. **Phase 2 (P0 Scalability)** - Weeks 9-12
4. **Phase 3 (P1 CI/CD)** - Weeks 13-16
5. **Phase 4 (P2 API)** - Weeks 17-18
6. **Phase 5 (P2 RAG)** - Weeks 19-20

**Total Duration:** 20 weeks with team of 4-6 engineers

**Production Deployment:** Week 21 (after all phases complete)

---

## Gating Criteria for Production

**DO NOT DEPLOY TO PRODUCTION UNTIL:**

1. ✅ Phase 0 complete (all P0 security issues resolved)
2. ✅ Phase 1 complete (all P0 reliability issues resolved)
3. ✅ Security audit passed
4. ✅ Penetration testing passed
5. ✅ Load tests pass (1000 users, p95<500ms)
6. ✅ 80%+ test coverage
7. ✅ Disaster recovery tested
8. ✅ On-call rotation established
9. ✅ Runbooks documented
10. ✅ Monitoring and alerting operational

---

## Resource Requirements

| Phase | Engineers | Duration | Total Engineer-Weeks |
|-------|-----------|----------|---------------------|
| Phase 0 (Security) | 2-3 | 4 weeks | 8-12 |
| Phase 1 (Reliability) | 2-3 | 4 weeks | 8-12 |
| Phase 2 (Scalability) | 2 | 4 weeks | 8 |
| Phase 3 (CI/CD) | 2 | 4 weeks | 8 |
| Phase 4 (API) | 1-2 | 2 weeks | 2-4 |
| Phase 5 (RAG) | 2 | 2 weeks | 4 |
| **TOTAL** | **4-6** | **20 weeks** | **38-46** |

---

## Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Security breach before Phase 0 complete | Critical | **DO NOT DEPLOY** - keep in staging |
| Team capacity constraints | High | Prioritize P0 only, defer P1/P2 |
| Technical debt accumulation | Medium | Code review for all changes |
| Production incidents | High | Feature flags, canary deployments |
| Cost overruns | Medium | Budget alerts, cost tracking |
| Timeline slippage | Medium | Weekly progress reviews |

---

## Commit Strategy

- **Per-task commits**: Each task completed and tested before moving to next
- **Feature branches**: `phase-X/task-Y-description`
- **Security reviews**: All Phase 0 changes require security team review
- **Integration commits**: After each phase, verify end-to-end flows
- **TDD workflow**: Tests written before implementation

---

## Next Steps

1. **Immediate (This Week):**
   - Review and approve this updated plan
   - Assemble Phase 0 team (2-3 engineers)
   - Set up project tracking
   - Begin Task 0.1 (Authentication)

2. **Week 1-4:**
   - Complete Phase 0 (Security Foundation)
   - Security audit scheduled for Week 4

3. **Week 5:**
   - Phase 0 review and sign-off
   - Begin Phase 1 (Reliability Foundation)

---

**CRITICAL REMINDER:** Do not deploy to production until Phase 0 and Phase 1 are complete. The current implementation has critical security vulnerabilities that pose existential risks to the business.

**Previous Version:** 1.0 (March 31, 2026)
**Current Version:** 2.0 (April 1, 2026)
**Next Review:** April 8, 2026 (weekly progress review)

#### Task 1.1: Go Ingestion Pipeline - Project Setup (2 days)
- Create directory structure
- Set up go.mod with dependencies
- Create environment configuration
- Implement basic CLI entry point

#### Task 1.2: Go Ingestion - JSON Parser (2 days)
- Implement Python JSON output parser
- Add schema validation
- Create parser tests

#### Task 1.3: Go Ingestion - Parent-Child Chunker (3 days)
- Implement recursive text splitter
- Handle tables atomically
- Add chunking tests

#### Task 1.4: Go Ingestion - Keyword Extractor (2 days)
- Implement Gemini API keyword extraction
- Add TF-IDF fallback
- Create keyword tests

#### Task 1.5: Go Ingestion - Vertex AI Embedder (2 days)
- Implement batch embedding
- Add retry logic with backoff
- Create embedder tests

#### Task 1.6: Go Ingestion - AlloyDB Writer (3 days)
- Implement transactional writes
- Add DLQ routing
- Create writer tests

#### Task 1.7: Go Ingestion - Integration (2 days)
- Wire all components together
- Add CLI flags and configuration
- Create end-to-end tests

#### Task 1.8: Real Data Evaluation (1 week)
- Ingest 100+ pages from real textbooks
- Create 100+ golden QA pairs
- Fix evaluation scripts to use real data
- Establish baseline metrics

#### Task 1.9: Rate Limiting Middleware (3 days)
- Implement Redis-based rate limiter
- Add per-endpoint limits
- Create rate limit tests

**Phase 1 Acceptance Criteria:**
- Go ingestion pipeline complete and tested
- Evaluation producing real metrics (≥0.75 target)
- Rate limiting protecting API
- 85%+ test coverage for new code

---

### Phase 2: P1 Core Features (Weeks 5-8)
**Goal:** Add essential enterprise features

#### Task 2.1: Multi-Format Support - DOCX Parser (1 week)
- Implement python-docx parser
- Add heading hierarchy detection
- Create DOCX tests

#### Task 2.2: Multi-Format Support - PPTX Parser (1 week)
- Implement python-pptx parser
- Add slide notes extraction
- Create PPTX tests

#### Task 2.3: Multi-Format Support - Markdown/HTML (1 week)
- Implement mistune Markdown parser
- Implement BeautifulSoup HTML parser
- Create format tests

#### Task 2.4: Content Deduplication (2 weeks)
- Implement exact hash deduplication (SHA-256)
- Implement semantic deduplication (embedding cosine)
- Implement near-duplicate detection (MinHash LSH)
- Add database schema for dedup tracking
- Create deduplication tests

#### Task 2.5: Incremental Ingestion (3 weeks)
- Implement document-level change detection
- Implement page-level change detection
- Add chunk versioning
- Implement orphaned chunk cleanup
- Create incremental ingestion tests

**Phase 2 Acceptance Criteria:**
- 5 document formats supported (PDF, DOCX, PPTX, MD, HTML)
- Deduplication preventing 95%+ duplicates
- Incremental ingestion detecting changes
- All features tested with 90%+ coverage

---

### Phase 3: P1 Test Coverage Expansion (Weeks 9-12)
**Goal:** Achieve 85%+ code coverage

#### Task 3.1: Python Unit Tests (2 weeks)
- Parser unit tests (20+ tests)
- Chunker unit tests (15+ tests)
- Keyword extractor tests (10+ tests)
- Embedder tests (10+ tests)
- Writer tests (15+ tests)

#### Task 3.2: Go Unit Tests (2 weeks)
- Handler tests (20+ tests)
- Retrieval tests (15+ tests)
- Session tests (10+ tests)
- Cache tests (10+ tests)
- Middleware tests (15+ tests)

#### Task 3.3: Integration Tests (1 week)
- Ingestion pipeline integration (5+ tests)
- Query flow integration (5+ tests)
- Feedback loop integration (3+ tests)

#### Task 3.4: E2E Tests (1 week)
- Teacher workflow (3+ scenarios)
- Student workflow (3+ scenarios)
- Admin workflow (2+ scenarios)

#### Task 3.5: Load Tests (1 week)
- Concurrent queries (1000 users)
- Sustained load (30 min)
- Spike load testing

**Phase 3 Acceptance Criteria:**
- 85%+ code coverage (Python + Go)
- All critical paths tested
- Load tests passing (1000 concurrent users, p99 < 500ms)

---

### Phase 4: P2 Enterprise Features (Weeks 13-16)
**Goal:** Enterprise-grade features

#### Task 4.1: Disaster Recovery (1 week)
- Implement automated backups (AlloyDB → GCS)
- Configure Redis persistence
- Create backup restoration runbook
- Test point-in-time recovery

#### Task 4.2: A/B Testing Framework (2 weeks)
- Implement experiment assignment service
- Add traffic splitting
- Create metric collection per experiment
- Build statistical significance calculator

#### Task 4.3: Compliance Features (2 weeks)
- Implement PII detection and redaction
- Add data retention policies
- Create user data export (GDPR Art. 15)
- Implement right to be forgotten (GDPR Art. 17)
- Add audit trail for data access

#### Task 4.4: Canary Deployments (1 week)
- Configure Cloud Run traffic splitting
- Add health check gates
- Implement automatic rollback
- Create deployment metrics dashboard

**Phase 4 Acceptance Criteria:**
- Disaster recovery tested (RTO < 4h, RPO < 1h)
- A/B testing framework operational
- GDPR compliance features implemented
- Canary deployments with automatic rollback

---

### Phase 5: P2 Scale Features (Weeks 17-20)
**Goal:** Multi-region, high availability

#### Task 5.1: Multi-Region Support (2 weeks)
- Configure database read replicas
- Set up Cloud Run multi-region
- Configure global load balancing
- Add DNS failover

#### Task 5.2: Observability Enhancements (1 week)
- Add custom metrics dashboards
- Configure alert thresholds
- Implement distributed tracing improvements
- Add log aggregation

#### Task 5.3: Performance Optimization (1 week)
- Profile and optimize hot paths
- Tune database queries
- Optimize embedding cache
- Reduce TTFT latency

**Phase 5 Acceptance Criteria:**
- Multi-region deployment active
- 99.99% availability SLA
- p99 latency < 400ms
- All success metrics achieved

---

## Success Metrics

| Metric | Current | Target | Phase |
|--------|---------|--------|-------|
| **Test Coverage** | ~40% | 85%+ | Phase 3 |
| **Evaluation Metrics** | 0.00 | ≥0.75 | Phase 1 |
| **Document Formats** | 1 (PDF) | 5 | Phase 2 |
| **Duplicate Rate** | Unknown | <1% | Phase 2 |
| **Ingestion Speed** | ~50 pg/min | 100 pg/min | Phase 2 |
| **Query Latency (p99)** | <500ms | <400ms | Phase 5 |
| **Error Rate** | <1% | <0.1% | Phase 5 |
| **Production Readiness** | 84% | 95%+ | Phase 5 |

---

## Execution Order

1. **Phase 1 (P0)** - Tasks 1.1 → 1.9 (sequential with some parallelization)
2. **Phase 2 (P1)** - Tasks 2.1 → 2.5 (parallel where possible)
3. **Phase 3 (P1)** - Tasks 3.1 → 3.5 (parallel execution)
4. **Phase 4 (P2)** - Tasks 4.1 → 4.4 (sequential)
5. **Phase 5 (P2)** - Tasks 5.1 → 5.3 (parallel)

---

## Commit Strategy

- **Per-task commits**: Each task completed and tested before moving to next
- **Feature branches**: `phase-X/task-Y-description`
- **Integration commits**: After each phase, verify end-to-end flows
- **TDD workflow**: Tests written before implementation

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Team capacity | Prioritize P0/P1 only, defer P2 |
| Technical debt | Code review for all changes |
| Data migration | Staged rollout with rollback |
| Production incidents | Feature flags, canary deployments |

---

**Ready to execute. Starting with Phase 1, Task 1.1.**
