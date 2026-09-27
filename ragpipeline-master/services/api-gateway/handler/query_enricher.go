package handler

import (
	"context"
	"fmt"
	"strings"

	ctxpkg "github.com/visionary/ragpipeline/services/api-gateway/context"
)

// EnrichedQuery holds a query with all enrichment data attached.
type EnrichedQuery struct {
	OriginalQuery    string
	EnrichedPrompt   string
	UserID           string
	Grade            string
	Board            string
	Language         string
	CurrentChapter   string
	CurrentSubject   string
	PracticeErrors   []ctxpkg.PracticeError
	SessionTurns     int
}

// QueryEnricher enrichs queries with user profile and session context.
type QueryEnricher struct {
	registry *ctxpkg.SessionRegistry
}

// NewQueryEnricher creates a new query enricher.
func NewQueryEnricher(registry *ctxpkg.SessionRegistry) *QueryEnricher {
	return &QueryEnricher{registry: registry}
}

// Enrich looks up session context and builds an enriched prompt for the LLM.
// If no session context is found, it returns the original query with minimal enrichment.
func (e *QueryEnricher) Enrich(ctx context.Context, sessionID, userID, query, grade string) (*EnrichedQuery, error) {
	enriched := &EnrichedQuery{
		OriginalQuery: query,
		UserID:        userID,
		Grade:         grade,
	}

	// Look up session context if sessionID is provided
	if sessionID != "" && e.registry != nil {
		ctxObj, err := e.registry.Get(ctx, sessionID)
		if err != nil {
			// Non-blocking: log and continue with available data
			// The query will still work, just without session enrichment
		} else if ctxObj != nil {
			enriched.Board = ctxObj.Board
			enriched.Grade = coalesceString(grade, ctxObj.Grade)
			enriched.Language = ctxObj.Language
			enriched.CurrentChapter = ctxObj.CurrentChapter
			enriched.CurrentSubject = ctxObj.CurrentSubject
			enriched.PracticeErrors = ctxObj.PracticeErrors
			enriched.SessionTurns = ctxObj.SessionTurns
		}
	}

	// Build the enriched prompt prefix
	enriched.EnrichedPrompt = BuildPromptPrefix(enriched)

	return enriched, nil
}

// coalesceString returns the first non-empty string.
func coalesceString(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// BuildPromptPrefix builds a context-aware prompt prefix for the LLM.
func BuildPromptPrefix(eq *EnrichedQuery) string {
	var sb strings.Builder

	// Build the system context prefix
	contextParts := make([]string, 0, 5)
	if eq.Board != "" {
		contextParts = append(contextParts, eq.Board+" curriculum")
	}
	if eq.Grade != "" {
		contextParts = append(contextParts, "Grade "+eq.Grade)
	}
	if eq.Language != "" {
		contextParts = append(contextParts, "in "+eq.Language)
	}

	if len(contextParts) > 0 {
		sb.WriteString(fmt.Sprintf("You are teaching a %s student studying %s.\n",
			eq.Grade,
			strings.Join(contextParts, " "),
		))
	}

	if eq.CurrentSubject != "" {
		sb.WriteString(fmt.Sprintf("Current subject: %s.\n", eq.CurrentSubject))
	}
	if eq.CurrentChapter != "" {
		sb.WriteString(fmt.Sprintf("Current chapter: %s.\n", eq.CurrentChapter))
	}

	// Add practice errors as context
	if len(eq.PracticeErrors) > 0 {
		sb.WriteString("\nRecent mistakes:\n")
		for _, pe := range eq.PracticeErrors {
			if pe.Count > 1 {
				sb.WriteString(fmt.Sprintf("- %s: %s (occurred %d times)\n", pe.Topic, pe.ErrorMsg, pe.Count))
			} else {
				sb.WriteString(fmt.Sprintf("- %s: %s\n", pe.Topic, pe.ErrorMsg))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("\nAnswer the following query: %s\n", eq.OriginalQuery))

	return sb.String()
}
