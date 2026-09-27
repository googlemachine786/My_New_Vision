# Gap Analysis: User Stories with No Implementation

**Report Date:** April 1, 2026
**Scope:** 24 User Stories (32.8% of total 73) with NO implementation
**Priority Focus:** P0 and P1 stories blocking MVP and production deployment

---

## Executive Summary

| Category | Count | Percentage of 73 Total |
|----------|-------|----------------------|
| **P0 Critical (MVP Blockers)** | 7 | 9.6% |
| **P1 High (Production Blockers)** | 9 | 12.3% |
| **P2 Medium (Enterprise Features)** | 8 | 10.9% |

**Total Gaps:** 24 stories requiring implementation

---

## 1. P0 Critical Gaps (MVP Blockers)

These stories are **CRITICAL** for MVP and must be completed before any production deployment.

### 1.1 US-023: Auto Dataset Creation

**Epic:** E-001 Core RAG & AI  
**Priority:** P0 Critical  
**Status:** ❌ Not Started  
**Blocking:** US-032 (Golden Response Generation), Quality Loop

#### Acceptance Criteria
- [ ] Automatically generate golden QA pairs from negative feedback
- [ ] Minimum 100 QA pairs per grade/subject combination
- [ ] Difficulty rating auto-assigned based on query complexity
- [ ] QA pairs stored in `golden_qa_dataset` table
- [ ] Integration with feedback loop pipeline

#### Current State
- `data/golden_qa_dataset.jsonl` exists with 30 manually created QA pairs
- `quality_loop/judge.py` has LLM judge but doesn't auto-create datasets
- No automated pipeline for dataset expansion

#### Required Implementation
```python
# NEW FILE: quality_loop/auto_dataset_creator.py
class AutoDatasetCreator:
    """Automatically create golden QA dataset from feedback."""
    
    def __init__(self, db_pool, gemini_client):
        self.db_pool = db_pool
        self.gemini_client = gemini_client
    
    async def process_negative_feedback(self, limit=100):
        """Fetch negative feedback and generate golden QA pairs."""
        async with self.db_pool.acquire() as conn:
            feedback_rows = await conn.fetch("""
                SELECT * FROM ai_feedback_loop
                WHERE feedback_score = -1
                  AND processed_for_tuning = FALSE
                ORDER BY created_at DESC
                LIMIT $1
            """, limit)
            
            for row in feedback_rows:
                # Reconstruct context from UUID[]
                context = await self._reconstruct_context(conn, row['retrieved_context'])
                
                # Generate golden response
                golden = await self.gemini_client.generate(
                    prompt=self._build_judge_prompt(row, context)
                )
                
                # Create QA pair
                qa_pair = {
                    'question': row['user_query'],
                    'answer': golden,
                    'grade': self._extract_grade(row),
                    'subject': self._extract_subject(row),
                    'difficulty': self._assess_difficulty(row, context),
                    'taxonomy_id': row['taxonomy_id'],
                    'source_feedback_id': row['feedback_id']
                }
                
                # Store in golden_qa_dataset
                await self._store_qa_pair(conn, qa_pair)
                
                # Mark feedback as processed
                await conn.execute("""
                    UPDATE ai_feedback_loop
                    SET processed_for_tuning = TRUE,
                        golden_response = $1
                    WHERE feedback_id = $2
                """, golden, row['feedback_id'])
```

#### Dependencies
- US-031 (LLM Judge) - ✅ Complete
- US-030 (Feedback Logging) - ✅ Complete
- Database schema - ✅ Complete

#### Estimated Effort
- **Time:** 3-4 days
- **Complexity:** Medium
- **Risk:** Low (uses existing components)

---

### 1.2 US-040: Hybrid Search Tuning

**Epic:** E-001 Core RAG & AI  
**Priority:** P0 Critical  
**Status:** ❌ Not Started  
**Blocking:** Optimal retrieval performance

#### Acceptance Criteria
- [ ] Dynamic RRF k parameter tuning based on query type
- [ ] A/B testing framework for retrieval parameters
- [ ] Performance metrics per parameter configuration
- [ ] Automatic parameter selection based on historical performance
- [ ] Configuration stored in database, not hardcoded

#### Current State
- `orchestrator/retrieval/hybrid_search.go` has hardcoded `rrf_k=60`
- `rag_config.py` has static RRF_K=60 configuration
- No A/B testing infrastructure
- No parameter performance tracking

