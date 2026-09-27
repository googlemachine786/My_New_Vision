package search

import (
	"testing"

	"github.com/google/uuid"

	"github.com/visionary/ragpipeline/services/vector-search-service/db"
)

func ptrInt(n int) *int     { return &n }
func ptrStr(s string) string { return s }

func TestReciprocalRankFuse(t *testing.T) {
	parent1 := uuid.New()
	parent2 := uuid.New()
	parent3 := uuid.New()
	parent4 := uuid.New()

	tests := []struct {
		name          string
		denseResults  []db.DenseSearchResult
		sparseResults []db.SparseSearchResult
		cfg           RRFConfig
		wantLen       int
		wantFirstID   uuid.UUID
	}{
		{
			name: "both dense and sparse with overlap",
			denseResults: []db.DenseSearchResult{
				{ParentID: parent1, ParentContent: "chunk1", CosineDist: 0.1},
				{ParentID: parent2, ParentContent: "chunk2", CosineDist: 0.2},
				{ParentID: parent3, ParentContent: "chunk3", CosineDist: 0.3},
			},
			sparseResults: []db.SparseSearchResult{
				{ParentID: parent2, Content: "chunk2", KeywordScore: 0.9},
				{ParentID: parent3, Content: "chunk3", KeywordScore: 0.8},
				{ParentID: parent4, Content: "chunk4", KeywordScore: 0.7},
			},
			cfg:         DefaultRRFConfig(),
			wantLen:     4,
			wantFirstID: parent2, // parent2 appears in both, should rank highest
		},
		{
			name: "dense only results",
			denseResults: []db.DenseSearchResult{
				{ParentID: parent1, ParentContent: "chunk1", CosineDist: 0.1},
				{ParentID: parent2, ParentContent: "chunk2", CosineDist: 0.2},
			},
			sparseResults: []db.SparseSearchResult{},
			cfg:           DefaultRRFConfig(),
			wantLen:       2,
			wantFirstID:   parent1,
		},
		{
			name:          "sparse only results",
			denseResults:  []db.DenseSearchResult{},
			sparseResults: []db.SparseSearchResult{
				{ParentID: parent1, Content: "chunk1", KeywordScore: 0.9},
				{ParentID: parent2, Content: "chunk2", KeywordScore: 0.8},
			},
			cfg:         DefaultRRFConfig(),
			wantLen:     2,
			wantFirstID: parent1,
		},
		{
			name:          "empty results",
			denseResults:  []db.DenseSearchResult{},
			sparseResults: []db.SparseSearchResult{},
			cfg:           DefaultRRFConfig(),
			wantLen:       0,
		},
		{
			name: "no overlap between dense and sparse",
			denseResults: []db.DenseSearchResult{
				{ParentID: parent1, ParentContent: "chunk1", CosineDist: 0.1},
			},
			sparseResults: []db.SparseSearchResult{
				{ParentID: parent4, Content: "chunk4", KeywordScore: 0.9},
			},
			cfg:         DefaultRRFConfig(),
			wantLen:     2,
			wantFirstID: parent1, // dense rank 1 gives higher RRF than sparse rank 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReciprocalRankFuse(tt.denseResults, tt.sparseResults, nil, tt.cfg)

			if len(got) != tt.wantLen {
				t.Errorf("ReciprocalRankFuse() returned %d entries, want %d", len(got), tt.wantLen)
			}

			if tt.wantLen > 0 && tt.wantFirstID != uuid.Nil {
				if got[0].parentID != tt.wantFirstID {
					t.Errorf("ReciprocalRankFuse() first ID = %v, want %v", got[0].parentID, tt.wantFirstID)
				}
			}

			// Verify sorting by RRF score descending
			for i := 1; i < len(got); i++ {
				if got[i].rrfScore > got[i-1].rrfScore {
					t.Errorf("ReciprocalRankFuse() not sorted descending: got[%d].rrfScore=%f > got[%d].rrfScore=%f",
						i, got[i].rrfScore, i-1, got[i-1].rrfScore)
				}
			}
		})
	}
}

func TestReciprocalRankFuse_Scores(t *testing.T) {
	parent1 := uuid.New()
	parent2 := uuid.New()

	denseResults := []db.DenseSearchResult{
		{ParentID: parent1, ParentContent: "chunk1", CosineDist: 0.1},
		{ParentID: parent2, ParentContent: "chunk2", CosineDist: 0.2},
	}
	sparseResults := []db.SparseSearchResult{
		{ParentID: parent1, Content: "chunk1", KeywordScore: 0.9},
		{ParentID: parent2, Content: "chunk2", KeywordScore: 0.8},
	}

	cfg := RRFConfig{K: 60.0, DenseWeight: 1.0, SparseWeight: 1.0, BM25Weight: 1.0}
	got := ReciprocalRankFuse(denseResults, sparseResults, nil, cfg)

	// Both appear in both lists, verify RRF scores are calculated
	if len(got) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(got))
	}

	// Both should have positive RRF scores
	for _, entry := range got {
		if entry.rrfScore <= 0 {
			t.Errorf("Entry %v has non-positive RRF score: %f", entry.parentID, entry.rrfScore)
		}
		if entry.denseScore <= 0 {
			t.Errorf("Entry %v has non-positive dense score: %f", entry.parentID, entry.denseScore)
		}
		if entry.sparseScore <= 0 {
			t.Errorf("Entry %v has non-positive sparse score: %f", entry.parentID, entry.sparseScore)
		}
	}
}

