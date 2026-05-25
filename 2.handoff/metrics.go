package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

type HTTPMetrics struct {
	HTTPRequestTotal    *prometheus.CounterVec
	HTTPDurationSeconds *prometheus.HistogramVec
}

type IncidentStoreMetrics struct {
	IncidentTotal          *prometheus.GaugeVec
	TotalEntries           prometheus.Counter
	DbQueryDurationSeconds *prometheus.HistogramVec
}

type RegistryMetric struct {
	wsConnections prometheus.Gauge
}

func NewHttpMetrics(promReg *prometheus.Registry) *HTTPMetrics {
	m := HTTPMetrics{}
	m.HTTPRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "handoff_http_requests_total",
			Help: "	Total requests",
		},
		[]string{"method", "path", "status_code"},
	)

	m.HTTPDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "handoff_http_request_duration_seconds",
			Help: "Request latency distribution",
			// Buckets: prometheus.DefBuckets,
			Buckets: []float64{.03, .1},
		},
		[]string{"method", "path"},
	)
	promReg.MustRegister(m.HTTPRequestTotal, m.HTTPDurationSeconds)
	return &m
}

func NewIncidentStoreMetric(promReg *prometheus.Registry) *IncidentStoreMetrics {
	m := IncidentStoreMetrics{}
	m.IncidentTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "handoff_incidents_total",
			Help: "Current number of incidents",
		},
		[]string{"status"},
	)

	m.TotalEntries = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "handoff_entries_total",
			Help: "Total timeline entries created",
		},
	)

	m.DbQueryDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "handoff_db_query_duration_seconds",
			Help:    "Database query latency",
			Buckets: []float64{.03, .1},
		},
		[]string{"operation"},
	)
	promReg.MustRegister(m.IncidentTotal, m.TotalEntries, m.DbQueryDurationSeconds)
	return &m
}

func NewRegistryMetric(promReg *prometheus.Registry) *RegistryMetric {
	m := RegistryMetric{}
	m.wsConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "handoff_websocket_connections",
			Help: "Current number of active WebSocket connections",
		},
	)
	promReg.MustRegister(m.wsConnections)
	return &m
}