#### Required Implementation
```go
// NEW FILE: orchestrator/retrieval/parameter_tuner.go
package retrieval

type RRFConfig struct {
    K float64
    QueryType string  // "factual", "conceptual", "multi-hop"
    Grade int
    Subject string
    PerformanceMetrics
}

type ParameterTuner struct {
    db *pgxpool.Pool
    redis *redis.Client
}

// GetOptimalRRFK returns the best k parameter for this query context
func (pt *ParameterTuner) GetOptimalRRFK(ctx context.Context, query string, grade int, subject string) (float64, error) {
    // 1. Classify query type
    queryType := pt.classifyQuery(query)
    
    // 2. Fetch historical performance for similar queries
    metrics, err := pt.getHistoricalMetrics(ctx, queryType, grade, subject)
    if err != nil {
        return 60.0, nil // fallback to default
    }
    
    // 3. Return best performing k
    return metrics.BestRRFK, nil
}

// RecordRetrievalMetrics logs retrieval performance for tuning
func (pt *ParameterTuner) RecordRetrievalMetrics(ctx context.Context, 
    query string, rrfK float64, results []Chunk, userSatisfied bool) error {
    // Store in database for analysis
    // Update running averages in Redis
}
```

#### Dependencies
- US-017 (Hybrid Search) - ✅ Complete
- US-041 (Response Quality Metrics) - ❌ Not Started (circular dependency)
- US-029 (Distributed Tracing) - ⚠️ Partial

#### Estimated Effort
- **Time:** 1 week
- **Complexity:** High
- **Risk:** Medium (requires production data)

---

### 1.3 US-042: Retrieval Analytics

**Epic:** E-001 Core RAG & AI  
**Priority:** P0 Critical  
**Status:** ❌ Not Started  
**Blocking:** Performance optimization, quality monitoring

#### Acceptance Criteria
- [ ] Chunk-level retrieval statistics (which chunks retrieved most)
- [ ] Query pattern analysis (common queries, failed queries)
- [ ] Taxonomy coverage metrics (which subjects/grades underrepresented)
- [ ] Retrieval latency breakdown (dense vs sparse vs RRF)
- [ ] Dashboard with real-time analytics

#### Current State
- Basic logging in `orchestrator/handler/rag_handler_real.go`
- No analytics database tables
- No aggregation queries
- No dashboard

#### Required Implementation
```sql
-- NEW FILE: schema/analytics_tables.sql

-- Retrieval analytics
CREATE TABLE retrieval_analytics (
    analytics_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    query_hash BYTEA NOT NULL,
    query_text TEXT NOT NULL,
    taxonomy_id INTEGER,
    retrieved_chunk_ids UUID[],
    retrieval_latency_ms INTEGER,
    rrf_k_used FLOAT,
    dense_score FLOAT,
    sparse_score FLOAT,
    user_satisfaction_score INTEGER,  -- from feedback
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_analytics_query_hash ON retrieval_analytics(query_hash);
CREATE INDEX idx_analytics_taxonomy ON retrieval_analytics(taxonomy_id);
CREATE INDEX idx_analytics_created ON retrieval_analytics(created_at DESC);

-- Chunk usage statistics
CREATE TABLE chunk_usage_stats (
    chunk_id UUID PRIMARY KEY,
    parent_id UUID,
    taxonomy_id INTEGER,
    times_retrieved INTEGER DEFAULT 0,
    avg_satisfaction_score FLOAT,
    last_retrieved_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_chunk_usage_taxonomy ON chunk_usage_stats(taxonomy_id);
CREATE INDEX idx_chunk_usage_times ON chunk_usage_stats(times_retrieved DESC);

-- Query pattern analysis
CREATE MATERIALIZED VIEW query_patterns AS
SELECT 
    DATE_TRUNC('hour', created_at) AS hour,
    taxonomy_id,
    COUNT(*) AS query_count,
    AVG(retrieval_latency_ms) AS avg_latency,
    COUNT(DISTINCT query_hash) AS unique_queries
FROM retrieval_analytics
GROUP BY hour, taxonomy_id
WITH DATA;

-- Refresh every 5 minutes
CREATE INDEX idx_query_patterns_hour ON query_patterns(hour DESC);
```

