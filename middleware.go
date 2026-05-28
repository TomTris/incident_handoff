package main

import (
	"bufio"
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

func ObservabilityMiddleware(httpMetrics *HTTPMetrics) func(http.Handler) http.Handler {
	return func(nextHandler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrappedWriter := &statusRecorder{ResponseWriter: w, StatusCode: 200}

			defer func() {
				duration := time.Since(start)
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
					httpMetrics.HTTPRequestTotal.WithLabelValues(r.Method, r.URL.Path, "500").Inc()
					httpMetrics.HTTPDurationSeconds.WithLabelValues(r.Method, r.URL.Path).Observe(duration.Seconds())
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

				httpMetrics.HTTPRequestTotal.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(wrappedWriter.StatusCode)).Inc()
				httpMetrics.HTTPDurationSeconds.WithLabelValues(r.Method, r.URL.Path).Observe(duration.Seconds())
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
}

func RequestIDMiddleware(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newRequestID := uuid.New().String()
		w.Header().Add("X-Request-ID", newRequestID)
		ctxWithNewRequestID := context.WithValue(r.Context(), requestIDKey, newRequestID)
		nextHandler.ServeHTTP(w, r.WithContext(ctxWithNewRequestID))
	})
}

func AuthMiddleware(JWT_SECRET []byte) func(http.Handler) http.Handler {
	return func(nextHandler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") == false {
				writeError(w, http.StatusUnauthorized, ErrorMessageJSON{
					ErrorCode: "NO_AUTHENTICATION_HEADER",
					Message:   "Bearer not found",
					RequestID: r.Context().Value(requestIDKey).(string),
				})
			}

			claims := CustomClaims{}
			tokenRaw := strings.TrimPrefix(auth, "Bearer ")
			token, err := jwt.ParseWithClaims(tokenRaw, &claims, func(t *jwt.Token) (any, error) {
				if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
					return nil, errors.New("unexpected signing method")
				}
				return JWT_SECRET, nil
			})

			if err != nil || token.Valid == false {
				msg := "jwt invalid"
				if err != nil {
					msg = err.Error()
				}

				writeError(w, http.StatusUnauthorized, ErrorMessageJSON{
					ErrorCode: "BAD_JWT_TOKEN",
					Message:   msg,
					RequestID: r.Context().Value(requestIDKey).(string),
				})
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, UserContext{
				ID:       claims.Subject,
				Username: claims.Username,
				Role:     claims.Role,
			})
			nextHandler.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ResponseMiddleware(next func(*http.Request) (*AppResponse, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Context().Value(requestIDKey).(string)
		res, err := next(r)
		if err != nil {
			var appErr *AppError
			if errors.As(err, &appErr) {
				writeError(w, appErr.Status, ErrorMessageJSON{
					ErrorCode: appErr.Code,
					Message:   appErr.Err.Error(),
					RequestID: requestID,
				})
			} else {
				writeError(w, http.StatusInternalServerError, ErrorMessageJSON{
					ErrorCode: "INTERNAL_SERVER_ERROR",
					Message:   "Error Type not detected",
					RequestID: requestID,
				})
			}
			return
		}
		writeJSON(w, res.Status, requestID, res.Body)
	}
}
