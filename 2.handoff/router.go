package main

import (
	"net/http"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func getRouter(incHandler *IncidentHandler, mongoClient *mongo.Client) http.Handler {
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
	router := RequestIDMiddleware(LoggingMiddleware(CORSMiddleware(TimeoutMiddleware(mux))))
	return router
}
