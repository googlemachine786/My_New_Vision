-- Visionary RAG Pipeline - Production Schema v2.0
-- AlloyDB for PostgreSQL 15 with pgvector + ScaNN
-- CBSE Science Grades 6-8 | 1,000+ Concurrent Students
--
-- CRITICAL: Apply to AlloyDB ONLY (not Cloud SQL - ScaNN requires AlloyDB)
-- Apply Order: Extensions → Tables → Indexes → Data Population
--
-- Usage: psql $ALLOYDB_DSN -f schema/v2_production.sql

-- =============================================================================
-- EXTENSIONS
-- =============================================================================

-- pgvector: Vector similarity search (required for embeddings)
CREATE EXTENSION IF NOT EXISTS vector;

-- btree_gin: GIN index support for B-tree operations (required for hybrid search)
CREATE EXTENSION IF NOT EXISTS btree_gin;

-- pgcrypto: UUID generation and cryptographic functions
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Verify extensions loaded
DO $$
BEGIN
  RAISE NOTICE 'Extensions installed: vector, btree_gin, pgcrypto';
END $$;

-- =============================================================================
-- TABLE: cbse_taxonomy
-- Purpose: Grade/subject/chapter hierarchy for content filtering
-- Index: B-tree on (grade, subject) for fast JWT claim lookups
-- =============================================================================

CREATE TABLE IF NOT EXISTS cbse_taxonomy (
  taxonomy_id   SERIAL PRIMARY KEY,
  grade         INTEGER NOT NULL CHECK (grade BETWEEN 6 AND 8),
  subject       VARCHAR(50) NOT NULL,
  chapter       VARCHAR(200) NOT NULL,
  chapter_order INTEGER NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  
  -- Ensure unique grade+subject+chapter combinations
  UNIQUE (grade, subject, chapter)
);

-- Index for JWT claim → taxonomy_id lookups
CREATE INDEX IF NOT EXISTS idx_taxonomy_grade_subject 
  ON cbse_taxonomy (grade, subject);

-- Index for chapter listing within grade+subject
CREATE INDEX IF NOT EXISTS idx_taxonomy_chapter_order 
  ON cbse_taxonomy (grade, subject, chapter_order);

-- =============================================================================
-- TABLE: parent_chunks
-- Purpose: Parent chunks (1500 chars max) with keyword extraction for sparse search
-- Critical: extracted_keywords is TEXT[] (not TEXT) for GIN index
-- =============================================================================

CREATE TABLE IF NOT EXISTS parent_chunks (
  parent_id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  taxonomy_id       INTEGER NOT NULL REFERENCES cbse_taxonomy(taxonomy_id) ON DELETE CASCADE,
  content           TEXT NOT NULL,
  extracted_keywords TEXT[] NOT NULL DEFAULT '{}',  -- CRITICAL: TEXT[] array type
  page_number       INTEGER,
  chapter           VARCHAR(200),
  section           VARCHAR(200),
  subsection        VARCHAR(200),
  content_type      VARCHAR(20) NOT NULL DEFAULT 'prose' CHECK (content_type IN ('prose', 'formula', 'table')),
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  
  -- Content length validation (1500 char max for parent chunks)
  CONSTRAINT chk_parent_content_length CHECK (LENGTH(content) <= 1500)
);

-- GIN index for keyword array intersection queries (&& operator)
-- CRITICAL: Uses array_ops for TEXT[] containment checks
CREATE INDEX IF NOT EXISTS idx_parent_keywords_gin 
  ON parent_chunks USING GIN (extracted_keywords);

-- Index for taxonomy filtering (applied on EVERY query)
CREATE INDEX IF NOT EXISTS idx_parent_taxonomy 
  ON parent_chunks (taxonomy_id);

-- Composite index for taxonomy + content_type filtering
CREATE INDEX IF NOT EXISTS idx_parent_taxonomy_content_type 
  ON parent_chunks (taxonomy_id, content_type);

-- =============================================================================
-- TABLE: child_chunks
-- Purpose: Child chunks (512 chars max) with embeddings for dense search
-- Index: ScaNN index on embedding (768-dim cosine similarity)
-- =============================================================================

CREATE TABLE IF NOT EXISTS child_chunks (
  child_id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  parent_id     UUID NOT NULL REFERENCES parent_chunks(parent_id) ON DELETE CASCADE,
  taxonomy_id   INTEGER NOT NULL REFERENCES cbse_taxonomy(taxonomy_id) ON DELETE CASCADE,
  content       TEXT NOT NULL,
  embedding     vector(768) NOT NULL,  -- text-embedding-005, 768 dimensions
  page_number   INTEGER,
  content_type  VARCHAR(20) NOT NULL DEFAULT 'prose' CHECK (content_type IN ('prose', 'formula', 'table')),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  
  -- Content length validation (512 char max for child chunks)
  CONSTRAINT chk_child_content_length CHECK (LENGTH(content) <= 512)
);

