package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMarshalNewEntryEvent(t *testing.T) {
	entryTimeline := TimelineEntry{
		ID:     "TLE-1",
		Time:   time.Now(),
		Author: "anh",
		Type:   OBSERVATION,
		Text:   "test entry",
	}
	rawMsg := marshalNewEntryEvent("INC-test1", entryTimeline)
	var event map[string]any
	json.Unmarshal(rawMsg, &event)

	if event["type"] != "new_entry" {
		t.Fatalf("type expected %v, get %v", OBSERVATION, event["type"])
	}
	if event["incident_id"] != "INC-test1" {
		t.Fatalf("incident_id %v, get %v", "INC-test1", event["incident_id"])
	}
	e := event["entry"].(map[string](any))
	if e["author"] != "anh" {
		t.Fatalf("author expected %v, get %v", "anh", e["author"])
	}
	if e["type"] != OBSERVATION {
		t.Fatalf("type expected %v, get %v", OBSERVATION, e["type"])
	}
}

func TestMarshalIncidentUpdateEvent(t *testing.T) {
	inc := Incident{
		ID:        "INC-test1",
		Title:     "test title",
		Service:   "test service",
		Severity:  "SEV1",
		Status:    TRIGGERED,
		OpenedBy:  "anh",
		OnCall:    "tom",
		CreatedAt: time.Now().Add(-15 * time.Minute),
		UpdatedAt: time.Now(),
		Entries:   []TimelineEntry{},
	}

	rawMsg := marshalIncidentUpdateEvent(inc)
	var event map[string]any
	json.Unmarshal(rawMsg, &event)
	if event["type"] != "incident_updated" {
		t.Fatalf("type expected %v, get %v", "incident_updated", event["type"])
	}
	if event["type"] != "incident_updated" {
		t.Fatalf("type expected %v, get %v", "incident_updated", event["type"])
	}
	e := event["incident"].(map[string]any)
	if e["id"] != "INC-test1" {
		t.Fatalf("id expected %v, get %v", "INC-test1", e["id"])
	}
	if e["service"] != "test service" {
		t.Fatalf("id expected %v, get %v", "test service", e["service"])
	}
}

func TestGetIncidentOK(t *testing.T) {
	store := NewMemoryIncidentStore()
	validIncRequest := validCreateIncidentRequest()
	store.CreateIncident(context.Background(), validIncRequest)

	handler := IncidentHandler{IncidentStore: store}
	req := httptest.NewRequest("GET", "/incident/INC-1", nil)
	req.SetPathValue("id", "INC-1")

	res, err := handler.GetIncident(req)

	if err != nil {
		t.Fatalf("expected no error, get error %v", err)
	}
	if res.Status != http.StatusOK {
		t.Fatalf("expected status %v, get %v", http.StatusOK, res.Status)
	}
	inc := res.Body.(Incident)
	if inc.ID != "INC-1" {
		t.Fatalf("expected id %v, get %v", "INC-1", inc.ID)
	}
	if inc.Title != validIncRequest.Title {
		t.Fatalf("expected Title %v, get %v", validIncRequest.Title, inc.Title)
	}
	if inc.Service != validIncRequest.Service {
		t.Fatalf("expected Service %v, get %v", validIncRequest.Service, inc.Service)
	}
	if inc.Severity != validIncRequest.Severity {
		t.Fatalf("expected Severity %v, get %v", validIncRequest.Severity, inc.Severity)
	}
	if inc.OpenedBy != validIncRequest.OpenedBy {
		t.Fatalf("expected OpenedBy %v, get %v", validIncRequest.OpenedBy, inc.OpenedBy)
	}
	if inc.OnCall != *validIncRequest.OnCall {
		t.Fatalf("expected OnCall %v, get %v", validIncRequest.OnCall, inc.OnCall)
	}
}

func TestGetIncident404(t *testing.T) {
	store := NewMemoryIncidentStore()
	handler := IncidentHandler{IncidentStore: store}
	req := httptest.NewRequest("GET", "/incident/INC-1", nil)
	req.SetPathValue("id", "INC-1")

	_, err := handler.GetIncident(req)

	if err != nil {
		var appErr *AppError
		if errors.As(err, &appErr) {
			if appErr.Status != 404 {
				t.Fatalf("expected code 404, get %v", appErr.Status)
			}
		} else {
			t.Fatalf("expected no *AppError")
		}
	}
}

// func newTestServer(t *testing.T) *httptest.Server {
// 	t.Helper()
// 	promRegistry := prometheus.NewRegistry()
// 	httpMetrics := NewHttpMetrics(promRegistry)
// 	registryMetric := NewRegistryMetric(promRegistry)
// 	incidentStoreMetric := NewIncidentStoreMetric(promRegistry)
// 	registry := NewRegistry(registryMetric)
// 	go registry.run()
// 	t.Cleanup(func() { close(registry.done) })

