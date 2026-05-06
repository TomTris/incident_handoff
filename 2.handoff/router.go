package main

import "net/http"

func getRouter(incHandler IncidentHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /incidents", incHandler.CreateIncident)
	mux.HandleFunc("POST /incidents/{id}/entries", incHandler.AddEntry)
	mux.HandleFunc("GET /incidents/{id}", incHandler.GetIncident)
	mux.HandleFunc("GET /incidents", incHandler.ListIncidents)
	mux.HandleFunc("GET /incidents/{id}/handoff", incHandler.GetHandoffBrief)
	mux.HandleFunc("GET /healthz", healthCheck)
	mux.HandleFunc("PATCH /incidents/{id}", incHandler.UpdateIncident)
	router := RequestIDMiddleware(LoggingMiddleware(CORSMiddleware(mux)))
	return router
}
