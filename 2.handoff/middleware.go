package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type statusRecorder struct {
	http.ResponseWriter
	StatusCode int
	Hijacked   bool
}

func (r *statusRecorder) WriteHeader(code int) {
	r.StatusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// https://pkg.go.dev/net/http#FS, Crtl+F "ResponseWriter wrapper" -> Hijack && Flush
// If in the future, this wrapper gets more complicated, recommend to use github.com/felixge/httpsnoop
func (r *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	r.Hijacked = true
	return h.Hijack()
}

func (r *statusRecorder) Flush() {
	f, ok := r.ResponseWriter.(http.Flusher)
	if ok {
		f.Flush()
	}
}

func TimeoutMiddleware(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		nextHandler.ServeHTTP(w, r.WithContext(ctx))
	})
}

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
		wrappedWriter := &statusRecorder{ResponseWriter: w, StatusCode: 200}
		defer func() {
			duration := fmt.Sprintf("%dms", time.Since(start).Milliseconds())
			requestID := r.Context().Value(requestIDKey).(string)
			err := recover()
			if err != nil {
				slog.Error("request panicked",
					"method", r.Method,
					"path", r.URL.Path,
					"err", err,
					"duration", duration,
					requestIDKey, requestID,
				)
				writeError(w, 500, ErrorMessageJSON{
					ErrorCode: INTERNAL_SERVER_ERROR,
					Message:   "Server panicked",
					RequestID: requestID,
				})
				return
			}
			if wrappedWriter.Hijacked {
				slog.Info("websocket connection",
					"method", r.Method,
					"path", r.URL.Path,
					"duration", duration,
					requestIDKey, requestID,
				)
				return
			}
			slog.Info("request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrappedWriter.StatusCode,
				"duration", duration,
				requestIDKey, requestID,
			)
		}()
		nextHandler.ServeHTTP(wrappedWriter, r)
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
