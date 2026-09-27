-- Visionary RAG Pipeline - Complete Self-Contained Initialization Script
-- This script contains ALL schema definitions inline - no external dependencies

-- =============================================================================
-- EXTENSIONS
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS btree_gin;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$
BEGIN
  RAISE NOTICE 'Extensions installed: vector, btree_gin, pgcrypto';
END $$;

-- =============================================================================
-- TABLE: cbse_taxonomy
-- =============================================================================

CREATE TABLE IF NOT EXISTS cbse_taxonomy (
  taxonomy_id   SERIAL PRIMARY KEY,
  grade         INTEGER NOT NULL CHECK (grade BETWEEN 6 AND 8),
  subject       VARCHAR(50) NOT NULL,
  chapter       VARCHAR(200) NOT NULL,
  chapter_order INTEGER NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (grade, subject, chapter)
);

CREATE INDEX IF NOT EXISTS idx_taxonomy_grade_subject
  ON cbse_taxonomy (grade, subject);

CREATE INDEX IF NOT EXISTS idx_taxonomy_chapter_order
  ON cbse_taxonomy (grade, subject, chapter_order);

-- =============================================================================
-- TABLE: parent_chunks
-- =============================================================================

CREATE TABLE IF NOT EXISTS parent_chunks (
  parent_id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  taxonomy_id       INTEGER NOT NULL REFERENCES cbse_taxonomy(taxonomy_id) ON DELETE CASCADE,
  content           TEXT NOT NULL,
  extracted_keywords TEXT[] NOT NULL DEFAULT '{}',
  page_number       INTEGER,
  chapter           VARCHAR(200),
  section           VARCHAR(200),
  subsection        VARCHAR(200),
  content_type      VARCHAR(20) NOT NULL DEFAULT 'prose' CHECK (content_type IN ('prose', 'formula', 'table')),
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  version           INTEGER DEFAULT 1,
  is_current        BOOLEAN DEFAULT TRUE,
  superseded_version INTEGER,
  superseded_by     UUID,
  version_ingested  TIMESTAMPTZ DEFAULT NOW(),
  CONSTRAINT chk_parent_content_length CHECK (LENGTH(content) <= 1500)
);

CREATE INDEX IF NOT EXISTS idx_parent_keywords_gin
  ON parent_chunks USING GIN (extracted_keywords);

CREATE INDEX IF NOT EXISTS idx_parent_taxonomy
  ON parent_chunks (taxonomy_id);

CREATE INDEX IF NOT EXISTS idx_parent_taxonomy_content_type
  ON parent_chunks (taxonomy_id, content_type);

CREATE INDEX IF NOT EXISTS idx_parent_chunks_current
  ON parent_chunks(taxonomy_id, page_number)
  WHERE is_current = TRUE;

CREATE INDEX IF NOT EXISTS idx_parent_chunks_version
  ON parent_chunks(taxonomy_id, version);

-- =============================================================================
-- TABLE: child_chunks
-- =============================================================================

CREATE TABLE IF NOT EXISTS child_chunks (
  child_id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  parent_id     UUID NOT NULL REFERENCES parent_chunks(parent_id) ON DELETE CASCADE,
  taxonomy_id   INTEGER NOT NULL REFERENCES cbse_taxonomy(taxonomy_id) ON DELETE CASCADE,
  content       TEXT NOT NULL,
  embedding     vector(768) NOT NULL,
  page_number   INTEGER,
  content_type  VARCHAR(20) NOT NULL DEFAULT 'prose' CHECK (content_type IN ('prose', 'formula', 'table')),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  version       INTEGER DEFAULT 1,
  is_current    BOOLEAN DEFAULT TRUE,
  superseded_version INTEGER,
  version_ingested TIMESTAMPTZ DEFAULT NOW(),
  CONSTRAINT chk_child_content_length CHECK (LENGTH(content) <= 512)
);

-- Use HNSW instead of ScaNN (ScaNN requires AlloyDB, HNSW works with pgvector)
CREATE INDEX IF NOT EXISTS idx_child_embedding_hnsw
  ON child_chunks USING hnsw (embedding vector_cosine_ops)
  WITH (m = 16, ef_construction = 128);

CREATE INDEX IF NOT EXISTS idx_child_taxonomy
  ON child_chunks (taxonomy_id);

CREATE INDEX IF NOT EXISTS idx_child_parent
  ON child_chunks (parent_id);

CREATE INDEX IF NOT EXISTS idx_child_taxonomy_parent
  ON child_chunks (taxonomy_id, parent_id);

CREATE INDEX IF NOT EXISTS idx_child_chunks_current
  ON child_chunks(taxonomy_id, page_number)
  WHERE is_current = TRUE;

CREATE INDEX IF NOT EXISTS idx_child_chunks_version
  ON child_chunks(taxonomy_id, version);

-- =============================================================================
-- TABLE: ingestion_dlq
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
  CONSTRAINT chk_dlq_failed CHECK (failed IN (TRUE, FALSE))
);

CREATE INDEX IF NOT EXISTS idx_dlq_failed
  ON ingestion_dlq (created_at)
  WHERE failed = TRUE;