```python
# NEW FILE: analytics/retrieval_analyzer.py
class RetrievalAnalyzer:
    """Analytics engine for retrieval patterns."""
    
    async def get_chunk_usage(self, taxonomy_id: int = None, limit: int = 100):
        """Get most/least retrieved chunks."""
        query = """
            SELECT * FROM chunk_usage_stats
            WHERE taxonomy_id = $1 OR $1 IS NULL
            ORDER BY times_retrieved DESC
            LIMIT $2
        """
        return await self.pool.fetch(query, taxonomy_id, limit)
    
    async def get_query_patterns(self, hours: int = 24):
        """Get query patterns for last N hours."""
        query = """
            SELECT * FROM query_patterns
            WHERE hour >= NOW() - INTERVAL '%s hours'
            ORDER BY hour DESC
        """
        return await self.pool.fetch(query, hours)
    
    async def get_taxonomy_coverage(self):
        """Analyze taxonomy coverage gaps."""
        query = """
            SELECT 
                t.grade,
                t.subject,
                t.chapter,
                COUNT(DISTINCT pc.parent_id) AS chunk_count,
                AVG(cus.times_retrieved) AS avg_retrievals
            FROM cbse_taxonomy t
            LEFT JOIN parent_chunks pc ON t.taxonomy_id = pc.taxonomy_id
            LEFT JOIN chunk_usage_stats cus ON pc.parent_id = cus.chunk_id
            GROUP BY t.grade, t.subject, t.chapter
            ORDER BY chunk_count ASC
        """
        return await self.pool.fetch(query)
```

#### Dependencies
- US-029 (Distributed Tracing) - ⚠️ Partial
- US-030 (Feedback Logging) - ✅ Complete
- Database schema - Requires new tables

#### Estimated Effort
- **Time:** 1 week
- **Complexity:** Medium
- **Risk:** Low

---

### 1.4 US-043: Cost Tracking

**Epic:** E-001 Core RAG & AI  
**Priority:** P0 Critical  
**Status:** ❌ Not Started  
**Blocking:** Budget management, cost optimization

#### Acceptance Criteria
- [ ] Token usage tracking per query (embedding + generation)
- [ ] Cost calculation per query (based on Vertex AI pricing)
- [ ] Daily/weekly/monthly cost aggregation
- [ ] Budget alerts (email/Slack when exceeding thresholds)
- [ ] Cost dashboard with trends and breakdowns

#### Current State
- No token counting
- No cost tracking
- No budget alerts
- Vertex AI usage unmonitored

#### Required Implementation
```go
// NEW FILE: orchestrator/observability/cost_tracker.go
package observability

type CostTracker struct {
    redis *redis.Client
    db *pgxpool.Pool
    config *CostConfig
}

type CostConfig struct {
    EmbeddingCostPer1KTokens float64  // $0.02 for text-embedding-005
    GenerationCostPer1KTokens float64 // $0.075 for Gemini 1.5 Flash
    DailyBudgetUSD float64
    AlertThresholdPercent float64  // Alert at 80% of budget
}

type TokenUsage struct {
    QueryID string
    SessionID string
    TaxonomyID int
    EmbeddingTokens int
    GenerationTokens int
    TotalCostUSD float64
    Timestamp time.Time
}

// TrackQuery records token usage and cost for a query
func (ct *CostTracker) TrackQuery(ctx context.Context, usage TokenUsage) error {
    // Store in database
    _, err := ct.db.Exec(ctx, `
        INSERT INTO cost_tracking (
            query_id, session_id, taxonomy_id,
            embedding_tokens, generation_tokens, total_cost_usd,
            created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, NOW())
    `, usage.QueryID, usage.SessionID, usage.TaxonomyID,
       usage.EmbeddingTokens, usage.GenerationTokens, usage.TotalCostUSD)
    
    // Update daily total in Redis
    dailyKey := fmt.Sprintf("cost:daily:%s", time.Now().Format("2006-01-02"))
    ct.redis.IncrByFloat(ctx, dailyKey, usage.TotalCostUSD)
    ct.redis.Expire(ctx, dailyKey, 7*24*time.Hour)  // Keep 7 days
    
    // Check budget alert
    if err := ct.checkBudgetAlert(ctx); err != nil {
        log.Error().Err(err).Msg("Budget alert check failed")
    }
    
    return err
}

// GetCostReport returns cost breakdown for period
func (ct *CostTracker) GetCostReport(ctx context.Context, days int) (*CostReport, error) {
    query := `
        SELECT 
            DATE(created_at) AS date,
            SUM(embedding_tokens) AS total_embedding_tokens,
            SUM(generation_tokens) AS total_generation_tokens,
            SUM(total_cost_usd) AS total_cost,
            AVG(total_cost_usd) AS avg_cost_per_query,
            COUNT(*) AS total_queries
        FROM cost_tracking
        WHERE created_at >= NOW() - INTERVAL '%d days'
        GROUP BY DATE(created_at)
        ORDER BY date DESC
    `
    // Execute and return report
}
```

