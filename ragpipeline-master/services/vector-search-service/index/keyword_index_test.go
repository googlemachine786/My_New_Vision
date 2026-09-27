package index

import "testing"

func TestNewKeywordIndex(t *testing.T) {
	idx := NewKeywordIndex()
	if idx == nil {
		t.Fatal("expected non-nil index")
	}
	if idx.NumDocs() != 0 {
		t.Errorf("expected 0 docs, got %d", idx.NumDocs())
	}
}

func TestAddDocument(t *testing.T) {
	idx := NewKeywordIndex()
	idx.AddDocument("doc1", "The quick brown fox jumps over the lazy dog")

	if idx.NumDocs() != 1 {
		t.Errorf("expected 1 doc, got %d", idx.NumDocs())
	}

	if idx.DocLength("doc1") == 0 {
		t.Error("expected non-zero doc length")
	}
}

func TestDocFrequency(t *testing.T) {
	idx := NewKeywordIndex()
	idx.AddDocument("doc1", "the quick brown fox")
	idx.AddDocument("doc2", "the lazy dog")
	idx.AddDocument("doc3", "the quick rabbit")

	// "the" appears in all 3 docs
	if df := idx.DocFrequency("the"); df != 3 {
		t.Errorf("expected df(th3) = 3, got %d", df)
	}

	// "quick" appears in 2 docs
	if df := idx.DocFrequency("quick"); df != 2 {
		t.Errorf("expected df(quick) = 2, got %d", df)
	}

	// "fox" appears in 1 doc
	if df := idx.DocFrequency("fox"); df != 1 {
		t.Errorf("expected df(fox) = 1, got %d", df)
	}

	// "nonexistent" appears in 0 docs
	if df := idx.DocFrequency("nonexistent"); df != 0 {
		t.Errorf("expected df(nonexistent) = 0, got %d", df)
	}
}

func TestGetPostings(t *testing.T) {
	idx := NewKeywordIndex()
	idx.AddDocument("doc1", "the quick brown fox")
	idx.AddDocument("doc2", "the quick dog")

	postings := idx.GetPostings("quick")
	if postings == nil {
		t.Fatal("expected non-nil postings for 'quick'")
	}
	if len(postings) != 2 {
		t.Errorf("expected 2 docs in postings, got %d", len(postings))
	}
}

func TestRemoveDocument(t *testing.T) {
	idx := NewKeywordIndex()
	idx.AddDocument("doc1", "the quick brown fox")
	idx.AddDocument("doc2", "the lazy dog")

	idx.RemoveDocument("doc1")

	if idx.NumDocs() != 1 {
		t.Errorf("expected 1 doc after removal, got %d", idx.NumDocs())
	}

	if idx.DocFrequency("fox") != 0 {
		t.Error("expected 'fox' to have df=0 after doc1 removal")
	}
}

func TestAverageDocLength(t *testing.T) {
	idx := NewKeywordIndex()
	idx.AddDocument("doc1", "one two three")
	idx.AddDocument("doc2", "four five six seven eight")

	avg := idx.AverageDocLength()
	expected := float64(3+5) / 2.0
	if avg != expected {
		t.Errorf("expected avg doc length %f, got %f", expected, avg)
	}
}

func TestAverageDocLength_Empty(t *testing.T) {
	idx := NewKeywordIndex()
	avg := idx.AverageDocLength()
	if avg != 1.0 {
		t.Errorf("expected default avg doc length 1.0 for empty index, got %f", avg)
	}
}

func TestContainsTerm(t *testing.T) {
	idx := NewKeywordIndex()
	idx.AddDocument("doc1", "the quick brown fox")

	if !idx.ContainsTerm("doc1", "quick") {
		t.Error("expected doc1 to contain 'quick'")
	}

	if idx.ContainsTerm("doc1", "nonexistent") {
		t.Error("expected doc1 to NOT contain 'nonexistent'")
	}

	if idx.ContainsTerm("doc2", "quick") {
		t.Error("expected doc2 to NOT exist")
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"Hello World", []string{"hello", "world"}},
		{"UPPERCASE lowercase", []string{"uppercase", "lowercase"}},
		{"Hello, World!", []string{"hello", "world"}},
		{"  spaced  out  ", []string{"spaced", "out"}},
		{"", nil},
	}

	for _, tt := range tests {
		result := tokenize(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("tokenize(%q) = %d tokens, want %d", tt.input, len(result), len(tt.expected))
			continue
		}
		for i, token := range result {
			if token != tt.expected[i] {
				t.Errorf("tokenize(%q)[%d] = %q, want %q", tt.input, i, token, tt.expected[i])
			}
		}
	}
}
