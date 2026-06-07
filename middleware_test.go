package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
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
		inner := func(r *http.Request) (*AppResponse, *AppError) {
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
		inner := func(r *http.Request) (*AppResponse, *AppError) {
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
		inner := func(r *http.Request) (*AppResponse, *AppError) {
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
		inner := func(r *http.Request) (*AppResponse, *AppError) {
			return nil, &AppError{
				Code:   "UNKNOWN",
				Status: 500,
				Err:    errors.New("Unknown Error")}
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

func TestObservabilityMiddleware(t *testing.T) {
	testRequestID := "Test-Request-ID"
	t.Run("Success: logging httpRequest and duration", func(t *testing.T) {
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(time.Duration(4 * time.Microsecond))
			w.WriteHeader(http.StatusOK)
		})
		prompReg := prometheus.NewRegistry()
		newHTTPMetrics := NewHttpMetrics(prompReg)
		rec := httptest.NewRecorder()
		rec.Header().Add("X-Request-ID", testRequestID)
		// Send request 1
		req1 := httptest.NewRequest("GET", "/", nil)
		ctxWithNewRequestID := context.WithValue(req1.Context(), requestIDKey, testRequestID)
		ObservabilityMiddleware(newHTTPMetrics)(inner).ServeHTTP(rec, req1.WithContext(ctxWithNewRequestID))

		// Send request 2
		req2 := httptest.NewRequest("GET", "/abc", nil)
		ctxWithNewRequestID = context.WithValue(req2.Context(), requestIDKey, testRequestID)
		ObservabilityMiddleware(newHTTPMetrics)(inner).ServeHTTP(rec, req2.WithContext(ctxWithNewRequestID))
		ObservabilityMiddleware(newHTTPMetrics)(inner).ServeHTTP(rec, req2.WithContext(ctxWithNewRequestID))
		ObservabilityMiddleware(newHTTPMetrics)(inner).ServeHTTP(rec, req2.WithContext(ctxWithNewRequestID))

		// Evaluate
		totalHTTPRequest := testutil.CollectAndCount(newHTTPMetrics.HTTPRequestTotal)
		if totalHTTPRequest != 2 {
			t.Fatalf("expected HTTPRequestTotal %v, %v", 2, totalHTTPRequest)
		}

		totaldbDurationQuerys := testutil.CollectAndCount(newHTTPMetrics.HTTPDurationSeconds)
		if totaldbDurationQuerys != 2 {
			t.Fatalf("expected HTTPRequestTotal %v, %v", 2, totaldbDurationQuerys)
		}
	})
	t.Run("Server Panicked", func(t *testing.T) {
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		})
		prompReg := prometheus.NewRegistry()
		newHTTPMetrics := NewHttpMetrics(prompReg)
		rec := httptest.NewRecorder()
		rec.Header().Add("X-Request-ID", testRequestID)
		// Send request 1
		req1 := httptest.NewRequest("GET", "/", nil)
		ctxWithNewRequestID := context.WithValue(req1.Context(), requestIDKey, testRequestID)
		ObservabilityMiddleware(newHTTPMetrics)(inner).ServeHTTP(rec, req1.WithContext(ctxWithNewRequestID))

		// Evaluate
		totalHTTPRequest := testutil.CollectAndCount(newHTTPMetrics.HTTPRequestTotal)
		if totalHTTPRequest != 1 {
			t.Fatalf("expected HTTPRequestTotal %v, %v", 1, totalHTTPRequest)
		}
		totaldbDurationQuerys := testutil.CollectAndCount(newHTTPMetrics.HTTPDurationSeconds)
		if totaldbDurationQuerys != 1 {
			t.Fatalf("expected HTTPRequestTotal %v, %v", 1, totaldbDurationQuerys)
		}
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expect code %v, get %v", http.StatusInternalServerError, rec.Code)
		}
	})
}

