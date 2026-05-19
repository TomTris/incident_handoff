package main

// import (
// 	"context"
// 	"encoding/json"
// 	"net/http"
// 	"net/http/httptest"
// 	"strings"
// 	"testing"

// 	"github.com/gorilla/websocket"
// 	"github.com/prometheus/client_golang/prometheus"
// )

// func TestCreateIncident(t *testing.T) {
// 	config := loadConfig()
// 	client, store := NewStore(config)
// 	incHandler := IncidentHandler{Store: store, Registry: NewRegistry()}
// 	router := getRouter(&incHandler, client, prometheus.NewRegistry())
// 	go incHandler.Registry.run()
// 	defer close(incHandler.Registry.done)
// 	if instrumented, ok := store.(*InstrumentedStore); ok {
// 		if ms, ok := instrumented.s.(*MongoStore); ok {
// 			ms.DropAll(context.Background())
// 		}
// 	}

// 	t.Run("Normal Request", func(t *testing.T) {

// 		body := `{"title": "order-service request drop", "service": "order-service", "severity": "SEV1", "opened_by": "Anh Nguyen"}`
// 		req := httptest.NewRequest("POST", "/incidents", strings.NewReader(body))
// 		req.Header.Set("Content-Type", "application/json")
// 		rec := httptest.NewRecorder()
// 		router.ServeHTTP(rec, req)

// 		if rec.Code != 201 {
// 			t.Errorf("Code expected %d, got %d", 200, rec.Code)
// 		}

// 		var got Incident
// 		json.NewDecoder(rec.Body).Decode(&got)

// 		expect := Incident{
// 			ID:       "INC-1",
// 			Title:    "order-service request drop",
// 			Service:  "order-service",
// 			Severity: "SEV1",
// 			OpenedBy: "Anh Nguyen",
// 		}

// 		if got.ID != expect.ID {
// 			t.Errorf("ID expect %s got %s", expect.ID, got.ID)
// 		}
// 		if got.Title != expect.Title {
// 			t.Errorf("Title expect %s got %s", expect.Title, got.Title)
// 		}
// 		if got.Service != expect.Service {
// 			t.Errorf("Service expect %s got %s", expect.Service, got.Service)
// 		}
// 		if got.Severity != expect.Severity {
// 			t.Errorf("Severity expect %s got %s", expect.Severity, got.Severity)
// 		}
// 		if got.OpenedBy != expect.OpenedBy {
// 			t.Errorf("OpenedBy expect %s got %s", expect.OpenedBy, got.OpenedBy)
// 		}
// 	})
// 	tests := []struct {
// 		name         string
// 		body         string
// 		responseCode int
// 		errorCode    string
// 		message      string
// 	}{
// 		{"Missing fields_title", `{"title": "", "service": "order-service", "severity": "SEV1", "opened_by": "Anh Nguyen"}`, 400, MISSING_FIELD, ErrNoTitle.Error()},
// 		{"Missing fields_service", `{"title": "order-service request drop", "service": "", "severity": "SEV1", "opened_by": "Anh Nguyen"}`, 400, MISSING_FIELD, ErrNoService.Error()},
// 		{"Missing fields_severity4", `{"title": "order-service request drop", "service": "order-service", "severity": "SEV4", "opened_by": "Anh Nguyen"}`, 400, MISSING_FIELD, ErrInvalidSeverity.Error()},
// 		{"Missing fields_open_by", `{"title": "order-service request drop", "service": "order-service", "severity": "SEV1", "opened_by": ""}`, 400, MISSING_FIELD, ErrOpenedBy.Error()},
// 		{"Missing fields_on_call", `{"title": "order-service request drop", "service": "order-service", "severity": "SEV1", "opened_by": ""}`, 400, MISSING_FIELD, ErrOpenedBy.Error()},
// 		{"Missing fields_on_call", `{"title": "order-service request drop", "service": "order-service", "severity": "SEV1", "opened_by": "Anh Nguyen", "on_call": ""}`, 400, BAD_REQUEST, ErrOnCall.Error()},
// 	}
// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			req := httptest.NewRequest("POST", "/incidents", strings.NewReader(test.body))
// 			req.Header.Set("Content-Type", "application/json")
// 			rec := httptest.NewRecorder()
// 			router.ServeHTTP(rec, req)

