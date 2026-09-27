-- ============================================================
-- Visionary RAG Pipeline - PRODUCTION Vector Database Schema
-- Enterprise-grade pgvector implementation for Supabase
-- ============================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS btree_gin;
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS pg_trgm;  -- For text similarity
CREATE EXTENSION IF NOT EXISTS fuzzystrmatch;  -- For fuzzy matching

-- ============================================================
-- CONFIGURATION TABLES
-- ============================================================

-- Vector database configuration
CREATE TABLE vector_db_config (
    config_key VARCHAR(100) PRIMARY KEY,
    config_value JSONB NOT NULL,
    description TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert default configuration
INSERT INTO vector_db_config (config_key, config_value, description) VALUES
('hnsw', '{"m": 16, "ef_construction": 64, "ef_search": 40}'::jsonb, 'HNSW index parameters'),
('embedding', '{"model": "text-embedding-005", "dimensions": 768, "task_type": "RETRIEVAL_DOCUMENT"}'::jsonb, 'Embedding model config'),
('chunking', '{"parent_max_chars": 1500, "child_max_chars": 512, "overlap_chars": 77}'::jsonb, 'Chunking parameters'),
('retrieval', '{"top_k": 5, "rrf_k": 60, "min_score": 0.5}'::jsonb, 'Retrieval parameters'),
('cache', '{"ttl_seconds": 300, "max_entries": 10000}'::jsonb, 'Cache configuration');

-- ============================================================
-- CORE ENUMS
-- ============================================================

CREATE TYPE content_type AS ENUM ('prose', 'formula', 'table', 'figure', 'equation', 'diagram', 'code');
CREATE TYPE feedback_score_type AS ENUM ('negative', 'neutral', 'positive');
CREATE TYPE user_action_type AS ENUM ('thumbs_up', 'thumbs_down', 'regenerate', 'share', 'bookmark');
CREATE TYPE ingestion_status AS ENUM ('pending', 'processing', 'completed', 'failed', 'retrying');
CREATE TYPE chunk_quality_score AS ENUM ('excellent', 'good', 'fair', 'poor');

-- ============================================================
-- TAXONOMY & CURRICULUM
-- ============================================================

CREATE TABLE cbse_taxonomy (
    taxonomy_id SERIAL PRIMARY KEY,
    grade INTEGER NOT NULL CHECK (grade BETWEEN 1 AND 12),
    subject VARCHAR(100) NOT NULL,
    chapter VARCHAR(300) NOT NULL,
    section VARCHAR(300),
    subsection VARCHAR(300),
    chapter_order INTEGER NOT NULL,
    learning_objectives TEXT[],
    keywords TEXT[],
    difficulty_level VARCHAR(20) CHECK (difficulty_level IN ('beginner', 'intermediate', 'advanced')),
    estimated_duration_minutes INTEGER,
    prerequisites INTEGER[] REFERENCES cbse_taxonomy(taxonomy_id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE (grade, subject, chapter, section)
);

-- ============================================================
-- CHUNK TABLES (Core Vector Storage)
-- ============================================================

-- Parent chunks (1500 chars max) - for context
CREATE TABLE parent_chunks (
    parent_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    taxonomy_id INTEGER NOT NULL REFERENCES cbse_taxonomy(taxonomy_id) ON DELETE CASCADE,
    source_document_id UUID,
    content TEXT NOT NULL,
    content_hash VARCHAR(64) NOT NULL,  -- SHA256 for deduplication
    extracted_keywords TEXT[] NOT NULL DEFAULT '{}',
    keyword_scores FLOAT[] DEFAULT '{}',  -- TF-IDF scores for keywords
    page_number INTEGER,
    chapter VARCHAR(300),
    section VARCHAR(300),
    subsection VARCHAR(300),
    content_type content_type NOT NULL DEFAULT 'prose',
    language VARCHAR(10) DEFAULT 'en',
    quality_score chunk_quality_score DEFAULT 'good',
    metadata JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_parent_content_length CHECK (LENGTH(content) <= 1500),
    CONSTRAINT chk_parent_content_not_empty CHECK (LENGTH(TRIM(content)) > 10)
);

-- Child chunks (512 chars max) - for embedding & retrieval
CREATE TABLE child_chunks (
    child_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id UUID NOT NULL REFERENCES parent_chunks(parent_id) ON DELETE CASCADE,
    taxonomy_id INTEGER NOT NULL REFERENCES cbse_taxonomy(taxonomy_id) ON DELETE CASCADE,
    source_document_id UUID,
    content TEXT NOT NULL,
    content_hash VARCHAR(64) NOT NULL,  -- SHA256 for deduplication
    embedding vector(768) NOT NULL,  -- text-embedding-005
    embedding_model VARCHAR(50) DEFAULT 'text-embedding-005',
    embedding_version INTEGER DEFAULT 1,
    page_number INTEGER,
    content_type content_type NOT NULL DEFAULT 'prose',
    language VARCHAR(10) DEFAULT 'en',
    quality_score chunk_quality_score DEFAULT 'good',
    retrieval_count INTEGER NOT NULL DEFAULT 0,
    last_retrieved_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_child_content_length CHECK (LENGTH(content) <= 512),
    CONSTRAINT chk_child_content_not_empty CHECK (LENGTH(TRIM(content)) > 5),
    CONSTRAINT chk_embedding_dimension CHECK (vector_dims(embedding) = 768)
);

-- ============================================================
-- DEDUPLICATION & VERSIONING
-- ============================================================

-- Content hash index for deduplication
CREATE UNIQUE INDEX idx_parent_content_hash ON parent_chunks (content_hash) WHERE is_active = TRUE;
CREATE UNIQUE INDEX idx_child_content_hash ON child_chunks (content_hash) WHERE is_active = TRUE;

-- Versioning table for audit trail
CREATE TABLE chunk_versions (
    version_id BIGSERIAL PRIMARY KEY,
    chunk_type VARCHAR(20) NOT NULL CHECK (chunk_type IN ('parent', 'child')),
    chunk_id UUID NOT NULL,
    version_number INTEGER NOT NULL,
    content TEXT NOT NULL,
    embedding vector(768),
    changed_by VARCHAR(100),
    change_reason VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE (chunk_type, chunk_id, version_number)
);

-- ============================================================
-- INDEXES (Production-Optimized)
-- ============================================================

-- Taxonomy indexes
CREATE INDEX idx_taxonomy_grade_subject ON cbse_taxonomy (grade, subject) INCLUDE (chapter, section);
CREATE INDEX idx_taxonomy_chapter_order ON cbse_taxonomy (grade, subject, chapter_order);
CREATE INDEX idx_taxonomy_keywords_gin ON cbse_taxonomy USING GIN (keywords);

-- Parent chunks indexes
CREATE INDEX idx_parent_taxonomy ON parent_chunks (taxonomy_id) INCLUDE (parent_id, content_type);
CREATE INDEX idx_parent_content_type ON parent_chunks (content_type);
CREATE INDEX idx_parent_keywords_gin ON parent_chunks USING GIN (extracted_keywords);
CREATE INDEX idx_parent_chapter_trgm ON parent_chunks USING GIN (chapter gin_trgm_ops);
CREATE INDEX idx_parent_section_trgm ON parent_chunks USING GIN (section gin_trgm_ops);
CREATE INDEX idx_parent_quality ON parent_chunks (quality_score, is_active);
CREATE INDEX idx_parent_created ON parent_chunks (created_at DESC);

-- Child chunks indexes (CRITICAL for retrieval performance)
CREATE INDEX idx_child_taxonomy ON child_chunks (taxonomy_id) INCLUDE (child_id, parent_id);
CREATE INDEX idx_child_parent ON child_chunks (parent_id);
CREATE INDEX idx_child_content_type ON child_chunks (content_type);
CREATE INDEX idx_child_quality ON child_chunks (quality_score, is_active);
CREATE INDEX idx_child_retrieval ON child_chunks (retrieval_count DESC, last_retrieved_at);

-- HNSW index for approximate nearest neighbor (ANN) search
-- Optimized for ~100k chunks with 768 dimensions
CREATE INDEX idx_child_embedding_hnsw ON child_chunks 
USING hnsw (embedding vector_cosine_ops) 
WITH (
    m = 16,              -- Number of connections per node
    ef_construction = 64 -- Size of dynamic candidate list during construction
);

-- GIN index for keyword sparse search
CREATE INDEX idx_child_keywords_gin ON child_chunks USING GIN (
    array_to_string(extracted_keywords, ' ') gin_trgm_ops
);

-- Composite index for taxonomy-filtered vector search
CREATE INDEX idx_child_taxonomy_embedding ON child_chunks (taxonomy_id, embedding);

-- Partial index for active chunks only (faster queries)
CREATE INDEX idx_child_active_embedding ON child_chunks (embedding) WHERE is_active = TRUE;

-- ============================================================
-- INGESTION PIPELINE
-- ============================================================

-- Ingestion queue for batch processing
CREATE TABLE ingestion_queue (
    queue_id BIGSERIAL PRIMARY KEY,
    batch_id UUID NOT NULL DEFAULT gen_random_uuid(),
    source_document_id UUID,
    taxonomy_id INTEGER REFERENCES cbse_taxonomy(taxonomy_id),
    content TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    status ingestion_status NOT NULL DEFAULT 'pending',
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    error_message TEXT,
    error_type VARCHAR(100),
    priority INTEGER NOT NULL DEFAULT 0,
    scheduled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_ingestion_status CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'retrying'))
);

