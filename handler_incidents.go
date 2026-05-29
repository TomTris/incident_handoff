package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type FlagEvaluator interface {
	Evaluate(flagName string, userID string) (*FlagEvaluateAnswer, error)
}

type CurrentOnCall interface {
	CurrentOnCall(ctx context.Context, service string) (string, error)
}

type IncidentHandler struct {
	IncidentStore IncidentStore
	Registry      Registry
	FlagEvaluator FlagEvaluator
	CurrentOnCall CurrentOnCall
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

func (h *IncidentHandler) CreateIncident(r *http.Request) (*AppResponse, *AppError) {
	req := CreateIncidentRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, BadRequest(err)
	}

	if err := req.Validate(); err != nil {
		return nil, BadRequest(err)
	}

	onCall, err := h.CurrentOnCall.CurrentOnCall(r.Context(), req.Service)
	if err != nil {
		if errors.Is(err, OnCallUserNotFound) {
			onCall = ""
		} else {
			return nil, InternalServerError(err)
		}
	}
	req.OnCall = onCall

	createdIncident, err := h.IncidentStore.CreateIncident(r.Context(), req)
	if err != nil {
		return nil, InternalServerError(err)
	}

	return newAppResponse(http.StatusCreated, createdIncident), nil
}

func (h *IncidentHandler) GetIncident(r *http.Request) (*AppResponse, *AppError) {
	incidentID := r.PathValue("id")
	inc, err := h.IncidentStore.GetIncident(r.Context(), incidentID)

	if err != nil {
		if errors.Is(err, ErrIncidentNotFound) {
			return nil, NotFound(err)
		}
		return nil, InternalServerError(err)
	}

	return newAppResponse(http.StatusOK, inc), nil
}

// TODO Handle error cases properly
func (h *IncidentHandler) AddEntry(r *http.Request) (*AppResponse, *AppError) {
	user := r.Context().Value(userContextKey).(UserContext)
	inc, err := h.IncidentStore.GetIncident(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, ErrIncidentNotFound) {
			return nil, NotFound(err)
		}
		return nil, InternalServerError(err)
	}
	if err := AuthorizeIncidentAction(user, inc, ActionAddEntry); err != nil {
		return nil, Forbidden(err)
	}

	timelineEntry := TimelineEntry{}
	if err := json.NewDecoder(r.Body).Decode(&timelineEntry); err != nil {
		return nil, BadRequest(err)
	}
	if err := timelineEntry.Validate(); err != nil {
		return nil, BadRequest(err)
	}
	incidentID := r.PathValue("id")

	newEntry, err := h.IncidentStore.AddEntry(r.Context(), incidentID, timelineEntry)

	if err != nil {
		if errors.Is(err, ErrIncidentNotFound) {
			return nil, NotFound(err)
		}
		if errors.Is(err, ErrIncidentConflict) {
			return nil, Conflict(err)
		}
		return nil, InternalServerError(err)
	}

	h.Registry.broadcast <- BroadcastMessage{
		incidentID: incidentID,
		msg:        marshalNewEntryEvent(incidentID, newEntry),
	}

	return newAppResponse(http.StatusCreated, newEntry), nil
}

func (h *IncidentHandler) ListIncidents(r *http.Request) (*AppResponse, *AppError) {
	incidentFilter := IncidentFilter{
		Status:  r.URL.Query().Get("status"),
		Service: r.URL.Query().Get("service"),
	}

	if err := incidentFilter.Validate(); err != nil {
		return nil, BadRequest(err)
	}

	filteredIncidents, err := h.IncidentStore.ListIncidents(r.Context(), incidentFilter)
	if err != nil {
		return nil, InternalServerError(err)
	}
	return newAppResponse(http.StatusOK, filteredIncidents), nil
}

func (h *IncidentHandler) UpdateIncident(r *http.Request) (*AppResponse, *AppError) {
	user := r.Context().Value(userContextKey).(UserContext)
	inc, err := h.IncidentStore.GetIncident(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, ErrIncidentNotFound) {
			return nil, NotFound(err)
		}
		return nil, InternalServerError(err)
	}
	if err := AuthorizeIncidentAction(user, inc, ActionUpdateIncident); err != nil {
		return nil, Forbidden(err)
	}

	incidentUpdate := IncidentUpdate{}
	if err := json.NewDecoder(r.Body).Decode(&incidentUpdate); err != nil {
		return nil, BadRequest(err)
	}
	if err := incidentUpdate.Validate(); err != nil {
		return nil, BadRequest(err)
	}

	incidentID := r.PathValue("id")
	incAfter, err := h.IncidentStore.UpdateIncident(r.Context(), incidentID, incidentUpdate)
	if err != nil {
		if errors.Is(err, ErrIncidentNotFound) {
			return nil, NotFound(err)
		}
		return nil, InternalServerError(err)
	}

	h.Registry.broadcast <- BroadcastMessage{
		msg:        marshalIncidentUpdateEvent(incAfter),
		incidentID: incidentID,
	}
	return newAppResponse(http.StatusNoContent, nil), nil
}

func (h *IncidentHandler) GetHandoffBrief(r *http.Request) (*AppResponse, *AppError) {
	incidentID := r.PathValue("id")
	inc, err := h.IncidentStore.GetIncident(r.Context(), incidentID)

	if err != nil {
		if errors.Is(err, ErrIncidentNotFound) {
			return nil, NotFound(err)
		}
		return nil, InternalServerError(err)
	}

	user := r.Context().Value(userContextKey).(UserContext)
	body := buildHandoffBrief(inc, h.FlagEvaluator, user.ID)
	return newAppResponse(http.StatusOK, body), nil
}

func (h *IncidentHandler) HandleIncidentWebSocket(w http.ResponseWriter, r *http.Request) {
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
	client.joinRegistry(&h.Registry)

	go client.writePump()
	go client.readPump()
}
