package main

// import (
// 	"bytes"
// 	"encoding/json"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/prometheus/client_golang/prometheus"
// )

// func TestHealthCheck(t *testing.T) {
// 	memoryStore := MemoryStore{incidents: make(map[string]Incident)}
// 	incHandler := &IncidentHandler{Store: &memoryStore}
// 	router := getRouter(incHandler, nil, prometheus.NewRegistry())

// 	req := httptest.NewRequest("GET", "/healthz", nil)
// 	req.Header.Set("Content-Type", "application/json")
// 	rec := httptest.NewRecorder()

// 	router.ServeHTTP(rec, req)

// 	if rec.Code != 200 {
// 		t.Errorf("Code expected %d, got %d", 200, rec.Code)
// 	}

// 	got := bytes.TrimSpace(rec.Body.Bytes())
// 	expect, _ := json.Marshal(map[string]string{"status": "ok"})

// 	if !bytes.Equal(got, expect) {
// 		t.Errorf("Body expected %s, got %s", expect, got)
// 	}
// }