// 			if rec.Code != test.responseCode {
// 				t.Errorf("Response Code expected §%d§, got %d", test.responseCode, rec.Code)
// 			}

// 			var got map[string]map[string]string
// 			json.NewDecoder(rec.Body).Decode(&got)

// 			if got["error"]["code"] != test.errorCode {
// 				t.Errorf("Error code expected §%s§, got §%s§", test.errorCode, got["error"]["code"])
// 			}
// 			if got["error"]["message"] != test.message {
// 				t.Errorf("Error message expected %s, got %s", test.message, got["error"]["message"])
// 			}
// 		})
// 	}
// }

// func TestGetIncident(t *testing.T) {
// 	config := loadConfig()
// 	client, store := NewStore(config)
// 	incHandler := IncidentHandler{Store: store, Registry: NewRegistry()}
// 	router := getRouter(&incHandler, client, prometheus.NewRegistry())
// 	go incHandler.Registry.run()
// 	defer close(incHandler.Registry.done)
// 	if instrumented, ok := store.(*InstrumentedStore); ok {
// 		if ms, ok := instrumented.s.(*MongoStore); ok {
// 			ms.DropAll(context.Background())
// 		}
// 	}

// 	rec1 := httptest.NewRecorder()
// 	body := `{"title": "order-service request drop", "service": "order-service", "severity": "SEV1", "opened_by": "Anh Nguyen"}`
// 	req1 := httptest.NewRequest("POST", "/incidents", strings.NewReader(body))
// 	req1.Header.Set("Content-Type", "application/json")
// 	router.ServeHTTP(rec1, req1)

// 	t.Run("Normal Get Incident", func(t *testing.T) {

// 		rec2 := httptest.NewRecorder()
// 		req2 := httptest.NewRequest("GET", "/incidents/INC-1", strings.NewReader(body))
// 		req2.Header.Set("Content-Type", "application/json")
// 		router.ServeHTTP(rec2, req2)

// 		if rec2.Code != 200 {
// 			t.Errorf("Code expected %d, got %d", 200, rec2.Code)
// 		}

// 		var got2 Incident
// 		json.NewDecoder(rec2.Body).Decode(&got2)

// 		expect := Incident{
// 			ID:       "INC-1",
// 			Title:    "order-service request drop",
// 			Service:  "order-service",
// 			Severity: "SEV1",
// 			OpenedBy: "Anh Nguyen",
// 		}

// 		if got2.ID != expect.ID {
// 			t.Errorf("ID expect %s got2 %s", expect.ID, got2.ID)
// 		}
// 		if got2.Title != expect.Title {
// 			t.Errorf("Title expect %s got2 %s", expect.Title, got2.Title)
// 		}
// 		if got2.Service != expect.Service {
// 			t.Errorf("Service expect %s got2 %s", expect.Service, got2.Service)
// 		}
// 		if got2.Severity != expect.Severity {
// 			t.Errorf("Severity expect %s got2 %s", expect.Severity, got2.Severity)
// 		}
// 		if got2.OpenedBy != expect.OpenedBy {
// 			t.Errorf("OpenedBy expect %s got2 %s", expect.OpenedBy, got2.OpenedBy)
// 		}
// 	})
// 	t.Run("Failed GetIncident", func(t *testing.T) {
// 		rec2 := httptest.NewRecorder()
// 		req2 := httptest.NewRequest("GET", "/incidents/INC-2", strings.NewReader(body))
// 		req2.Header.Set("Content-Type", "application/json")
// 		router.ServeHTTP(rec2, req2)

// 		if rec2.Code != 404 {
// 			t.Errorf("Code expected %d, got %d", 404, rec2.Code)
// 		}
// 	})
// }

// func TestListIncident(t *testing.T) {
// 	config := loadConfig()
// 	client, store := NewStore(config)
// 	incHandler := IncidentHandler{Store: store, Registry: NewRegistry()}
// 	router := getRouter(&incHandler, client, prometheus.NewRegistry())
// 	go incHandler.Registry.run()
// 	defer close(incHandler.Registry.done)
// 	if instrumented, ok := store.(*InstrumentedStore); ok {
// 		if ms, ok := instrumented.s.(*MongoStore); ok {
// 			ms.DropAll(context.Background())
// 		}
// 	}