func TestDeduplicateAndLimit(t *testing.T) {
	parent1 := uuid.New()
	parent2 := uuid.New()
	parent3 := uuid.New()

	entries := []rrfEntry{
		{parentID: parent1, rrfScore: 0.03},
		{parentID: parent1, rrfScore: 0.02}, // duplicate
		{parentID: parent2, rrfScore: 0.025},
		{parentID: parent3, rrfScore: 0.02},
	}

	tests := []struct {
		name    string
		entries []rrfEntry
		topK    int
		wantLen int
	}{
		{
			name:    "limit to top 2",
			entries: entries,
			topK:    2,
			wantLen: 2,
		},
		{
			name:    "limit to top 5 (more than available)",
			entries: entries,
			topK:    5,
			wantLen: 3, // 3 unique parents
		},
		{
			name:    "topK=0 returns empty",
			entries: entries,
			topK:    0,
			wantLen: 0,
		},
		{
			name:    "negative topK returns empty",
			entries: entries,
			topK:    -1,
			wantLen: 0,
		},
		{
			name:    "empty entries",
			entries: []rrfEntry{},
			topK:    5,
			wantLen: 0,
		},
		{
			name:    "all unique entries",
			entries: []rrfEntry{
				{parentID: parent1, rrfScore: 0.03},
				{parentID: parent2, rrfScore: 0.025},
				{parentID: parent3, rrfScore: 0.02},
			},
			topK:    2,
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeduplicateAndLimit(tt.entries, tt.topK)

			if len(got) != tt.wantLen {
				t.Errorf("DeduplicateAndLimit() returned %d entries, want %d", len(got), tt.wantLen)
			}

			// Verify no duplicates
			seen := make(map[uuid.UUID]bool)
			for _, entry := range got {
				if seen[entry.parentID] {
					t.Errorf("DeduplicateAndLimit() returned duplicate entry for %v", entry.parentID)
				}
				seen[entry.parentID] = true
			}
		})
	}
}

func TestToSearchResults(t *testing.T) {
	parent1 := uuid.New()
	parent2 := uuid.New()

	entries := []rrfEntry{
		{
			parentID:   parent1,
			denseRank:  1,
			sparseRank: 2,
			denseScore: 0.0164,
			sparseScore: 0.0161,
			rrfScore:   0.0325,
			denseResult: &db.DenseSearchResult{
				ParentID:    parent1,
				Content:     "dense content 1",
				PageNumber:  ptrInt(5),
				ContentType: "prose",
				Chapter:     "Chapter 1",
				Section:     "Section A",
				Subsection:  "Subsection i",
				Grade:       "7",
				Subject:     "Science",
				CosineDist:  0.1,
			},
			sparseResult: &db.SparseSearchResult{
				ParentID:     parent1,
				Content:      "sparse content 1",
				PageNumber:   ptrInt(5),
				ContentType:  "prose",
				KeywordScore: 0.9,
			},
		},
		{
			parentID:     parent2,
			denseRank:    0,
			sparseRank:   1,
			denseScore:   0,
			sparseScore:  0.0164,
			rrfScore:     0.0164,
			denseResult:  nil,
			sparseResult: &db.SparseSearchResult{
				ParentID:     parent2,
				Content:      "sparse only content",
				PageNumber:   ptrInt(10),
				ContentType:  "table",
				KeywordScore: 0.8,
			},
		},
	}

	got := ToSearchResults(entries)

	if len(got) != 2 {
		t.Fatalf("ToSearchResults() returned %d results, want 2", len(got))
	}

	// Verify first result (has both dense and sparse)
	if got[0].ParentID != parent1.String() {
		t.Errorf("Result[0].ParentID = %v, want %v", got[0].ParentID, parent1.String())
	}
	if got[0].Content != "dense content 1" {
		t.Errorf("Result[0].Content = %q, want %q", got[0].Content, "dense content 1")
	}
	if got[0].RRFScore != 0.0325 {
		t.Errorf("Result[0].RRFScore = %f, want 0.0325", got[0].RRFScore)
	}
	if got[0].DenseRank == nil || *got[0].DenseRank != 1 {
		t.Errorf("Result[0].DenseRank = %v, want 1", got[0].DenseRank)
	}
	if got[0].SparseRank == nil || *got[0].SparseRank != 2 {
		t.Errorf("Result[0].SparseRank = %v, want 2", got[0].SparseRank)
	}
	if val, ok := got[0].Metadata["grade"]; !ok || val != "7" {
		t.Errorf("Result[0].Metadata[grade] = %v, want '7'", val)
	}
	if val, ok := got[0].Metadata["page_number"]; !ok || val != 5 {
		t.Errorf("Result[0].Metadata[page_number] = %v, want 5", val)
	}

	// Verify second result (sparse only)
	if got[1].ParentID != parent2.String() {
		t.Errorf("Result[1].ParentID = %v, want %v", got[1].ParentID, parent2.String())
	}
	if got[1].Content != "sparse only content" {
		t.Errorf("Result[1].Content = %q, want %q", got[1].Content, "sparse only content")
	}
	if got[1].DenseRank != nil {
		t.Errorf("Result[1].DenseRank should be nil for sparse-only entry")
	}
	if val, ok := got[1].Metadata["content_type"]; !ok || val != "table" {
		t.Errorf("Result[1].Metadata[content_type] = %v, want 'table'", val)
	}
}

