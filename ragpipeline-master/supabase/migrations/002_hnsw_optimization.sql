-- ============================================================
-- Migration 002: HNSW Index Optimization
-- Production-tuned parameters for 100k+ chunks
-- ============================================================

-- Drop existing HNSW index if exists
DROP INDEX IF EXISTS idx_child_embedding_hnsw;

-- Recreate with production-optimized parameters
-- For dataset size 100k+ chunks with 768 dimensions:
-- m = 16 (number of connections per node)
-- ef_construction = 64 (size of dynamic candidate list during construction)
-- Higher values = better accuracy but slower indexing
CREATE INDEX idx_child_embedding_hnsw ON child_chunks 
USING hnsw (embedding vector_cosine_ops) 
WITH (m = 16, ef_construction = 64);

-- Set ef_search for query time (can be adjusted per-query)
-- Default: 40 (good balance of speed/accuracy)
-- Range: 10-100 (higher = more accurate but slower)
SET hnsw.ef_search = 40;

-- Create composite index for taxonomy-filtered search
-- This is CRITICAL for grade/subject isolation
CREATE INDEX idx_child_taxonomy_embedding ON child_chunks (taxonomy_id, embedding);

-- Create index for page-based retrieval
CREATE INDEX idx_child_page ON child_chunks (page_number) INCLUDE (child_id, parent_id);

-- Create index for content type filtering
CREATE INDEX idx_child_content_type ON child_chunks (content_type) INCLUDE (child_id);

-- Create index for quality-based filtering
CREATE INDEX idx_child_quality_active ON child_chunks (quality_score, is_active) 
    WHERE is_active = TRUE;

-- Analyze tables for query planner
ANALYZE child_chunks;
ANALYZE parent_chunks;
ANALYZE cbse_taxonomy;

-- Add index statistics comment
COMMENT ON INDEX idx_child_embedding_hnsw IS 'HNSW index for approximate nearest neighbor search. m=16, ef_construction=64 optimized for 100k+ chunks with 768 dimensions. Use SET hnsw.ef_search to tune query accuracy.';

-- Create function to dynamically adjust ef_search
CREATE OR REPLACE FUNCTION set_hnsw_ef_search(p_ef_search INTEGER)
RETURNS VOID AS $$
BEGIN
    EXECUTE format('SET LOCAL hnsw.ef_search = %L', p_ef_search);
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION set_hnsw_ef_search IS 'Dynamically adjust HNSW ef_search parameter for query-time accuracy tuning. Higher values = more accurate but slower.';
