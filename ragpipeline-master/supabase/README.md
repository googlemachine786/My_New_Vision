# 🏗️ Production-Grade Vector Database Schema

**Enterprise-ready pgvector implementation for Visionary RAG Pipeline on Supabase**

---

## 🎯 Key Features

### ✅ Production-Ready

- **HNSW Index Optimization**: m=16, ef_construction=64 for 100k+ chunks
- **Deduplication**: SHA256 content hashing prevents duplicate chunks
- **Versioning**: Full audit trail with `chunk_versions` table
- **DLQ**: Dead letter queue with automatic retry and exponential backoff
- **RLS**: Row Level Security on all tables
- **Caching**: Response cache with TTL and atomic hit counting
- **Monitoring**: Performance metrics tracking and analytics

### ✅ Robust Data Model

- **9 Core Tables**: taxonomy, parent_chunks, child_chunks, DLQ, feedback, sessions, cache, golden dataset, evaluation
- **28+ Indexes**: HNSW, GIN, B-tree, partial, composite
- **19 Functions**: Hybrid search, ingestion, cleanup, statistics
- **9 RPC Functions**: Supabase Edge Functions integration
- **3 Materialized Views**: Pre-computed statistics

### ✅ Enterprise Features

- **Content Deduplication**: SHA256 hash-based deduplication
- **Version Tracking**: Full history of chunk changes
- **Quality Scoring**: Chunk quality assessment (excellent/good/fair/poor)
- **Popularity Tracking**: Retrieval count and last retrieved timestamp
- **Taxonomy Filtering**: Grade/subject/chapter isolation
- **Multi-language Support**: Language field for future internationalization

---

## 📁 Migration Files

| Migration | Lines | Tables | Functions | Indexes | Description |
|-----------|-------|--------|-----------|---------|-------------|
| **001_production_schema.sql** | 800+ | 9 | 5 | 28+ | Base schema with all tables, indexes, RLS |
| **002_hnsw_optimization.sql** | 50 | - | 1 | 5 | HNSW tuning, composite indexes |
| **003_ingestion_functions.sql** | 300 | - | 6 | - | Atomic ingestion with DLQ |
| **004_rpc_functions.sql** | 350 | - | 9 | - | Supabase RPC functions |
| **TOTAL** | **~1,500** | **9** | **21** | **33+** | **Production-ready** |

---

## 🚀 Deployment

### Prerequisites

```bash
# Install Supabase CLI
npm install -g supabase

# Login
supabase login

# Link to your project
supabase link --project-ref your-project-ref
```

### Apply Migrations

```bash
# Push all migrations
supabase db push

# Or apply manually in Supabase Dashboard:
# SQL Editor → New Query → Paste migration → Run
```

### Verify Deployment

```sql
-- Check tables
SELECT table_name FROM information_schema.tables 
WHERE table_schema = 'public' ORDER BY table_name;

-- Check indexes
SELECT indexname, indexdef FROM pg_indexes 
WHERE schemaname = 'public' ORDER BY indexname;

-- Check functions
SELECT routine_name FROM information_schema.routines 
WHERE routine_schema = 'public' ORDER BY routine_name;

-- Check HNSW index
SELECT indexname, indexdef FROM pg_indexes 
WHERE indexname = 'idx_child_embedding_hnsw';
```

---

## 📊 Schema Overview

### Core Tables

