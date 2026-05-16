package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func getRouter(incHandler *IncidentHandler, mongoClient *mongo.Client, promRegistry *prometheus.Registry) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /incidents", incHandler.CreateIncident)
	mux.HandleFunc("POST /incidents/{id}/entries", incHandler.AddEntry)
	mux.HandleFunc("GET /incidents/{id}", incHandler.GetIncident)
	mux.HandleFunc("GET /incidents", incHandler.ListIncidents)
	mux.HandleFunc("GET /incidents/{id}/handoff", incHandler.GetHandoffBrief)
	mux.HandleFunc("PATCH /incidents/{id}", incHandler.UpdateIncident)
	mux.HandleFunc("GET /incidents/{id}/ws", incHandler.HandleIncidentWebSocket)

	mux.HandleFunc("GET /healthz", healthCheck)
	mux.HandleFunc("GET /readyz", readyCheck(mongoClient))
	mux.Handle("/metrics", promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{Registry: promRegistry}))

	mux.HandleFunc("POST /flags", incHandler.CreateFlag)
	mux.HandleFunc("GET /flags", incHandler.ListAllFlag)
	mux.HandleFunc("PATCH /flags/{name}", incHandler.UpdateFlag)
	mux.HandleFunc("GET /flags/{name}/evaluate", incHandler.Evaluate)
	router := RequestIDMiddleware(ObservabilityMiddleware(CORSMiddleware(TimeoutMiddleware(mux))))
	return router
}
