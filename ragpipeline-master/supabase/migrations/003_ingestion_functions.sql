-- ============================================================
-- Migration 003: Production Ingestion Functions
-- Atomic operations with full error handling and DLQ routing
-- ============================================================

-- Function to insert parent-child chunk pairs atomically
CREATE OR REPLACE FUNCTION insert_parent_child(
    p_parent_id UUID,
    p_taxonomy_id INTEGER,
    p_content TEXT,
    p_keywords TEXT[],
    p_page_number INTEGER,
    p_chapter VARCHAR,
    p_section VARCHAR,
    p_subsection VARCHAR,
    p_content_type content_type,
    p_children JSONB,  -- Array of {content, embedding, page_number, content_type, metadata}
    p_metadata JSONB DEFAULT '{}',
    p_quality_score chunk_quality_score DEFAULT 'good'
)
RETURNS UUID AS $$
DECLARE
    v_parent_id UUID;
    v_child JSONB;
    v_content_hash VARCHAR(64);
BEGIN
    -- Generate content hash for deduplication
    v_content_hash := generate_content_hash(p_content);
    
    -- Check for duplicate
    SELECT parent_id INTO v_parent_id
    FROM parent_chunks
    WHERE content_hash = v_content_hash
      AND is_active = TRUE
    LIMIT 1;
    
    IF v_parent_id IS NOT NULL THEN
        -- Duplicate found, return existing
        RETURN v_parent_id;
    END IF;
    
    -- Generate new parent_id if not provided
    IF p_parent_id IS NULL THEN
        p_parent_id := gen_random_uuid();
    END IF;
    
    -- Insert parent
    INSERT INTO parent_chunks (
        parent_id, taxonomy_id, content, content_hash, extracted_keywords,
        page_number, chapter, section, subsection, content_type,
        metadata, quality_score
    ) VALUES (
        p_parent_id, p_taxonomy_id, p_content, v_content_hash, p_keywords,
        p_page_number, p_chapter, p_section, p_subsection, p_content_type,
        p_metadata, p_quality_score
    )
    RETURNING parent_id INTO v_parent_id;
    
    -- Insert children
    FOR v_child IN SELECT * FROM jsonb_array_elements(p_children)
    LOOP
        INSERT INTO child_chunks (
            parent_id, taxonomy_id, content, content_hash, embedding,
            page_number, content_type, metadata, quality_score
        ) VALUES (
            v_parent_id,
            p_taxonomy_id,
            v_child->>'content',
            generate_content_hash(v_child->>'content'),
            (v_child->>'embedding')::vector,
            COALESCE((v_child->>'page_number')::INTEGER, p_page_number),
            COALESCE((v_child->>'content_type')::content_type, p_content_type),
            COALESCE((v_child->>'metadata')::jsonb, '{}'),
            COALESCE((v_child->>'quality_score')::chunk_quality_score, p_quality_score)
        );
    END LOOP;
    
    RETURN v_parent_id;
    
EXCEPTION
    WHEN OTHERS THEN
        -- Log to DLQ with full context
        INSERT INTO ingestion_dlq (
            payload, original_content, error_message, error_type, error_stack
        ) VALUES (
            jsonb_build_object(
                'parent_id', p_parent_id,
                'taxonomy_id', p_taxonomy_id,
                'content_length', LENGTH(p_content),
                'keywords_count', array_length(p_keywords, 1),
                'children_count', jsonb_array_length(p_children),
                'metadata', p_metadata
            ),
            p_content,
            SQLERRM,
            pg_exception_context(),
            pg_exception_detail()
        );
        
        -- Re-raise for caller to handle
        RAISE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function for bulk upsert (incremental ingestion)
CREATE OR REPLACE FUNCTION upsert_parent_child(
    p_parent_id UUID,
    p_taxonomy_id INTEGER,
    p_content TEXT,
    p_keywords TEXT[],
    p_page_number INTEGER,
    p_chapter VARCHAR,
    p_section VARCHAR,
    p_subsection VARCHAR,
    p_content_type content_type,
    p_children JSONB,
    p_metadata JSONB DEFAULT '{}',
    p_change_reason VARCHAR DEFAULT 'Content update'
)
RETURNS UUID AS $$
DECLARE
    v_parent_id UUID;
    v_action VARCHAR;
    v_old_version INTEGER;