```
cbse_taxonomy (Curriculum hierarchy)
├── taxonomy_id (PK)
├── grade (1-12)
├── subject
├── chapter
├── section
├── learning_objectives (TEXT[])
├── keywords (TEXT[])
└── difficulty_level (beginner/intermediate/advanced)

parent_chunks (Context chunks - 1500 chars)
├── parent_id (PK, UUID)
├── taxonomy_id (FK)
├── content (≤1500 chars)
├── content_hash (SHA256 for dedup)
├── extracted_keywords (TEXT[])
├── keyword_scores (FLOAT[])
├── quality_score (excellent/good/fair/poor)
├── is_active (BOOLEAN)
├── version (INTEGER)
└── GIN index on keywords

child_chunks (Embedding chunks - 512 chars)
├── child_id (PK, UUID)
├── parent_id (FK)
├── taxonomy_id (FK)
├── content (≤512 chars)
├── content_hash (SHA256 for dedup)
├── embedding (vector(768))
├── embedding_model (VARCHAR)
├── retrieval_count (INTEGER)
├── last_retrieved_at (TIMESTAMPTZ)
├── quality_score (excellent/good/fair/poor)
├── is_active (BOOLEAN)
└── HNSW index on embedding

ingestion_queue (Batch processing)
├── queue_id (BIGSERIAL)
├── batch_id (UUID)
├── status (pending/processing/completed/failed/retrying)
├── retry_count
├── priority
└── scheduled_at

ingestion_dlq (Dead letter queue)
├── dlq_id (BIGSERIAL)
├── queue_id (FK)
├── payload (JSONB)
├── error_message
├── error_type
├── error_stack
└── resolution_notes

ai_feedback_loop (Quality loop)
├── feedback_id (UUID)
├── session_id
├── user_id (FK to auth.users)
├── user_query
├── query_embedding (vector(768))
├── retrieved_context (UUID[])
├── retrieved_context_scores (FLOAT[])
├── llm_response
├── feedback_score (negative/neutral/positive)
├── user_action (thumbs_up/thumbs_down/regenerate/share/bookmark)
├── golden_response
├── quality_metrics (JSONB)
└── processed_for_tuning

session_history (Session backup)
├── session_id
├── turn_id (UUID)
├── user_id (FK to auth.users)
├── role (user/assistant/system)
├── content
├── content_embedding (vector(768))
└── token_count

query_cache (Response cache)
├── cache_id (UUID)
├── query_hash (SHA256 - unique)
├── query_text
├── query_embedding (vector(768))
├── answer
├── sources (JSONB)
├── hits
├── last_hit_at
├── expires_at
└── TTL constraint

golden_qa_dataset (Evaluation)
├── qa_id
├── question
├── question_embedding (vector(768))
├── answer
├── expected_child_ids (UUID[])
├── expected_child_scores (FLOAT[])
└── difficulty

evaluation_results (Metrics tracking)
├── eval_id
├── evaluation_date
├── test_name
├── metric_name
├── metric_value
├── target_value
├── passed (BOOLEAN)
└── details (JSONB)
```

---

## 🔧 RPC Functions (Call from Edge Functions)

### Hybrid Search

```typescript
// TypeScript (Supabase Edge Function)
const { data, error } = await supabase.rpc('rpc_hybrid_search', {
  query_embedding: embedding,
  query_keywords: ['photosynthesis', 'chloroplast'],
  p_taxonomy_id: 42,
  p_top_k: 5,
  p_ef_search: 40  // Higher = more accurate but slower
});

// Returns: Array of {
//   child_id, parent_id, content, parent_content,
//   page_number, section, cosine_score, keyword_score, rrf_score
// }
```

### Semantic Search

```typescript
const { data } = await supabase.rpc('rpc_semantic_search', {
  query_embedding: embedding,
  p_taxonomy_id: 42,
  p_top_k: 5,
  p_min_score: 0.5  // Minimum cosine similarity
});
```

### Keyword Search

```typescript
const { data } = await supabase.rpc('rpc_keyword_search', {
  query_keywords: ['cell', 'membrane', 'nucleus'],
  p_taxonomy_id: 42,
  p_top_k: 5
});
```

### Log Feedback

```typescript
const { data } = await supabase.rpc('rpc_log_feedback', {
  p_session_id: 'session-123',
  p_user_query: 'What is photosynthesis?',
  p_retrieved_context: ['uuid-1', 'uuid-2', 'uuid-3'],
  p_retrieved_context_scores: [0.85, 0.78, 0.72],
  p_llm_response: 'Photosynthesis is...',
  p_feedback_score: 'positive',
  p_user_action: 'thumbs_up',
  p_quality_metrics: {
    relevance: 0.9,
    faithfulness: 0.95,
    completeness: 0.85
  }
});
```

### Session Management

