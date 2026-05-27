package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
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
		t.Fatalf("expected no error, get error %v", err.Error())
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

func TestCreateIncident(t *testing.T) {

	incCreateRequest := validCreateIncidentRequest()
	store := NewMemoryIncidentStore()
	store.CreateIncident(context.Background(), incCreateRequest)
	handler := IncidentHandler{IncidentStore: store}

	bodyRaw, _ := json.Marshal(incCreateRequest)
	req := httptest.NewRequest("POST", "/incident", bytes.NewReader(bodyRaw))
	appRes, err := handler.CreateIncident(req)
	if err != nil {
		t.Fatalf("expected no error, get %v", err.Error())
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
		t.Fatalf("expected no error, get %v", err.Error())
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

// init a server with an incident available
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	promRegistry := prometheus.NewRegistry()
	httpMetrics := NewHttpMetrics(promRegistry)
	registryMetric := NewRegistryMetric(promRegistry)
	incidentStoreMetric := NewIncidentStoreMetric(promRegistry)

	registry := NewRegistry(registryMetric)
	go registry.run()
	t.Cleanup(func() { close(registry.done) })

	flagHandler := FlagHandler{store: CreateFlagStore()}
	memStore := NewMemoryIncidentStore()
	instrumentedIncidentStore := InstrumentedIncidentStore{
		inner:   memStore,
		metrics: incidentStoreMetric,
	}
	incHandler := IncidentHandler{
		IncidentStore: &instrumentedIncidentStore,
		Registry:      registry,
		FlagEvaluator: &flagHandler.store,
	}

	memStore.CreateIncident(context.Background(), validCreateIncidentRequest())
	router := getRouter(&incHandler, &flagHandler, nil, promRegistry, httpMetrics)
	return httptest.NewServer(router)
}

func TestAddEntry(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/incidents/INC-1/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("expected no error, get error %v", err.Error())
	}
	defer conn.Close()

	entry := TimelineEntry{
		Author: "tom",
		Type:   OBSERVATION,
		Text:   "looking into A",
	}
	bodyRaw, _ := json.Marshal(entry)

	// HTTP Respsone
	resp, err := http.Post(srv.URL+"/incidents/INC-1/entries", "application/json", bytes.NewReader(bodyRaw))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status code expected %v, got %v", http.StatusCreated, resp.StatusCode)
	}
	var resEntry1 TimelineEntry
	json.NewDecoder(resp.Body).Decode(&resEntry1)

	if resEntry1.ID != "TLE-1" {
		t.Fatalf("entry ID expected %v, got %v", "TLE-1", resEntry1.ID)
	}
	if resEntry1.Type != OBSERVATION {
		t.Fatalf("entry type expected %v, got %v", OBSERVATION, resEntry1.Type)
	}

	// Websocket Response
	_, msgRaw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Expected no error, get %v", err)
	}
	var wsMsg map[string]any
	json.Unmarshal(msgRaw, &wsMsg)
	if wsMsg["type"] != "new_entry" {
		t.Fatalf("expected type %v, get %v", "new_entry", wsMsg["type"])
	}
	if wsMsg["incident_id"] != "INC-1" {
		t.Fatalf("expected Incident ID %v, get %v", "INC-1", wsMsg["incident_id"])
	}

	entryRaw, err := json.Marshal(wsMsg["entry"])
	if err != nil {
		t.Fatalf("Expected no error, got %v", err.Error())
	}

	var respEntry2 map[string]string
	json.Unmarshal(entryRaw, &respEntry2)

	if respEntry2["author"] != "tom" {
		t.Fatalf("author expected %v, get %v", "tom", respEntry2["author"])
	}
	if respEntry2["type"] != OBSERVATION {
		t.Fatalf("type expected %v, get %v", OBSERVATION, respEntry2["type"])
	}
}

func TestUpdateIncident(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/incidents/INC-1/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("expected no error, get error %v", err.Error())
	}
	defer conn.Close()

	incidentUpdate := IncidentUpdate{
		Status:   new(RESOLVED),
		Severity: new("SEV2"),
	}
	bodyRaw, _ := json.Marshal(incidentUpdate)

	// HTTP Respsone
	req, err := http.NewRequest(http.MethodPatch, srv.URL+"/incidents/INC-1", bytes.NewReader(bodyRaw))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status code expected %v, got %v", http.StatusNoContent, resp.StatusCode)
	}

	// Websocket Response
	_, msgRaw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Expected no error, get %v", err)
	}
	var wsMsg map[string]any
	json.Unmarshal(msgRaw, &wsMsg)
	if wsMsg["type"] != "incident_updated" {
		t.Fatalf("expected type %v, get %v", "incident_updated", wsMsg["type"])
	}

	incidentRaw, err := json.Marshal(wsMsg["incident"])
	if err != nil {
		t.Fatalf("Expected no error, got %v", err.Error())
	}

	var resIncident map[string]string
	json.Unmarshal(incidentRaw, &resIncident)

	if resIncident["id"] != "INC-1" {
		t.Fatalf("expected Incident ID %v, get %v", "INC-1", wsMsg["incident_id"])
	}
	if resIncident["status"] != RESOLVED {
		t.Fatalf("author expected %v, get %v", RESOLVED, resIncident["status"])
	}
	if resIncident["severity"] != "SEV2" {
		t.Fatalf("type expected %v, get %v", "SEV2", resIncident["severity"])
	}
}

func TestGetHandoffBrief(t *testing.T) {
	store := NewMemoryIncidentStore()
	validIncRequest := validCreateIncidentRequest()
	store.CreateIncident(context.Background(), validIncRequest)

	handler := IncidentHandler{IncidentStore: store}
	req := httptest.NewRequest("GET", "/incidents/INC-1/handoff?user_id=tom", nil)
	req.SetPathValue("id", "INC-1")

	res, err := handler.GetHandoffBrief(req)
	if err != nil {
		t.Fatalf("expected no error, get %v", err.Error())
	}
	if res.Status != http.StatusOK {
		t.Fatalf("expected status %v, get %v", http.StatusOK, res.Status)
	}

	bodyRaw, err := json.Marshal(res.Body)
	if err != nil {
		t.Fatalf("expected not nil, get %v", err.Error())
	}
	var body HandoffBrief
	err = json.Unmarshal(bodyRaw, &body)
	if err != nil {
		t.Fatalf("expected not nil, get %v", err.Error())
	}
}