// 	flagHandler := FlagHandler{store: CreateFlagStore()}
// 	instrumentedIncidentStore := InstrumentedIncidentStore{
// 		inner:   NewMemoryIncidentStore(),
// 		metrics: incidentStoreMetric,
// 	}
// 	incHandler := IncidentHandler{
// 		IncidentStore: &instrumentedIncidentStore,
// 		Registry:      NewRegistry(registryMetric),
// 		FlagEvaluator: &flagHandler.store,
// 	}

// 	router := getRouter(&incHandler, &flagHandler, nil, promRegistry, httpMetrics)
// 	return httptest.NewServer(router)
// }

func TestCreateIncident(t *testing.T) {

	incCreateRequest := validCreateIncidentRequest()
	store := NewMemoryIncidentStore()
	store.CreateIncident(context.Background(), incCreateRequest)
	handler := IncidentHandler{IncidentStore: store}

	bodyRaw, _ := json.Marshal(incCreateRequest)
	req := httptest.NewRequest("POST", "/incident", bytes.NewReader(bodyRaw))
	appRes, err := handler.CreateIncident(req)
	if err != nil {
		t.Fatalf("expected no error, get %v", err)
	}
	if appRes.Status != http.StatusCreated {
		t.Fatalf("status code expected %v, get %v", http.StatusCreated, appRes.Status)
	}
	// Evaluate
	response := appRes.Body.(Incident)
	if response.ID != "INC-2" {
		t.Fatalf("status code expected %v, get %v", "INC-2", response.ID)
	}
	if response.Title != incCreateRequest.Title {
		t.Fatalf("title expected %v, got %v", incCreateRequest.Title, response.Title)
	}
	if response.Severity != incCreateRequest.Severity {
		t.Fatalf("Severity expected %v, got %v", incCreateRequest.Severity, response.Severity)
	}
	if response.Service != incCreateRequest.Service {
		t.Fatalf("Service expected %v, got %v", incCreateRequest.Service, response.Service)
	}
}

func TestListIncident(t *testing.T) {

	store := NewMemoryIncidentStore()
	incCreateRequest := validCreateIncidentRequest()
	store.CreateIncident(context.Background(), incCreateRequest)
	incCreateRequest.Title = "123"
	store.CreateIncident(context.Background(), incCreateRequest)
	incCreateRequest.Service = "no services"
	store.CreateIncident(context.Background(), incCreateRequest)
	incCreateRequest.Severity = "SEV3"
	store.CreateIncident(context.Background(), incCreateRequest)

	handler := IncidentHandler{IncidentStore: store}

	req := httptest.NewRequest("GET", "/incidents", nil)
	appRes, err := handler.ListIncidents(req)
	if err != nil {
		t.Fatalf("expected no error, get %v", err)
	}
	if appRes.Status != http.StatusOK {
		t.Fatalf("status code expected %v, get %v", http.StatusOK, appRes.Status)
	}

	// Evaluate
	response := appRes.Body.([]Incident)
	if len(response) != 4 {
		t.Fatalf("len expect %v, get %v", 4, len(response))
	}
}

// func TestAddEntry(t *testing.T) {

// 	store := NewMemoryIncidentStore()
// 	incCreateRequest := validCreateIncidentRequest()
// 	inc, _ := store.CreateIncident(context.Background(), incCreateRequest)
// 	fmt.Println(inc)
// 	handler := IncidentHandler{IncidentStore: store}
// 	go  <- handler.Registry.broadcast (t)
// 	entry := TimelineEntry{
// 		Author: "tom",
// 		Type:   ACTION,
// 		Text:   "test action entry",
// 	}

// 	bodyRaw, _ := json.Marshal(entry)
// 	req := httptest.NewRequest("POST", "/incident/INC-1", bytes.NewReader(bodyRaw))
// 	req.SetPathValue("id", "INC-1")
// 	appRes, err := handler.AddEntry(req)
// 	if err != nil {
// 		t.Fatalf("expected no error, get %v", err)
// 	}
// 	if appRes.Status != http.StatusCreated {
// 		t.Fatalf("status code expected %v, get %v", http.StatusCreated, appRes.Status)
// 	}

// 	// Evaluate
// 	response := appRes.Body.(TimelineEntry)
// 	if response.ID != "TLE-1" {
// 		t.Fatalf("status code expected %v, get %v", "TLE-1", response.ID)
// 	}
// 	if response.Type != ACTION {
// 		t.Fatalf("title expected %v, got %v", ACTION, response.Type)
// 	}
// }
