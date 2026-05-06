package main

import (
	"errors"
	"net/http"
)

type ErrorMessageJSON struct {
	ErrorCode string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func writeError(w http.ResponseWriter, status int, e ErrorMessageJSON) {
	writeJSON(w, status, map[string]ErrorMessageJSON{"error": e})
}

var ErrIncidentNotFound = errors.New("Incident not found")
var ErrConflict = errors.New("The Incident is already resolved")

// Will use if we have database
var ErrInternal = errors.New("Internal Error")
