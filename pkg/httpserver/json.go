package httpserver

import (
	"encoding/json"
	"net/http"
)

// APIError is a small JSON error envelope for HTTP APIs.
type APIError struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// WriteJSON writes v as JSON with the given status.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// WriteError writes a JSON API error. msg is a short machine-oriented code
// (e.g. "bad_request"); details is a human-readable explanation.
func WriteError(w http.ResponseWriter, status int, msg string, details string) {
	WriteJSON(w, status, APIError{Error: msg, Details: details})
}

func OK(w http.ResponseWriter, data any) {
	WriteJSON(w, http.StatusOK, HTTPResponse{Status: http.StatusOK, Message: "Success", Data: data})
}

func BadRequest(w http.ResponseWriter, details string) {
	WriteError(w, http.StatusBadRequest, "bad_request", details)
}

func Unauthorized(w http.ResponseWriter, details string) {
	WriteError(w, http.StatusUnauthorized, "unauthorized", details)
}
