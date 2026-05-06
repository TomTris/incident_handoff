package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type IncidentHandler struct {
	Store Store
}

func (incHandler *IncidentHandler) GetIncident(w http.ResponseWriter, r *http.Request) {
	incidentID := r.PathValue("id")
	inc, err := incHandler.Store.GetIncident(r.Context(), incidentID)
	if err != nil {
		if errors.Is(err, ErrIncidentNotFound) {
			writeError(w, http.StatusNotFound, ErrorMessageJSON{
				ErrorCode: INCIDENT_NOT_FOUND,
				Message:   err.Error(),
				RequestID: r.Context().Value(requestIDKey).(string),
			})
			return
		}
		writeError(w, http.StatusInternalServerError, ErrorMessageJSON{
			ErrorCode: INTERNAL_SERVER_ERROR,
			Message:   err.Error(),
			RequestID: r.Context().Value(requestIDKey).(string),
		})
		return
	}
	writeJSON(w, http.StatusOK, inc)
}

func (incHandler *IncidentHandler) AddEntry(w http.ResponseWriter, r *http.Request) {
	timelineEntry := TimelineEntry{}
	err := json.NewDecoder(r.Body).Decode(&timelineEntry)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrorMessageJSON{
			ErrorCode: BAD_REQUEST,
			Message:   err.Error(),
			RequestID: r.Context().Value(requestIDKey).(string),
		})
		return
	}
	err = timelineEntry.Validate()
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrorMessageJSON{
			ErrorCode: BAD_REQUEST,
			Message:   err.Error(),
			RequestID: r.Context().Value(requestIDKey).(string),
		})
		return
	}
	incidentID := r.PathValue("id")
	newEntry, err := incHandler.Store.AddEntry(r.Context(), incidentID, timelineEntry)
	if err != nil {
		if errors.Is(err, ErrIncidentNotFound) {
			writeError(w, http.StatusNotFound, ErrorMessageJSON{
				ErrorCode: INCIDENT_NOT_FOUND,
				Message:   err.Error(),
				RequestID: r.Context().Value(requestIDKey).(string),
			})
			return
		}
		// If the incident is already resolved
		if errors.Is(err, ErrIncidentConflict) {
			writeError(w, http.StatusConflict, ErrorMessageJSON{
				ErrorCode: CONFLICT,
				Message:   err.Error(),
				RequestID: r.Context().Value(requestIDKey).(string),
			})
			return
		}
		writeError(w, http.StatusInternalServerError, ErrorMessageJSON{
			ErrorCode: INTERNAL_SERVER_ERROR,
			Message:   err.Error(),
			RequestID: r.Context().Value(requestIDKey).(string),
		})
		return
	}
	writeJSON(w, http.StatusCreated, newEntry)
}

func (incHandler *IncidentHandler) CreateIncident(w http.ResponseWriter, r *http.Request) {
	req := CreateIncidentRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrorMessageJSON{
			ErrorCode: BAD_REQUEST,
			Message:   err.Error(),
			RequestID: r.Context().Value(requestIDKey).(string),
		})
		return
	}

	err = req.Validate()
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrorMessageJSON{
			ErrorCode: "MISSING_FIELD",
			Message:   err.Error(),
			RequestID: r.Context().Value(requestIDKey).(string),
		})
		return
	}

	onCall := req.OpenedBy
	if req.OnCall != nil {
		onCall = *req.OnCall
	}
	createdIncident, err := incHandler.Store.CreateIncident(r.Context(), Incident{
		Title:    req.Title,
		Service:  req.Service,
		Severity: req.Severity,
		OpenedBy: req.OpenedBy,
		OnCall:   onCall,
	})

	if err != nil {
		writeError(w, http.StatusInternalServerError, ErrorMessageJSON{
			ErrorCode: INTERNAL_SERVER_ERROR,
			Message:   err.Error(),
			RequestID: r.Context().Value(requestIDKey).(string),
		})
		return
	}

	writeJSON(w, http.StatusCreated, createdIncident)
	return
}

