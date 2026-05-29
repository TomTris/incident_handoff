package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type OnCallHandler struct {
	Store OnCallStore
}

type OnCallEntry struct {
	ID       string    `json:"id" bson:"_id"`
	Service  string    `json:"service" bson:"service"`
	Username string    `json:"username" bson:"username"`
	StartsAt time.Time `json:"starts_at" bson:"starts_at"`
	EndsAt   time.Time `json:"ends_at" bson:"ends_at"`
}

func (entry *OnCallEntry) Validation() error {
	if strings.TrimSpace(entry.Service) == "" {
		return ErrBadRequest
	}
	if strings.TrimSpace(entry.Username) == "" {
		return ErrBadRequest
	}
	if entry.StartsAt.After(entry.EndsAt) == true {
		return ErrBadRequest
	}
	return nil
}

// TODO: Define error code and have clear return value
func (h *OnCallHandler) CreateShift(r *http.Request) (*AppResponse, *AppError) {
	entry := OnCallEntry{}
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		return nil, BadRequest(err)
	}

	if err := entry.Validation(); err != nil {
		return nil, BadRequest(err)
	}

	created_entry, err := h.Store.Create(r.Context(), entry)
	if err != nil {
		return nil, InternalServerError(err)
	}
	return newAppResponse(http.StatusCreated, created_entry), nil
}

// TODO: Define error code and have clear return value
func (h *OnCallHandler) CurrentOnCall(r *http.Request) (*AppResponse, *AppError) {
	service := r.URL.Query().Get("service")
	if service == "" {
		return nil, BadRequest(ErrServiceRequired)
	}
	username, err := h.Store.CurrentOnCall(r.Context(), service)
	if err != nil {
		return nil, NotFound(err)
	}
	return newAppResponse(http.StatusOK, map[string]string{"username": username}), nil
}