```typescript
// Get session history
const { data: history } = await supabase.rpc('rpc_get_session_history', {
  p_session_id: 'session-123',
  p_last_n: 10
});

// Append turn
const { data: turn } = await supabase.rpc('rpc_append_session_turn', {
  p_session_id: 'session-123',
  p_role: 'user',
  p_content: 'What is photosynthesis?',
  p_metadata: { grade: 8, subject: 'Science' }
});
```

### Response Caching

```typescript
// Get cached response
const { data: cached } = await supabase.rpc('rpc_get_cached_response', {
  p_query_hash: 'sha256-of-query-embedding'
});

if (cached) {
  // Cache hit - use cached answer
  console.log(`Cache hits: ${cached.hits}`);
} else {
  // Cache miss - generate new answer and cache it
  const { data: cacheId } = await supabase.rpc('rpc_cache_response', {
    p_query_hash: 'sha256-hash',
    p_query_text: 'What is photosynthesis?',
    p_answer: 'Photosynthesis is...',
    p_sources: [{child_id: 'uuid', content: '...', page: 85, section: 'Photosynthesis'}],
    p_ttft_ms: 150,
    p_total_latency_ms: 450,
    p_ttl_seconds: 300
  });
}
```

### Cleanup Expired Data

```typescript
// Run cleanup (scheduled job)
const { data: stats } = await supabase.rpc('rpc_cleanup_expired', {
  p_cleanup_cache: true,
  p_cleanup_sessions: true,
  p_session_ttl_minutes: 35
});

console.log(`Deleted ${stats.cache_deleted} cache entries`);
console.log(`Deleted ${stats.sessions_deleted} old sessions`);
```

---

## 🐍 Python Client Example

```python
from supabase import create_client, Client

supabase: Client = create_client(
    supabase_url,
    supabase_key
)

# Hybrid search
result = supabase.rpc('rpc_hybrid_search', {
    'query_embedding': embedding.tolist(),
    'query_keywords': ['cell', 'membrane'],
    'p_taxonomy_id': 42,
    'p_top_k': 5
}).execute()

print(result.data)

# Log feedback
result = supabase.rpc('rpc_log_feedback', {
    'p_session_id': session_id,
    'p_user_query': query,
    'p_retrieved_context': chunk_ids,
    'p_llm_response': answer,
    'p_feedback_score': 'positive'
}).execute()

# Get ingestion stats
result = supabase.rpc('rpc_get_ingestion_stats', {
    'p_taxonomy_id': 42,
    'p_include_details': True
}).execute()

print(result.data)
```

---

## 📈 Indexes Created

| Table | Index | Type | Purpose |
|-------|-------|------|---------|
| `child_chunks` | `idx_child_embedding_hnsw` | HNSW | Dense vector search (m=16, ef=64) |
| `parent_chunks` | `idx_parent_keywords_gin` | GIN | Keyword sparse search |
| `child_chunks` | `idx_child_taxonomy_embedding` | B-tree + vector | Taxonomy-filtered search |
| `child_chunks` | `idx_child_active_embedding` | Partial HNSW | Active chunks only |
| `ai_feedback_loop` | `idx_feedback_unprocessed` | Partial | Quality loop queries |
| `query_cache` | `idx_query_cache_expires` | B-tree | Cache cleanup |
| `cbse_taxonomy` | `idx_taxonomy_keywords_gin` | GIN | Keyword search |
| `parent_chunks` | `idx_parent_chapter_trgm` | GIN trgm | Fuzzy chapter matching |

---

## 🔒 Row Level Security (RLS)

All tables have RLS enabled with these policies:

| Table | Policy | Access |
|-------|--------|--------|
| `cbse_taxonomy` | Public read | Anyone can read |
| `parent_chunks` | Authenticated read + is_active | Logged-in users, active only |
| `child_chunks` | Authenticated read + is_active | Logged-in users, active only |
| `ai_feedback_loop` | User-specific | Users can only see their own |
| `session_history` | User-specific | Users can only access their sessions |
| `query_cache` | Public read | Anyone can read cached responses |
| `golden_qa_dataset` | Service role only | Protected for evaluation |

---

## 🧪 Testing

### Test Hybrid Search

```sql
-- Test with random vector
SELECT * FROM rpc_hybrid_search(
  array_fill(random(), array[768])::vector,
  ARRAY['test', 'keyword'],
  1,  -- taxonomy_id
  5,  -- top_k
  40  -- ef_search
);
```

