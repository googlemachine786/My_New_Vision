// Package db provides PostgreSQL database connection and query operations
// for the vector search service.
package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"github.com/rs/zerolog/log"

	"github.com/visionary/ragpipeline/services/vector-search-service/config"
)

// DenseSearchResult represents a result from dense vector similarity search.
type DenseSearchResult struct {
	ChildID       uuid.UUID         `json:"child_id"`
	ParentID      uuid.UUID         `json:"parent_id"`
	CosineDist    float64           `json:"cosine_dist"`
	Content       string            `json:"content"`
	ParentContent string            `json:"parent_content"`
	PageNumber    *int              `json:"page_number"`
	ContentType   string            `json:"content_type"`
	Chapter       string            `json:"chapter"`
	Section       string            `json:"section"`
	Subsection    string            `json:"subsection"`
	Grade         string            `json:"grade"`
	Subject       string            `json:"subject"`
}

// SparseSearchResult represents a result from sparse keyword search.
type SparseSearchResult struct {
	ParentID      uuid.UUID `json:"parent_id"`
	Content       string    `json:"content"`
	PageNumber    *int      `json:"page_number"`
	ContentType   string    `json:"content_type"`
	Chapter       string    `json:"chapter"`
	Section       string    `json:"section"`
	Subsection    string    `json:"subsection"`
	Grade         string    `json:"grade"`
	Subject       string    `json:"subject"`
	KeywordScore float64   `json:"keyword_score"`
	MatchCount    int       `json:"match_count"`
}

// QueryStats holds statistics about a search query execution.
type QueryStats struct {
	DenseTimeMS   int64 `json:"dense_time_ms"`
	SparseTimeMS  int64 `json:"sparse_time_ms"`
	BM25TimeMS    int64 `json:"bm25_time_ms"`
	RRFTimeMS     int64 `json:"rrf_time_ms"`
	TotalTimeMS   int64 `json:"total_time_ms"`
	DenseResults  int   `json:"dense_results"`
	SparseResults int   `json:"sparse_results"`
	BM25Results   int   `json:"bm25_results"`
	FusedResults  int   `json:"fused_results"`
}

// MetadataFilter represents metadata filtering criteria.
type MetadataFilter struct {
	Grades      []string `json:"grades"`
	Subjects    []string `json:"subjects"`
	ContentTypes []string `json:"content_types"`
	Chapters    []string `json:"chapters"`
}

// SearchQuery represents a complete search query request.
type SearchQuery struct {
	Embedding []float32       `json:"embedding"`
	Keywords  []string        `json:"keywords"`
	TopK      int             `json:"top_k"`
	Filters   *MetadataFilter `json:"filters"`
}

// Store provides database operations for the vector search service.
type Store struct {
	pool *pgxpool.Pool
	cfg  *config.Config

	// Prepared statements
	denseSearchStmt  string
	sparseSearchStmt string
}

// NewStore creates a new database store with the given configuration.
func NewStore(ctx context.Context, cfg *config.Config) (*Store, error) {
	pool, err := createPool(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	store := &Store{
		pool: pool,
		cfg:  cfg,
	}

	log.Info().
		Int("max_conns", int(cfg.MaxConns)).
		Int("min_conns", int(cfg.MinConns)).
		Msg("Database connection pool created")

	return store, nil
}

// createPool creates a configured connection pool.
func createPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolCfg.HealthCheckPeriod = cfg.HealthCheckPeriod

	// Set connection timeouts
	poolCfg.ConnConfig.ConnectTimeout = 10 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// Ping checks if the database connection is alive.
func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// Close closes the connection pool.
func (s *Store) Close() {
	s.pool.Close()
	log.Info().Msg("Database connection pool closed")
}

// Pool returns the underlying connection pool (for advanced usage).
func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

// Stats returns pool statistics.
func (s *Store) Stats() map[string]interface{} {
	stats := s.pool.Stat()
	return map[string]interface{}{
		"acquire_count":          stats.AcquireCount(),
		"acquired_conns":         stats.AcquiredConns(),
		"canceled_acquire_count": stats.CanceledAcquireCount(),
		"constructing_conns":     stats.ConstructingConns(),
		"empty_acquire_count":    stats.EmptyAcquireCount(),
		"idle_conns":             stats.IdleConns(),
		"max_conns":              stats.MaxConns(),
		"total_conns":            stats.TotalConns(),
	}
}

// DenseSearch performs ScaNN cosine similarity search on child_chunks.
func (s *Store) DenseSearch(ctx context.Context, query *SearchQuery) ([]DenseSearchResult, time.Duration, error) {
	start := time.Now()

	// Apply query timeout
	if s.cfg.QueryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.cfg.QueryTimeout)
		defer cancel()
	}

	limit := s.cfg.DenseSearchLimit(query.TopK)
	querySQL, args := s.buildDenseSearchQuery(query, limit)

	rows, err := s.pool.Query(ctx, querySQL, args...)
	if err != nil {
		return nil, time.Since(start), fmt.Errorf("dense search query failed: %w", err)
	}
	defer rows.Close()

	results := make([]DenseSearchResult, 0, limit)
	for rows.Next() {
		var r DenseSearchResult
		var cosineDist float64

		err := rows.Scan(
			&r.ChildID,
			&r.ParentID,
			&cosineDist,
			&r.Content,
			&r.PageNumber,
			&r.ContentType,
			&r.ParentContent,
			&r.Chapter,
			&r.Section,
			&r.Subsection,
			&r.Grade,
			&r.Subject,
		)
		if err != nil {
			return nil, time.Since(start), fmt.Errorf("failed to scan dense result: %w", err)
		}

		r.CosineDist = cosineDist
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, time.Since(start), fmt.Errorf("error iterating dense results: %w", err)
	}

	log.Debug().
		Int("results", len(results)).
		Int("limit", limit).
		Int64("duration_ms", time.Since(start).Milliseconds()).
		Msg("Dense search completed")

	return results, time.Since(start), nil
}

