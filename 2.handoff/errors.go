package main

import (
	"errors"
	"net/http"
)

type ErrorMessageJSON struct {
	ErrorCode string `json:"code" bson:"code"`
	Message   string `json:"message" bson:"message"`
	RequestID string `json:"request_id" bson:"request_id"`
}

func writeError(w http.ResponseWriter, status int, e ErrorMessageJSON) {
	writeJSON(w, status, e.RequestID, map[string]ErrorMessageJSON{"error": e})
}

var ErrNoAuthor = errors.New("Invalid Author")
var ErrBadEntryType = errors.New("Bad Timeline Entry Type")
var ErrNoText = errors.New("Invalid Text")

var ErrBadRequest = errors.New("Bad Request")
var ErrBadIncidentStatus = errors.New("Bad Service")
var ErrIncidentNotFound = errors.New("Incident not found")
var ErrIncidentConflict = errors.New("Incident resolved")
var ErrNoTitle = errors.New("Invalid Title")
var ErrNoService = errors.New("Invalid Service")
var ErrInvalidSeverity = errors.New("Invalid Severity")
var ErrOpenedBy = errors.New("Invalid open_by")
var ErrOnCall = errors.New("Invalid on_call") //The variable on_call must be either empty or non-existent

var ErrInternal = errors.New("Internal Error")

var ErrFlagNotfound = errors.New("Flag Not Found")

const (
	INCIDENT_NOT_FOUND    = "INCIDENT_NOT_FOUND"
	INTERNAL_SERVER_ERROR = "INTERNAL_SERVER_ERROR"
	BAD_REQUEST           = "BAD_REQUEST"
	CONFLICT              = "CONFLICT"
	MISSING_FIELD         = "MISSING_FIELD"

	FLAG_NOT_FOUND = "FLAG_NOT_FOUND"
)
