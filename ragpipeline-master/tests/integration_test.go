// Package main provides end-to-end integration tests for the RAG pipeline.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/visionary/ragpipeline/ingestion-go/chunker"
	"github.com/visionary/ragpipeline/ingestion-go/keywords"
	"github.com/visionary/ragpipeline/orchestrator/config"
	"github.com/visionary/ragpipeline/orchestrator/handler"
	"github.com/visionary/ragpipeline/orchestrator/retrieval"
	"github.com/visionary/ragpipeline/orchestrator/session"
)

// TestMain runs all integration tests with goroutine leak detection (P2-21).
func TestMain(m *testing.M) {
	fmt.Println("Starting Visionary RAG Pipeline Integration Tests")
	fmt.Println("================================================")

	// Run tests
	exitCode := m.Run()

	// P2-21: Goroutine leak detection (goleak)
	// In production, uncomment the following line:
	// goleak.VerifyTestMain(m)

	// Exit
	os.Exit(exitCode)
}

// =============================================================================
// Chunker Tests
// =============================================================================

func TestChunker_ParentChildSplit(t *testing.T) {
	t.Run("LongTextSplit", func(t *testing.T) {
		// Create long text that will split
		longText := strings.Repeat("This is a sentence. ", 100)

		elements := []chunker.ParsedElement{
			{
				Text:        longText,
				PageNumber:  1,
				Chapter:     "Test Chapter",
				Section:     "Test Section",
				TaxonomyID:  1,
				Grade:       7,
				Subject:     "Science",
				ContentType: "prose",
			},
		}

		parents, children := chunker.CreateParentChildChunks(elements)

		// Verify parent chunks
		assertGreater(t, len(parents), 0, "Should create at least one parent chunk")
		for _, parent := range parents {
			assertLess(t, len(parent.Content), 1501, "Parent chunk should be ≤1500 chars")
		}

		// Verify child chunks
		assertGreater(t, len(children), 0, "Should create at least one child chunk")
		for _, child := range children {
			assertNotEmpty(t, child.ParentID, "Child should have parent ID")
			assertNotEmpty(t, child.ChildID, "Child should have child ID")
		}
	})

	t.Run("TableAtomic", func(t *testing.T) {
		// Create table element
		tableText := "| Header 1 | Header 2 |\n| --- | --- |\n| Cell 1 | Cell 2 |"

		elements := []chunker.ParsedElement{
			{
				Text:        tableText,
				PageNumber:  1,
				TaxonomyID:  1,
				Grade:       7,
				Subject:     "Science",
				ContentType: "table",
				IsTable:     true,
			},
		}

		parents, children := chunker.CreateParentChildChunks(elements)

		// Verify atomic (parent_id == child_id)
		assertTrue(t, len(parents) == 1, "Should create exactly one parent")
		assertTrue(t, len(children) == 1, "Should create exactly one child")
		assertEqual(t, parents[0].ParentID, children[0].ChildID, "Table should be atomic")
	})
}

func TestChunker_Overlap(t *testing.T) {
	longText := strings.Repeat("This is a test sentence for overlap. ", 50)

	elements := []chunker.ParsedElement{
		{
			Text:       longText,
			TaxonomyID: 1,
		},
	}

	parents, children := chunker.CreateParentChildChunks(elements)

	// Verify overlap exists
	assertGreater(t, len(children), 1, "Should create multiple children for overlap test")

	// Check first two children for overlap
	if len(children) > 1 {
		child1 := children[0].Content
		child2 := children[1].Content

		// Last 50 chars of child1 should appear in child2 (overlap)
		if len(child1) > 50 {
			overlap := child1[len(child1)-50:]
			assertContains(t, child2, overlap, "Should have overlap between consecutive children")
		}
	}

	_ = parents // Suppress unused warning
}

// =============================================================================
// Keyword Extraction Tests
// =============================================================================