BEGIN
    -- Check if parent exists
    SELECT parent_id, version INTO v_parent_id, v_old_version
    FROM parent_chunks
    WHERE parent_id = p_parent_id;
    
    IF v_parent_id IS NULL THEN
        -- Insert new
        v_parent_id := insert_parent_child(
            p_parent_id, p_taxonomy_id, p_content, p_keywords,
            p_page_number, p_chapter, p_section, p_subsection,
            p_content_type, p_children, p_metadata
        );
        v_action := 'INSERT';
    ELSE
        -- Archive old version
        INSERT INTO chunk_versions (
            chunk_type, chunk_id, version_number, content,
            changed_by, change_reason
        )
        SELECT 
            'parent',
            parent_id,
            version,
            content,
            current_user,
            p_change_reason
        FROM parent_chunks
        WHERE parent_id = p_parent_id;
        
        -- Update existing parent
        UPDATE parent_chunks
        SET 
            content = p_content,
            content_hash = generate_content_hash(p_content),
            extracted_keywords = p_keywords,
            page_number = p_page_number,
            chapter = p_chapter,
            section = p_section,
            subsection = p_subsection,
            content_type = p_content_type,
            metadata = p_metadata,
            version = version + 1,
            updated_at = NOW()
        WHERE parent_id = p_parent_id;
        
        -- Delete old children
        DELETE FROM child_chunks WHERE parent_id = p_parent_id;
        
        -- Insert new children
        INSERT INTO child_chunks (
            parent_id, taxonomy_id, content, content_hash, embedding,
            page_number, content_type, metadata
        )
        SELECT 
            p_parent_id,
            p_taxonomy_id,
            child->>'content',
            generate_content_hash(child->>'content'),
            (child->>'embedding')::vector,
            COALESCE((child->>'page_number')::INTEGER, p_page_number),
            COALESCE((child->>'content_type')::content_type, p_content_type),
            COALESCE((child->>'metadata')::jsonb, '{}')
        FROM jsonb_array_elements(p_children) AS child;
        
        v_action := 'UPDATE';
    END IF;
    
    -- Log version change
    INSERT INTO performance_metrics (metric_name, metric_value, dimensions)
    VALUES ('chunk_version_change', 1, jsonb_build_object('action', v_action, 'old_version', v_old_version));
    
    RETURN v_parent_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to get taxonomy ID from grade/subject/chapter
CREATE OR REPLACE FUNCTION get_taxonomy_id(
    p_grade INTEGER,
    p_subject VARCHAR,
    p_chapter VARCHAR,
    p_section VARCHAR DEFAULT NULL
)
RETURNS INTEGER AS $$
DECLARE
    v_taxonomy_id INTEGER;
BEGIN
    SELECT taxonomy_id INTO v_taxonomy_id
    FROM cbse_taxonomy
    WHERE grade = p_grade
      AND subject = p_subject
      AND chapter = p_chapter
      AND (p_section IS NULL OR section = p_section)
    LIMIT 1;
    
    RETURN v_taxonomy_id;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Function to validate chunk before insertion
CREATE OR REPLACE FUNCTION validate_chunk(
    p_content TEXT,
    p_embedding vector,
    p_content_type content_type,
    p_max_length INTEGER DEFAULT 512
)
RETURNS BOOLEAN AS $$
BEGIN
    -- Check content length
    IF LENGTH(p_content) > p_max_length THEN
        RAISE EXCEPTION 'Chunk content exceeds maximum length (% > %)', LENGTH(p_content), p_max_length;
    END IF;
    
    -- Check content not empty
    IF LENGTH(TRIM(p_content)) < 5 THEN
        RAISE EXCEPTION 'Chunk content too short (minimum 5 characters)';
    END IF;
    
    -- Check embedding dimension
    IF vector_dims(p_embedding) != 768 THEN
        RAISE EXCEPTION 'Embedding dimension must be 768, got %', vector_dims(p_embedding);
    END IF;
    
    -- Check embedding not all zeros
    IF p_embedding = array_fill(0, array[768])::vector THEN
        RAISE EXCEPTION 'Embedding cannot be all zeros';
    END IF;
    
    -- Check content type
    IF p_content_type NOT IN ('prose', 'formula', 'table', 'figure', 'equation', 'diagram', 'code') THEN
        RAISE EXCEPTION 'Invalid content type: %', p_content_type;
    END IF;
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- Function to get ingestion statistics
CREATE OR REPLACE FUNCTION get_ingestion_stats(
    p_taxonomy_id INTEGER DEFAULT NULL,
    p_include_details BOOLEAN DEFAULT FALSE
)
RETURNS JSONB AS $$
DECLARE
    v_stats JSONB;
    v_details JSONB;
