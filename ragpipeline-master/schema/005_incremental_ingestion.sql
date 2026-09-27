-- Visionary RAG Pipeline - Incremental Ingestion Schema
-- Adds document versioning and change tracking support

-- Document versions table for tracking incremental changes
CREATE TABLE IF NOT EXISTS document_versions (
    taxonomy_id INTEGER PRIMARY KEY,
    content_hash BYTEA NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    ingested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    page_hashes JSONB NOT NULL DEFAULT '{}',
    metadata JSONB,

    CONSTRAINT version_positive CHECK (version > 0)
);

-- Index for version lookups
CREATE INDEX IF NOT EXISTS idx_document_versions_taxonomy_version
ON document_versions(taxonomy_id, version DESC);

-- Add version tracking to parent_chunks
ALTER TABLE parent_chunks
ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1,
ADD COLUMN IF NOT EXISTS is_current BOOLEAN DEFAULT TRUE,
ADD COLUMN IF NOT EXISTS superseded_version INTEGER,
ADD COLUMN IF NOT EXISTS superseded_by UUID,
ADD COLUMN IF NOT EXISTS version_ingested TIMESTAMPTZ DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();

-- Add version tracking to child_chunks
ALTER TABLE child_chunks
ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1,
ADD COLUMN IF NOT EXISTS is_current BOOLEAN DEFAULT TRUE,
ADD COLUMN IF NOT EXISTS superseded_version INTEGER,
ADD COLUMN IF NOT EXISTS version_ingested TIMESTAMPTZ DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();

-- Indexes for current chunks only (for queries)
CREATE INDEX IF NOT EXISTS idx_parent_chunks_current
ON parent_chunks(taxonomy_id, page_number)
WHERE is_current = TRUE;

CREATE INDEX IF NOT EXISTS idx_child_chunks_current
ON child_chunks(taxonomy_id, page_number)
WHERE is_current = TRUE;

-- Index for version lookups
CREATE INDEX IF NOT EXISTS idx_parent_chunks_version
ON parent_chunks(taxonomy_id, version);

CREATE INDEX IF NOT EXISTS idx_child_chunks_version
ON child_chunks(taxonomy_id, version);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add triggers for parent_chunks
DROP TRIGGER IF EXISTS update_parent_chunks_updated_at ON parent_chunks;
CREATE TRIGGER update_parent_chunks_updated_at
    BEFORE UPDATE ON parent_chunks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add triggers for child_chunks
DROP TRIGGER IF EXISTS update_child_chunks_updated_at ON child_chunks;
CREATE TRIGGER update_child_chunks_updated_at
    BEFORE UPDATE ON child_chunks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function to get current version of a document
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

-- Function to get chunk statistics by version
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

-- Function to cleanup old versions (keep last N versions)
CREATE OR REPLACE FUNCTION rpc_cleanup_old_versions(
    p_taxonomy_id INTEGER,
    keep_versions INTEGER DEFAULT 3
)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    -- Delete old document version records
    DELETE FROM document_versions
    WHERE taxonomy_id = p_taxonomy_id
    AND version NOT IN (
        SELECT version
        FROM document_versions
        WHERE taxonomy_id = p_taxonomy_id
        ORDER BY version DESC
        LIMIT keep_versions
    );

    GET DIAGNOSTICS deleted_count = ROW_COUNT;

    -- Delete old chunks (older than keep_versions)
    DELETE FROM parent_chunks
    WHERE taxonomy_id = p_taxonomy_id
    AND is_current = FALSE
    AND version < (
        SELECT MIN(version)
        FROM (
            SELECT version
            FROM document_versions
            WHERE taxonomy_id = p_taxonomy_id
            ORDER BY version DESC
            LIMIT keep_versions
        ) AS recent_versions
    );

    DELETE FROM child_chunks
    WHERE taxonomy_id = p_taxonomy_id
    AND is_current = FALSE
    AND version < (
        SELECT MIN(version)
        FROM (
            SELECT version
            FROM document_versions
            WHERE taxonomy_id = p_taxonomy_id
            ORDER BY version DESC
            LIMIT keep_versions
        ) AS recent_versions
    );

    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Function to rollback to a specific version
CREATE OR REPLACE FUNCTION rpc_rollback_to_version(
    p_taxonomy_id INTEGER,
    p_target_version INTEGER
)
RETURNS BOOLEAN AS $$
DECLARE
    target_record RECORD;
BEGIN
    -- Get target version
    SELECT * INTO target_record
    FROM document_versions
    WHERE taxonomy_id = p_taxonomy_id
    AND version = p_target_version;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Version % not found for taxonomy_id %', p_target_version, p_taxonomy_id;
    END IF;

    -- Mark current chunks as not current
    UPDATE parent_chunks
    SET is_current = FALSE
    WHERE taxonomy_id = p_taxonomy_id
    AND is_current = TRUE;

    UPDATE child_chunks
    SET is_current = FALSE
    WHERE taxonomy_id = p_taxonomy_id
    AND is_current = TRUE;

    -- Restore target version chunks
    UPDATE parent_chunks
    SET is_current = TRUE
    WHERE taxonomy_id = p_taxonomy_id
    AND version_ingested = target_record.ingested_at;

    UPDATE child_chunks
    SET is_current = TRUE
    WHERE taxonomy_id = p_taxonomy_id
    AND version_ingested = target_record.ingested_at;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- Comments
COMMENT ON TABLE document_versions IS 'Tracks document versions for incremental ingestion';
COMMENT ON COLUMN document_versions.content_hash IS 'SHA-256 hash of entire document';
COMMENT ON COLUMN document_versions.page_hashes IS 'JSONB map of page_number -> SHA-256 hash';
COMMENT ON COLUMN parent_chunks.is_current IS 'True if this is the current version of the chunk';
COMMENT ON COLUMN parent_chunks.superseded_version IS 'Version that superseded this chunk';
COMMENT ON COLUMN child_chunks.is_current IS 'True if this is the current version of the chunk';

-- Grant permissions
GRANT SELECT ON document_versions TO authenticated;
GRANT SELECT ON parent_chunks TO authenticated;
GRANT SELECT ON child_chunks TO authenticated;
