package main

import (
	"encoding/json"
	"errors"
	"net/http"
)

type IncidentHandler struct {
	Store    Store
	Registry Registry
}

func marshalNewEntryEvent(incidentID string, entry TimelineEntry) json.RawMessage {
	event := struct {
		Type       string        `json:"type"`
		IncidentID string        `json:"incident_id"`
		Entry      TimelineEntry `json:"entry"`
	}{
		Type:       "new_entry",
		IncidentID: incidentID,
		Entry:      entry,
	}
	data, _ := json.Marshal(event)
	return data
}

func marshalStateChangeEvent(incidentID string, update IncidentUpdate) json.RawMessage {
	// One message per changed field
	// Or one message with all changes — your call
	event := struct {
		Type       string         `json:"type"`
		IncidentID string         `json:"incident_id"`
		Update     IncidentUpdate `json:"update"`
	}{
		Type:       "state_change",
		IncidentID: incidentID,
		Update:     update,
	}
	data, _ := json.Marshal(event)
	return data
}

func (incHandler *IncidentHandler) CreateIncident(w http.ResponseWriter, r *http.Request) {
	RequestID := r.Context().Value(requestIDKey).(string)
	req := CreateIncidentRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrorMessageJSON{
			ErrorCode: BAD_REQUEST,
			Message:   err.Error(),
			RequestID: RequestID,
		})
		return
	}

	err = req.Validate()
	if err != nil {
		if errors.Is(err, ErrOnCall) {
			writeError(w, http.StatusBadRequest, ErrorMessageJSON{
				ErrorCode: BAD_REQUEST,
				Message:   err.Error(),
				RequestID: RequestID,
			})
			return
		}
		writeError(w, http.StatusBadRequest, ErrorMessageJSON{
			ErrorCode: MISSING_FIELD,
			Message:   err.Error(),
			RequestID: RequestID,
		})
		return
	}

	createdIncident, err := incHandler.Store.CreateIncident(r.Context(), Incident{
		Title:    req.Title,
		Service:  req.Service,
		Severity: req.Severity,
		OpenedBy: req.OpenedBy,
		OnCall:   derefOrEmpty(req.OnCall),
	})

	if err != nil {
		writeError(w, http.StatusInternalServerError, ErrorMessageJSON{
			ErrorCode: INTERNAL_SERVER_ERROR,
			Message:   err.Error(),
			RequestID: RequestID,
		})
		return
	}

	writeJSON(w, http.StatusCreated, RequestID, createdIncident)
}

func (incHandler *IncidentHandler) GetIncident(w http.ResponseWriter, r *http.Request) {
	RequestID := r.Context().Value(requestIDKey).(string)
	incidentID := r.PathValue("id")
	inc, err := incHandler.Store.GetIncident(r.Context(), incidentID)
	if err != nil {
		if errors.Is(err, ErrIncidentNotFound) {
			writeError(w, http.StatusNotFound, ErrorMessageJSON{
				ErrorCode: INCIDENT_NOT_FOUND,
				Message:   err.Error(),
				RequestID: RequestID,
			})
			return
		}
		writeError(w, http.StatusInternalServerError, ErrorMessageJSON{
			ErrorCode: INTERNAL_SERVER_ERROR,
			Message:   err.Error(),
			RequestID: RequestID,
		})
		return
	}
	writeJSON(w, http.StatusOK, RequestID, inc)
}

func (incHandler *IncidentHandler) AddEntry(w http.ResponseWriter, r *http.Request) {
	RequestID := r.Context().Value(requestIDKey).(string)
	timelineEntry := TimelineEntry{}
	err := json.NewDecoder(r.Body).Decode(&timelineEntry)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrorMessageJSON{
			ErrorCode: BAD_REQUEST,
			Message:   err.Error(),
			RequestID: RequestID,
		})
		return
	}
	err = timelineEntry.Validate()
	if err != nil {
		if errors.Is(err, ErrBadEntryType) {
			writeError(w, http.StatusBadRequest, ErrorMessageJSON{
				ErrorCode: BAD_REQUEST,
				Message:   err.Error(),
				RequestID: RequestID,
			})
			return
		}
		writeError(w, http.StatusBadRequest, ErrorMessageJSON{
			ErrorCode: MISSING_FIELD,
			Message:   err.Error(),
			RequestID: RequestID,
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
				RequestID: RequestID,
			})
			return
		}
		// If the incident is already resolved
		if errors.Is(err, ErrIncidentConflict) {
			writeError(w, http.StatusConflict, ErrorMessageJSON{
				ErrorCode: CONFLICT,
				Message:   err.Error(),
				RequestID: RequestID,
			})
			return
		}
		writeError(w, http.StatusInternalServerError, ErrorMessageJSON{
			ErrorCode: INTERNAL_SERVER_ERROR,
			Message:   err.Error(),
			RequestID: RequestID,
		})
		return
	}
	incHandler.Registry.broadcast <- BroadcastMessage{
		incidentID: incidentID,
		msg:        marshalNewEntryEvent(incidentID, newEntry),
	}
	writeJSON(w, http.StatusCreated, RequestID, newEntry)
}

