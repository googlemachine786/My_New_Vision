-- ============================================================
-- Self-Query Retriever - Database Index Setup
-- ============================================================
-- Creates pgvector payload indices for efficient metadata filtering
-- Run on AlloyDB production or local PostgreSQL with pgvector
-- ============================================================

-- ============================================================
-- 1. Grade Index
-- ============================================================
-- Optimizes queries like: "grade 8 questions"
-- Cardinality: Low (3 values: 6, 7, 8)
-- Selectivity: High (filters to ~33% of data)
CREATE INDEX IF NOT EXISTS idx_chunks_grade
ON chunks USING gin ((metadata->'grade'));

-- ============================================================
-- 2. Chapter Number Index
-- ============================================================
-- Optimizes queries like: "chapter 9 questions"
-- Cardinality: Medium (18 values: 1-18)
-- Selectivity: High (filters to ~5.5% of data)
CREATE INDEX IF NOT EXISTS idx_chunks_chapter_number
ON chunks USING gin ((metadata->'chapter_number'));

-- ============================================================
-- 3. Content Type Index
-- ============================================================
-- Optimizes queries like: "show me diagrams", "find tables"
-- Cardinality: Low (4 values: Text, Table, Figure, Formula)
-- Selectivity: Medium (varies by content distribution)
CREATE INDEX IF NOT EXISTS idx_chunks_content_type
ON chunks USING gin ((metadata->'content_type'));

-- ============================================================
-- 4. Subject Index
-- ============================================================
-- Optimizes queries like: "science questions", "biology topics"
-- Cardinality: Low (1-3 values: Science, Physics, Chemistry, Biology)
-- Selectivity: Low (most data is Science)
CREATE INDEX IF NOT EXISTS idx_chunks_subject
ON chunks USING gin ((metadata->'subject'));

-- ============================================================
-- 5. Chapter Name Index (Text Search)
-- ============================================================
-- Optimizes queries like: "photosynthesis questions", "force and pressure"
-- Uses ILIKE for case-insensitive partial matching
-- Cardinality: Medium (~50 unique chapter names)
CREATE INDEX IF NOT EXISTS idx_chunks_chapter_name
ON chunks USING gin ((metadata->>'chapter') gin_trgm_ops);

-- Note: Requires pg_trgm extension
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- ============================================================
-- 6. Composite Index: Grade + Chapter Number
-- ============================================================
-- Optimizes queries like: "grade 8 chapter 9 questions"
-- Most common filter combination
-- Selectivity: Very High (filters to ~1.8% of data)
CREATE INDEX IF NOT EXISTS idx_chunks_grade_chapter
ON chunks USING gin (
    (metadata->'grade'),
    (metadata->'chapter_number')
);

-- ============================================================
-- 7. Composite Index: Grade + Content Type
-- ============================================================
-- Optimizes queries like: "grade 8 diagrams", "class 7 tables"
-- Common for visual content requests
CREATE INDEX IF NOT EXISTS idx_chunks_grade_content_type
ON chunks USING gin (
    (metadata->'grade'),
    (metadata->'content_type')
);

-- ============================================================
-- 8. Composite Index: Grade + Chapter Name
-- ============================================================
-- Optimizes queries like: "grade 8 photosynthesis"
-- Common for topic-specific grade queries
CREATE INDEX IF NOT EXISTS idx_chunks_grade_chapter_name
ON chunks USING gin (
    (metadata->'grade'),
    ((metadata->>'chapter') gin_trgm_ops)
);

-- ============================================================
-- 9. Page Number Index (Range Queries)
-- ============================================================
-- Optimizes queries like: "pages 90-100", "chapter near page 95"
-- Supports range scans for page_min/page_max filters
-- Note: Using btree instead of gin for efficient range queries
CREATE INDEX IF NOT EXISTS idx_chunks_page_number
ON chunks USING btree ((metadata->>'page_number')::int);

-- ============================================================
-- 10. Taxonomy ID Index
-- ============================================================
-- Optimizes queries like: "physics questions" (via taxonomy)
-- Supports curriculum-aligned filtering
CREATE INDEX IF NOT EXISTS idx_chunks_taxonomy_id
ON chunks USING gin ((metadata->'taxonomy_id'));

-- ============================================================
-- Index Usage Statistics Query
-- ============================================================
-- Run this to check which indices are being used:
-- SELECT
--     schemaname,
--     tablename,
--     indexname,
--     idx_scan,
--     idx_tup_read,
--     idx_tup_fetch
-- FROM pg_stat_user_indexes
-- WHERE tablename = 'chunks'
-- ORDER BY idx_scan DESC;

-- ============================================================
-- Index Size Analysis Query
-- ============================================================
-- Run this to check index sizes:
-- SELECT
--     indexname,
--     pg_size_pretty(pg_relation_size(indexname::regclass)) as size
-- FROM pg_indexes
-- WHERE tablename = 'chunks'
-- ORDER BY pg_relation_size(indexname::regclass) DESC;

-- ============================================================
-- Verify Indices Created
-- ============================================================
SELECT
    indexname,
    indexdef
FROM pg_indexes
WHERE tablename = 'chunks'
  AND indexname LIKE 'idx_chunks_%'
ORDER BY indexname;

-- ============================================================
-- Expected Output:
-- ============================================================
-- idx_chunks_chapter_name | CREATE INDEX ... USING gin ((metadata->>'chapter') gin_trgm_ops)
-- idx_chunks_chapter_number | CREATE INDEX ... USING gin ((metadata->'chapter_number'))
-- idx_chunks_content_type | CREATE INDEX ... USING gin ((metadata->'content_type'))
-- idx_chunks_grade | CREATE INDEX ... USING gin ((metadata->'grade'))
-- idx_chunks_grade_chapter | CREATE INDEX ... USING gin ((metadata->'grade'), (metadata->'chapter_number'))
-- idx_chunks_grade_chapter_name | CREATE INDEX ... USING gin ((metadata->'grade'), ((metadata->>'chapter') gin_trgm_ops))
-- idx_chunks_grade_content_type | CREATE INDEX ... USING gin ((metadata->'grade'), (metadata->'content_type'))
-- idx_chunks_page_number | CREATE INDEX ... USING btree ((metadata->>'page_number')::int)
-- idx_chunks_subject | CREATE INDEX ... USING gin ((metadata->'subject'))
-- idx_chunks_taxonomy_id | CREATE INDEX ... USING gin ((metadata->'taxonomy_id'))

-- ============================================================
-- Performance Notes:
-- ============================================================
-- - GIN indices are optimal for JSONB key-value lookups
-- - B-tree is better for range queries (page numbers)
-- - Trigram indices enable efficient ILIKE '%pattern%' searches
-- - Composite indices should match common filter combinations
-- - Index creation time: ~1-2 minutes for 100K rows
-- - Index size: ~10-20% of table size per index

-- ============================================================
-- Production Deployment:
-- ============================================================
-- 1. Run during low-traffic window
-- 2. Monitor pg_stat_progress_create_index for long operations
-- 3. ANALYZE chunks after index creation
-- 4. Verify query plans with EXPLAIN ANALYZE

-- Run ANALYZE to update statistics
ANALYZE chunks;

-- Verify index usage with sample query
EXPLAIN ANALYZE
SELECT chunk_id, content, metadata
FROM chunks
WHERE (metadata->>'grade')::int = 8
  AND (metadata->>'chapter_number')::int = 9
LIMIT 5;
