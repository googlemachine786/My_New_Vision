package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/visionary/ragpipeline/services/api-gateway/model"
	"github.com/visionary/ragpipeline/pkg/types"

	"github.com/google/uuid"
)

const (
	// ConfidenceThreshold is the minimum confidence score to answer directly
	ConfidenceThreshold = 0.7
)

// ClarificationHandler handles clarification question generation and processing
type ClarificationHandler struct{}

// NewClarificationHandler creates a new clarification handler
func NewClarificationHandler() *ClarificationHandler {
	return &ClarificationHandler{}
}

// ServeHTTP handles POST /clarify requests
func (h *ClarificationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "method_not_allowed", Message: "Only POST is allowed"},
		})
		return
	}

	var req model.ClarificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "invalid_request", Message: "Invalid JSON body"},
		})
		return
	}

	if err := req.Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "validation_error", Message: err.Error()},
		})
		return
	}

	// Generate clarification options
	options := h.generateClarificationOptions(req)

	resp := model.ClarificationResponse{
		NeedsClarification: true,
		Options:            options,
		Message:            h.generateClarificationMessage(req, options),
		OriginalQuery:      req.Query,
		Intent:             req.Intent,
		Confidence:         req.Confidence,
	}

	writeJSON(w, http.StatusOK, resp)
}

// HandleSelection handles POST /clarify/select when user selects an option
func (h *ClarificationHandler) HandleSelection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "method_not_allowed", Message: "Only POST is allowed"},
		})
		return
	}

	var selection model.ClarificationSelection
	if err := json.NewDecoder(r.Body).Decode(&selection); err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "invalid_request", Message: "Invalid JSON body"},
		})
		return
	}

	if err := selection.Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "validation_error", Message: err.Error()},
		})
		return
	}

	// Find the selected option and build enhanced query
	option := h.findOptionByID(selection.OptionID)
	if option == nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "invalid_option", Message: "Selected option not found"},
		})
		return
	}

	enhancedQuery := h.buildEnhancedQuery(selection.OriginalQuery, option)

	result := model.ClarificationResult{
		EnhancedQuery:    enhancedQuery,
		SelectedOption:   *option,
		OriginalQuery:    selection.OriginalQuery,
	}

	writeJSON(w, http.StatusOK, result)
}

// NeedsClarification checks if the confidence score requires clarification
func NeedsClarification(confidence float64) bool {
	return confidence < ConfidenceThreshold
}

// generateClarificationOptions creates 2-3 clarification options based on the request
func (h *ClarificationHandler) generateClarificationOptions(req model.ClarificationRequest) []model.ClarificationOption {
	var options []model.ClarificationOption

	// Generate topic clarification options if subject is unclear
	if req.Subject == "" {
		options = append(options, h.generateTopicOptions(req)...)
	}

	// Generate grade clarification options if grade is unclear
	if req.Grade == "" {
		options = append(options, h.generateGradeOptions(req)...)
	}

	// Generate intent clarification options
	options = append(options, h.generateIntentOptions(req)...)

	// Limit to 3 options maximum
	if len(options) > 3 {
		options = options[:3]
	}

	// If no options generated, provide generic options
	if len(options) == 0 {
		options = []model.ClarificationOption{
			{
				ID:          uuid.New().String(),
				Label:       "General explanation",
				Description: "Provide a general explanation of the topic",
				Type:        model.ClarificationTypeIntent,
			},
			{
				ID:          uuid.New().String(),
				Label:       "Specific example",
				Description: "Provide a specific example to illustrate the concept",
				Type:        model.ClarificationTypeIntent,
			},
		}
	}

	return options
}

// generateTopicOptions creates topic clarification options
func (h *ClarificationHandler) generateTopicOptions(req model.ClarificationRequest) []model.ClarificationOption {
	subjects := h.inferPossibleSubjects(req.Query)
	var options []model.ClarificationOption

	for _, subject := range subjects {
		options = append(options, model.ClarificationOption{
			ID:          uuid.New().String(),
			Label:       fmt.Sprintf("About %s", subject),
			Description: fmt.Sprintf("Are you asking about %s?", subject),
			Type:        model.ClarificationTypeTopic,
		})
	}

	return options
}

// generateGradeOptions creates grade clarification options
func (h *ClarificationHandler) generateGradeOptions(req model.ClarificationRequest) []model.ClarificationOption {
	grade := h.inferPossibleGrade(req.Query)
	if grade != "" {
		return []model.ClarificationOption{
			{
				ID:          uuid.New().String(),
				Label:       fmt.Sprintf("Grade %s level", grade),
				Description: fmt.Sprintf("Is this for Grade %s?", grade),
				Type:        model.ClarificationTypeGrade,
			},
		}
	}

	// Default grade options for common grades
	grades := []string{"6", "7", "8"}
	var options []model.ClarificationOption
	for _, g := range grades {
		options = append(options, model.ClarificationOption{
			ID:          uuid.New().String(),
			Label:       fmt.Sprintf("Grade %s", g),
			Description: fmt.Sprintf("Are you in Grade %s?", g),
			Type:        model.ClarificationTypeGrade,
		})
	}

	return options
}

