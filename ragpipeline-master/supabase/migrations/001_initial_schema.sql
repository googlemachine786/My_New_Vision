-- ============================================================
-- Visionary RAG Pipeline - Supabase Migration Scripts
-- Complete database schema for RAG pipeline with pgvector
-- ============================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS btree_gin;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ============================================================
-- ENUMS
-- ============================================================

-- Content types for chunks
CREATE TYPE content_type AS ENUM ('prose', 'formula', 'table', 'figure', 'equation');

-- Feedback scores
CREATE TYPE feedback_score_type AS ENUM ('negative', 'neutral', 'positive');

-- User actions
CREATE TYPE user_action_type AS ENUM ('thumbs_up', 'thumbs_down', 'regenerate', 'share');

-- ============================================================
-- TABLES
-- ============================================================

-- CBSE Taxonomy - Grade/Subject/Chapter hierarchy
CREATE TABLE cbse_taxonomy (
    taxonomy_id SERIAL PRIMARY KEY,
    grade INTEGER NOT NULL CHECK (grade BETWEEN 6 AND 12),
    subject VARCHAR(100) NOT NULL,
    chapter VARCHAR(300) NOT NULL,
    section VARCHAR(300),
    subsection VARCHAR(300),
    chapter_order INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE (grade, subject, chapter, section)
);