```sql
-- NEW FILE: schema/cost_tracking.sql
CREATE TABLE cost_tracking (
    cost_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    query_id UUID NOT NULL,
    session_id TEXT,
    taxonomy_id INTEGER,
    embedding_tokens INTEGER NOT NULL,
    generation_tokens INTEGER NOT NULL,
    total_cost_usd NUMERIC(10, 6) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_cost_date ON cost_tracking(DATE(created_at));
CREATE INDEX idx_cost_taxonomy ON cost_tracking(taxonomy_id);
CREATE INDEX idx_cost_session ON cost_tracking(session_id);

-- Daily cost summary view
CREATE VIEW daily_cost_summary AS
SELECT 
    DATE(created_at) AS date,
    SUM(embedding_tokens) AS embedding_tokens,
    SUM(generation_tokens) AS generation_tokens,
    SUM(total_cost_usd) AS total_cost,
    COUNT(*) AS queries
FROM cost_tracking
GROUP BY DATE(created_at);
```

#### Dependencies
- US-029 (Distributed Tracing) - ⚠️ Partial (for query ID tracking)
- US-016 (Embedding) - ✅ Complete
- US-018 (Generation) - ✅ Complete

#### Estimated Effort
- **Time:** 3-4 days
- **Complexity:** Medium
- **Risk:** Low

---

### 1.5 US-044: Multi-hop Retrieval

**Epic:** E-001 Core RAG & AI  
**Priority:** P0 Critical  
**Status:** ❌ Not Started  
**Blocking:** Complex query answering

#### Acceptance Criteria
- [ ] Iterative retrieval (retrieve → analyze → retrieve again)
- [ ] Query decomposition for multi-part questions
- [ ] Context accumulation across retrieval hops
- [ ] Maximum 3 hops with diminishing returns check
- [ ] Latency budget: <800ms for 3-hop queries

#### Current State
- Single-pass retrieval only
- No query decomposition
- No iterative retrieval logic

#### Required Implementation
```python
# NEW FILE: retrievers/multi_hop_retriever.py
class MultiHopRetriever:
    """Multi-hop retrieval for complex queries."""
    
    def __init__(self, base_retriever, llm_client, max_hops=3):
        self.base_retriever = base_retriever
        self.llm_client = llm_client
        self.max_hops = max_hops
    
    async def search(self, query: str, taxonomy_id: int) -> List[Chunk]:
        """Perform multi-hop retrieval."""
        all_chunks = []
        current_query = query
        context_so_far = ""
        
        for hop in range(self.max_hops):
            # Retrieve with current query
            chunks = await self.base_retriever.search(
                current_query, 
                taxonomy_id,
                top_k=5
            )
            
            if not chunks:
                break
            
            all_chunks.extend(chunks)
            
            # Check if we have enough information
            context_so_far = self._build_context(all_chunks)
            answerability = await self._assess_answerability(
                query, context_so_far
            )
            
            if answerability.confident:
                break
            
            # Generate next query
            current_query = await self._generate_next_query(
                query, current_query, context_so_far, answerability.gaps
            )
            
            logger.info(f"Multi-hop: Hop {hop+1} complete, "
                       f"generating next query: {current_query}")
        
        # Deduplicate and re-rank
        return self._deduplicate_and_rerank(all_chunks)
    
    async def _assess_answerability(self, query: str, context: str):
        """Use LLM to assess if context answers query."""
        prompt = f"""
        Query: {query}
        
        Context: {context}
        
        Can this query be answered confidently with the provided context?
        If not, what information is missing?
        
        Respond in JSON:
        {{
            "confident": true/false,
            "gaps": ["missing info 1", "missing info 2"]
        }}
        """
        response = await self.llm_client.generate(prompt)
        return AnswerabilityResult.from_json(response)
```

#### Dependencies
- US-017 (Hybrid Search) - ✅ Complete
- US-021 (Contextual Grounding) - ⚠️ Partial
- US-026 (Query Rewriting) - ⚠️ Partial

