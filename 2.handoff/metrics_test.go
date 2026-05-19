package main

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"io"
// 	"net/http"
// 	"net/http/httptest"
// 	"strings"
// 	"testing"

// 	"github.com/gorilla/websocket"
// 	"github.com/prometheus/client_golang/prometheus"
// )

// func createWsConns(srv *httptest.Server, cnt int) ([]*websocket.Conn, error) {
// 	wsURLs := []string{}
// 	conns := []*websocket.Conn{}
// 	for i := 0; i < cnt; i++ {
// 		wsURLs = append(wsURLs, "ws"+strings.TrimPrefix(srv.URL, "http")+"/incidents/INC-1/ws")
// 		conn, _, err := websocket.DefaultDialer.Dial(wsURLs[i], nil)
// 		if err != nil {
// 			for _, conn := range conns {
// 				conn.Close()
// 			}
// 			return nil, err
// 		}
// 		conns = append(conns, conn)
// 	}
// 	return conns, nil
// }

// func CreateIncident(srv *httptest.Server, cnt int) error {
// 	body, err := json.Marshal(CreateIncidentRequest{
// 		Title:    "order-service request drop",
// 		Service:  "order-service",
// 		Severity: "SEV1",
// 		OpenedBy: "Anh Nguyen",
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	for i := 0; i < cnt; i++ {
// 		bodyIOReader := bytes.NewReader(body)
// 		res, err := http.Post(
// 			srv.URL+"/incidents",
// 			"application/json",
// 			bodyIOReader)
// 		if err != nil {
// 			return err
// 		}
// 		if res.StatusCode == 400 {
// 			_, err := io.ReadAll(res.Body)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 		res.Body.Close()
// 	}
// 	return nil
// }

// func metricReinit() {
// 	httpRequestsTotal = prometheus.NewCounterVec(
// 		prometheus.CounterOpts{
// 			Name: "handoff_http_requests_total",
// 			Help: "	Total requests",
// 		},
// 		[]string{"method", "path", "status_code"},
// 	)

// 	httpDurationSeconds = prometheus.NewHistogramVec(
// 		prometheus.HistogramOpts{
// 			Name: "handoff_http_request_duration_seconds",
// 			Help: "Request latency distribution",
// 			// Buckets: prometheus.DefBuckets,
// 			Buckets: []float64{.03, .1},
// 		},
// 		[]string{"method", "path"},
// 	)

// 	incidentTotal = prometheus.NewGaugeVec(
// 		prometheus.GaugeOpts{
// 			Name: "handoff_incidents_total",
// 			Help: "Current number of incidents",
// 		},
// 		[]string{"status"},
// 	)

// 	totalEntries = prometheus.NewCounter(
// 		prometheus.CounterOpts{
// 			Name: "handoff_entries_total",
// 			Help: "Total timeline entries created",
// 		},
// 	)

// 	dbQueryDurationSeconds = prometheus.NewHistogramVec(
// 		prometheus.HistogramOpts{
// 			Name:    "handoff_db_query_duration_seconds",
// 			Help:    "Database query latency",
// 			Buckets: []float64{.03, .1},
// 		},
// 		[]string{"operation"},
// 	)

// 	wsConnections = prometheus.NewGauge(
// 		prometheus.GaugeOpts{
// 			Name: "handoff_websocket_connections",
// 			Help: "Current number of active WebSocket connections",
// 		},
// 	)
// }

// func TestMetric(t *testing.T) {
// 	// Reset metrics
// 	metricReinit()

// 	// Normal initializtion
// 	config := loadConfig()
// 	client, store := NewStore(config)
// 	registry := NewRegistry()
// 	promRegistry := prometheus.NewRegistry()
// 	NewMetrics(promRegistry)

// 	incHandler := IncidentHandler{Store: store, Registry: registry}
// 	router := getRouter(&incHandler, client, promRegistry)
// 	go incHandler.Registry.run()
// 	defer close(incHandler.Registry.done)
// 	if instrumented, ok := store.(*InstrumentedStore); ok {
// 		if ms, ok := instrumented.s.(*MongoStore); ok {
// 			ms.DropAll(context.Background())
// 		}
// 	}

// 	// Start server and init
// 	srv := httptest.NewServer(router)
// 	defer srv.Close()
// 	err := CreateIncident(srv, 5)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	conns, err := createWsConns(srv, 5)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	conns[0].Close()
// 	for idx, conn := range conns {
// 		if idx > 0 {
// 			defer conn.Close()
// 		}
// 	}

// 	req, err := http.NewRequest("PATCH",
// 		srv.URL+"/incidents/INC-1",
// 		strings.NewReader(`{"status":"resolved"}`))
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	res, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if res.StatusCode != 204 {
// 		t.Fatalf("status code expected 204, got %d", res.StatusCode)
// 	}

// 	res, err = http.Post(
// 		srv.URL+"/incidents/INC-5/entries",
// 		"application/json",
// 		strings.NewReader(`{"author":"Anh Nguyen","type":"observation","text":"Connection pool exhaustion. Pool at 100/100."}`))
// 	if res.StatusCode != 201 {
// 		t.Fatalf("res.StatusCode expect 201, got %d", res.StatusCode)
// 	}

// 	// Check result
// 	res, err = http.Get(srv.URL + "/metrics")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer res.Body.Close()
// 	body, err := io.ReadAll(res.Body)
// 	bodyString := string(body)
// 	if !strings.Contains(bodyString, `handoff_http_requests_total{method="POST",path="/incidents",status_code="201"} 5`) {
// 		t.Error(`metric [handoff_http_requests_total{method="POST",path="/incidents",status_code="201"}] not correct`)
// 	}
// 	if !strings.Contains(bodyString, `handoff_incidents_total{status="resolved"} 1`) {
// 		t.Error(`metric [handoff_incidents_total{status="resolved"}] not correct`)
// 	}
// 	if !strings.Contains(bodyString, `handoff_incidents_total{status="triggered"} 4`) {
// 		t.Error(`metric [handoff_incidents_total{status="triggered"}] not correct`)
// 	}
// 	if !strings.Contains(bodyString, `handoff_entries_total 1`) {
// 		t.Error(`metric [handoff_entries_total] not correct`)
// 	}
// 	if !strings.Contains(bodyString, `handoff_websocket_connections 4`) {
// 		t.Error(`metric [handoff_websocket_connections] not correct`)
// 	}
// }
