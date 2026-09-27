-- ============================================================
-- Migration 004: Production RPC Functions
-- Supabase Edge Functions integration with full error handling
-- ============================================================

-- RPC function for hybrid search (callable from Edge Functions)
CREATE OR REPLACE FUNCTION rpc_hybrid_search(
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
    rrf_score FLOAT
) SECURITY DEFINER STABLE AS $$
BEGIN
    -- Set ef_search for better accuracy
    PERFORM set_hnsw_ef_search(p_ef_search);
    
    RETURN QUERY
    SELECT * FROM hybrid_search(
        query_embedding,
        query_keywords,
        p_taxonomy_id,
        p_top_k,
        p_ef_search
    );
END;
$$ LANGUAGE plpgsql;

-- RPC function for semantic similarity search
CREATE OR REPLACE FUNCTION rpc_semantic_search(
    query_embedding vector(768),
    p_taxonomy_id INTEGER,
    p_top_k INTEGER DEFAULT 5,
    p_min_score FLOAT DEFAULT 0.5
)
RETURNS TABLE (
    child_id UUID,
    parent_id UUID,
    content TEXT,
    parent_content TEXT,
    page_number INTEGER,
    section VARCHAR,
    cosine_score FLOAT
) SECURITY DEFINER STABLE AS $$
BEGIN
    RETURN QUERY
    SELECT 
        cc.child_id,
        cc.parent_id,
        cc.content,
        pc.content AS parent_content,
        cc.page_number,
        pc.section,
        (1 - (cc.embedding <=> query_embedding)) AS cosine_score
    FROM child_chunks cc
    JOIN parent_chunks pc ON cc.parent_id = pc.parent_id
    WHERE pc.taxonomy_id = p_taxonomy_id
      AND cc.is_active = TRUE
      AND pc.is_active = TRUE
      AND (1 - (cc.embedding <=> query_embedding)) >= p_min_score
    ORDER BY cc.embedding <=> query_embedding
    LIMIT p_top_k;
END;
$$ LANGUAGE plpgsql;

-- RPC function for keyword search
CREATE OR REPLACE FUNCTION rpc_keyword_search(
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
    keyword_overlap INTEGER
) SECURITY DEFINER STABLE AS $$
BEGIN
    RETURN QUERY
    SELECT 
        cc.child_id,
        cc.parent_id,
        cc.content,
        pc.content AS parent_content,
        cc.page_number,
        pc.section,
        cardinality(pc.extracted_keywords & query_keywords) AS keyword_overlap
    FROM child_chunks cc
    JOIN parent_chunks pc ON cc.parent_id = pc.parent_id
    WHERE pc.taxonomy_id = p_taxonomy_id
      AND pc.is_active = TRUE
      AND cc.is_active = TRUE
      AND pc.extracted_keywords & query_keywords <> '{}'
    ORDER BY keyword_overlap DESC
    LIMIT p_top_k;
END;
$$ LANGUAGE plpgsql;

-- RPC function to log feedback
CREATE OR REPLACE FUNCTION rpc_log_feedback(
    p_session_id VARCHAR,
    p_user_query TEXT,
    p_retrieved_context UUID[],
    p_llm_response TEXT,
    p_feedback_score feedback_score_type DEFAULT NULL,
    p_user_action user_action_type DEFAULT NULL,
    p_quality_metrics JSONB DEFAULT '{}'
)
RETURNS UUID SECURITY DEFINER AS $$
DECLARE
    v_feedback_id UUID;
BEGIN
    INSERT INTO ai_feedback_loop (
        session_id, user_query, retrieved_context, llm_response,
        feedback_score, user_action, quality_metrics
    ) VALUES (
        p_session_id, p_user_query, p_retrieved_context, p_llm_response,
        p_feedback_score, p_user_action, p_quality_metrics
    )
    RETURNING feedback_id INTO v_feedback_id;
    
    RETURN v_feedback_id;
END;
$$ LANGUAGE plpgsql;

-- RPC function to get session history
CREATE OR REPLACE FUNCTION rpc_get_session_history(
    p_session_id VARCHAR,
    p_last_n INTEGER DEFAULT 10
)
RETURNS TABLE (
    turn_id UUID,
    role VARCHAR,
    content TEXT,
    created_at TIMESTAMPTZ
) SECURITY DEFINER STABLE AS $$
BEGIN
    RETURN QUERY
    SELECT 
        sh.turn_id,
        sh.role,
        sh.content,
        sh.created_at
    FROM session_history sh
    WHERE sh.session_id = p_session_id
    ORDER BY sh.created_at DESC
    LIMIT p_last_n;
END;
$$ LANGUAGE plpgsql;

-- RPC function to append session turn
CREATE OR REPLACE FUNCTION rpc_append_session_turn(
    p_session_id VARCHAR,
    p_role VARCHAR,
    p_content TEXT,
    p_metadata JSONB DEFAULT '{}'
)
RETURNS UUID SECURITY DEFINER AS $$
DECLARE
    v_turn_id UUID;
BEGIN
    INSERT INTO session_history (
        session_id, role, content, metadata
    ) VALUES (
        p_session_id, p_role, p_content, p_metadata
    )
    RETURNING turn_id INTO v_turn_id;
    
    RETURN v_turn_id;
END;
$$ LANGUAGE plpgsql;

-- RPC function to get cached response
CREATE OR REPLACE FUNCTION rpc_get_cached_response(
    p_query_hash VARCHAR
)
RETURNS TABLE (
    answer TEXT,
    sources JSONB,
    hits INTEGER
) SECURITY DEFINER STABLE AS $$
BEGIN
    -- Increment hits atomically
    UPDATE query_cache
    SET 
        hits = hits + 1,
        last_hit_at = NOW()
    WHERE query_hash = p_query_hash
      AND expires_at > NOW();
    
    -- Return cached response
    RETURN QUERY
    SELECT 
        qc.answer,
        qc.sources,
        qc.hits
    FROM query_cache qc
    WHERE qc.query_hash = p_query_hash
      AND qc.expires_at > NOW();