#### Estimated Effort
- **Time:** 1-2 weeks
- **Complexity:** High
- **Risk:** Medium (latency concerns)

---

### 1.6 US-045: Clarification Generation

**Epic:** E-001 Core RAG & AI  
**Priority:** P0 Critical  
**Status:** ❌ Not Started  
**Blocking:** Ambiguous query handling

#### Acceptance Criteria
- [ ] Detect ambiguous queries (low retrieval confidence)
- [ ] Generate clarifying questions
- [ ] Present multiple interpretation options
- [ ] Re-retrieve based on clarification
- [ ] Track clarification patterns for learning

#### Current State
- No ambiguity detection
- No clarification logic
- All queries treated as unambiguous

#### Required Implementation
```python
# NEW FILE: retrievers/clarification_generator.py
class ClarificationGenerator:
    """Generate clarifying questions for ambiguous queries."""
    
    def __init__(self, llm_client, retriever):
        self.llm_client = llm_client
        self.retriever = retriever
    
    async def process_query(self, query: str, taxonomy_id: int):
        """Process query with clarification if needed."""
        # Initial retrieval
        chunks = await self.retriever.search(query, taxonomy_id, top_k=5)
        
        # Assess retrieval quality
        quality = self._assess_retrieval_quality(chunks, query)
        
        if quality.confidence < 0.6:
            # Low confidence - generate clarification
            interpretations = await self._generate_interpretations(query, chunks)
            
            if len(interpretations) > 1:
                return ClarificationResponse(
                    needs_clarification=True,
                    interpretations=interpretations,
                    suggested_question=self._build_clarification_question(
                        query, interpretations
                    )
                )
        
        # High confidence - proceed with answer
        return StandardResponse(chunks=chunks)
    
    async def _generate_interpretations(self, query: str, chunks: List[Chunk]):
        """Generate possible interpretations of ambiguous query."""
        prompt = f"""
        Query: {query}
        
        Retrieved context (may be insufficient):
        {self._build_context(chunks)}
        
        This query appears ambiguous. List 2-3 possible interpretations:
        
        Example format:
        1. [Interpretation 1]: Student asking about [topic A]
        2. [Interpretation 2]: Student asking about [topic B]
        """
        response = await self.llm_client.generate(prompt)
        return self._parse_interpretations(response)
```

#### Dependencies
- US-017 (Hybrid Search) - ✅ Complete
- US-021 (Contextual Grounding) - ⚠️ Partial

#### Estimated Effort
- **Time:** 4-5 days
- **Complexity:** Medium
- **Risk:** Low

---

### 1.7 US-014: JWT Auth with RBAC (Incomplete)

**Epic:** E-005 Security & Compliance  
**Priority:** P0 Critical  
**Status:** ⚠️ Partial (needs RBAC)  
**Blocking:** Production security, grade isolation

#### Acceptance Criteria
- [ ] Role-based access control (admin/teacher/student)
- [ ] Grade-level isolation enforced in ALL queries
- [ ] Subject-level filtering based on JWT claims
- [ ] Token invalidation/refresh mechanism
- [ ] Audit logging for auth events

#### Current State
- JWT validation middleware exists (`orchestrator/middleware/auth.go`)
- Claims extraction implemented
- **MISSING:** RBAC enforcement
- **MISSING:** Grade isolation in database queries
- **MISSING:** Token invalidation

#### Required Implementation
```go
// MODIFY: orchestrator/middleware/auth.go
package middleware

type Role string
const (
    RoleAdmin    Role = "admin"
    RoleTeacher  Role = "teacher"
    RoleStudent  Role = "student"
)

type Claims struct {
    UserID    string `json:"user_id"`
    SessionID string `json:"session_id"`
    Grade     int    `json:"grade"`
    Subject   string `json:"subject"`
    TaxonomyID int   `json:"taxonomy_id"`
    Role      Role   `json:"role"`
    jwt.RegisteredClaims
}

// EnforceGradeIsolation ensures query only accesses allowed grade content
func EnforceGradeIsolation(ctx context.Context, claims *Claims, query string) (string, error) {
    // Inject taxonomy_id filter into ALL database queries
    // This should be called in HybridSearch, not optional
    if claims.TaxonomyID == 0 {
        return "", fmt.Errorf("taxonomy_id missing from JWT")
    }
    
    // Log for audit
    log.Info().
        Str("user_id", claims.UserID).
        Int("grade", claims.Grade).
        Int("taxonomy_id", claims.TaxonomyID).
        Msg("Grade isolation enforced")
    
    return query, nil
}

// EnforceRBAC checks if user role has permission for action
func EnforceRBAC(ctx context.Context, claims *Claims, requiredRole Role, action string) error {
    roleHierarchy := map[Role]int{
        RoleStudent: 1,
        RoleTeacher: 2,
        RoleAdmin:   3,
    }
    
    if roleHierarchy[claims.Role] < roleHierarchy[requiredRole] {
        log.Warn().
            Str("user_id", claims.UserID).
            Str("role", string(claims.Role)).
            Str("required_role", string(requiredRole)).
            Str("action", action).
            Msg("RBAC violation attempted")
        
        return fmt.Errorf("insufficient permissions")
    }
    
    return nil
}
```

