// Package httperrors provides standardized HTTP error response helpers.
package httperrors

import (
	"encoding/json"
	"net/http"

	"github.com/visionary/ragpipeline/pkg/types"
)

// WriteError writes a structured error response
func WriteError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(types.ErrorResponse{
		Error: types.ErrorDetail{Code: code, Message: message},
	})
}

// WriteValidationError writes a validation error response
func WriteValidationError(w http.ResponseWriter, field, message string) {
	WriteError(w, http.StatusBadRequest, "validation_error", field+": "+message)
}

// WriteInternalError writes an internal server error response
func WriteInternalError(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusInternalServerError, "internal_error", message)
}

// WriteBadRequest writes a bad request error response
func WriteBadRequest(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusBadRequest, "bad_request", message)
}

// WriteUnauthorized writes an unauthorized error response
func WriteUnauthorized(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusUnauthorized, "unauthorized", message)
}

// WriteTooManyRequests writes a rate limit error response
func WriteTooManyRequests(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusTooManyRequests, "rate_limit_exceeded", message)
}

// WriteServiceUnavailable writes a service unavailable error response
func WriteServiceUnavailable(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusServiceUnavailable, "service_unavailable", message)
}
