package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	sessionctx "github.com/visionary/ragpipeline/services/api-gateway/context"
	"github.com/visionary/ragpipeline/services/api-gateway/middleware"
)

// ContextHandler handles session context CRUD endpoints.
type ContextHandler struct {
	registry *sessionctx.SessionRegistry
}

// NewContextHandler creates a new context handler.
func NewContextHandler(registry *sessionctx.SessionRegistry) *ContextHandler {
	return &ContextHandler{registry: registry}
}

// GetContext handles GET /context/{sessionID}
func (h *ContextHandler) GetContext(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionID"]

	ctxObj, err := h.registry.Get(r.Context(), sessionID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to retrieve context",
		})
		return
	}

	if ctxObj == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "context not found",
		})
		return
	}

	writeJSON(w, http.StatusOK, ctxObj)
}

// UpdateContext handles PUT /context/{sessionID}
func (h *ContextHandler) UpdateContext(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionID"]

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON body",
		})
		return
	}

	// Get existing context or create new one
	userID := middleware.GetUserID(r.Context())
	ctxObj, err := h.registry.EnsureContext(r.Context(), sessionID, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to ensure context",
		})
		return
	}

	// Apply updates
	if board, ok := updates["board"].(string); ok {
		ctxObj.UpdateBoard(board)
	}
	if grade, ok := updates["grade"].(string); ok {
		ctxObj.UpdateGrade(grade)
	}
	if language, ok := updates["language"].(string); ok {
		ctxObj.UpdateLanguage(language)
	}
	if chapter, ok := updates["current_chapter"].(string); ok {
		ctxObj.UpdateCurrentChapter(chapter)
	}
	if subject, ok := updates["current_subject"].(string); ok {
		ctxObj.UpdateCurrentSubject(subject)
	}
	if practiceErrors, ok := updates["practice_errors"].([]interface{}); ok {
		ctxObj.PracticeErrors = make([]sessionctx.PracticeError, 0, len(practiceErrors))
		for _, pe := range practiceErrors {
			if m, ok := pe.(map[string]interface{}); ok {
				err := sessionctx.PracticeError{}
				if topic, ok := m["topic"].(string); ok {
					err.Topic = topic
				}
				if msg, ok := m["error_msg"].(string); ok {
					err.ErrorMsg = msg
				}
				if count, ok := m["count"].(float64); ok {
					err.Count = int(count)
				}
				ctxObj.PracticeErrors = append(ctxObj.PracticeErrors, err)
			}
		}
	}
	if scores, ok := updates["practice_scores"].(map[string]interface{}); ok {
		ctxObj.PracticeScores = make(map[string]float64)
		for k, v := range scores {
			if score, ok := v.(float64); ok {
				ctxObj.PracticeScores[k] = score
			}
		}
	}

	ctxObj.IncrementTurns()

	if err := h.registry.Set(r.Context(), ctxObj); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to update context",
		})
		return
	}

	writeJSON(w, http.StatusOK, ctxObj)
}
