package httpserver

import (
	"encoding/json"
	"net/http"
)

type APIError struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	Data    any    `json:"data"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, details string) {
	WriteJSON(w, status, APIError{Error: http.StatusText(status), Details: details})
}

func OK(w http.ResponseWriter, data any) {
	WriteJSON(w, http.StatusOK, HTTPResponse{Status: http.StatusOK, Message: "Success", Data: data})
}

func BadRequest(w http.ResponseWriter, details string) {
	WriteError(w, http.StatusBadRequest, details)
}

func Unauthorized(w http.ResponseWriter, details string) {
	WriteError(w, http.StatusUnauthorized, details)
}