// 	rec1 := httptest.NewRecorder()
// 	body := `{"title": "order-service request drop", "service": "order-service", "severity": "SEV1", "opened_by": "Anh Nguyen"}`
// 	req1 := httptest.NewRequest("POST", "/incidents", strings.NewReader(body))
// 	req1.Header.Set("Content-Type", "application/json")
// 	router.ServeHTTP(rec1, req1)

// 	rec9 := httptest.NewRecorder()
// 	body = `{"title": "title2", "service": "service2", "severity": "SEV2", "opened_by": "tom2"}`
// 	req9 := httptest.NewRequest("POST", "/incidents", strings.NewReader(body))
// 	req9.Header.Set("Content-Type", "application/json")
// 	router.ServeHTTP(rec9, req9)

// 	t.Run("Normal ListIncident", func(t *testing.T) {

// 		rec2 := httptest.NewRecorder()
// 		req2 := httptest.NewRequest("GET", "/incidents?status=active", nil)
// 		req2.Header.Set("Content-Type", "application/json")
// 		router.ServeHTTP(rec2, req2)

// 		if rec2.Code != 200 {
// 			t.Errorf("Code expected %d, got %d", 200, rec2.Code)
// 		}

// 		var got []Incident
// 		json.NewDecoder(rec2.Body).Decode(&got)

// 		if len(got) != 2 {
// 			t.Errorf("len expect %d got %d", 2, len(got))
// 		}
// 	})
// 	t.Run("Failed ListIncident", func(t *testing.T) {

// 		rec2 := httptest.NewRecorder()
// 		req2 := httptest.NewRequest("GET", "/incidents?status=activ1e&service=order-service", nil)
// 		req2.Header.Set("Content-Type", "application/json")
// 		router.ServeHTTP(rec2, req2)

// 		if rec2.Code != 400 {
// 			t.Errorf("Code expected %d, got %d", 400, rec2.Code)
// 		}

// 		var got []Incident
// 		json.NewDecoder(rec2.Body).Decode(&got)

// 		if rec2.Code != 400 {
// 			t.Errorf("Code expected %d, got %d", 400, rec2.Code)
// 		}
// 	})
// }

// func TestUpdateIncident(t *testing.T) {
// 	config := loadConfig()
// 	client, store := NewStore(config)
// 	incHandler := IncidentHandler{Store: store, Registry: NewRegistry()}
// 	router := getRouter(&incHandler, client, prometheus.NewRegistry())
// 	go incHandler.Registry.run()
// 	defer close(incHandler.Registry.done)
// 	if instrumented, ok := store.(*InstrumentedStore); ok {
// 		if ms, ok := instrumented.s.(*MongoStore); ok {
// 			ms.DropAll(context.Background())
// 		}
// 	}

// 	rec1 := httptest.NewRecorder()
// 	body1 := `{"title": "order-service request drop", "service": "order-service", "severity": "SEV1", "opened_by": "Anh Nguyen"}`
// 	req1 := httptest.NewRequest("POST", "/incidents", strings.NewReader(body1))
// 	req1.Header.Set("Content-Type", "application/json")
// 	router.ServeHTTP(rec1, req1)

// 	t.Run("Normal UpdateIncident ", func(t *testing.T) {
// 		rec2 := httptest.NewRecorder()
// 		body := `{"status":"resolved"}`
// 		req2 := httptest.NewRequest("PATCH", "/incidents/INC-1", strings.NewReader(body))
// 		req2.Header.Set("Content-Type", "application/json")
// 		router.ServeHTTP(rec2, req2)

// 		if rec2.Code != 204 {
// 			t.Errorf("Code expected %d, got %d", 204, rec2.Code)
// 		}
// 	})

// 	t.Run("fail UpdateIncident ", func(t *testing.T) {
// 		rec2 := httptest.NewRecorder()
// 		body := `{"status":"resolve"}`
// 		req2 := httptest.NewRequest("PATCH", "/incidents/INC-1", strings.NewReader(body))
// 		req2.Header.Set("Content-Type", "application/json")
// 		router.ServeHTTP(rec2, req2)