// generateIntentOptions creates intent clarification options
func (h *ClarificationHandler) generateIntentOptions(req model.ClarificationRequest) []model.ClarificationOption {
	return []model.ClarificationOption{
		{
			ID:          uuid.New().String(),
			Label:       "Theory explanation",
			Description: "Explain the underlying theory and concepts",
			Type:        model.ClarificationTypeIntent,
		},
		{
			ID:          uuid.New().String(),
			Label:       "Practical application",
			Description: "Show how this is applied in practice",
			Type:        model.ClarificationTypeIntent,
		},
	}
}

// generateClarificationMessage creates a human-readable clarification message
func (h *ClarificationHandler) generateClarificationMessage(req model.ClarificationRequest, options []model.ClarificationOption) string {
	if len(options) == 0 {
		return "Could you please clarify your question?"
	}

	var labels []string
	for _, opt := range options {
		labels = append(labels, opt.Label)
	}

	return fmt.Sprintf("Did you mean one of these: %s?", strings.Join(labels, ", "))
}

// findOptionByID finds an option by its ID from the predefined set
func (h *ClarificationHandler) findOptionByID(optionID string) *model.ClarificationOption {
	// Search through all predefined option types
	allOptions := h.getAllPossibleOptions()
	for _, opt := range allOptions {
		if opt.ID == optionID {
			return &opt
		}
	}
	return nil
}

// getAllPossibleOptions returns all possible clarification options
func (h *ClarificationHandler) getAllPossibleOptions() []model.ClarificationOption {
	return []model.ClarificationOption{
		// Topic options
		{Label: "About Mathematics", Type: model.ClarificationTypeTopic},
		{Label: "About Science", Type: model.ClarificationTypeTopic},
		{Label: "About History", Type: model.ClarificationTypeTopic},
		{Label: "About Language", Type: model.ClarificationTypeTopic},
		// Grade options
		{Label: "Grade 6", Type: model.ClarificationTypeGrade},
		{Label: "Grade 7", Type: model.ClarificationTypeGrade},
		{Label: "Grade 8", Type: model.ClarificationTypeGrade},
		// Intent options
		{Label: "Theory explanation", Type: model.ClarificationTypeIntent},
		{Label: "Practical application", Type: model.ClarificationTypeIntent},
	}
}

// buildEnhancedQuery builds a query enhanced with the user's clarification
func (h *ClarificationHandler) buildEnhancedQuery(originalQuery string, option *model.ClarificationOption) string {
	switch option.Type {
	case model.ClarificationTypeTopic:
		// Extract subject from label and append to query
		subject := strings.TrimPrefix(option.Label, "About ")
		return fmt.Sprintf("%s in the context of %s", originalQuery, subject)
	case model.ClarificationTypeGrade:
		// Extract grade and append to query
		grade := strings.TrimPrefix(option.Label, "Grade ")
		return fmt.Sprintf("%s for grade %s level", originalQuery, grade)
	case model.ClarificationTypeIntent:
		// Append intent clarification to query
		return fmt.Sprintf("%s - focusing on %s", originalQuery, strings.ToLower(option.Label))
	default:
		return originalQuery
	}
}

// inferPossibleSubjects attempts to infer possible subjects from the query
func (h *ClarificationHandler) inferPossibleSubjects(query string) []string {
	queryLower := strings.ToLower(query)
	subjectKeywords := map[string][]string{
		"algebra":     {"Mathematics", "Algebra"},
		"geometry":    {"Mathematics", "Geometry"},
		"calculus":    {"Mathematics", "Calculus"},
		"physics":     {"Science", "Physics"},
		"chemistry":   {"Science", "Chemistry"},
		"biology":     {"Science", "Biology"},
		"history":     {"History"},
		"grammar":     {"Language Arts", "Grammar"},
		"literature":  {"Language Arts", "Literature"},
	}

	var subjects []string
	for keyword, subs := range subjectKeywords {
		if strings.Contains(queryLower, keyword) {
			subjects = append(subjects, subs...)
			break
		}
	}

	if len(subjects) == 0 {
		// Default: offer broad subject options
		subjects = []string{"Mathematics", "Science", "Language Arts"}
	}

	return subjects
}

// inferPossibleGrade attempts to infer a grade from the query
func (h *ClarificationHandler) inferPossibleGrade(query string) string {
	gradeKeywords := map[string]string{
		"grade 6":  "6",
		"grade 7":  "7",
		"grade 8":  "8",
		"6th":      "6",
		"7th":      "7",
		"8th":      "8",
	}

	queryLower := strings.ToLower(query)
	for keyword, grade := range gradeKeywords {
		if strings.Contains(queryLower, keyword) {
			return grade
		}
	}

	return ""
}