-- Parent chunks (1500 chars max)
CREATE TABLE parent_chunks (
    parent_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    taxonomy_id INTEGER NOT NULL REFERENCES cbse_taxonomy(taxonomy_id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    extracted_keywords TEXT[] NOT NULL DEFAULT '{}',
    page_number INTEGER,
    chapter VARCHAR(300),
    section VARCHAR(300),
    subsection VARCHAR(300),
    content_type content_type NOT NULL DEFAULT 'prose',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_parent_content_length CHECK (LENGTH(content) <= 1500)
);

-- Child chunks (512 chars max) with embeddings
CREATE TABLE child_chunks (
    child_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id UUID NOT NULL REFERENCES parent_chunks(parent_id) ON DELETE CASCADE,
    taxonomy_id INTEGER NOT NULL REFERENCES cbse_taxonomy(taxonomy_id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    embedding vector(768) NOT NULL,  -- text-embedding-005 dimensions
    page_number INTEGER,
    content_type content_type NOT NULL DEFAULT 'prose',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_child_content_length CHECK (LENGTH(content) <= 512)
);

-- Ingestion Dead Letter Queue
CREATE TABLE ingestion_dlq (
    dlq_id BIGSERIAL PRIMARY KEY,
    payload JSONB NOT NULL,
    error_message TEXT NOT NULL,
    error_type VARCHAR(100) NOT NULL,
    retry_count INTEGER NOT NULL DEFAULT 0,
    failed BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    
    CONSTRAINT chk_dlq_failed CHECK (failed IN (TRUE, FALSE))
);

-- AI Feedback Loop
CREATE TABLE ai_feedback_loop (
    feedback_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id VARCHAR(200) NOT NULL,
    user_id UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    user_query TEXT NOT NULL,
    retrieved_context UUID[] NOT NULL,  -- Array of child_ids
    llm_response TEXT NOT NULL,
    feedback_score feedback_score_type,
    user_action user_action_type,
    golden_response TEXT,
    processed_for_tuning BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    
    CONSTRAINT chk_feedback_context_length CHECK (array_length(retrieved_context, 1) BETWEEN 1 AND 10)
);

-- Session History (Redis backup in Postgres)
CREATE TABLE session_history (
    session_id VARCHAR(200) NOT NULL,
    turn_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    content TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_session_ttl CHECK (created_at > NOW() - INTERVAL '35 minutes')
);

-- Query Cache
CREATE TABLE query_cache (
    cache_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    query_hash VARCHAR(64) NOT NULL UNIQUE,  -- SHA256 of query embedding
    query_text TEXT NOT NULL,
    answer TEXT NOT NULL,
    sources JSONB NOT NULL,  -- Array of {content, page, section, score}
    ttft_ms INTEGER,
    total_latency_ms INTEGER,
    hits INTEGER NOT NULL DEFAULT 0,
    cached_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    
    CONSTRAINT chk_cache_ttl CHECK (expires_at > NOW())
);

-- Evaluation Golden Dataset
CREATE TABLE golden_qa_dataset (
    qa_id SERIAL PRIMARY KEY,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    grade INTEGER NOT NULL CHECK (grade BETWEEN 6 AND 12),
    subject VARCHAR(100) NOT NULL,
    taxonomy_id INTEGER REFERENCES cbse_taxonomy(taxonomy_id),
    expected_child_ids UUID[],
    difficulty VARCHAR(20) CHECK (difficulty IN ('easy', 'medium', 'hard')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Evaluation Results
CREATE TABLE evaluation_results (
    eval_id SERIAL PRIMARY KEY,
    evaluation_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    test_name VARCHAR(100) NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    metric_value FLOAT NOT NULL,
    target_value FLOAT,
    passed BOOLEAN NOT NULL,
    details JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- INDEXES
-- ============================================================

-- Taxonomy indexes
CREATE INDEX idx_taxonomy_grade_subject ON cbse_taxonomy (grade, subject);
CREATE INDEX idx_taxonomy_chapter_order ON cbse_taxonomy (grade, subject, chapter_order);

-- Parent chunks indexes
CREATE INDEX idx_parent_taxonomy ON parent_chunks (taxonomy_id);
CREATE INDEX idx_parent_content_type ON parent_chunks (content_type);
CREATE INDEX idx_parent_keywords_gin ON parent_chunks USING GIN (extracted_keywords);
CREATE INDEX idx_parent_chapter ON parent_chunks (chapter);
CREATE INDEX idx_parent_section ON parent_chunks (section);

-- Child chunks indexes (CRITICAL for retrieval)
CREATE INDEX idx_child_taxonomy ON child_chunks (taxonomy_id);
CREATE INDEX idx_child_parent ON child_chunks (parent_id);
CREATE INDEX idx_child_content_type ON child_chunks (content_type);

-- HNSW index for dense vector search (ScaNN alternative)
-- num_neighbors should be set based on your dataset size
CREATE INDEX idx_child_embedding_hnsw ON child_chunks 
USING hnsw (embedding vector_cosine_ops) 
WITH (m = 16, ef_construction = 64);

-- GIN index for keyword sparse search
CREATE INDEX idx_child_keywords_gin ON child_chunks USING GIN (extracted_keywords);

-- DLQ indexes
CREATE INDEX idx_dlq_failed ON ingestion_dlq (created_at) WHERE failed = TRUE;
CREATE INDEX idx_dlq_retry ON ingestion_dlq (retry_count, created_at) WHERE failed = TRUE;

-- Feedback loop indexes
CREATE INDEX idx_feedback_session ON ai_feedback_loop (session_id, created_at);
CREATE INDEX idx_feedback_unprocessed ON ai_feedback_loop (created_at) 
    WHERE feedback_score = 'negative' AND processed_for_tuning = FALSE;
CREATE INDEX idx_feedback_score ON ai_feedback_loop (feedback_score, created_at);
CREATE INDEX idx_feedback_context_gin ON ai_feedback_loop USING GIN (retrieved_context);
CREATE INDEX idx_feedback_user ON ai_feedback_loop (user_id, created_at);

-- Session history indexes
CREATE INDEX idx_session_history_session ON session_history (session_id, created_at);
CREATE INDEX idx_session_history_created ON session_history (created_at);

-- Query cache indexes
CREATE INDEX idx_query_cache_hash ON query_cache (query_hash);
CREATE INDEX idx_query_cache_expires ON query_cache (expires_at);
CREATE INDEX idx_query_cache_hits ON query_cache (hits DESC);

-- Golden dataset indexes
CREATE INDEX idx_golden_grade_subject ON golden_qa_dataset (grade, subject);
CREATE INDEX idx_golden_taxonomy ON golden_qa_dataset (taxonomy_id);

-- Evaluation results indexes
CREATE INDEX idx_evaluation_date ON evaluation_results (evaluation_date);
CREATE INDEX idx_evaluation_test ON evaluation_results (test_name, evaluation_date);

-- ============================================================
-- FUNCTIONS
-- ============================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function to calculate cosine similarity
CREATE OR REPLACE FUNCTION cosine_similarity(vec1 vector, vec2 vector)
RETURNS FLOAT AS $$
BEGIN
    RETURN 1 - (vec1 <=> vec2);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function to find similar chunks (hybrid search)
CREATE OR REPLACE FUNCTION hybrid_search(
    query_embedding vector(768),
    query_keywords TEXT[],
    p_taxonomy_id INTEGER,
    p_top_k INTEGER DEFAULT 5
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
    rrf_score FLOAT
) AS $$
BEGIN
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
            0.0 AS cosine_score,
            cardinality(pc.extracted_keywords & query_keywords) AS keyword_overlap,
            ROW_NUMBER() OVER (ORDER BY cardinality(pc.extracted_keywords & query_keywords) DESC) AS sparse_rank
        FROM parent_chunks pc
        WHERE pc.taxonomy_id = p_taxonomy_id
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
            (1.0 / (60.0 + COALESCE(d.dense_rank, 999999))) + 
            (1.0 / (60.0 + COALESCE(s.sparse_rank, 999999))) AS rrf_score
        FROM dense d
        FULL OUTER JOIN sparse s ON d.child_id = s.child_id
    )
    SELECT *
    FROM rrf
    ORDER BY rrf_score DESC
    LIMIT p_top_k;
END;
$$ LANGUAGE plpgsql STABLE;

-- Function to increment cache hits
CREATE OR REPLACE FUNCTION increment_cache_hits(p_cache_id UUID)
RETURNS VOID AS $$
BEGIN
    UPDATE query_cache
    SET hits = hits + 1
    WHERE cache_id = p_cache_id;
END;
$$ LANGUAGE plpgsql;

-- Function to cleanup expired cache entries
CREATE OR REPLACE FUNCTION cleanup_expired_cache()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM query_cache
    WHERE expires_at < NOW();
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Function to cleanup old sessions
CREATE OR REPLACE FUNCTION cleanup_old_sessions()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM session_history
    WHERE created_at < NOW() - INTERVAL '35 minutes';
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- TRIGGERS
-- ============================================================

-- Update updated_at for parent_chunks
CREATE TRIGGER update_parent_chunks_updated_at
    BEFORE UPDATE ON parent_chunks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Update updated_at for cbse_taxonomy
CREATE TRIGGER update_cbse_taxonomy_updated_at
    BEFORE UPDATE ON cbse_taxonomy
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- ROW LEVEL SECURITY (RLS)
-- ============================================================

-- Enable RLS on all tables
ALTER TABLE cbse_taxonomy ENABLE ROW LEVEL SECURITY;
ALTER TABLE parent_chunks ENABLE ROW LEVEL SECURITY;
ALTER TABLE child_chunks ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_feedback_loop ENABLE ROW LEVEL SECURITY;
ALTER TABLE session_history ENABLE ROW LEVEL SECURITY;
ALTER TABLE query_cache ENABLE ROW LEVEL SECURITY;

-- Policies for cbse_taxonomy (public read)
CREATE POLICY "Taxonomy is publicly readable"
    ON cbse_taxonomy FOR SELECT
    USING (true);

-- Policies for parent_chunks (authenticated read for retrieval)
CREATE POLICY "Parent chunks readable by authenticated users"
    ON parent_chunks FOR SELECT
    TO authenticated
    USING (true);

-- Policies for child_chunks (authenticated read for retrieval)
CREATE POLICY "Child chunks readable by authenticated users"
    ON child_chunks FOR SELECT
    TO authenticated
    USING (true);

-- Policies for feedback loop (users can insert their own feedback)
CREATE POLICY "Users can insert their own feedback"
    ON ai_feedback_loop FOR INSERT
    TO authenticated
    WITH CHECK (auth.uid() = user_id OR user_id IS NULL);

CREATE POLICY "Users can view their own feedback"
    ON ai_feedback_loop FOR SELECT
    TO authenticated
    USING (auth.uid() = user_id OR user_id IS NULL);

-- Policies for session history (users can only access their own sessions)
CREATE POLICY "Users can manage their own sessions"
    ON session_history FOR ALL
    TO authenticated
    USING (session_id LIKE '%' || auth.uid()::text || '%');

-- Policies for query cache (public read for caching)
CREATE POLICY "Query cache is publicly readable"
    ON query_cache FOR SELECT
    USING (true);

-- ============================================================
-- SEED DATA
-- ============================================================

-- Insert CBSE Grade 6-8 Science taxonomy
INSERT INTO cbse_taxonomy (grade, subject, chapter, chapter_order) VALUES
-- Grade 6 Science
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

-- Grade 7 Science
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

-- Grade 8 Science
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

-- ============================================================
-- VIEWS
-- ============================================================

-- View for chunk statistics
CREATE VIEW v_chunk_statistics AS
SELECT 
    ct.grade,
    ct.subject,
    ct.chapter,
    COUNT(DISTINCT pc.parent_id) AS parent_count,
    COUNT(DISTINCT cc.child_id) AS child_count,
    AVG(LENGTH(cc.content)) AS avg_child_length,
    AVG(LENGTH(pc.content)) AS avg_parent_length
FROM cbse_taxonomy ct
LEFT JOIN parent_chunks pc ON ct.taxonomy_id = pc.taxonomy_id
LEFT JOIN child_chunks cc ON pc.parent_id = cc.parent_id
GROUP BY ct.grade, ct.subject, ct.chapter
ORDER BY ct.grade, ct.subject, ct.chapter_order;

-- View for feedback statistics
CREATE VIEW v_feedback_statistics AS
SELECT 
    DATE(created_at) AS feedback_date,
    feedback_score,
    user_action,
    COUNT(*) AS count,
    COUNT(*) FILTER (WHERE processed_for_tuning = TRUE) AS processed_count
FROM ai_feedback_loop
GROUP BY DATE(created_at), feedback_score, user_action
ORDER BY feedback_date DESC;

-- View for cache performance
CREATE VIEW v_cache_performance AS
SELECT 
    DATE(cached_at) AS cache_date,
    COUNT(*) AS total_cached,
    SUM(hits) AS total_hits,
    AVG(hits) AS avg_hits_per_query,
    COUNT(*) FILTER (WHERE hits > 0) AS hit_queries,
    COUNT(*) FILTER (WHERE hits = 0) AS miss_queries
FROM query_cache
GROUP BY DATE(cached_at)
ORDER BY cache_date DESC;

-- ============================================================
-- SCHEDULED JOBS (using pg_cron extension)
-- ============================================================

-- Enable pg_cron if available (Supabase Pro feature)
-- CREATE EXTENSION IF NOT EXISTS pg_cron;

-- Cleanup expired cache every hour
-- SELECT cron.schedule(
--     'cleanup-cache-hourly',
--     '0 * * * *',
--     'SELECT cleanup_expired_cache()'
-- );

-- Cleanup old sessions every 30 minutes
-- SELECT cron.schedule(
--     'cleanup-sessions',
--     '*/30 * * * *',
--     'SELECT cleanup_old_sessions()'
-- );

-- ============================================================
-- COMMENTS
-- ============================================================

COMMENT ON TABLE cbse_taxonomy IS 'CBSE curriculum taxonomy for grades 6-12';
COMMENT ON TABLE parent_chunks IS 'Parent chunks (1500 chars max) with keywords for sparse search';
COMMENT ON TABLE child_chunks IS 'Child chunks (512 chars max) with embeddings for dense search';
COMMENT ON TABLE ingestion_dlq IS 'Dead letter queue for failed ingestion attempts';
COMMENT ON TABLE ai_feedback_loop IS 'Student feedback and LLM judge golden responses';
COMMENT ON TABLE session_history IS 'Session history backup (Redis is primary)';
COMMENT ON TABLE query_cache IS 'Response cache for frequent queries';
COMMENT ON FUNCTION hybrid_search IS 'Hybrid search combining dense (HNSW) and sparse (GIN) retrieval with RRF fusion';

-- ============================================================
-- MIGRATION COMPLETE
-- ============================================================
