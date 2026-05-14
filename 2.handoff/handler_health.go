package main

import (
	"context"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func healthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, r.Context().Value(requestIDKey).(string), map[string]string{"status": "ok"})
}

func readyCheck(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if client == nil {
			writeJSON(w, 200, r.Context().Value(requestIDKey).(string), map[string]string{"status": "ready"})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		err := client.Ping(ctx, nil)
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, ErrorMessageJSON{
				ErrorCode: "DATABASE_UNAVAILABLE",
				Message:   err.Error(),
				RequestID: r.Context().Value(requestIDKey).(string),
			})
		}
		writeJSON(w, 200, r.Context().Value(requestIDKey).(string), map[string]string{"status": "ready"})
	}
}