// 		if rec2.Code != 400 {
// 			t.Errorf("Code expected %d, got %d", 400, rec2.Code)
// 		}
// 	})
// }

// func TestAddTimelineEntry(t *testing.T) {
// 	config := loadConfig()
// 	client, store := NewStore(config)
// 	incHandler := IncidentHandler{Store: store, Registry: NewRegistry()}
// 	router := getRouter(&incHandler, client, prometheus.NewRegistry())
// 	go incHandler.Registry.run()
// 	defer close(incHandler.Registry.done)
// 	if instrumented, ok := store.(*InstrumentedStore); ok {
// 		if ms, ok := instrumented.s.(*MongoStore); ok {
// 			ms.DropAll(context.Background())
// 		}
// 	}

// 	rec1 := httptest.NewRecorder()
// 	body1 := `{"title": "order-service request drop", "service": "order-service", "severity": "SEV1", "opened_by": "Anh Nguyen"}`
// 	req1 := httptest.NewRequest("POST", "/incidents", strings.NewReader(body1))
// 	req1.Header.Set("Content-Type", "application/json")
// 	router.ServeHTTP(rec1, req1)

// 	t.Run("Normal addEntry ", func(t *testing.T) {
// 		rec2 := httptest.NewRecorder()
// 		body := `{"author":"Anh Nguyen","type":"observation","text":"Connection pool exhaustion. Pool at 100/100."}`
// 		req2 := httptest.NewRequest("POST", "/incidents/INC-1/entries", strings.NewReader(body))
// 		req2.Header.Set("Content-Type", "application/json")
// 		router.ServeHTTP(rec2, req2)

// 		if rec2.Code != 201 {
// 			t.Errorf("Code expected %d, got %d", 201, rec2.Code)
// 		}

// 		var got TimelineEntry
// 		json.NewDecoder(rec2.Body).Decode(&got)

// 		expect := TimelineEntry{
// 			Author: "Anh Nguyen",
// 			Type:   "observation",
// 			Text:   "Connection pool exhaustion. Pool at 100/100.",
// 		}

// 		if got.Author != expect.Author {
// 			t.Errorf("Author got %s, want %s", got.Author, expect.Author)
// 		}
// 		if got.Type != expect.Type {
// 			t.Errorf("Type got %s, want %s", got.Type, expect.Type)
// 		}
// 		if got.Text != expect.Text {
// 			t.Errorf("Text got %s, want %s", got.Text, expect.Text)
// 		}
// 	})
// 	t.Run("Fail addEntry Missing field", func(t *testing.T) {
// 		rec2 := httptest.NewRecorder()
// 		body := `{"author":"","type":"observation","text":"Connection pool exhaustion. Pool at 100/100."}`
// 		req2 := httptest.NewRequest("POST", "/incidents/INC-1/entries", strings.NewReader(body))
// 		req2.Header.Set("Content-Type", "application/json")
// 		router.ServeHTTP(rec2, req2)

// 		if rec2.Code != 400 {
// 			t.Errorf("Code expected %d, got %d", 400, rec2.Code)
// 		}
// 	})
// 	t.Run("Fail addEntry bad entry type", func(t *testing.T) {
// 		rec2 := httptest.NewRecorder()
// 		body := `{"author":"Anhh","type":"observa","text":"Connection pool exhaustion. Pool at 100/100."}`
// 		req2 := httptest.NewRequest("POST", "/incidents/INC-1/entries", strings.NewReader(body))
// 		req2.Header.Set("Content-Type", "application/json")
// 		router.ServeHTTP(rec2, req2)

// 		if rec2.Code != 400 {
// 			t.Errorf("Code expected %d, got %d", 400, rec2.Code)
// 		}
// 	})
// 	t.Run("Fail addEntry non-existent incident", func(t *testing.T) {
// 		rec2 := httptest.NewRecorder()
// 		body := `{"author":"Anhh","type":"observation","text":"Connection pool exhaustion. Pool at 100/100."}`
// 		req2 := httptest.NewRequest("POST", "/incidents/INC-2/entries", strings.NewReader(body))
// 		req2.Header.Set("Content-Type", "application/json")
// 		router.ServeHTTP(rec2, req2)