// buildDenseSearchQuery constructs the SQL query and arguments for dense search.
func (s *Store) buildDenseSearchQuery(query *SearchQuery, limit int) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	// Base query with ScaNN cosine distance operator
	baseQuery := fmt.Sprintf(`
		SELECT
			cc.child_id,
			cc.parent_id,
			cc.embedding <=> $%d AS cosine_dist,
			cc.content,
			cc.page_number,
			cc.content_type,
			pc.content AS parent_content,
			pc.chapter,
			pc.section,
			pc.subsection,
			pc.grade,
			pc.subject
		FROM %s cc
		JOIN %s pc ON cc.parent_id = pc.parent_id`,
		argIdx, s.cfg.DenseChildTable, s.cfg.DenseParentTable)
	args = append(args, pgvector.NewVector(query.Embedding))
	argIdx++

	// Add metadata filters
	if query.Filters != nil {
		if len(query.Filters.Grades) > 0 {
			conditions = append(conditions, fmt.Sprintf("pc.grade = ANY($%d)", argIdx))
			args = append(args, query.Filters.Grades)
			argIdx++
		}
		if len(query.Filters.Subjects) > 0 {
			conditions = append(conditions, fmt.Sprintf("pc.subject = ANY($%d)", argIdx))
			args = append(args, query.Filters.Subjects)
			argIdx++
		}
		if len(query.Filters.ContentTypes) > 0 {
			conditions = append(conditions, fmt.Sprintf("cc.content_type = ANY($%d)", argIdx))
			args = append(args, query.Filters.ContentTypes)
			argIdx++
		}
		if len(query.Filters.Chapters) > 0 {
			conditions = append(conditions, fmt.Sprintf("pc.chapter = ANY($%d)", argIdx))
			args = append(args, query.Filters.Chapters)
			argIdx++
		}
	}

	// Build WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Complete query with ordering and limit
	querySQL := fmt.Sprintf(`%s%s
		ORDER BY cc.embedding <=> $1
		LIMIT $%d`, baseQuery, whereClause, argIdx)
	args = append(args, limit)

	return querySQL, args
}

// SparseSearch performs GIN keyword intersection search on parent_chunks.
func (s *Store) SparseSearch(ctx context.Context, query *SearchQuery) ([]SparseSearchResult, time.Duration, error) {
	if len(query.Keywords) == 0 {
		return nil, 0, nil
	}

	start := time.Now()

	// Apply query timeout
	if s.cfg.QueryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.cfg.QueryTimeout)
		defer cancel()
	}

	limit := s.cfg.SparseSearchLimit(query.TopK)
	querySQL, args := s.buildSparseSearchQuery(query, limit)

	rows, err := s.pool.Query(ctx, querySQL, args...)
	if err != nil {
		return nil, time.Since(start), fmt.Errorf("sparse search query failed: %w", err)
	}
	defer rows.Close()

	results := make([]SparseSearchResult, 0, limit)
	for rows.Next() {
		var r SparseSearchResult
		var keywordScore float64
		var matchCount int

		err := rows.Scan(
			&r.ParentID,
			&r.Content,
			&r.PageNumber,
			&r.ContentType,
			&r.Chapter,
			&r.Section,
			&r.Subsection,
			&r.Grade,
			&r.Subject,
			&keywordScore,
			&matchCount,
		)
		if err != nil {
			return nil, time.Since(start), fmt.Errorf("failed to scan sparse result: %w", err)
		}

		r.KeywordScore = keywordScore
		r.MatchCount = matchCount
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, time.Since(start), fmt.Errorf("error iterating sparse results: %w", err)
	}

	log.Debug().
		Int("results", len(results)).
		Int("limit", limit).
		Int64("duration_ms", time.Since(start).Milliseconds()).
		Msg("Sparse search completed")

	return results, time.Since(start), nil
}