CREATE INDEX IF NOT EXISTS idx_dlq_retry
  ON ingestion_dlq (retry_count, created_at)
  WHERE failed = TRUE;

-- =============================================================================
-- TABLE: ai_feedback_loop
-- =============================================================================

CREATE TABLE IF NOT EXISTS ai_feedback_loop (
  feedback_id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id            VARCHAR(100) NOT NULL,
  user_query            TEXT NOT NULL,
  retrieved_context     UUID[] NOT NULL,
  llm_response          TEXT NOT NULL,
  feedback_score        INTEGER CHECK (feedback_score IN (-1, 0, 1)),
  user_action           VARCHAR(50),
  golden_response       TEXT,
  processed_for_tuning  BOOLEAN NOT NULL DEFAULT FALSE,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  processed_at          TIMESTAMPTZ,
  CONSTRAINT chk_feedback_context_length CHECK (array_length(retrieved_context, 1) BETWEEN 1 AND 10)
);

CREATE INDEX IF NOT EXISTS idx_feedback_session
  ON ai_feedback_loop (session_id, created_at);

CREATE INDEX IF NOT EXISTS idx_feedback_unprocessed
  ON ai_feedback_loop (created_at)
  WHERE feedback_score = -1 AND processed_for_tuning = FALSE;

CREATE INDEX IF NOT EXISTS idx_feedback_score
  ON ai_feedback_loop (feedback_score, created_at);

CREATE INDEX IF NOT EXISTS idx_feedback_context_gin
  ON ai_feedback_loop USING GIN (retrieved_context);

-- =============================================================================
-- TABLE: document_versions
-- =============================================================================

CREATE TABLE IF NOT EXISTS document_versions (
    taxonomy_id INTEGER PRIMARY KEY,
    content_hash BYTEA NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    ingested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    page_hashes JSONB NOT NULL DEFAULT '{}',
    metadata JSONB,
    CONSTRAINT version_positive CHECK (version > 0)
);

CREATE INDEX IF NOT EXISTS idx_document_versions_taxonomy_version
  ON document_versions(taxonomy_id, version DESC);

-- =============================================================================
-- TRIGGERS
-- =============================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER update_parent_chunks_updated_at
    BEFORE UPDATE ON parent_chunks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE OR REPLACE TRIGGER update_child_chunks_updated_at
    BEFORE UPDATE ON child_chunks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- FUNCTIONS
-- =============================================================================

CREATE OR REPLACE FUNCTION rpc_get_document_version(p_taxonomy_id INTEGER)
RETURNS TABLE (
    taxonomy_id INTEGER,
    content_hash BYTEA,
    version INTEGER,
    ingested_at TIMESTAMPTZ,
    page_hashes JSONB,
    metadata JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        dv.taxonomy_id,
        dv.content_hash,
        dv.version,
        dv.ingested_at,
        dv.page_hashes,
        dv.metadata
    FROM document_versions dv
    WHERE dv.taxonomy_id = p_taxonomy_id
    ORDER BY dv.version DESC
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION rpc_get_version_stats(p_taxonomy_id INTEGER)
RETURNS TABLE (
    version INTEGER,
    parent_count BIGINT,
    child_count BIGINT,
    ingested_at TIMESTAMPTZ,
    is_current BOOLEAN
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        pc.version,
        COUNT(DISTINCT pc.parent_id)::BIGINT,
        COUNT(DISTINCT cc.child_id)::BIGINT,
        pc.version_ingested,
        pc.is_current
    FROM parent_chunks pc
    LEFT JOIN child_chunks cc ON pc.parent_id = cc.parent_id
    WHERE pc.taxonomy_id = p_taxonomy_id
    GROUP BY pc.version, pc.version_ingested, pc.is_current
    ORDER BY pc.version DESC;
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- DATA POPULATION: CBSE Grade 6-8 Science Taxonomy
-- =============================================================================

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
(6, 'Science', 'Garbage In, Garbage Out', 16),
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
(7, 'Science', 'Wastewater Story', 18),
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
-- VERIFICATION
-- =============================================================================

DO $$
DECLARE
  v_table_count INTEGER;
  v_taxonomy_count INTEGER;
BEGIN
  SELECT COUNT(*) INTO v_table_count
  FROM information_schema.tables
  WHERE table_schema = 'public'
  AND table_name IN ('cbse_taxonomy', 'parent_chunks', 'child_chunks', 
                     'ingestion_dlq', 'ai_feedback_loop', 'document_versions');
  
  RAISE NOTICE 'Tables created: % (expected: 6)', v_table_count;
  
  SELECT COUNT(*) INTO v_taxonomy_count FROM cbse_taxonomy;
  RAISE NOTICE 'Taxonomy entries: % (expected: 52)', v_taxonomy_count;
  
  IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector') THEN
    RAISE NOTICE 'vector extension installed';
  ELSE
    RAISE EXCEPTION 'vector extension NOT installed';
  END IF;
  
  IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'btree_gin') THEN
    RAISE NOTICE 'btree_gin extension installed';
  ELSE
    RAISE EXCEPTION 'btree_gin extension NOT installed';
  END IF;
  
  RAISE NOTICE 'Database initialization complete!';
END $$;
