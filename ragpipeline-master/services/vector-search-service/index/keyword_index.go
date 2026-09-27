// Package index provides an inverted keyword index for BM25 calculations.
package index

import "strings"

// KeywordIndex maintains an inverted index mapping terms to document frequencies.
type KeywordIndex struct {
	// postings maps term -> docID -> term frequency
	postings map[string]map[string]int
	// docLengths maps docID -> document length (token count)
	docLengths map[string]int
	// docTerms maps docID -> set of unique terms in that document
	docTerms map[string]map[string]struct{}
}

// NewKeywordIndex creates a new empty keyword index.
func NewKeywordIndex() *KeywordIndex {
	return &KeywordIndex{
		postings:   make(map[string]map[string]int),
		docLengths: make(map[string]int),
		docTerms:   make(map[string]map[string]struct{}),
	}
}

// AddDocument indexes a document with the given ID and content.
func (ki *KeywordIndex) AddDocument(docID, content string) {
	tokens := tokenize(content)
	ki.docLengths[docID] = len(tokens)

	termFreq := make(map[string]int)
	termSet := make(map[string]struct{})

	for _, token := range tokens {
		termFreq[token]++
		termSet[token] = struct{}{}
	}

	ki.docTerms[docID] = termSet

	for term, freq := range termFreq {
		if ki.postings[term] == nil {
			ki.postings[term] = make(map[string]int)
		}
		ki.postings[term][docID] = freq
	}
}

// RemoveDocument removes a document from the index.
func (ki *KeywordIndex) RemoveDocument(docID string) {
	// Remove from postings
	for term, docs := range ki.postings {
		if _, exists := docs[docID]; exists {
			delete(docs, docID)
			if len(docs) == 0 {
				delete(ki.postings, term)
			}
		}
	}
	delete(ki.docLengths, docID)
	delete(ki.docTerms, docID)
}

// DocFrequency returns the number of documents containing the given term.
func (ki *KeywordIndex) DocFrequency(term string) int {
	if docs, exists := ki.postings[term]; exists {
		return len(docs)
	}
	return 0
}

// GetPostings returns the term frequency map for a given term.
func (ki *KeywordIndex) GetPostings(term string) map[string]int {
	if docs, exists := ki.postings[term]; exists {
		return docs
	}
	return nil
}

// DocLength returns the token count for a document.
func (ki *KeywordIndex) DocLength(docID string) int {
	if length, exists := ki.docLengths[docID]; exists {
		return length
	}
	return 0
}

// NumDocs returns the total number of indexed documents.
func (ki *KeywordIndex) NumDocs() int {
	return len(ki.docLengths)
}

// AverageDocLength returns the average document length in tokens.
func (ki *KeywordIndex) AverageDocLength() float64 {
	if len(ki.docLengths) == 0 {
		return 1.0
	}

	total := 0
	for _, length := range ki.docLengths {
		total += length
	}

	return float64(total) / float64(len(ki.docLengths))
}

// ContainsTerm checks if a document contains a given term.
func (ki *KeywordIndex) ContainsTerm(docID, term string) bool {
	if terms, exists := ki.docTerms[docID]; exists {
		_, found := terms[term]
		return found
	}
	return false
}

// tokenize splits content into lowercase tokens.
func tokenize(content string) []string {
	fields := strings.Fields(strings.ToLower(content))
	tokens := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.Trim(f, ".,;:!?\"'()[]{}")
		if len(f) > 0 {
			tokens = append(tokens, f)
		}
	}
	return tokens
}