// buildSparseSearchQuery constructs the SQL query and arguments for sparse search.
func (s *Store) buildSparseSearchQuery(query *SearchQuery, limit int) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	// Base query with GIN index intersection
	baseQuery := fmt.Sprintf(`
		SELECT
			pc.parent_id,
			pc.content,
			pc.page_number,
			pc.content_type,
			pc.chapter,
			pc.section,
			pc.subsection,
			pc.grade,
			pc.subject,
			cardinality(
				array(SELECT unnest(pc.%s)
					  INTERSECT
					  SELECT unnest($%d::TEXT[]))
			) AS keyword_score,
			cardinality(
				array(SELECT unnest(pc.%s)
					  INTERSECT
					  SELECT unnest($%d::TEXT[]))
			) AS match_count
		FROM %s pc
		WHERE pc.%s && $%d::TEXT[]`,
		s.cfg.KeywordsColumn, argIdx,
		s.cfg.KeywordsColumn, argIdx,
		s.cfg.SparseTable,
		s.cfg.KeywordsColumn, argIdx)
	args = append(args, query.Keywords)
	argIdx++

	// Add metadata filters
	if query.Filters != nil {
		if len(query.Filters.Grades) > 0 {
			conditions = append(conditions, fmt.Sprintf("pc.grade = ANY($%d)", argIdx))
			args = append(args, query.Filters.Grades)
			argIdx++
		}
		if len(query.Filters.Subjects) > 0 {
			conditions = append(conditions, fmt.Sprintf("pc.subject = ANY($%d)", argIdx))
			args = append(args, query.Filters.Subjects)
			argIdx++
		}
		if len(query.Filters.ContentTypes) > 0 {
			conditions = append(conditions, fmt.Sprintf("pc.content_type = ANY($%d)", argIdx))
			args = append(args, query.Filters.ContentTypes)
			argIdx++
		}
		if len(query.Filters.Chapters) > 0 {
			conditions = append(conditions, fmt.Sprintf("pc.chapter = ANY($%d)", argIdx))
			args = append(args, query.Filters.Chapters)
			argIdx++
		}
	}

	// Complete query with ordering and limit
	querySQL := fmt.Sprintf(`%s
		ORDER BY keyword_score DESC
		LIMIT $%d`, baseQuery, argIdx)
	args = append(args, limit)

	return querySQL, args
}

// ExplainAnalyze runs EXPLAIN ANALYZE on the dense search query for debugging.
func (s *Store) ExplainAnalyze(ctx context.Context, query *SearchQuery) (string, error) {
	limit := s.cfg.DenseSearchLimit(query.TopK)
	querySQL, args := s.buildDenseSearchQuery(query, limit)

	// Prepend EXPLAIN ANALYZE
	explainSQL := "EXPLAIN ANALYZE " + querySQL

	var result strings.Builder
	rows, err := s.pool.Query(ctx, explainSQL, args...)
	if err != nil {
		return "", fmt.Errorf("explain analyze failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			return "", fmt.Errorf("failed to scan explain result: %w", err)
		}
		result.WriteString(line + "\n")
	}

	return result.String(), nil
}

// IsUniqueViolation checks if the error is a unique constraint violation.
func IsUniqueViolation(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		return pgErr.Code == "23505"
	}
	return false
}

// IsConnectionError checks if the error is a connection-related error.
func IsConnectionError(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		// Connection failure class
		return pgErr.Severity == "FATAL" || pgErr.Code[:2] == "08"
	}
	return false
}

// IsTimeoutError checks if the error is a timeout error.
func IsTimeoutError(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		return pgErr.Code == "57014" // query_canceled
	}
	return false
}