func TestToSearchResults_Empty(t *testing.T) {
	got := ToSearchResults([]rrfEntry{})
	if len(got) != 0 {
		t.Errorf("ToSearchResults([]) returned %d results, want 0", len(got))
	}
}

func TestDefaultRRFConfig(t *testing.T) {
	cfg := DefaultRRFConfig()

	if cfg.K != 60.0 {
		t.Errorf("DefaultRRFConfig().K = %f, want 60.0", cfg.K)
	}
	if cfg.DenseWeight != 1.0 {
		t.Errorf("DefaultRRFConfig().DenseWeight = %f, want 1.0", cfg.DenseWeight)
	}
	if cfg.SparseWeight != 1.0 {
		t.Errorf("DefaultRRFConfig().SparseWeight = %f, want 1.0", cfg.SparseWeight)
	}
	if cfg.BM25Weight != 1.0 {
		t.Errorf("DefaultRRFConfig().BM25Weight = %f, want 1.0", cfg.BM25Weight)
	}
}

func TestReciprocalRankFuse_ThreeWay(t *testing.T) {
	parent1 := uuid.New()
	parent2 := uuid.New()
	parent3 := uuid.New()

	denseResults := []db.DenseSearchResult{
		{ParentID: parent1, ParentContent: "chunk1", CosineDist: 0.1},
		{ParentID: parent2, ParentContent: "chunk2", CosineDist: 0.2},
	}
	sparseResults := []db.SparseSearchResult{
		{ParentID: parent2, Content: "chunk2", KeywordScore: 0.9},
		{ParentID: parent3, Content: "chunk3", KeywordScore: 0.8},
	}
	bm25Results := []BM25Result{
		{DocID: parent1.String(), Score: 2.5, Rank: 1},
		{DocID: parent3.String(), Score: 1.8, Rank: 2},
	}

	cfg := DefaultRRFConfig()
	got := ReciprocalRankFuse(denseResults, sparseResults, bm25Results, cfg)

	if len(got) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(got))
	}

	// Verify all entries have positive RRF scores
	for _, entry := range got {
		if entry.rrfScore <= 0 {
			t.Errorf("Entry %v has non-positive RRF score: %f", entry.parentID, entry.rrfScore)
		}
	}

	// Verify BM25 ranks are set
	bm25RanksFound := 0
	for _, entry := range got {
		if entry.bm25Rank > 0 {
			bm25RanksFound++
		}
	}
	if bm25RanksFound != 2 {
		t.Errorf("Expected 2 entries with BM25 rank, got %d", bm25RanksFound)
	}
}

func BenchmarkReciprocalRankFuse(b *testing.B) {
	parentIDs := make([]uuid.UUID, 100)
	for i := range parentIDs {
		parentIDs[i] = uuid.New()
	}

	denseResults := make([]db.DenseSearchResult, 100)
	for i := range denseResults {
		denseResults[i] = db.DenseSearchResult{
			ParentID:   parentIDs[i],
			CosineDist: float64(i) / 100.0,
		}
	}

	sparseResults := make([]db.SparseSearchResult, 100)
	for i := range sparseResults {
		sparseResults[i] = db.SparseSearchResult{
			ParentID:     parentIDs[(i+50)%100], // 50% overlap
			KeywordScore: float64(100-i) / 100.0,
		}
	}

	cfg := DefaultRRFConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ReciprocalRankFuse(denseResults, sparseResults, nil, cfg)
	}
}

func BenchmarkDeduplicateAndLimit(b *testing.B) {
	entries := make([]rrfEntry, 1000)
	for i := range entries {
		entries[i] = rrfEntry{
			parentID: uuid.New(),
			rrfScore: float64(1000-i) / 1000.0,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DeduplicateAndLimit(entries, 10)
	}
}

func BenchmarkToSearchResults(b *testing.B) {
	parentIDs := make([]uuid.UUID, 100)
	for i := range parentIDs {
		parentIDs[i] = uuid.New()
	}

	entries := make([]rrfEntry, 100)
	for i := range entries {
		entries[i] = rrfEntry{
			parentID:  parentIDs[i],
			rrfScore:  0.01,
			denseRank: i + 1,
			denseResult: &db.DenseSearchResult{
				ParentID:    parentIDs[i],
				Content:     "content",
				PageNumber:  ptrInt(i + 1),
				ContentType: "prose",
				Chapter:     "Chapter 1",
				Section:     "Section A",
				Grade:       "7",
				Subject:     "Science",
				CosineDist:  0.1,
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ToSearchResults(entries)
	}
}
