package response

import (
	"encoding/json"
	"net/http"
)

// APIResponse is the standard API response envelope.
type APIResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Data    any               `json:"data,omitempty"`
	Errors  map[string]string `json:"errors,omitempty"`
	Meta    *Meta             `json:"meta,omitempty"`
}

// Meta holds pagination metadata.
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// JSON writes a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, resp APIResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, `{"success":false,"message":"internal server error"}`, http.StatusInternalServerError)
	}
}

// Success sends a successful response.
func Success(w http.ResponseWriter, status int, message string, data any) {
	JSON(w, status, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessWithMeta sends a successful response with pagination metadata.
func SuccessWithMeta(w http.ResponseWriter, status int, message string, data any, meta Meta) {
	JSON(w, status, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    &meta,
	})
}

// Created sends a 201 Created response.
func Created(w http.ResponseWriter, message string, data any) {
	Success(w, http.StatusCreated, message, data)
}

// OK sends a 200 OK response.
func OK(w http.ResponseWriter, message string, data any) {
	Success(w, http.StatusOK, message, data)
}

// NoContent sends a 204 No Content response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error sends an error response.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, APIResponse{
		Success: false,
		Message: message,
	})
}

// ValidationError sends a 422 Unprocessable Entity response.
func ValidationError(w http.ResponseWriter, message string, errors map[string]string) {
	JSON(w, http.StatusUnprocessableEntity, APIResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

// BadRequest sends a 400 Bad Request response.
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, message)
}

// Unauthorized sends a 401 Unauthorized response.
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, message)
}

// Forbidden sends a 403 Forbidden response.
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, message)
}

// NotFound sends a 404 Not Found response.
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, message)
}

// Conflict sends a 409 Conflict response.
func Conflict(w http.ResponseWriter, message string) {
	Error(w, http.StatusConflict, message)
}

// InternalError sends a 500 Internal Server Error response.
// The message should not leak sensitive information.
func InternalError(w http.ResponseWriter) {
	Error(w, http.StatusInternalServerError, "Terjadi kesalahan pada server")
}