BEGIN
    -- Base statistics
    SELECT jsonb_build_object(
        'total_parents', COUNT(DISTINCT pc.parent_id),
        'total_children', COUNT(DISTINCT cc.child_id),
        'avg_parent_length', ROUND(AVG(LENGTH(pc.content)), 0),
        'avg_child_length', ROUND(AVG(LENGTH(cc.content)), 0),
        'total_keywords', SUM(COALESCE(array_length(pc.extracted_keywords, 1), 0)),
        'active_chunks', COUNT(DISTINCT cc.child_id) FILTER (WHERE cc.is_active = TRUE),
        'inactive_chunks', COUNT(DISTINCT cc.child_id) FILTER (WHERE cc.is_active = FALSE)
    ) INTO v_stats
    FROM parent_chunks pc
    LEFT JOIN child_chunks cc ON pc.parent_id = cc.parent_id
    WHERE p_taxonomy_id IS NULL OR pc.taxonomy_id = p_taxonomy_id;
    
    -- Detailed breakdown if requested
    IF p_include_details THEN
        SELECT jsonb_build_object(
            'content_types', (
                SELECT jsonb_object_agg(COALESCE(content_type::text, 'unknown'), count)
                FROM (
                    SELECT content_type, COUNT(*) 
                    FROM parent_chunks 
                    WHERE p_taxonomy_id IS NULL OR taxonomy_id = p_taxonomy_id
                    GROUP BY content_type
                ) t
            ),
            'quality_scores', (
                SELECT jsonb_object_agg(COALESCE(quality_score::text, 'unknown'), count)
                FROM (
                    SELECT quality_score, COUNT(*) 
                    FROM child_chunks 
                    WHERE p_taxonomy_id IS NULL OR taxonomy_id = p_taxonomy_id
                    GROUP BY quality_score
                ) t
            ),
            'recent_ingestions', (
                SELECT COUNT(*) 
                FROM parent_chunks 
                WHERE created_at > NOW() - INTERVAL '24 hours'
                  AND (p_taxonomy_id IS NULL OR taxonomy_id = p_taxonomy_id)
            )
        ) INTO v_details;
        
        v_stats := v_stats || v_details;
    END IF;
    
    RETURN COALESCE(v_stats, '{}'::jsonb);
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- Function to process ingestion queue
CREATE OR REPLACE FUNCTION process_ingestion_queue(
    p_batch_size INTEGER DEFAULT 10,
    p_max_retries INTEGER DEFAULT 3
)
RETURNS INTEGER AS $$
DECLARE
    v_queue_item RECORD;
    v_children JSONB;
    v_processed_count INTEGER := 0;
BEGIN
    -- Fetch pending items
    FOR v_queue_item IN 
        SELECT * FROM ingestion_queue
        WHERE status = 'pending'
          AND retry_count < p_max_retries
          AND scheduled_at <= NOW()
        ORDER BY priority DESC, created_at ASC
        LIMIT p_batch_size
        FOR UPDATE SKIP LOCKED
    LOOP
        -- Update status to processing
        UPDATE ingestion_queue
        SET 
            status = 'processing',
            started_at = NOW()
        WHERE queue_id = v_queue_item.queue_id;
        
        -- Process item
        BEGIN
            -- Parse children from metadata
            v_children := v_queue_item.metadata->'children';
            
            -- Insert parent-child
            PERFORM insert_parent_child(
                NULL,  -- Generate new parent_id
                v_queue_item.taxonomy_id,
                v_queue_item.content,
                v_queue_item.metadata->'keywords',
                (v_queue_item.metadata->>'page_number')::INTEGER,
                v_queue_item.metadata->>'chapter',
                v_queue_item.metadata->>'section',
                NULL,  -- subsection
                'prose',
                v_children,
                v_queue_item.metadata
            );
            
            -- Mark as completed
            UPDATE ingestion_queue
            SET 
                status = 'completed',
                completed_at = NOW()
            WHERE queue_id = v_queue_item.queue_id;
            
            v_processed_count := v_processed_count + 1;
            
        EXCEPTION
            WHEN OTHERS THEN
                -- Mark as failed or retry
                IF v_queue_item.retry_count >= p_max_retries - 1 THEN
                    UPDATE ingestion_queue
                    SET 
                        status = 'failed',
                        error_message = SQLERRM,
                        error_type = pg_exception_context()
                    WHERE queue_id = v_queue_item.queue_id;
                ELSE
                    UPDATE ingestion_queue
                    SET 
                        status = 'retrying',
                        retry_count = retry_count + 1,
                        scheduled_at = NOW() + (POWER(2, retry_count) || ' minutes')::INTERVAL,
                        error_message = SQLERRM
                    WHERE queue_id = v_queue_item.queue_id;
                END IF;
        END;
    END LOOP;
    
    RETURN v_processed_count;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION insert_parent_child IS 'Atomically insert parent-child chunk pairs with deduplication and DLQ routing on failure';
COMMENT ON FUNCTION upsert_parent_child IS 'Upsert parent-child pairs for incremental ingestion with version tracking';
COMMENT ON FUNCTION get_taxonomy_id IS 'Get taxonomy ID from grade/subject/chapter hierarchy';
COMMENT ON FUNCTION validate_chunk IS 'Validate chunk before insertion (length, embedding dim, content type, zeros check)';
COMMENT ON FUNCTION get_ingestion_stats IS 'Get ingestion statistics with optional detailed breakdown';
COMMENT ON FUNCTION process_ingestion_queue IS 'Process ingestion queue with retry logic and exponential backoff';