### Test Ingestion

```sql
-- Test parent-child insertion with deduplication
SELECT insert_parent_child(
  NULL,                -- Generate new parent_id
  1,                   -- taxonomy_id
  'Test content',      -- content
  ARRAY['test'],       -- keywords
  1,                   -- page_number
  'Test Chapter',      -- chapter
  'Test Section',      -- section
  NULL,                -- subsection
  'prose',             -- content_type
  '[{"content": "Child content", "embedding": array_fill(0.1, array[768])::vector}]'::jsonb,
  '{}'::jsonb,         -- metadata
  'good'::chunk_quality_score
);
```

### Test Feedback Loop

```sql
-- Test feedback logging
SELECT rpc_log_feedback(
  'test-session',
  'What is force?',
  ARRAY['child-uuid-1', 'child-uuid-2'],
  'Force is a push or pull...',
  'positive',
  'thumbs_up',
  '{"relevance": 0.9, "faithfulness": 0.95}'::jsonb
);
```

### Test Cleanup

```sql
-- Test expired data cleanup
SELECT rpc_cleanup_expired(
  true,   -- cleanup_cache
  true,   -- cleanup_sessions
  35      -- session_ttl_minutes
);
```

---

## 📊 Views & Materialized Views

### Chunk Statistics

```sql
SELECT * FROM v_chunk_statistics
WHERE grade = 8 AND subject = 'Science';
```

### Feedback Statistics

```sql
SELECT * FROM v_feedback_statistics
ORDER BY feedback_date DESC
LIMIT 30;
```

### Cache Performance

```sql
SELECT * FROM v_cache_performance
ORDER BY cache_date DESC
LIMIT 30;
```

### Feedback Statistics (Materialized)

```sql
-- Refresh materialized view
REFRESH MATERIALIZED VIEW mv_feedback_statistics;

-- Query
SELECT * FROM mv_feedback_statistics
WHERE feedback_date > NOW() - INTERVAL '7 days';
```

---

## 🧹 Maintenance

### Cleanup Expired Cache

```sql
SELECT cleanup_expired_cache();
-- Returns number of deleted entries
```

### Cleanup Old Sessions

```sql
SELECT cleanup_old_sessions(35);
-- Returns number of deleted sessions (older than 35 minutes)
```

### Analyze Tables

```sql
ANALYZE child_chunks;
ANALYZE parent_chunks;
ANALYZE cbse_taxonomy;
ANALYZE ai_feedback_loop;
```

### Adjust HNSW Accuracy

```sql
-- For higher accuracy (slower)
SET hnsw.ef_search = 80;

-- For faster search (less accurate)
SET hnsw.ef_search = 20;

-- Reset to default
SET hnsw.ef_search = 40;
```

---

## 📁 File Structure

```
supabase/
├── migrations/
│   ├── 001_production_schema.sql       (800+ lines)
│   ├── 002_hnsw_optimization.sql       (50 lines)
│   ├── 003_ingestion_functions.sql     (300 lines)
│   └── 004_rpc_functions.sql           (350 lines)
├── functions/                          # Supabase Edge Functions
│   ├── hybrid_search/
│   │   └── index.ts
│   ├── log_feedback/
│   │   └── index.ts
│   ├── evaluate/
│   │   └── index.ts
│   └── cleanup/
│       └── index.ts
└── config.toml
```

---

## 🚀 Next Steps

1. ✅ Apply migrations to Supabase
2. ✅ Test RPC functions with sample data
3. ✅ Update Go/Python clients to use Supabase RPC
4. ✅ Deploy Edge Functions for hybrid search
5. ✅ Set up scheduled cleanup jobs (pg_cron)
6. ✅ Configure monitoring and alerts
7. ✅ Load test with production data volume

---

## 📞 Support

For issues or questions:
- Check Supabase Dashboard → Logs
- Review function comments in migrations
- Test with provided SQL examples
- Monitor `performance_metrics` table

---

**Production-ready vector database schema with enterprise-grade features!** 🎉

**Total**: ~1,500 lines of production SQL, 9 tables, 21 functions, 33+ indexes