// 		if rec2.Code != 404 {
// 			t.Errorf("Code expected %d, got %d", 404, rec2.Code)
// 		}
// 	})
// 	t.Run("Fail addEntry conflict", func(t *testing.T) {
// 		rec2 := httptest.NewRecorder()
// 		body := `{"status":"resolved"}`
// 		req2 := httptest.NewRequest("PATCH", "/incidents/INC-1", strings.NewReader(body))
// 		req2.Header.Set("Content-Type", "application/json")
// 		router.ServeHTTP(rec2, req2)

// 		if rec2.Code != 204 {
// 			t.Errorf("Code expected %d, got %d", 204, rec2.Code)
// 		}

// 		rec3 := httptest.NewRecorder()
// 		body3 := `{"author":"abc","type":"observation","text":"Connection pool exhaustion. Pool at 100/100."}`
// 		req3 := httptest.NewRequest("POST", "/incidents/INC-1/entries", strings.NewReader(body3))
// 		req3.Header.Set("Content-Type", "application/json")
// 		router.ServeHTTP(rec3, req3)

// 		if rec3.Code != 409 {
// 			t.Errorf("Code expected %d, got %d", 409, rec3.Code)
// 		}

// 	})
// }

// func TestHandleIncidentWebSocket(t *testing.T) {
// 	config := loadConfig()
// 	client, store := NewStore(config)
// 	incHandler := IncidentHandler{Store: store, Registry: NewRegistry()}
// 	router := getRouter(&incHandler, client, prometheus.NewRegistry())
// 	go incHandler.Registry.run()
// 	defer close(incHandler.Registry.done)
// 	if instrumented, ok := store.(*InstrumentedStore); ok {
// 		if ms, ok := instrumented.s.(*MongoStore); ok {
// 			ms.DropAll(context.Background())
// 		}
// 	}

// 	srv := httptest.NewServer(router)
// 	defer srv.Close()

// 	_, err := http.Post(
// 		srv.URL+"/incidents",
// 		"application/json",
// 		strings.NewReader(`{"title": "order-service request drop", "service": "order-service", "severity": "SEV1", "opened_by": "Anh Nguyen"}`))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/incidents/INC-1/ws"
// 	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer conn.Close()

// 	t.Run("Websocket Normal addEntry ", func(t *testing.T) {
// 		_, err := http.Post(
// 			srv.URL+"/incidents/INC-1/entries",
// 			"application/json",
// 			strings.NewReader(`{"author":"Anh Nguyen","type":"observation","text":"Connection pool exhaustion. Pool at 100/100."}`))
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		_, msg, err := conn.ReadMessage()
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		var got map[string]any
// 		json.Unmarshal(msg, &got)

// 		if got["type"] != "new_entry" {
// 			t.Errorf("type: got %s", got["type"])
// 		}
// 		if got["incident_id"] != "INC-1" {
// 			t.Errorf("incident_id: got %s", got["incident_id"])
// 		}

// 		entry := got["entry"].(map[string]interface{})
// 		if entry["author"] != "Anh Nguyen" {
// 			t.Errorf("author: got %s", entry["author"])
// 		}
// 		if entry["type"] != "observation" {
// 			t.Errorf("type: got %s", entry["type"])
// 		}
// 	})
// 	t.Run("Websocket UpdateIncident ", func(t *testing.T) {
// 		req, err := http.NewRequest("PATCH",
// 			srv.URL+"/incidents/INC-1",
// 			strings.NewReader(`{"status":"resolved"}`))
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		req.Header.Set("Content-Type", "application/json")

// 		resp, err := http.DefaultClient.Do(req)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		defer resp.Body.Close()

// 		_, msg, err := conn.ReadMessage()
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		if string(msg) != `{"type":"incident_updated","incident_id":"INC-1","field":"status","old_value":"triggered","new_value":"resolved"}` {
// 			t.Errorf("wrong msg")
// 		}

// 	})
// }