```sql
-- MODIFY: schema/v2_production.sql
-- Add RLS (Row Level Security) policies
ALTER TABLE parent_chunks ENABLE ROW LEVEL SECURITY;

-- Students can only access their grade's content
CREATE POLICY student_grade_isolation ON parent_chunks
    FOR SELECT
    USING (
        taxonomy_id IN (
            SELECT taxonomy_id FROM cbse_taxonomy
            WHERE grade = current_setting('app.user_grade')::int
        )
    );

-- Teachers can access multiple grades
CREATE POLICY teacher_access ON parent_chunks
    FOR SELECT
    USING (
        taxonomy_id IN (
            SELECT taxonomy_id FROM cbse_taxonomy
            WHERE grade BETWEEN 
                current_setting('app.teacher_grade_min')::int AND
                current_setting('app.teacher_grade_max')::int
        )
    );

-- Admins can access everything
CREATE POLICY admin_access ON parent_chunks
    FOR ALL
    USING (true);
```

#### Dependencies
- US-013 (Vector Database) - ✅ Complete
- US-036 (Taxonomy Filtering) - ✅ Complete

#### Estimated Effort
- **Time:** 1 week
- **Complexity:** High
- **Risk:** **CRITICAL** (security vulnerability if incomplete)

---

## 2. P1 High Priority Gaps (Production Blockers)

### 2.1 US-035: Incremental Ingestion

**Epic:** E-001 Core RAG & AI  
**Priority:** P1 High  
**Status:** ❌ Not Started  
**Blocking:** Production ingestion at scale

#### Acceptance Criteria
- [ ] Document-level change detection (SHA-256)
- [ ] Page-level change detection
- [ ] Chunk versioning with lineage tracking
- [ ] Orphaned chunk cleanup
- [ ] Rollback support (revert to previous version)

#### Required Implementation
```python
# NEW FILE: ingestion/incremental.py
class IncrementalIngestion:
    """Incremental ingestion with change detection."""
    
    async def ingest_document(self, file_path: str, taxonomy_id: int):
        # Calculate document fingerprint
        doc_hash = self._calculate_document_hash(file_path)
        
        # Check if document changed
        old_hash = await self._get_stored_hash(taxonomy_id)
        if old_hash == doc_hash:
            logger.info(f"Document {taxonomy_id} unchanged, skipping")
            return
        
        # Identify changed pages
        changed_pages = await self._detect_changed_pages(file_path, taxonomy_id)
        
        if changed_pages:
            # Incremental update
            await self._update_chunks(changed_pages, taxonomy_id)
            await self._mark_orphaned_chunks(taxonomy_id, changed_pages)
            await self._update_document_hash(taxonomy_id, doc_hash)
```

**Estimated Effort:** 3 weeks

---

### 2.2 US-022: Multi-agent Routing

**Epic:** E-001 Core RAG & AI  
**Priority:** P1 High  
**Status:** ⚠️ Partial (structure exists, no routing)  
**Blocking:** Advanced query handling

**Estimated Effort:** 1-2 weeks

---

### 2.3 US-038: Model Routing

**Epic:** E-001 Core RAG & AI  
**Priority:** P1 High  
**Status:** ⚠️ Partial (single model only)  
**Blocking:** Cost optimization, fallback

**Estimated Effort:** 1 week

---

### 2.4 US-039: Hallucination Control

**Epic:** E-001 Core RAG & AI  
**Priority:** P1 High  
**Status:** ⚠️ Partial (basic grounding only)  
**Blocking:** Answer quality, trust

**Estimated Effort:** 1-2 weeks

---

### 2.5 US-059: RAG Filtering by Curriculum

