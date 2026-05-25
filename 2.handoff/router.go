package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func getRouter(incHandler *IncidentHandler, flagHander *FlagHandler,
	mongoClient *mongo.Client, promRegistry *prometheus.Registry, httpMatrics *HTTPMetrics) http.Handler {
	mux := http.NewServeMux()

	// Incident handler
	mux.HandleFunc("POST /incidents", ResponseMiddleware(incHandler.CreateIncident))
	mux.HandleFunc("POST /incidents/{id}/entries", ResponseMiddleware(incHandler.AddEntry))
	mux.HandleFunc("GET /incidents/{id}", ResponseMiddleware(incHandler.GetIncident))
	mux.HandleFunc("GET /incidents", ResponseMiddleware(incHandler.ListIncidents))
	mux.HandleFunc("GET /incidents/{id}/handoff", ResponseMiddleware(incHandler.GetHandoffBrief))
	mux.HandleFunc("PATCH /incidents/{id}", ResponseMiddleware(incHandler.UpdateIncident))
	// WebsocketHandler
	mux.HandleFunc("GET /incidents/{id}/ws", incHandler.HandleIncidentWebSocket)

	// Flag Handler
	mux.HandleFunc("POST /flags", ResponseMiddleware(flagHander.CreateFlag))
	mux.HandleFunc("GET /flags", ResponseMiddleware(flagHander.ListAllFlag))
	mux.HandleFunc("PATCH /flags/{name}", ResponseMiddleware(flagHander.UpdateFlag))
	mux.HandleFunc("GET /flags/{name}/evaluate", ResponseMiddleware(flagHander.Evaluate))

	// metrics, health and ready
	mux.Handle("/metrics", promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{Registry: promRegistry}))
	// mux.HandleFunc("GET /healthz", healthCheck)
	// mux.HandleFunc("GET /readyz", readyCheck(mongoClient))

	router := RequestIDMiddleware(ObservabilityMiddleware(httpMatrics)(CORSMiddleware(TimeoutMiddleware(mux))))
	return router
}