func TestAuthMiddleware(t *testing.T) {
	testRequestID := "Test-Request-ID"
	secret := "jwt-testing"
	now := time.Now()
	ttl := time.Duration(15 * time.Minute)
	user := User{
		ID:       "123",
		Username: "anh",
		Role:     "engineer",
	}
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("no access_token in Cookie", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		ctx := req.Context()
		ctx = context.WithValue(ctx, requestIDKey, testRequestID)
		AuthMiddleware([]byte(secret))(inner).ServeHTTP(rec, req.WithContext(ctx))

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expect error code %v, get %v", http.StatusUnauthorized, rec.Code)
		}
		var res map[string]map[string]string
		json.NewDecoder(rec.Body).Decode(&res)
		if res["error"]["code"] != "NO_AUTH_COOKIE" {
			t.Fatalf("expect error with code %v, get %v", "NO_AUTH_COOKIE", res["error"]["code"])
		}
	})

	t.Run("access_token in Cookie", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		ctx := req.Context()
		ctx = context.WithValue(ctx, requestIDKey, testRequestID)
		tokenSigned, _ := IssueToken(user, []byte(secret), ttl, now)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: tokenSigned})
		AuthMiddleware([]byte(secret))(inner).ServeHTTP(rec, req.WithContext(ctx))

		if rec.Code != http.StatusOK {
			var res map[string]map[string]string
			json.NewDecoder(rec.Body).Decode(&res)
			t.Errorf("expect error code %v, get %v, reason $%v$", http.StatusOK, rec.Code, res["error"]["code"])
		}
	})
	t.Run("fake JWT in access_token", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		ctx := req.Context()
		ctx = context.WithValue(ctx, requestIDKey, testRequestID)
		fakeTokenSigned := "fake-jwts"
		req.AddCookie(&http.Cookie{Name: "access_token", Value: fakeTokenSigned})
		AuthMiddleware([]byte(secret))(inner).ServeHTTP(rec, req.WithContext(ctx))

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expect error code %v, get %v", http.StatusUnauthorized, rec.Code)
		}
		var res map[string]map[string]string
		json.NewDecoder(rec.Body).Decode(&res)
		if res["error"]["code"] != "BAD_JWT_TOKEN" {
			t.Fatalf("expect error with code %v, get %v", "BAD_JWT_TOKEN", res["error"]["code"])
		}
	})
}

func TestAuthAdminOnlyMiddleware(t *testing.T) {
	inner := func(r *http.Request) (*AppResponse, *AppError) {
		return newAppResponse(http.StatusOK, nil), nil
	}
	t.Run("test with admin role", func(t *testing.T) {
		adminOnlyMW := AuthAdminOnlyMiddleware(inner)
		userAdminCtx := UserContext{
			Role: "admin",
		}
		req := httptest.NewRequest("GET", "/", nil)
		ctx := context.WithValue(context.Background(), userContextKey, userAdminCtx)
		appRes, err := adminOnlyMW(req.WithContext(ctx))
		if err != nil {
			t.Fatalf("expect no error, get %v", err.Error())
		}
		if appRes.Status != http.StatusOK {
			t.Fatalf("expect status %v, get %v", http.StatusOK, appRes.Status)
		}
	})
	t.Run("test with engineer role", func(t *testing.T) {
		adminOnlyMW := AuthAdminOnlyMiddleware(inner)
		userEngineerCtx := UserContext{
			Role: "engineer",
		}
		req := httptest.NewRequest("GET", "/", nil)
		ctx := context.WithValue(context.Background(), userContextKey, userEngineerCtx)
		_, err := adminOnlyMW(req.WithContext(ctx))
		if err == nil {
			t.Fatalf("expect error, get no error")
		}
		var appErr *AppError
		if errors.As(err, &appErr) == false {
			t.Fatalf("expect error has type *AppError")
		}
		if appErr.Status != http.StatusForbidden {
			t.Fatalf("expect status %v, get %v", http.StatusForbidden, appErr.Status)
		}
		if appErr.Code != "FORBIDDEN" {
			t.Fatalf("expect Code %v, get %v", "FORBIDDEN", appErr.Code)
		}
	})

}
