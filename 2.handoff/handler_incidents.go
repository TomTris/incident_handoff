package main

import (
	"encoding/json"
	"errors"
	"net/http"
)

type FlagEvaluator interface {
	Evaluate(flagName string, userID string) (*FlagEvaluateAnswer, error)
}

type IncidentHandler struct {
	IncidentStore IncidentStore
	Registry      Registry
	FlagEvaluator FlagEvaluator
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

func marshalIncidentUpdateEvent(incAfter Incident) json.RawMessage {
	event := struct {
		Type     string   `json:"type"`
		Incident Incident `json:"incident"`
	}{
		Type:     "incident_updated",
		Incident: incAfter,
	}

	data, _ := json.Marshal(event)
	return data
}

func (incHandler *IncidentHandler) CreateIncident(r *http.Request) (*AppResponse, error) {
	req := CreateIncidentRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, BadRequest(err)
	}

	if err := req.Validate(); err != nil {
		return nil, BadRequest(err)
	}

	createdIncident, err := incHandler.IncidentStore.CreateIncident(r.Context(), req)
	if err != nil {
		return nil, InternalServerError(err)
	}

	return newAppResponse(http.StatusCreated, createdIncident), nil
}

func (incHandler *IncidentHandler) GetIncident(r *http.Request) (*AppResponse, error) {
	incidentID := r.PathValue("id")
	inc, err := incHandler.IncidentStore.GetIncident(r.Context(), incidentID)

	if err != nil {
		if errors.Is(err, ErrIncidentNotFound) {
			return nil, NotFound(err)
		}
		return nil, InternalServerError(err)
	}

	return newAppResponse(http.StatusOK, inc), nil
}

func (incHandler *IncidentHandler) AddEntry(r *http.Request) (*AppResponse, error) {
	timelineEntry := TimelineEntry{}

	if err := json.NewDecoder(r.Body).Decode(&timelineEntry); err != nil {
		return nil, BadRequest(err)
	}
	if err := timelineEntry.Validate(); err != nil {
		return nil, BadRequest(err)
	}
	incidentID := r.PathValue("id")
	newEntry, err := incHandler.IncidentStore.AddEntry(r.Context(), incidentID, timelineEntry)

	if err != nil {
		if errors.Is(err, ErrIncidentNotFound) {
			return nil, NotFound(err)
		}
		if errors.Is(err, ErrIncidentConflict) {
			return nil, Conflict(err)
		}
		return nil, InternalServerError(err)
	}

	incHandler.Registry.broadcast <- BroadcastMessage{
		incidentID: incidentID,
		msg:        marshalNewEntryEvent(incidentID, newEntry),
	}

	return newAppResponse(http.StatusCreated, newEntry), nil
}

func (incHandler *IncidentHandler) ListIncidents(r *http.Request) (*AppResponse, error) {
	incidentFilter := IncidentFilter{
		Status:  r.URL.Query().Get("status"),
		Service: r.URL.Query().Get("service"),
	}

	if err := incidentFilter.Validate(); err != nil {
		return nil, BadRequest(err)
	}

	filteredIncidents, err := incHandler.IncidentStore.ListIncidents(r.Context(), incidentFilter)
	if err != nil {
		return nil, InternalServerError(err)
	}
	return newAppResponse(http.StatusOK, filteredIncidents), nil
}

func (incHandler *IncidentHandler) UpdateIncident(r *http.Request) (*AppResponse, error) {
	incidentUpdate := IncidentUpdate{}
	if err := json.NewDecoder(r.Body).Decode(&incidentUpdate); err != nil {
		return nil, BadRequest(err)
	}
	if err := incidentUpdate.Validate(); err != nil {
		return nil, BadRequest(err)
	}

	incidentID := r.PathValue("id")
	incAfter, err := incHandler.IncidentStore.UpdateIncident(r.Context(), incidentID, incidentUpdate)
	if err != nil {
		if errors.Is(err, ErrIncidentNotFound) {
			return nil, NotFound(err)
		}
		return nil, InternalServerError(err)
	}

	incHandler.Registry.broadcast <- BroadcastMessage{
		msg:        marshalIncidentUpdateEvent(incAfter),
		incidentID: incidentID,
	}
	return newAppResponse(http.StatusNoContent, nil), nil
}

func (incHandler *IncidentHandler) GetHandoffBrief(r *http.Request) (*AppResponse, error) {
	incidentID := r.PathValue("id")
	inc, err := incHandler.IncidentStore.GetIncident(r.Context(), incidentID)

	if err != nil {
		if errors.Is(err, ErrIncidentNotFound) {
			return nil, NotFound(err)
		}
		return nil, InternalServerError(err)
	}

	userID := r.URL.Query().Get("user_id")
	body := buildHandoffBrief(inc, incHandler.FlagEvaluator, userID)
	return newAppResponse(http.StatusOK, body), nil
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
