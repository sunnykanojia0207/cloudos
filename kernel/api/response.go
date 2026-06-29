// Package api provides the CloudOS Control Plane — the HTTP API that exposes
// kernel status, health, version, and system information to dashboards, CLIs,
// and external integrations.
package api

import (
	"encoding/json"
	"net/http"
)

// Response is the standard JSON envelope for all API responses.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo carries a machine-readable code and human-readable message.
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// writeJSON serialises v as JSON and writes it with the given status code.
// It sets Content-Type: application/json and handles the write error gracefully
// by simply discarding it — once headers are sent there is no way to signal
// failure to the client.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

// OK writes a 200 response with the given payload wrapped in the success envelope.
func OK(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// Created writes a 201 response.
func Created(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// NoContent writes a 204 response with an empty body.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// BadRequest writes a 400 error response.
func BadRequest(w http.ResponseWriter, code, message string) {
	writeJSON(w, http.StatusBadRequest, Response{
		Success: false,
		Error:   &ErrorInfo{Code: code, Message: message},
	})
}

// NotFound writes a 404 error response.
func NotFound(w http.ResponseWriter, code, message string) {
	writeJSON(w, http.StatusNotFound, Response{
		Success: false,
		Error:   &ErrorInfo{Code: code, Message: message},
	})
}

// InternalError writes a 500 error response.
func InternalError(w http.ResponseWriter, code, message string) {
	writeJSON(w, http.StatusInternalServerError, Response{
		Success: false,
		Error:   &ErrorInfo{Code: code, Message: message},
	})
}

// ServiceUnavailable writes a 503 error response.
func ServiceUnavailable(w http.ResponseWriter, code, message string) {
	writeJSON(w, http.StatusServiceUnavailable, Response{
		Success: false,
		Error:   &ErrorInfo{Code: code, Message: message},
	})
}