-- ScaNN index for cosine similarity search (dense retrieval)
-- CRITICAL: Requires AlloyDB (will fail on Cloud SQL)
-- num_leaves = sqrt(row_count) formula for optimal performance
CREATE INDEX IF NOT EXISTS idx_child_embedding_scann 
  ON child_chunks USING ScaNN (embedding vector_cosine_ops)
  WITH (num_leaves = 1000, max_num_levels = 2);

-- Index for taxonomy filtering (applied on EVERY query)
CREATE INDEX IF NOT EXISTS idx_child_taxonomy 
  ON child_chunks (taxonomy_id);

-- Index for parent lookup (JOIN optimization)
CREATE INDEX IF NOT EXISTS idx_child_parent 
  ON child_chunks (parent_id);

-- Composite index for taxonomy + parent lookup
CREATE INDEX IF NOT EXISTS idx_child_taxonomy_parent 
  ON child_chunks (taxonomy_id, parent_id);

-- =============================================================================
-- TABLE: ingestion_dlq
-- Purpose: Dead letter queue for failed ingestion attempts
-- Index: Partial index on failed=true for quick retry queries
-- =============================================================================

CREATE TABLE IF NOT EXISTS ingestion_dlq (
  dlq_id        BIGSERIAL PRIMARY KEY,
  payload       JSONB NOT NULL,
  error_message TEXT NOT NULL,
  error_type    VARCHAR(100) NOT NULL,
  retry_count   INTEGER NOT NULL DEFAULT 0,
  failed        BOOLEAN NOT NULL DEFAULT TRUE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  processed_at  TIMESTAMPTZ,
  
  -- Index for querying failed items
  CONSTRAINT chk_dlq_failed CHECK (failed IN (TRUE, FALSE))
);

-- Partial index: Only index failed items (smaller, faster)
CREATE INDEX IF NOT EXISTS idx_dlq_failed 
  ON ingestion_dlq (created_at) 
  WHERE failed = TRUE;

-- Index for retry count (prioritize low-retry items)
CREATE INDEX IF NOT EXISTS idx_dlq_retry 
  ON ingestion_dlq (retry_count, created_at) 
  WHERE failed = TRUE;

-- =============================================================================
-- TABLE: ai_feedback_loop
-- Purpose: Student feedback and LLM judge golden responses
-- Critical: retrieved_context is UUID[] (not UUID) for storing k chunk IDs
-- =============================================================================

CREATE TABLE IF NOT EXISTS ai_feedback_loop (
  feedback_id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id            VARCHAR(100) NOT NULL,
  user_query            TEXT NOT NULL,
  retrieved_context     UUID[] NOT NULL,  -- CRITICAL: UUID[] array type (stores k chunk IDs)
  llm_response          TEXT NOT NULL,
  feedback_score        INTEGER CHECK (feedback_score IN (-1, 0, 1)),  -- -1=negative, 0=neutral, 1=positive
  user_action           VARCHAR(50),  -- thumbs_up, thumbs_down, regenerate, etc.
  golden_response       TEXT,  -- Populated by LLM Judge for negative feedback
  processed_for_tuning  BOOLEAN NOT NULL DEFAULT FALSE,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  processed_at          TIMESTAMPTZ,
  
  -- Validate UUID array length (should match TopK configuration)
  CONSTRAINT chk_feedback_context_length CHECK (array_length(retrieved_context, 1) BETWEEN 1 AND 10)
);

-- Index for session lookup (conversation history)
CREATE INDEX IF NOT EXISTS idx_feedback_session 
  ON ai_feedback_loop (session_id, created_at);

-- Index for unprocessed negative feedback (quality loop queries)
CREATE INDEX IF NOT EXISTS idx_feedback_unprocessed 
  ON ai_feedback_loop (created_at) 
  WHERE feedback_score = -1 AND processed_for_tuning = FALSE;

-- Index for feedback score analysis
CREATE INDEX IF NOT EXISTS idx_feedback_score 
  ON ai_feedback_loop (feedback_score, created_at);

-- GIN index for UUID array containment (find feedback containing specific chunk)
CREATE INDEX IF NOT EXISTS idx_feedback_context_gin 
  ON ai_feedback_loop USING GIN (retrieved_context);

-- =============================================================================
-- DATA POPULATION: CBSE Grade 6-8 Science Taxonomy
-- Sample chapters for each grade (expand as needed)
-- =============================================================================

