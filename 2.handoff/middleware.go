package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func CORSMiddleware(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Method", "GET, POST, PATCH, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-type, Authorization")
		nextHandler.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		panicked := true
		defer func() {
			if panicked {
				slog.Error("request panicked",
					"method", r.Method,
					"path", r.URL.Path,
					"duration", fmt.Sprintf("%dms", time.Since(start).Milliseconds()),
					requestIDKey, r.Context().Value(requestIDKey).(string))
			}
		}()
		nextHandler.ServeHTTP(w, r)
		slog.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", fmt.Sprintf("%dms", time.Since(start).Milliseconds()),
			requestIDKey, r.Context().Value(requestIDKey),
		)
		panicked = false
	})
}

func RequestIDMiddleware(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newRequestID := uuid.New().String()
		w.Header().Add("X-Request-ID", newRequestID)
		ctxWithNewRequestID := context.WithValue(r.Context(), requestIDKey, newRequestID)
		nextHandler.ServeHTTP(w, r.WithContext(ctxWithNewRequestID))
	})
}
