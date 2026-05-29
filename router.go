package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func getRouter(
	incHandler *IncidentHandler,
	flagHandler *FlagHandler,
	authHandler *AuthHandler,
	onCallHandler *OnCallHandler,

	mongoClient *mongo.Client,
	promRegistry *prometheus.Registry,
	httpMetrics *HTTPMetrics) http.Handler {

	// Incident handler
	protected := http.NewServeMux()
	protected.HandleFunc("POST /incidents", ResponseMiddleware(incHandler.CreateIncident))
	protected.HandleFunc("POST /incidents/{id}/entries", ResponseMiddleware(incHandler.AddEntry))
	protected.HandleFunc("GET /incidents/{id}", ResponseMiddleware(incHandler.GetIncident))
	protected.HandleFunc("GET /incidents", ResponseMiddleware(incHandler.ListIncidents))
	protected.HandleFunc("GET /incidents/{id}/handoff", ResponseMiddleware(incHandler.GetHandoffBrief))
	protected.HandleFunc("PATCH /incidents/{id}", ResponseMiddleware(incHandler.UpdateIncident))
	// auth
	protected.HandleFunc("GET/auth/me", ResponseMiddleware(authHandler.WhoAmI))
	// WebsocketHandler
	protected.HandleFunc("GET /incidents/{id}/ws", incHandler.HandleIncidentWebSocket)

	// Flag Handler
	admin := http.NewServeMux()
	admin.HandleFunc("POST /flags", ResponseMiddleware(AuthAdminOnlyMiddleware(flagHandler.CreateFlag)))
	admin.HandleFunc("GET /flags", ResponseMiddleware(AuthAdminOnlyMiddleware(flagHandler.ListAllFlag)))
	admin.HandleFunc("PATCH /flags/{name}", ResponseMiddleware(AuthAdminOnlyMiddleware(flagHandler.UpdateFlag)))
	admin.HandleFunc("GET /flags/{name}/evaluate", ResponseMiddleware(AuthAdminOnlyMiddleware(flagHandler.Evaluate)))
	admin.HandleFunc("POST /oncall", ResponseMiddleware(AuthAdminOnlyMiddleware(onCallHandler.CreateShift)))
	admin.HandleFunc("GET /oncall/current", ResponseMiddleware(AuthAdminOnlyMiddleware(onCallHandler.CurrentOnCall)))

	// metrics, health and ready
	public := http.NewServeMux()
	public.Handle("GET /metrics", promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{Registry: promRegistry}))
	public.HandleFunc("GET /healthz", healthCheck)
	public.HandleFunc("GET /readyz", readyCheck(mongoClient))
	public.HandleFunc("POST /login", authHandler.LoginHandler)

	root := http.NewServeMux()
	authMW := AuthMiddleware(authHandler.Secret)
	root.Handle("/api/", http.StripPrefix("/api", authMW(protected)))
	root.Handle("/admin/", http.StripPrefix("/admin", authMW(admin)))
	root.Handle("/", public)

	router := RequestIDMiddleware(ObservabilityMiddleware(httpMetrics)(CORSMiddleware(TimeoutMiddleware(root))))
	return router
}