-- Grade 6 Science
INSERT INTO cbse_taxonomy (grade, subject, chapter, chapter_order) VALUES
(6, 'Science', 'Food: Where Does It Come From?', 1),
(6, 'Science', 'Components of Food', 2),
(6, 'Science', 'Fibre to Fabric', 3),
(6, 'Science', 'Sorting Materials into Groups', 4),
(6, 'Science', 'Separation of Substances', 5),
(6, 'Science', 'Changes Around Us', 6),
(6, 'Science', 'Getting to Know Plants', 7),
(6, 'Science', 'Body Movements', 8),
(6, 'Science', 'The Living Organisms and Their Surroundings', 9),
(6, 'Science', 'Motion and Measurement of Distances', 10),
(6, 'Science', 'Light, Shadows and Reflections', 11),
(6, 'Science', 'Electricity and Circuits', 12),
(6, 'Science', 'Fun with Magnets', 13),
(6, 'Science', 'Water', 14),
(6, 'Science', 'Air Around Us', 15),
(6, 'Science', 'Garbage In, Garbage Out', 16)
ON CONFLICT (grade, subject, chapter) DO NOTHING;

-- Grade 7 Science
INSERT INTO cbse_taxonomy (grade, subject, chapter, chapter_order) VALUES
(7, 'Science', 'Nutrition in Plants', 1),
(7, 'Science', 'Nutrition in Animals', 2),
(7, 'Science', 'Fibre to Fabric', 3),
(7, 'Science', 'Heat', 4),
(7, 'Science', 'Acids, Bases and Salts', 5),
(7, 'Science', 'Physical and Chemical Changes', 6),
(7, 'Science', 'Weather, Climate and Adaptations of Animals to Climate', 7),
(7, 'Science', 'Winds, Storms and Cyclones', 8),
(7, 'Science', 'Soil', 9),
(7, 'Science', 'Respiration in Organisms', 10),
(7, 'Science', 'Transportation in Animals and Plants', 11),
(7, 'Science', 'Reproduction in Plants', 12),
(7, 'Science', 'Motion and Time', 13),
(7, 'Science', 'Electric Current and Its Effects', 14),
(7, 'Science', 'Light', 15),
(7, 'Science', 'Water: A Precious Resource', 16),
(7, 'Science', 'Forests: Our Lifeline', 17),
(7, 'Science', 'Wastewater Story', 18)
ON CONFLICT (grade, subject, chapter) DO NOTHING;

-- Grade 8 Science
INSERT INTO cbse_taxonomy (grade, subject, chapter, chapter_order) VALUES
(8, 'Science', 'Crop Production and Management', 1),
(8, 'Science', 'Microorganisms: Friend and Foe', 2),
(8, 'Science', 'Synthetic Fibres and Plastics', 3),
(8, 'Science', 'Materials: Metals and Non-Metals', 4),
(8, 'Science', 'Coal and Petroleum', 5),
(8, 'Science', 'Combustion and Flame', 6),
(8, 'Science', 'Conservation of Plants and Animals', 7),
(8, 'Science', 'Cell — Structure and Functions', 8),
(8, 'Science', 'Reproduction in Animals', 9),
(8, 'Science', 'Reaching the Age of Adolescence', 10),
(8, 'Science', 'Force and Pressure', 11),
(8, 'Science', 'Friction', 12),
(8, 'Science', 'Sound', 13),
(8, 'Science', 'Chemical Effects of Electric Current', 14),
(8, 'Science', 'Some Natural Phenomena', 15),
(8, 'Science', 'Light', 16),
(8, 'Science', 'Stars and the Solar System', 17),
(8, 'Science', 'Pollution of Air and Water', 18)
ON CONFLICT (grade, subject, chapter) DO NOTHING;

-- =============================================================================
-- VERIFICATION QUERIES
-- Run these after schema application to verify correctness
-- =============================================================================

DO $$
DECLARE
  v_count INTEGER;
BEGIN
  -- Verify taxonomy population
  SELECT COUNT(*) INTO v_count FROM cbse_taxonomy;
  RAISE NOTICE 'Taxonomy entries: % (expected: 52)', v_count;
  
  -- Verify extensions
  PERFORM 1 FROM pg_extension WHERE extname = 'vector' LIMIT 1;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'vector extension not installed';
  END IF;
  
  PERFORM 1 FROM pg_extension WHERE extname = 'btree_gin' LIMIT 1;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'btree_gin extension not installed';
  END IF;
  
  RAISE NOTICE 'Schema v2.0 applied successfully';
END $$;

-- =============================================================================
-- SCHEMA APPLICATION COMPLETE
-- Next: Run scripts/verify-schema.sql for detailed verification
-- =============================================================================