func (incHandler *IncidentHandler) ListIncidents(w http.ResponseWriter, r *http.Request) {
	incidentFilter := IncidentFilter{
		Status:  r.URL.Query().Get("status"),
		Service: r.URL.Query().Get("service"),
	}

	err := incidentFilter.Validate()
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorMessageJSON{
			ErrorCode: BAD_REQUEST,
			Message:   err.Error(),
			RequestID: r.Context().Value(requestIDKey).(string),
		})
		return
	}

	filteredIncidents, err := incHandler.Store.ListIncidents(r.Context(), incidentFilter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, ErrorMessageJSON{
			ErrorCode: INTERNAL_SERVER_ERROR,
			Message:   err.Error(),
			RequestID: r.Context().Value(requestIDKey).(string),
		})
		return
	}
	writeJSON(w, http.StatusOK, filteredIncidents)
}

func (incHandler *IncidentHandler) UpdateIncident(w http.ResponseWriter, r *http.Request) {
	incidentUpdate := IncidentUpdate{}
	err := json.NewDecoder(r.Body).Decode(&incidentUpdate)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorMessageJSON{
			ErrorCode: BAD_REQUEST,
			Message:   err.Error(),
			RequestID: r.Context().Value(requestIDKey).(string),
		})
		return
	}
	err = incidentUpdate.Validate()
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorMessageJSON{
			ErrorCode: BAD_REQUEST,
			Message:   err.Error(),
			RequestID: r.Context().Value(requestIDKey).(string),
		})
		return
	}
	incidentID := r.PathValue("id")
	err = incHandler.Store.UpdateIncident(r.Context(), incidentID, incidentUpdate)
	if err != nil {
		if errors.Is(err, ErrIncidentNotFound) {
			writeError(w, http.StatusNotFound, ErrorMessageJSON{
				ErrorCode: INCIDENT_NOT_FOUND,
				Message:   err.Error(),
				RequestID: r.Context().Value(requestIDKey).(string),
			})
			return
		}
		writeError(w, http.StatusInternalServerError, ErrorMessageJSON{
			ErrorCode: INTERNAL_SERVER_ERROR,
			Message:   err.Error(),
			RequestID: r.Context().Value(requestIDKey).(string),
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (incHandler *IncidentHandler) GetHandoffBrief(w http.ResponseWriter, r *http.Request) {
	incidentID := r.PathValue("id")
	inc, err := incHandler.Store.GetIncident(r.Context(), incidentID)
	if err != nil {
		if errors.Is(err, ErrIncidentNotFound) {
			writeError(w, http.StatusNotFound, ErrorMessageJSON{
				ErrorCode: INCIDENT_NOT_FOUND,
				Message:   err.Error(),
				RequestID: r.Context().Value(requestIDKey).(string),
			})
			return
		}
		writeError(w, http.StatusInternalServerError, ErrorMessageJSON{
			ErrorCode: INTERNAL_SERVER_ERROR,
			Message:   err.Error(),
			RequestID: r.Context().Value(requestIDKey).(string),
		})
		return
	}
	writeJSON(w, http.StatusOK, buildHandoffBrief(inc))
}

func buildHandoffBrief(inc Incident) HandoffBrief {
	actions := []TimelineEntry{}
	openQuestions := []TimelineEntry{}

	for _, entry := range inc.Entries {
		switch entry.Type {
		case ACTION:
			actions = append(actions, entry)
		case OPEN_QUESTION:
			openQuestions = append(openQuestions, entry)
		}
	}

	return HandoffBrief{
		Severity:      inc.Severity,
		Status:        inc.Status,
		Service:       inc.Service,
		ElapsedMinute: int(time.Since(inc.CreatedAt).Minutes()),
		TotalEntry:    len(inc.Entries),
		TakenActions:  actions,
		OpenQuestion:  openQuestions,
		CreatedAt:     inc.CreatedAt,
	}
}