END;
$$ LANGUAGE plpgsql;

-- RPC function to cache response
CREATE OR REPLACE FUNCTION rpc_cache_response(
    p_query_hash VARCHAR,
    p_query_text TEXT,
    p_answer TEXT,
    p_sources JSONB,
    p_ttft_ms INTEGER,
    p_total_latency_ms INTEGER,
    p_ttl_seconds INTEGER DEFAULT 300
)
RETURNS UUID SECURITY DEFINER AS $$
DECLARE
    v_cache_id UUID;
BEGIN
    INSERT INTO query_cache (
        query_hash, query_text, answer, sources,
        ttft_ms, total_latency_ms, expires_at
    ) VALUES (
        p_query_hash, p_query_text, p_answer, p_sources,
        p_ttft_ms, p_total_latency_ms,
        NOW() + (p_ttl_seconds || ' seconds')::INTERVAL
    )
    RETURNING cache_id INTO v_cache_id;
    
    RETURN v_cache_id;
END;
$$ LANGUAGE plpgsql;

-- RPC function to get ingestion stats
CREATE OR REPLACE FUNCTION rpc_get_ingestion_stats(
    p_taxonomy_id INTEGER DEFAULT NULL,
    p_include_details BOOLEAN DEFAULT FALSE
)
RETURNS JSONB SECURITY DEFINER STABLE AS $$
BEGIN
    RETURN get_ingestion_stats(p_taxonomy_id, p_include_details);
END;
$$ LANGUAGE plpgsql;

-- RPC function to increment retrieval count
CREATE OR REPLACE FUNCTION rpc_increment_retrieval(
    p_child_id UUID
)
RETURNS VOID SECURITY DEFINER AS $$
BEGIN
    PERFORM increment_retrieval_count(p_child_id);
END;
$$ LANGUAGE plpgsql;

-- RPC function to cleanup expired data
CREATE OR REPLACE FUNCTION rpc_cleanup_expired(
    p_cleanup_cache BOOLEAN DEFAULT TRUE,
    p_cleanup_sessions BOOLEAN DEFAULT TRUE,
    p_session_ttl_minutes INTEGER DEFAULT 35
)
RETURNS JSONB SECURITY DEFINER AS $$
DECLARE
    v_cache_deleted INTEGER := 0;
    v_sessions_deleted INTEGER := 0;
BEGIN
    -- Cleanup cache
    IF p_cleanup_cache THEN
        v_cache_deleted := cleanup_expired_cache();
    END IF;
    
    -- Cleanup sessions
    IF p_cleanup_sessions THEN
        v_sessions_deleted := cleanup_old_sessions(p_session_ttl_minutes);
    END IF;
    
    RETURN jsonb_build_object(
        'cache_deleted', v_cache_deleted,
        'sessions_deleted', v_sessions_deleted
    );
END;
$$ LANGUAGE plpgsql;

-- Grant execute permissions to appropriate roles
GRANT EXECUTE ON FUNCTION rpc_hybrid_search TO authenticated, service_role;
GRANT EXECUTE ON FUNCTION rpc_semantic_search TO authenticated, service_role;
GRANT EXECUTE ON FUNCTION rpc_keyword_search TO authenticated, service_role;
GRANT EXECUTE ON FUNCTION rpc_log_feedback TO authenticated;
GRANT EXECUTE ON FUNCTION rpc_get_session_history TO authenticated;
GRANT EXECUTE ON FUNCTION rpc_append_session_turn TO authenticated;
GRANT EXECUTE ON FUNCTION rpc_get_cached_response TO authenticated, service_role;
GRANT EXECUTE ON FUNCTION rpc_cache_response TO service_role;
GRANT EXECUTE ON FUNCTION rpc_get_ingestion_stats TO authenticated, service_role;
GRANT EXECUTE ON FUNCTION rpc_increment_retrieval TO service_role;
GRANT EXECUTE ON FUNCTION rpc_cleanup_expired TO service_role;

COMMENT ON FUNCTION rpc_hybrid_search IS 'RPC: Hybrid search with RRF fusion (dense HNSW + sparse GIN). Supports configurable ef_search for accuracy tuning.';
COMMENT ON FUNCTION rpc_semantic_search IS 'RPC: Dense vector search using HNSW index with minimum score threshold.';
COMMENT ON FUNCTION rpc_keyword_search IS 'RPC: Sparse keyword search using GIN index on extracted_keywords.';
COMMENT ON FUNCTION rpc_log_feedback IS 'RPC: Log user feedback for quality loop with optional quality metrics.';
COMMENT ON FUNCTION rpc_get_session_history IS 'RPC: Get last N turns from session history.';
COMMENT ON FUNCTION rpc_append_session_turn IS 'RPC: Append turn to session history.';
COMMENT ON FUNCTION rpc_get_cached_response IS 'RPC: Get cached response by query hash with atomic hit increment.';
COMMENT ON FUNCTION rpc_cache_response IS 'RPC: Cache response with configurable TTL.';
COMMENT ON FUNCTION rpc_get_ingestion_stats IS 'RPC: Get ingestion statistics with optional detailed breakdown.';
COMMENT ON FUNCTION rpc_increment_retrieval IS 'RPC: Increment retrieval count for popularity tracking.';
COMMENT ON FUNCTION rpc_cleanup_expired IS 'RPC: Cleanup expired cache entries and old sessions.';
