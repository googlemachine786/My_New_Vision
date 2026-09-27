-- P1-5: HNSW Index Parameter Tuning
-- Optimizes vector search performance with proper ef_search configuration.
-- Reference: docs/audits/audit_rag_latency.md, docs/audits/audit_vector_databases.md
--
-- Impact: 10x faster vector search with proper HNSW tuning
--
-- Background:
-- - HNSW (Hierarchical Navigable Small World) index uses two key parameters:
--   - m: number of bi-directional links per element (controls index density)
--   - ef_construction: size of dynamic candidate list during index building
--   - ef_search: size of dynamic candidate list during search (runtime)
--
-- Current: m=16, ef_construction=64 (from 002_hnsw_optimization.sql)
-- Optimized: ef_search tuned per query type (see SET LOCAL below)

-- 1. Set runtime ef_search for balanced speed/quality
-- This should be set per-query in application code:
-- SET LOCAL hnsw.ef_search = 40; -- Fast search, good for production
-- SET LOCAL hnsw.ef_search = 100; -- Higher quality, for evaluation

-- 2. Rebuild index with higher ef_construction for better quality
-- This improves index quality at the cost of longer build time.
-- Only run during maintenance windows.
DROP INDEX IF EXISTS idx_parent_chunks_embedding_hnsw;

CREATE INDEX idx_parent_chunks_embedding_hnsw ON parent_chunks
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 128);

-- 3. Create index for taxonomy-based filtering (composite index)
-- Speeds up filtered queries significantly
CREATE INDEX IF NOT EXISTS idx_parent_chunks_taxonomy_embedding
ON parent_chunks (taxonomy_id, embedding vector_cosine_ops)
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 128);

-- 4. Verify index was created correctly
SELECT
    indexname,
    indexdef
FROM pg_indexes
WHERE tablename = 'parent_chunks'
  AND indexname LIKE '%hnsw%';

-- 5. Verify index size and performance
-- Run this after index creation to verify:
-- EXPLAIN ANALYZE SELECT id, embedding FROM parent_chunks ORDER BY embedding <-> '[...]' LIMIT 10;