**Epic:** E-002 Curriculum & Learning  
**Priority:** P1 High  
**Status:** ❌ Not Started  
**Blocking:** Curriculum-aligned retrieval

**Estimated Effort:** 1 week

---

### 2.6 US-061: Next-Concept Recommendation

**Epic:** E-002 Curriculum & Learning  
**Priority:** P1 High  
**Status:** ❌ Not Started  
**Blocking:** Personalized learning

**Estimated Effort:** 2 weeks

---

### 2.7 US-063: Progress Tracking

**Epic:** E-002 Curriculum & Learning  
**Priority:** P1 High  
**Status:** ❌ Not Started  
**Blocking:** Learning analytics

**Estimated Effort:** 1-2 weeks

---

### 2.8 US-042 & US-043: Analytics (covered in P0)

---

### 2.9 US-008 to US-018: UX Engagement Features

**Epic:** E-003 User Experience  
**Priority:** P2 Medium (can defer for MVP)  
**Status:** ❌ Not Started  

**Features:** Gamification, notifications, social sharing, bookmarks, notes, highlights, search history, favorites, avatars, themes, accessibility

**Estimated Effort:** 6-8 weeks total

---

## 3. P2 Medium Priority Gaps (Enterprise Features)

### 3.1 US-064 to US-073: Advanced Learning Features

**Epic:** E-002 Curriculum & Learning  
**Priority:** P2 Medium  
**Status:** ❌ Not Started  

**Features:**
- US-064: Mastery Assessment
- US-065: Adaptive Difficulty
- US-066: Prerequisite Warnings
- US-067: Search Curriculum
- US-068: Multilingual Support (Hindi/English)
- US-069: Learning Objectives API
- US-070: Teacher Dashboard
- US-071: Student Dashboard
- US-072: Parent Reports
- US-073: LMS Integration

**Estimated Effort:** 8-10 weeks total

---

## 4. Gap Summary by Category

### 4.1 Technical Gaps

| Category | Count | Stories |
|----------|-------|---------|
| **Retrieval & Search** | 5 | US-040, US-042, US-044, US-045, US-059 |
| **Quality & Evaluation** | 3 | US-023, US-039, US-041 |
| **Cost & Operations** | 2 | US-043, US-035 |
| **Architecture** | 2 | US-022, US-038 |

### 4.2 Feature Gaps

| Category | Count | Stories |
|----------|-------|---------|
| **Learning Management** | 10 | US-061 to US-073 |
| **User Engagement** | 11 | US-008 to US-018 |
| **Security** | 1 | US-014 (RBAC portion) |

---

## 5. Recommended Implementation Order

### Sprint 1-2 (Immediate - P0 Critical)
1. **US-014 RBAC** - Security critical
2. **US-023 Auto Dataset** - Quality loop dependency
3. **US-043 Cost Tracking** - Budget management
4. **US-042 Retrieval Analytics** - Performance visibility

### Sprint 3-4 (P1 High)
1. **US-035 Incremental Ingestion** - Production scale
2. **US-040 Hybrid Search Tuning** - Performance optimization
3. **US-044 Multi-hop Retrieval** - Complex queries
4. **US-045 Clarification** - Ambiguity handling

### Sprint 5-8 (P2 Medium)
1. **US-059 RAG Filtering** - Curriculum alignment
2. **US-061 Recommendations** - Personalization
3. **US-063 Progress Tracking** - Learning analytics
4. **US-022 Multi-agent** - Architecture improvement

### Sprint 9+ (Enterprise)
1. **US-064 to US-073** - Advanced learning features
2. **US-008 to US-018** - Engagement features

---

## 6. Risk Assessment

### High Risk (Block Production)
- **US-014 RBAC** - Security vulnerability without grade isolation
- **US-043 Cost Tracking** - Budget overruns without monitoring
- **US-035 Incremental Ingestion** - Cannot scale with full re-ingestion

### Medium Risk (Degrade Quality)
- **US-023 Auto Dataset** - Quality loop manual without automation
- **US-040 Hybrid Search Tuning** - Suboptimal retrieval
- **US-044 Multi-hop** - Cannot answer complex queries

### Low Risk (Nice to Have)
- **US-008 to US-018** - Engagement features
- **US-064 to US-073** - Advanced learning features

---

**Report Generated:** April 1, 2026  
**Next Review:** April 8, 2026  
**Owner:** Engineering Team
