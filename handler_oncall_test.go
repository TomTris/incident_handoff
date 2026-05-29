package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOnCallEntryValidation(t *testing.T) {
	entry := OnCallEntry{
		Service:  "payment",
		Username: "anh",
		StartsAt: time.Now(),
		EndsAt:   time.Now().Add(1 * time.Hour),
	}

	t.Run("normal OnCallEntry", func(t *testing.T) {
		if err := entry.Validation(); err != nil {
			t.Errorf("expect no error, get %v", err.Error())
		}
	})
	t.Run("OnCallEntry with empty Service", func(t *testing.T) {
		entry.Service = ""
		if err := entry.Validation(); err != ErrBadRequest {
			t.Errorf("expect %v, get %v", ErrBadRequest, err.Error())
		}
		entry.Service = "payment"
	})
	t.Run("OnCallEntry with empty username", func(t *testing.T) {
		entry.Username = ""
		if err := entry.Validation(); err != ErrBadRequest {
			t.Errorf("expect %v, get %v", ErrBadRequest, err.Error())
		}
		entry.Username = "anh"
	})
	t.Run("OnCallEntry with startsAt after EndsAt", func(t *testing.T) {
		entry.StartsAt = entry.EndsAt.Add(1 * time.Hour)
		if err := entry.Validation(); err != ErrBadRequest {
			t.Errorf("expect %v, get %v", ErrBadRequest, err.Error())
		}
		entry.StartsAt = entry.EndsAt.Add(-1 * time.Hour)
	})
}

func TestCreateShift201(t *testing.T) {
	onCallHandler := &OnCallHandler{Store: NewOnCallStore()}
	entry := OnCallEntry{
		Service:  "payment",
		Username: "anh",
		StartsAt: time.Now().Add(-1 * time.Hour),
		EndsAt:   time.Now().Add(1 * time.Hour),
	}
	entryRaw, _ := json.Marshal(entry)
	req := httptest.NewRequest("POST", "/", bytes.NewReader(entryRaw))
	req.SetPathValue("service", entry.Service)
	appRes, err := onCallHandler.CreateShift(req)
	if err != nil {
		t.Fatalf("expect no error, get %v", err)
	}
	if appRes.Status != http.StatusCreated {
		t.Fatalf("expect status %v, get %v", http.StatusCreated, appRes.Status)
	}
	body := appRes.Body.(OnCallEntry)
	if body.Service != entry.Service {
		t.Fatalf("expect same service")
	}
	if body.Username != entry.Username {
		t.Fatalf("expect same Username")
	}
	if body.StartsAt.Equal(entry.StartsAt) == false {
		t.Fatalf("expect same StartsAt")
	}
	if body.EndsAt.Equal(entry.EndsAt) == false {
		t.Fatalf("expect same EndsAt")
	}
}

func TestCurrentOnCall(t *testing.T) {
	onCallHandler := &OnCallHandler{Store: &InMemoryOnCallStore{
		OnCallEntries: make(map[string]OnCallEntry),
		currentID:     0,
	}}

	t.Run("no suitable CurrentOnCall", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/?service=payment", nil)
		_, err := onCallHandler.CurrentOnCall(req)
		if err == nil {
			t.Fatalf("expect error, get no err")
		}
	})

	t.Run("no suitable CurrentOnCall", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		_, appErr := onCallHandler.CurrentOnCall(req)
		if appErr == nil {
			t.Fatalf("expect error, get no err")
		}
		if appErr.Err != ErrServiceRequired {
			t.Fatalf("expect error %v, get %v", ErrServiceRequired, appErr.Err)
		}
	})

	t.Run("sucessful CurrentOnCall", func(t *testing.T) {
		entry := OnCallEntry{
			Service:  "payment",
			Username: "anh",
			StartsAt: time.Now().Add(-1 * time.Minute),
			EndsAt:   time.Now().Add(10 * time.Minute),
		}
		onCallHandler.Store.Create(context.Background(), entry)
		req := httptest.NewRequest("GET", "/?service=payment", nil)
		appRes, err := onCallHandler.CurrentOnCall(req)
		if err != nil {
			t.Fatalf("expect no error, get error %v", err)
		}
		if appRes.Status != http.StatusOK {
			t.Fatalf("expect status %v, get %v", http.StatusOK, appRes.Status)
		}
		body := appRes.Body.(map[string]string)
		if body["username"] != "anh" {
			t.Fatalf("expect username %v, get %v", "anh", body["username"])
		}
	})

}
