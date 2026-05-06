package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type IncidentHandler struct {
	Store Store
}

func (incHandler *IncidentHandler) GetIncident(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetIncident")
	incidentID := r.PathValue("id")
	inc, err := incHandler.Store.GetIncident(r.Context(), incidentID)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrorMessageJSON{
			ErrorCode: "BAD_REQUEST",
			Message:   err.Error(),
			RequestID: r.Context().Value(requestIDKey).(string),
		})
		return
	}
	writeJSON(w, http.StatusOK, inc)
}

func (incHandler *IncidentHandler) AddEntry(w http.ResponseWriter, r *http.Request) {
	fmt.Println("AddEntry")
	timelineEntry := TimelineEntry{}
	err := json.NewDecoder(r.Body).Decode(&timelineEntry)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrorMessageJSON{
			ErrorCode: "BAD_REQUEST",
			Message:   err.Error(),
			RequestID: r.Context().Value(requestIDKey).(string),
		})
		return
	}
	err = timelineEntry.Validate()
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrorMessageJSON{
			ErrorCode: "BAD_REQUEST",
			Message:   err.Error(),
			RequestID: r.Context().Value(requestIDKey).(string),
		})
		return
	}
	incidentID := r.PathValue("id")
	newEntry, err := incHandler.Store.AddEntry(r.Context(), incidentID, timelineEntry)
	if err != nil {
		if errors.Is(err, ErrIncidentNotFound) {
			writeError(w, http.StatusBadRequest, ErrorMessageJSON{
				ErrorCode: "BAD_REQUEST",
				Message:   err.Error(),
				RequestID: r.Context().Value(requestIDKey).(string),
			})
			return
		}
		// If the incident is already resolved
		if errors.Is(err, ErrConflict) {
			writeError(w, http.StatusConflict, ErrorMessageJSON{
				ErrorCode: "CONFLICT",
				Message:   err.Error(),
				RequestID: r.Context().Value(requestIDKey).(string),
			})
			return
		}
		writeError(w, http.StatusBadRequest, ErrorMessageJSON{
			ErrorCode: "INTERNAL_ERROR",
			Message:   err.Error(),
			RequestID: r.Context().Value(requestIDKey).(string),
		})
		return
	}
	writeJSON(w, http.StatusCreated, newEntry)
}

func (incHandler *IncidentHandler) CreateIncident(w http.ResponseWriter, r *http.Request) {
	fmt.Println("CreateIncident")
	newCreateIncidentRequest := CreateIncidentRequest{}
	err := json.NewDecoder(r.Body).Decode(&newCreateIncidentRequest)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorMessageJSON{
			ErrorCode: "BAD_REQUEST",
			Message:   fmt.Sprintf("Validation invalid: %s", err),
			RequestID: r.Context().Value(requestIDKey).(string),
		})
		return
	}

	err = newCreateIncidentRequest.Validate()
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorMessageJSON{
			ErrorCode: "BAD_REQUEST",
			Message:   fmt.Sprintf("Validation invalid: %s", err),
			RequestID: r.Context().Value(requestIDKey).(string),
		})
		return
	}

	createdIncident, err := incHandler.Store.CreateIncident(r.Context(), Incident{
		Title:    newCreateIncidentRequest.Title,
		Service:  newCreateIncidentRequest.Service,
		Severity: newCreateIncidentRequest.Severity,
		OpenedBy: newCreateIncidentRequest.OpenedBy,
	})

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorMessageJSON{
			ErrorCode: "INTERNAL_ERROR",
			Message:   "failed to create incident",
			RequestID: r.Context().Value(requestIDKey).(string),
		})
		return
	}

	writeJSON(w, http.StatusCreated, createdIncident)
	return
}
