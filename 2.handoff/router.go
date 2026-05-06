package main

import "net/http"

func getRouter(incHandler IncidentHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /incidents", incHandler.CreateIncident)
	mux.HandleFunc("POST /incidents/{id}/entries", incHandler.AddEntry)
	mux.HandleFunc("GET /incidents/{id}", incHandler.GetIncident)
	// mux.HandleFunc("GET /incidents", listAllIncidents)
	// mux.HandleFunc("GET /incidents/{id}/handoff", listHandoff)
	// mux.HandleFunc("GET /healthz", healthCheck)
	// mux.HandleFunc("PATCH /incidents", UpdateOneIncident)
	router := RequestIDMiddleware(LoggingMiddleware(CORSMiddleware(mux)))
	return router
}