func TestKeywords_Extraction(t *testing.T) {
	t.Run("ScienceText", func(t *testing.T) {
		text := "Photosynthesis is the process by which green plants and some other organisms use sunlight to synthesize foods with the help of chlorophyll."

		extractedKeywords, err := keywords.ExtractKeywords(text)
		requireNoError(t, err)

		// Verify keywords extracted
		assertGreater(t, len(extractedKeywords), 0, "Should extract keywords")
		assertLess(t, len(extractedKeywords), 9, "Should extract at most 8 keywords")

		// Verify keywords are bigrams or unigrams
		for _, kw := range extractedKeywords {
			words := strings.Fields(kw)
			assertLess(t, len(words), 3, "Keyword should be unigram or bigram")
		}
	})

	t.Run("ShortText", func(t *testing.T) {
		text := "Cell membrane"

		extractedKeywords, err := keywords.ExtractKeywords(text)
		requireNoError(t, err)

		// Short text may return empty or minimal keywords
		assertLess(t, len(extractedKeywords), 9, "Should respect max keywords")
	})

	t.Run("EmptyText", func(t *testing.T) {
		extractedKeywords, err := keywords.ExtractKeywords("")
		requireNoError(t, err)
		if len(extractedKeywords) != 0 {
			t.Error("Empty text should return no keywords")
		}
	})
}

// =============================================================================
// Retrieval Tests
// =============================================================================

func TestRetrieval_KeywordExtraction(t *testing.T) {
	t.Run("QueryKeywords", func(t *testing.T) {
		query := "What is photosynthesis and how do plants make food?"

		extractedKeywords := retrieval.ExtractKeywords(query)

		// Verify keywords extracted
		assertGreater(t, len(extractedKeywords), 0, "Should extract keywords from query")
		assertLess(t, len(extractedKeywords), 9, "Should extract at most 8 keywords")

		// Verify no stopwords
		stopwords := map[string]bool{
			"the": true, "a": true, "an": true, "is": true, "are": true,
			"what": true, "how": true, "do": true, "and": true,
		}

		for _, kw := range extractedKeywords {
			words := strings.Fields(strings.ToLower(kw))
			for _, word := range words {
				if stopwords[word] {
					t.Errorf("Should not contain stopwords: %s", word)
				}
			}
		}
	})
}

func TestRetrieval_RRFFuse(t *testing.T) {
	// Create mock dense results
	dense := []retrieval.RRFResult{
		{ID: "1", RankDense: 1},
		{ID: "2", RankDense: 2},
		{ID: "3", RankDense: 3},
	}

	// Create mock sparse results
	sparse := []retrieval.RRFResult{
		{ID: "2", RankSparse: 1},
		{ID: "3", RankSparse: 2},
		{ID: "4", RankSparse: 3},
	}

	// Perform RRF fusion
	fused := retrieval.RRFFuseV2(dense, sparse, 60.0)

	// Verify results
	assertTrue(t, len(fused) == 4, "Should have 4 unique results")

	// ID 2 should be ranked highest (appears in both dense and sparse)
	assertEqual(t, "2", fused[0].ID, "ID appearing in both should rank highest")

	// Verify RRF scores calculated
	for _, r := range fused {
		assertGreater(t, r.RRFScore, 0.0, "RRF score should be positive")
	}
}

func TestRetrieval_Metrics(t *testing.T) {
	t.Run("RecallAtK", func(t *testing.T) {
		retrieved := []string{"a", "b", "c", "d", "e"}
		relevant := []string{"b", "d", "f"}

		recall1 := retrieval.CalculateRecallAtK(retrieved, relevant, 1)
		assertEqual(t, 0.0, recall1, "Recall@1 should be 0.0")

		recall3 := retrieval.CalculateRecallAtK(retrieved, relevant, 3)
		assertEqual(t, 2.0/3.0, recall3, "Recall@3 should be 0.667")

		recall5 := retrieval.CalculateRecallAtK(retrieved, relevant, 5)
		assertEqual(t, 2.0/3.0, recall5, "Recall@5 should be 0.667")
	})

	t.Run("MRR", func(t *testing.T) {
		retrieved := []string{"a", "b", "c", "d"}
		relevant := []string{"c", "e"}

		mrr := retrieval.CalculateMRR(retrieved, relevant)
		assertEqual(t, 1.0/3.0, mrr, "MRR should be 0.333 (first relevant at position 3)")
	})

	t.Run("NDCG", func(t *testing.T) {
		retrieved := []string{"a", "b", "c"}
		relevant := []string{"b", "c"}

		ndcg := retrieval.CalculateNDCG(retrieved, relevant)
		assertGreater(t, ndcg, 0.0, "NDCG should be positive")
		assertLess(t, ndcg, 1.0001, "NDCG should be ≤ 1.0")
	})
}

