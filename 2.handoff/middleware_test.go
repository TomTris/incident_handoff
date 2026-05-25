package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDMiddleware(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Context().Value(requestIDKey)
		if id == nil {
			t.Fatal("no requestID in context")
		}
		w.WriteHeader(http.StatusOK)
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	RequestIDMiddleware(inner).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status expected %v, got %v", http.StatusOK, rec.Code)
	}
	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatalf("Header expected %v, got %v", "X-Request-ID", "empty")
	}
}

func TestCORSMiddleware(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	CORSMiddleware(inner).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status expected %v, got %v", http.StatusOK, rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("Header expected %v, got %v", "*", rec.Header().Get("Access-Control-Allow-Origin"))
	}
	if rec.Header().Get("Access-Control-Allow-Method") != "GET, POST, PATCH, DELETE" {
		t.Fatalf("Header expected %v, got %v", "GET, POST, PATCH, DELETE", rec.Header().Get("Access-Control-Allow-Method"))
	}
	if rec.Header().Get("Access-Control-Allow-Headers") != "Content-type, Authorization" {
		t.Fatalf("Header expected %v, got %v", "Content-type, Authorization", rec.Header().Get("Access-Control-Allow-Headers"))
	}
}

func TestTimeoutMiddleware(t *testing.T) {
	var gotDeadline bool
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, gotDeadline = r.Context().Deadline()
		w.WriteHeader(200)
	})

	handler := TimeoutMiddleware(inner)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(rec, req)

	if !gotDeadline {
		t.Error("expected context to have a deadline")
	}
}

func TestResponseMiddleware(t *testing.T) {
	testRequestID := "Test-Request-ID"
	t.Run("Success", func(t *testing.T) {
		inner := func(r *http.Request) (*AppResponse, error) {
			return newAppResponse(http.StatusOK, Incident{Title: "Title"}), nil
		}
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		rec.Header().Add("X-Request-ID", testRequestID)
		ctxWithNewRequestID := context.WithValue(req.Context(), requestIDKey, testRequestID)
		ResponseMiddleware(inner).ServeHTTP(rec, req.WithContext(ctxWithNewRequestID))

		if rec.Code != http.StatusOK {
			t.Fatalf("expected code %v, get %v", http.StatusOK, rec.Code)
		}
		var body map[string]any
		json.Unmarshal(rec.Body.Bytes(), &body)

		if body["title"] != "Title" {
			t.Fatalf("expected Title %v, get %v", "Title", body["title"])
		}
	})
	t.Run("Success Nil-Body", func(t *testing.T) {
		inner := func(r *http.Request) (*AppResponse, error) {
			return newAppResponse(http.StatusNoContent, nil), nil
		}
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		rec.Header().Add("X-Request-ID", testRequestID)
		ctxWithNewRequestID := context.WithValue(req.Context(), requestIDKey, testRequestID)
		ResponseMiddleware(inner).ServeHTTP(rec, req.WithContext(ctxWithNewRequestID))

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected code %v, get %v", http.StatusNoContent, rec.Code)
		}
	})

	t.Run("error with AppError", func(t *testing.T) {
		inner := func(r *http.Request) (*AppResponse, error) {
			return nil, BadRequest(errors.New("test-Error"))
		}
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		rec.Header().Add("X-Request-ID", testRequestID)
		ctxWithNewRequestID := context.WithValue(req.Context(), requestIDKey, testRequestID)
		ResponseMiddleware(inner).ServeHTTP(rec, req.WithContext(ctxWithNewRequestID))

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected code %v, get %v", http.StatusBadRequest, rec.Code)
		}
		var body map[string](map[string]any)
		json.Unmarshal(rec.Body.Bytes(), &body)
		if body["error"]["code"] != BAD_REQUEST {
			t.Fatalf("expected Code %v, get %v", BAD_REQUEST, body["error"]["code"])
		}
		if body["error"]["message"] != "test-Error" {
			t.Fatalf("expected Code %v, get %v", "test-Error", body["error"]["message"])
		}
	})

	t.Run("error with Unknown Error", func(t *testing.T) {
		inner := func(r *http.Request) (*AppResponse, error) {
			return nil, errors.New("Unknown Error")
		}
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		rec.Header().Add("X-Request-ID", testRequestID)
		ctxWithNewRequestID := context.WithValue(req.Context(), requestIDKey, testRequestID)
		ResponseMiddleware(inner).ServeHTTP(rec, req.WithContext(ctxWithNewRequestID))

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected code %v, get %v", http.StatusInternalServerError, rec.Code)
		}
	})
}