-- Ingestion dead letter queue
CREATE TABLE ingestion_dlq (
    dlq_id BIGSERIAL PRIMARY KEY,
    queue_id BIGINT REFERENCES ingestion_queue(queue_id) ON DELETE SET NULL,
    payload JSONB NOT NULL,
    original_content TEXT,
    error_message TEXT NOT NULL,
    error_type VARCHAR(100) NOT NULL,
    error_stack TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    failed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    resolution_notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- FEEDBACK & QUALITY LOOP
-- ============================================================

-- AI Feedback loop
CREATE TABLE ai_feedback_loop (
    feedback_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id VARCHAR(200) NOT NULL,
    user_id UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    user_query TEXT NOT NULL,
    query_embedding vector(768),
    retrieved_context UUID[] NOT NULL,  -- Array of child_ids
    retrieved_context_scores FLOAT[] DEFAULT '{}',
    llm_response TEXT NOT NULL,
    llm_model VARCHAR(100),
    feedback_score feedback_score_type,
    user_action user_action_type,
    user_action_metadata JSONB DEFAULT '{}',
    golden_response TEXT,
    golden_response_model VARCHAR(100),
    processed_for_tuning BOOLEAN NOT NULL DEFAULT FALSE,
    tuning_job_id VARCHAR(100),
    quality_metrics JSONB DEFAULT '{}',  -- {relevance, faithfulness, completeness}
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    
    CONSTRAINT chk_feedback_context_length CHECK (array_length(retrieved_context, 1) BETWEEN 1 AND 10)
);

-- Feedback statistics materialized view
CREATE MATERIALIZED VIEW mv_feedback_statistics AS
SELECT 
    DATE(created_at) AS feedback_date,
    feedback_score,
    user_action,
    COUNT(*) AS count,
    COUNT(*) FILTER (WHERE processed_for_tuning = TRUE) AS processed_count,
    AVG(quality_metrics->>'relevance') AS avg_relevance,
    AVG(quality_metrics->>'faithfulness') AS avg_faithfulness,
    AVG(quality_metrics->>'completeness') AS avg_completeness
FROM ai_feedback_loop
GROUP BY DATE(created_at), feedback_score, user_action
WITH DATA;

-- ============================================================
-- SESSION & CACHE MANAGEMENT
-- ============================================================

-- Session history (Postgres backup, Redis is primary)
CREATE TABLE session_history (
    session_id VARCHAR(200) NOT NULL,
    turn_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    content TEXT NOT NULL,
    content_embedding vector(768),
    metadata JSONB DEFAULT '{}',
    token_count INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_session_role CHECK (role IN ('user', 'assistant', 'system'))
);

-- Query response cache
CREATE TABLE query_cache (
    cache_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    query_hash VARCHAR(64) NOT NULL UNIQUE,  -- SHA256 of query embedding
    query_text TEXT NOT NULL,
    query_embedding vector(768),
    answer TEXT NOT NULL,
    answer_tokens INTEGER,
    sources JSONB NOT NULL,  -- Array of {child_id, content, page, section, score}
    ttft_ms INTEGER,
    total_latency_ms INTEGER,
    hits INTEGER NOT NULL DEFAULT 0,
    last_hit_at TIMESTAMPTZ,
    cached_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    
    CONSTRAINT chk_cache_ttl CHECK (expires_at > NOW())
);

-- ============================================================
-- EVALUATION & TESTING
-- ============================================================

-- Golden QA dataset for evaluation
CREATE TABLE golden_qa_dataset (
    qa_id SERIAL PRIMARY KEY,
    question TEXT NOT NULL,
    question_embedding vector(768),
    answer TEXT NOT NULL,
    grade INTEGER NOT NULL CHECK (grade BETWEEN 1 AND 12),
    subject VARCHAR(100) NOT NULL,
    taxonomy_id INTEGER REFERENCES cbse_taxonomy(taxonomy_id),
    expected_child_ids UUID[],
    expected_child_scores FLOAT[] DEFAULT '{}',
    difficulty VARCHAR(20) CHECK (difficulty IN ('easy', 'medium', 'hard')),
    tags TEXT[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Evaluation results tracking
CREATE TABLE evaluation_results (
    eval_id SERIAL PRIMARY KEY,
    evaluation_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    test_name VARCHAR(100) NOT NULL,
    test_run_id VARCHAR(100),
    metric_name VARCHAR(100) NOT NULL,
    metric_value FLOAT NOT NULL,
    target_value FLOAT,
    passed BOOLEAN NOT NULL,
    details JSONB DEFAULT '{}',
    qa_id INTEGER REFERENCES golden_qa_dataset(qa_id),
    retrieved_count INTEGER,
    retrieval_latency_ms INTEGER,
    generation_latency_ms INTEGER,
    total_latency_ms INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- ANALYTICS & MONITORING
-- ============================================================

-- Query analytics
CREATE TABLE query_analytics (
    analytics_id BIGSERIAL PRIMARY KEY,
    query_hash VARCHAR(64),
    query_text TEXT,
    user_id UUID,
    session_id VARCHAR(200),
    taxonomy_id INTEGER,
    retrieval_count INTEGER,
    cache_hit BOOLEAN DEFAULT FALSE,
    ttft_ms INTEGER,
    total_latency_ms INTEGER,
    feedback_score feedback_score_type,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Performance metrics
CREATE TABLE performance_metrics (
    metric_id BIGSERIAL PRIMARY KEY,
    metric_name VARCHAR(100) NOT NULL,
    metric_value FLOAT NOT NULL,
    metric_unit VARCHAR(50),
    dimensions JSONB DEFAULT '{}',
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- FUNCTIONS (Production-Grade)
-- ============================================================

-- Update timestamp trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Generate content hash for deduplication
CREATE OR REPLACE FUNCTION generate_content_hash(content TEXT)
RETURNS VARCHAR(64) AS $$
BEGIN
    RETURN encode(sha256(convert_to(TRIM(LOWER(content)), 'UTF8')), 'hex');
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Calculate cosine similarity
CREATE OR REPLACE FUNCTION cosine_similarity(vec1 vector, vec2 vector)
RETURNS FLOAT AS $$
BEGIN
    RETURN 1 - (vec1 <=> vec2);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Hybrid search with RRF fusion (production-optimized)
CREATE OR REPLACE FUNCTION hybrid_search(
    query_embedding vector(768),
    query_keywords TEXT[],
    p_taxonomy_id INTEGER,
    p_top_k INTEGER DEFAULT 5,
    p_ef_search INTEGER DEFAULT 40
)
RETURNS TABLE (
    child_id UUID,
    parent_id UUID,
    content TEXT,
    parent_content TEXT,
    page_number INTEGER,
    section VARCHAR,
    cosine_score FLOAT,
    keyword_score FLOAT,
    rrf_score FLOAT,
    rank_dense BIGINT,
    rank_sparse BIGINT
) AS $$
DECLARE
    v_ef_search_prev INTEGER;
BEGIN
    -- Set ef_search for this session (better accuracy)
    SELECT current_setting('hnsw.ef_search')::INTEGER INTO v_ef_search_prev;
    EXECUTE format('SET LOCAL hnsw.ef_search = %L', p_ef_search);
    
    RETURN QUERY
    WITH dense AS (
        SELECT 
            cc.child_id,
            cc.parent_id,
            cc.content,
            pc.content AS parent_content,
            cc.page_number,
            pc.section,
            (1 - (cc.embedding <=> query_embedding)) AS cosine_score,
            ROW_NUMBER() OVER (ORDER BY cc.embedding <=> query_embedding) AS dense_rank
        FROM child_chunks cc
        JOIN parent_chunks pc ON cc.parent_id = pc.parent_id
        WHERE pc.taxonomy_id = p_taxonomy_id
          AND cc.is_active = TRUE
          AND pc.is_active = TRUE
        ORDER BY cc.embedding <=> query_embedding
        LIMIT p_top_k * 20
    ),
    sparse AS (
        SELECT 
            pc.parent_id AS child_id,
            pc.parent_id,
            pc.content,
            pc.content AS parent_content,
            pc.page_number,
            pc.section,
            cardinality(pc.extracted_keywords & query_keywords) AS keyword_overlap,
            ROW_NUMBER() OVER (ORDER BY cardinality(pc.extracted_keywords & query_keywords) DESC) AS sparse_rank
        FROM parent_chunks pc
        WHERE pc.taxonomy_id = p_taxonomy_id
          AND pc.is_active = TRUE
          AND pc.extracted_keywords & query_keywords <> '{}'
        LIMIT p_top_k * 20
    ),
    rrf AS (
        SELECT 
            COALESCE(d.child_id, s.child_id) AS child_id,
            COALESCE(d.parent_id, s.parent_id) AS parent_id,
            COALESCE(d.content, s.content) AS content,
            COALESCE(d.parent_content, s.parent_content) AS parent_content,
            COALESCE(d.page_number, s.page_number) AS page_number,
            COALESCE(d.section, s.section) AS section,
            COALESCE(d.cosine_score, 0.0) AS cosine_score,
            COALESCE(s.keyword_overlap, 0) AS keyword_score,
            (1.0 / (60.0 + COALESCE(d.dense_rank, 999999)::FLOAT)) + 
            (1.0 / (60.0 + COALESCE(s.sparse_rank, 999999)::FLOAT)) AS rrf_score,
            d.dense_rank AS rank_dense,
            s.sparse_rank AS rank_sparse
        FROM dense d
        FULL OUTER JOIN sparse s ON d.child_id = s.child_id
    )
    SELECT *
    FROM rrf
    ORDER BY rrf_score DESC
    LIMIT p_top_k;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Increment retrieval count (for popularity tracking)
CREATE OR REPLACE FUNCTION increment_retrieval_count(p_child_id UUID)
RETURNS VOID AS $$
BEGIN
    UPDATE child_chunks
    SET 
        retrieval_count = retrieval_count + 1,
        last_retrieved_at = NOW()
    WHERE child_id = p_child_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Cleanup expired cache entries
CREATE OR REPLACE FUNCTION cleanup_expired_cache()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM query_cache
    WHERE expires_at < NOW();
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    -- Log cleanup metrics
    INSERT INTO performance_metrics (metric_name, metric_value, metric_unit, dimensions)
    VALUES ('cache_cleanup_count', deleted_count, 'entries', '{"operation": "cleanup"}');
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Cleanup old sessions
CREATE OR REPLACE FUNCTION cleanup_old_sessions(p_older_than_minutes INTEGER DEFAULT 35)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM session_history
    WHERE created_at < NOW() - (p_older_than_minutes || ' minutes')::INTERVAL;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Get ingestion statistics
CREATE OR REPLACE FUNCTION get_ingestion_stats(p_taxonomy_id INTEGER DEFAULT NULL)
RETURNS JSONB AS $$
DECLARE
    v_stats JSONB;
BEGIN
    SELECT jsonb_build_object(
        'total_parents', COUNT(DISTINCT pc.parent_id),
        'total_children', COUNT(DISTINCT cc.child_id),
        'avg_parent_length', AVG(LENGTH(pc.content)),
        'avg_child_length', AVG(LENGTH(cc.content)),
        'total_keywords', SUM(COALESCE(array_length(pc.extracted_keywords, 1), 0)),
        'active_chunks', COUNT(DISTINCT cc.child_id) FILTER (WHERE cc.is_active = TRUE),
        'content_types', jsonb_object_agg(
            COALESCE(pc.content_type::text, 'unknown'),
            COALESCE(type_count, 0)
        )
    ) INTO v_stats
    FROM parent_chunks pc
    LEFT JOIN child_chunks cc ON pc.parent_id = cc.parent_id
    LEFT JOIN (
        SELECT content_type, COUNT(*) as type_count
        FROM parent_chunks
        WHERE p_taxonomy_id IS NULL OR taxonomy_id = p_taxonomy_id
        GROUP BY content_type
    ) tc ON pc.content_type = tc.content_type
    WHERE p_taxonomy_id IS NULL OR pc.taxonomy_id = p_taxonomy_id;
    
    RETURN COALESCE(v_stats, '{}'::jsonb);
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- ============================================================
-- TRIGGERS
-- ============================================================

-- Auto-update timestamps
CREATE TRIGGER update_parent_chunks_updated_at
    BEFORE UPDATE ON parent_chunks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_cbse_taxonomy_updated_at
    BEFORE UPDATE ON cbse_taxonomy
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Auto-generate content hash before insert
CREATE TRIGGER generate_parent_content_hash
    BEFORE INSERT OR UPDATE OF content ON parent_chunks
    FOR EACH ROW
    EXECUTE FUNCTION (
        NEW.content_hash := generate_content_hash(NEW.content);
        RETURN NEW;
    );

CREATE TRIGGER generate_child_content_hash
    BEFORE INSERT OR UPDATE OF content ON child_chunks
    FOR EACH ROW
    EXECUTE FUNCTION (
        NEW.content_hash := generate_content_hash(NEW.content);
        RETURN NEW;
    );

-- ============================================================
-- ROW LEVEL SECURITY (RLS)
-- ============================================================

-- Enable RLS
ALTER TABLE cbse_taxonomy ENABLE ROW LEVEL SECURITY;
ALTER TABLE parent_chunks ENABLE ROW LEVEL SECURITY;
ALTER TABLE child_chunks ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_feedback_loop ENABLE ROW LEVEL SECURITY;
ALTER TABLE session_history ENABLE ROW LEVEL SECURITY;
ALTER TABLE query_cache ENABLE ROW LEVEL SECURITY;
ALTER TABLE golden_qa_dataset ENABLE ROW LEVEL SECURITY;

-- Taxonomy: Public read
CREATE POLICY "Taxonomy is publicly readable"
    ON cbse_taxonomy FOR SELECT
    USING (true);

-- Chunks: Authenticated read for retrieval
CREATE POLICY "Parent chunks readable by authenticated users"
    ON parent_chunks FOR SELECT
    TO authenticated
    USING (is_active = TRUE);

CREATE POLICY "Child chunks readable by authenticated users"
    ON child_chunks FOR SELECT
    TO authenticated
    USING (is_active = TRUE);

-- Feedback: User-specific
CREATE POLICY "Users can insert their own feedback"
    ON ai_feedback_loop FOR INSERT
    TO authenticated
    WITH CHECK (auth.uid() = user_id OR user_id IS NULL);

CREATE POLICY "Users can view their own feedback"
    ON ai_feedback_loop FOR SELECT
    TO authenticated
    USING (auth.uid() = user_id OR user_id IS NULL);

-- Sessions: User-specific
CREATE POLICY "Users can manage their own sessions"
    ON session_history FOR ALL
    TO authenticated
    USING (session_id LIKE '%' || auth.uid()::text || '%');

-- Cache: Public read
CREATE POLICY "Query cache is publicly readable"
    ON query_cache FOR SELECT
    USING (true);

-- Golden dataset: Service role only
CREATE POLICY "Golden dataset readable by service role"
    ON golden_qa_dataset FOR SELECT
    TO service_role
    USING (true);

-- ============================================================
-- SEED DATA
-- ============================================================

-- Insert CBSE Grade 6-8 Science taxonomy
INSERT INTO cbse_taxonomy (grade, subject, chapter, chapter_order, difficulty_level) VALUES
-- Grade 6 Science
(6, 'Science', 'Food: Where Does It Come From?', 1, 'beginner'),
(6, 'Science', 'Components of Food', 2, 'beginner'),
(6, 'Science', 'Fibre to Fabric', 3, 'beginner'),
(6, 'Science', 'Sorting Materials into Groups', 4, 'beginner'),
(6, 'Science', 'Separation of Substances', 5, 'intermediate'),
(6, 'Science', 'Changes Around Us', 6, 'beginner'),
(6, 'Science', 'Getting to Know Plants', 7, 'beginner'),
(6, 'Science', 'Body Movements', 8, 'beginner'),
(6, 'Science', 'The Living Organisms and Their Surroundings', 9, 'intermediate'),
(6, 'Science', 'Motion and Measurement of Distances', 10, 'intermediate'),
(6, 'Science', 'Light, Shadows and Reflections', 11, 'intermediate'),
(6, 'Science', 'Electricity and Circuits', 12, 'intermediate'),
(6, 'Science', 'Fun with Magnets', 13, 'beginner'),
(6, 'Science', 'Water', 14, 'beginner'),
(6, 'Science', 'Air Around Us', 15, 'beginner'),
(6, 'Science', 'Garbage In, Garbage Out', 16, 'beginner'),

-- Grade 7 Science
(7, 'Science', 'Nutrition in Plants', 1, 'intermediate'),
(7, 'Science', 'Nutrition in Animals', 2, 'intermediate'),
(7, 'Science', 'Fibre to Fabric', 3, 'beginner'),
(7, 'Science', 'Heat', 4, 'intermediate'),
(7, 'Science', 'Acids, Bases and Salts', 5, 'intermediate'),
(7, 'Science', 'Physical and Chemical Changes', 6, 'intermediate'),
(7, 'Science', 'Weather, Climate and Adaptations of Animals to Climate', 7, 'advanced'),
(7, 'Science', 'Winds, Storms and Cyclones', 8, 'advanced'),
(7, 'Science', 'Soil', 9, 'beginner'),
(7, 'Science', 'Respiration in Organisms', 10, 'intermediate'),
(7, 'Science', 'Transportation in Animals and Plants', 11, 'intermediate'),
(7, 'Science', 'Reproduction in Plants', 12, 'intermediate'),
(7, 'Science', 'Motion and Time', 13, 'intermediate'),
(7, 'Science', 'Electric Current and Its Effects', 14, 'intermediate'),
(7, 'Science', 'Light', 15, 'intermediate'),
(7, 'Science', 'Water: A Precious Resource', 16, 'beginner'),
(7, 'Science', 'Forests: Our Lifeline', 17, 'intermediate'),
(7, 'Science', 'Wastewater Story', 18, 'beginner'),

-- Grade 8 Science
(8, 'Science', 'Crop Production and Management', 1, 'intermediate'),
(8, 'Science', 'Microorganisms: Friend and Foe', 2, 'intermediate'),
(8, 'Science', 'Synthetic Fibres and Plastics', 3, 'intermediate'),
(8, 'Science', 'Materials: Metals and Non-Metals', 4, 'intermediate'),
(8, 'Science', 'Coal and Petroleum', 5, 'beginner'),
(8, 'Science', 'Combustion and Flame', 6, 'intermediate'),
(8, 'Science', 'Conservation of Plants and Animals', 7, 'beginner'),
(8, 'Science', 'Cell — Structure and Functions', 8, 'advanced'),
(8, 'Science', 'Reproduction in Animals', 9, 'intermediate'),
(8, 'Science', 'Reaching the Age of Adolescence', 10, 'intermediate'),
(8, 'Science', 'Force and Pressure', 11, 'intermediate'),
(8, 'Science', 'Friction', 12, 'beginner'),
(8, 'Science', 'Sound', 13, 'intermediate'),
(8, 'Science', 'Chemical Effects of Electric Current', 14, 'advanced'),
(8, 'Science', 'Some Natural Phenomena', 15, 'intermediate'),
(8, 'Science', 'Light', 16, 'intermediate'),
(8, 'Science', 'Stars and the Solar System', 17, 'beginner'),
(8, 'Science', 'Pollution of Air and Water', 18, 'intermediate')
ON CONFLICT (grade, subject, chapter) DO UPDATE SET
    difficulty_level = EXCLUDED.difficulty_level;

-- ============================================================
-- VIEWS & MATERIALIZED VIEWS
-- ============================================================

-- Chunk statistics view
CREATE OR REPLACE VIEW v_chunk_statistics AS
SELECT 
    ct.grade,
    ct.subject,
    ct.chapter,
    ct.section,
    COUNT(DISTINCT pc.parent_id) FILTER (WHERE pc.is_active = TRUE) AS parent_count,
    COUNT(DISTINCT cc.child_id) FILTER (WHERE cc.is_active = TRUE) AS child_count,
    AVG(LENGTH(cc.content)) AS avg_child_length,
    AVG(LENGTH(pc.content)) AS avg_parent_length,
    AVG(cc.retrieval_count) AS avg_retrieval_count,
    MAX(cc.last_retrieved_at) AS last_retrieved_at
FROM cbse_taxonomy ct
LEFT JOIN parent_chunks pc ON ct.taxonomy_id = pc.taxonomy_id
LEFT JOIN child_chunks cc ON pc.parent_id = cc.parent_id
GROUP BY ct.grade, ct.subject, ct.chapter, ct.section
ORDER BY ct.grade, ct.subject, ct.chapter_order;

-- Feedback statistics view
CREATE OR REPLACE VIEW v_feedback_statistics AS
SELECT 
    DATE(created_at) AS feedback_date,
    feedback_score,
    user_action,
    COUNT(*) AS count,
    COUNT(*) FILTER (WHERE processed_for_tuning = TRUE) AS processed_count,
    ROUND(100.0 * COUNT(*) FILTER (WHERE feedback_score = 'positive') / NULLIF(COUNT(*), 0), 2) AS positive_rate,
    ROUND(100.0 * COUNT(*) FILTER (WHERE feedback_score = 'negative') / NULLIF(COUNT(*), 0), 2) AS negative_rate
FROM ai_feedback_loop
GROUP BY DATE(created_at), feedback_score, user_action
ORDER BY feedback_date DESC;

-- Cache performance view
CREATE OR REPLACE VIEW v_cache_performance AS
SELECT 
    DATE(cached_at) AS cache_date,
    COUNT(*) AS total_cached,
    SUM(hits) AS total_hits,
    ROUND(AVG(hits), 2) AS avg_hits_per_query,
    COUNT(*) FILTER (WHERE hits > 0) AS hit_queries,
    COUNT(*) FILTER (WHERE hits = 0) AS miss_queries,
    ROUND(100.0 * COUNT(*) FILTER (WHERE hits > 0) / NULLIF(COUNT(*), 0), 2) AS hit_rate_percent
FROM query_cache
GROUP BY DATE(cached_at)
ORDER BY cache_date DESC;

-- ============================================================
-- COMMENTS
-- ============================================================

COMMENT ON TABLE parent_chunks IS 'Parent chunks (1500 chars max) with keywords for sparse search. Supports versioning and deduplication.';
COMMENT ON TABLE child_chunks IS 'Child chunks (512 chars max) with 768-dim embeddings for dense search. HNSW index optimized for ANN search.';
COMMENT ON TABLE ingestion_queue IS 'Ingestion queue for batch processing with retry logic and priority scheduling.';
COMMENT ON TABLE ingestion_dlq IS 'Dead letter queue for failed ingestion attempts with resolution tracking.';
COMMENT ON TABLE ai_feedback_loop IS 'Student feedback and LLM judge golden responses for quality loop and fine-tuning.';
COMMENT ON FUNCTION hybrid_search IS 'Hybrid search combining HNSW dense + GIN sparse retrieval with RRF fusion (k=60). Supports configurable ef_search.';
COMMENT ON INDEX idx_child_embedding_hnsw IS 'HNSW index for approximate nearest neighbor search. m=16, ef_construction=64 optimized for ~100k chunks with 768 dimensions.';

-- ============================================================
-- MIGRATION COMPLETE
-- ============================================================