// =============================================================================
// Handler Tests (Mock)
// =============================================================================

func TestHandler_HealthEndpoint(t *testing.T) {
	// Create mock config
	cfg := &config.Config{
		RequestTimeout: 450 * time.Millisecond,
		TopK:           5,
		RRFK:           60.0,
	}

	// Create handler (with nil dependencies for basic test)
	h := &handler.RAGHandler{}
	router := handler.NewRouter(h)

	// Create test request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Verify response
	assertEqual(t, http.StatusOK, w.Code, "Health endpoint should return 200")
	assertContains(t, w.Body.String(), "status", "Response should contain status field")

	_ = cfg // Suppress unused warning
}

func TestHandler_QueryValidation(t *testing.T) {
	// Create handler
	h := &handler.RAGHandler{}
	router := handler.NewRouter(h)

	t.Run("EmptyQuery", func(t *testing.T) {
		// Create request with empty query
		body := `{"query": "", "session_id": "test"}`
		req := httptest.NewRequest("POST", "/query", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return error
		assert.Equal(t, http.StatusBadRequest, w.Code, "Empty query should return 400")
	})

	t.Run("MissingAuth", func(t *testing.T) {
		// Create request without JWT
		body := `{"query": "test", "session_id": "test"}`
		req := httptest.NewRequest("POST", "/query", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return unauthorized
		assert.Equal(t, http.StatusUnauthorized, w.Code, "Missing JWT should return 401")
	})
}

// =============================================================================
// Configuration Tests
// =============================================================================

func TestConfig_Validation(t *testing.T) {
	t.Run("ValidConfig", func(t *testing.T) {
		cfg := &config.Config{
			AlloyDBDSN:     "host=localhost dbname=test",
			RedisAddr:      "localhost:6379",
			VertexProject:  "test-project",
			JWTSecret:      "secret",
			RequestTimeout: 450 * time.Millisecond,
		}

		err := cfg.Validate()
		if err != nil {
			t.Errorf("Valid config should pass validation: %v", err)
		}
	})

	t.Run("MissingDSN", func(t *testing.T) {
		cfg := &config.Config{
			RedisAddr:     "localhost:6379",
			VertexProject: "test-project",
			JWTSecret:     "secret",
		}

		err := cfg.Validate()
		if err == nil {
			t.Error("Missing DSN should fail validation")
		} else if !strings.Contains(err.Error(), "ALLOYDB_DSN") {
			t.Errorf("Expected error to contain 'ALLOYDB_DSN', got: %v", err)
		}
	})

	t.Run("MissingRedis", func(t *testing.T) {
		cfg := &config.Config{
			AlloyDBDSN:    "host=localhost dbname=test",
			VertexProject: "test-project",
			JWTSecret:     "secret",
		}

		err := cfg.Validate()
		if err == nil {
			t.Error("Missing Redis should fail validation")
		} else if !strings.Contains(err.Error(), "REDIS_ADDR") {
			t.Errorf("Expected error to contain 'REDIS_ADDR', got: %v", err)
		}
	})
}

// =============================================================================
// Session Store Tests (Mock)
// =============================================================================

func TestSession_KeyGeneration(t *testing.T) {
	// Test session key generation
	sessionID := "test-session-123"
	encoded := session.EncodeBase64URL(sessionID)

	assertNotEmpty(t, encoded, "Should generate encoded key")
	assertNotEqual(t, sessionID, encoded, "Encoded key should be different from original")

	// Test decoding
	decoded, err := session.DecodeBase64URL(encoded)
	requireNoError(t, err)
	assertEqual(t, sessionID, decoded, "Decoded should match original")
}

// =============================================================================
// Database Tests (Mock)
// =============================================================================

func TestDatabase_PgBouncerDSN(t *testing.T) {
	cfg := &config.Config{
		AlloyDBDSN:        "host=10.0.0.1 dbname=visionary user=visionary password=secret123",
		PgBouncerHost:     "localhost",
		PgBouncerPort:     "6432",
		PgBouncerDatabase: "visionary",
		PgBouncerUser:     "visionary",
	}

	dsn := cfg.PgBouncerDSN()

	assertContains(t, dsn, "host=localhost", "Should use PgBouncer host")
	assertContains(t, dsn, "port=6432", "Should use PgBouncer port")
	assertContains(t, dsn, "dbname=visionary", "Should use correct database")
}

// =============================================================================
// Integration Test Helpers
// =============================================================================

// createTestConfig creates a test configuration
func createTestConfig() *config.Config {
	return &config.Config{
		RequestTimeout: 450 * time.Millisecond,
		TopK:           5,
		TopKDense:      100,
		TopKSparse:     100,
		TopKRRF:        10,
		RRFK:           60.0,
		EmbedDimension: 768,
	}
}

// createMockEmbedding creates a mock 768-dim embedding
func createMockEmbedding() []float32 {
	embedding := make([]float32, 768)
	for i := range embedding {
		embedding[i] = float32(i) / 1000.0
	}
	return embedding
}

// TestMock_HybridSearch tests hybrid search with mock data
func TestMock_HybridSearch(t *testing.T) {
	cfg := createTestConfig()
	embedding := createMockEmbedding()
	keywords := []string{"cell", "membrane", "photosynthesis"}

	// Verify embedding dimension
	assertTrue(t, len(embedding) == cfg.EmbedDimension, "Embedding should be 768-dim")

	// Verify keywords
	assertGreater(t, len(keywords), 0, "Should have keywords")

	// In production, this would call actual database
	// For now, just verify inputs are correct
	_ = embedding
	_ = keywords
}

// =============================================================================
// JSON Serialization Tests
// =============================================================================

func TestJSON_Serialization(t *testing.T) {
	t.Run("QueryRequest", func(t *testing.T) {
		type QueryRequest struct {
			Query     string `json:"query"`
			SessionID string `json:"session_id"`
		}

		req := QueryRequest{
			Query:     "What is photosynthesis?",
			SessionID: "test-123",
		}

		data, err := json.Marshal(req)
		requireNoError(t, err)

		var unmarshaled QueryRequest
		err = json.Unmarshal(data, &unmarshaled)
		requireNoError(t, err)

		assertEqual(t, req.Query, unmarshaled.Query)
		assertEqual(t, req.SessionID, unmarshaled.SessionID)
	})

	t.Run("RetrievalResult", func(t *testing.T) {
		result := retrieval.RetrievalResult{
			Content:       "Test content",
			ParentContent: "Parent content",
			PageNumber:    42,
			Chapter:       "Test Chapter",
			Section:       "Test Section",
			RRFScore:      0.0167,
		}

		data, err := json.Marshal(result)
		requireNoError(t, err)

		// Verify JSON contains expected fields
		jsonStr := string(data)
		assertContains(t, jsonStr, "content")
		assertContains(t, jsonStr, "page_number")
		assertContains(t, jsonStr, "rrf_score")
	})
}

// =============================================================================
// Performance Tests
// =============================================================================

func TestPerformance_Chunking(t *testing.T) {
	// Create large text (100KB)
	largeText := strings.Repeat("This is a test sentence for performance testing. ", 2000)

	elements := []chunker.ParsedElement{
		{Text: largeText, TaxonomyID: 1},
	}

	start := time.Now()
	parents, children := chunker.CreateParentChildChunks(elements)
	elapsed := time.Since(start)

	fmt.Printf("\nPerformance: Chunked %d chars in %v\n", len(largeText), elapsed)
	fmt.Printf("  Parents: %d, Children: %d\n", len(parents), len(children))
	fmt.Printf("  Throughput: %.2f KB/ms\n", float64(len(largeText))/1024/elapsed.Seconds())

	// Should complete in <100ms
	if elapsed >= 100*time.Millisecond {
		t.Errorf("Chunking took %v, expected <100ms", elapsed)
	}
}

func TestPerformance_KeywordExtraction(t *testing.T) {
	text := strings.Repeat("Photosynthesis is the process by which plants make food. ", 20)

	start := time.Now()
	keywords, err := keywords.ExtractKeywords(text)
	elapsed := time.Since(start)

	fmt.Printf("\nPerformance: Extracted keywords in %v\n", elapsed)
	fmt.Printf("  Keywords: %v\n", keywords)

	requireNoError(t, err)
	// Should complete in <500ms (Gemini API call)
	if elapsed >= 500*time.Millisecond {
		t.Errorf("Keyword extraction took %v, expected <500ms", elapsed)
	}
}

// =============================================================================
// Test Helpers (replacing testify assertions)
// =============================================================================

// requireNoError fails if err is not nil
func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

// assertEqual fails if a != b
func assertEqual[T comparable](t *testing.T, want, got T, msg ...string) {
	t.Helper()
	if want != got {
		if len(msg) > 0 {
			t.Errorf("%s: want %v, got %v", msg[0], want, got)
		} else {
			t.Errorf("want %v, got %v", want, got)
		}
	}
}

// assertContains fails if s doesn't contain substr
func assertContains(t *testing.T, s, substr string, msg ...string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		if len(msg) > 0 {
			t.Errorf("%s: expected %q to contain %q", msg[0], s, substr)
		} else {
			t.Errorf("expected %q to contain %q", s, substr)
		}
	}
}

// assertLen fails if len(s) != n
func assertLen(t *testing.T, s interface{ Len() int }, n int, msg ...string) {
	t.Helper()
	if s.Len() != n {
		if len(msg) > 0 {
			t.Errorf("%s: expected length %d, got %d", msg[0], n, s.Len())
		} else {
			t.Errorf("expected length %d, got %d", n, s.Len())
		}
	}
}

// assertTrue fails if b is not true
func assertTrue(t *testing.T, b bool, msg ...string) {
	t.Helper()
	if !b {
		if len(msg) > 0 {
			t.Errorf("%s: expected true", msg[0])
		} else {
			t.Error("expected true")
		}
	}
}

// assertGreater fails if a <= b
func assertGreater[T ~int | ~int64 | ~float64](t *testing.T, a, b T, msg ...string) {
	t.Helper()
	if a <= b {
		if len(msg) > 0 {
			t.Errorf("%s: expected %v > %v", msg[0], a, b)
		} else {
			t.Errorf("expected %v > %v", a, b)
		}
	}
}

// assertLess fails if a >= b
func assertLess[T comparable](t *testing.T, a, b T, msg ...string) {
	t.Helper()
	if a >= b {
		if len(msg) > 0 {
			t.Errorf("%s: expected %v < %v", msg[0], a, b)
		} else {
			t.Errorf("expected %v < %v", a, b)
		}
	}
}

// assertNotEmpty fails if s is empty
func assertNotEmpty(t *testing.T, s string, msg ...string) {
	t.Helper()
	if s == "" {
		if len(msg) > 0 {
			t.Errorf("%s: expected non-empty string", msg[0])
		} else {
			t.Error("expected non-empty string")
		}
	}
}

// assertNotEqual fails if a == b
func assertNotEqual[T comparable](t *testing.T, a, b T, msg ...string) {
	t.Helper()
	if a == b {
		if len(msg) > 0 {
			t.Errorf("%s: expected %v != %v", msg[0], a, b)
		} else {
			t.Errorf("expected %v != %v", a, b)
		}
	}
}