func (incHandler *IncidentHandler) ListIncidents(w http.ResponseWriter, r *http.Request) {
	RequestID := r.Context().Value(requestIDKey).(string)
	incidentFilter := IncidentFilter{
		Status:  r.URL.Query().Get("status"),
		Service: r.URL.Query().Get("service"),
	}

	err := incidentFilter.Validate()
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrorMessageJSON{
			ErrorCode: BAD_REQUEST,
			Message:   err.Error(),
			RequestID: RequestID,
		})
		return
	}

	filteredIncidents, err := incHandler.Store.ListIncidents(r.Context(), incidentFilter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, ErrorMessageJSON{
			ErrorCode: INTERNAL_SERVER_ERROR,
			Message:   err.Error(),
			RequestID: RequestID,
		})
		return
	}
	writeJSON(w, http.StatusOK, RequestID, filteredIncidents)
}

// TODO:I can return state before update from UpdateIncident, but it shouldn't be best practice and make the code less clean
// So i do not do it.
// Do Change stream in the future instead.
func (incHandler *IncidentHandler) UpdateIncident(w http.ResponseWriter, r *http.Request) {
	RequestID := r.Context().Value(requestIDKey).(string)
	incidentUpdate := IncidentUpdate{}
	err := json.NewDecoder(r.Body).Decode(&incidentUpdate)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrorMessageJSON{
			ErrorCode: BAD_REQUEST,
			Message:   err.Error(),
			RequestID: RequestID,
		})
		return
	}
	err = incidentUpdate.Validate()
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrorMessageJSON{
			ErrorCode: BAD_REQUEST,
			Message:   err.Error(),
			RequestID: RequestID,
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
				RequestID: RequestID,
			})
			return
		}
		writeError(w, http.StatusInternalServerError, ErrorMessageJSON{
			ErrorCode: INTERNAL_SERVER_ERROR,
			Message:   err.Error(),
			RequestID: RequestID,
		})
		return
	}
	incHandler.Registry.broadcast <- BroadcastMessage{
		incidentID: incidentID,
		msg:        marshalStateChangeEvent(incidentID, incidentUpdate),
	}
	w.WriteHeader(http.StatusNoContent)
}

func (incHandler *IncidentHandler) GetHandoffBrief(w http.ResponseWriter, r *http.Request) {
	RequestID := r.Context().Value(requestIDKey).(string)
	incidentID := r.PathValue("id")
	inc, err := incHandler.Store.GetIncident(r.Context(), incidentID)
	if err != nil {
		if errors.Is(err, ErrIncidentNotFound) {
			writeError(w, http.StatusNotFound, ErrorMessageJSON{
				ErrorCode: INCIDENT_NOT_FOUND,
				Message:   err.Error(),
				RequestID: RequestID,
			})
			return
		}
		writeError(w, http.StatusInternalServerError, ErrorMessageJSON{
			ErrorCode: INTERNAL_SERVER_ERROR,
			Message:   err.Error(),
			RequestID: RequestID,
		})
		return
	}
	writeJSON(w, http.StatusOK, RequestID, buildHandoffBrief(inc))
}

func (incHandler *IncidentHandler) HandleIncidentWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		RequestID := r.Context().Value(requestIDKey).(string)
		writeError(w, http.StatusInternalServerError, ErrorMessageJSON{
			ErrorCode: INTERNAL_SERVER_ERROR,
			Message:   err.Error(),
			RequestID: RequestID,
		})
		return
	}
	incidentID := r.PathValue("id")
	client := newClient(incidentID, conn)
	client.joinRegistry(&incHandler.Registry)

	go client.writePump()
	go client.readPump()
}
