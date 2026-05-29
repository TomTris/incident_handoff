package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func doLogin(t *testing.T, authHandler *AuthHandler, body any) *httptest.ResponseRecorder {
	t.Helper()
	userLoginBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte(userLoginBytes)))
	rec := httptest.NewRecorder()
	authHandler.LoginHandler(rec, req)
	return rec
}

func TestLoginHandler(t *testing.T) {
	var seedUsers = []User{
		{ID: "u1", Username: "anh", Password: hashPassword("anh123"), Role: "engineer"},
		{ID: "u2", Username: "bernd", Password: hashPassword("bernd123"), Role: "engineer"},
		{ID: "u3", Username: "admin", Password: hashPassword("admin123"), Role: "admin"},
	}
	userStore := NewInMemoryUserStore(seedUsers)
	authHandler := NewAuthHandler(userStore, []byte("JWT-secret"), time.Duration(15*time.Minute))
	t.Run("Normal login", func(t *testing.T) {
		userLogin := UserLogin{
			Username: "anh",
			Password: "anh123",
		}
		rec := doLogin(t, authHandler, userLogin)
		if rec.Code != http.StatusOK {
			t.Fatalf("expect status %v, get %v", http.StatusOK, rec.Code)
		}
		cookies := rec.Result().Cookies()
		accessToken := ""
		for _, c := range cookies {
			if c.Name == "access_token" {
				accessToken = c.Value
				break
			}
		}
		if accessToken == "" {
			t.Fatalf("expect access_token, but empty")
		}
		// Check if the token is valid or not
		claims := CustomClaims{}
		token, err := jwt.ParseWithClaims(accessToken, &claims, func(t *jwt.Token) (any, error) {
			if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, errors.New("unexpected signing method")
			}
			return []byte("JWT-secret"), nil
		})
		if err != nil || token.Valid == false {
			if err != nil {
				t.Fatalf("expect no error, get %v", err.Error())
			}
			t.Fatalf("Invalid token")
		}
	})
	t.Run("non-exist username", func(t *testing.T) {
		userLogin := UserLogin{
			Username: "anh11111",
			Password: "anh123",
		}
		rec := doLogin(t, authHandler, userLogin)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expect status %v, get %v", http.StatusUnauthorized, rec.Code)
		}
	})
	t.Run("wrong password", func(t *testing.T) {
		userLogin := UserLogin{
			Username: "anh",
			Password: "anh1423",
		}
		rec := doLogin(t, authHandler, userLogin)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expect status %v, get %v", http.StatusUnauthorized, rec.Code)
		}
	})
	t.Run("Token Expired", func(t *testing.T) {
		authHandler = NewAuthHandler(userStore, []byte("JWT-secret"), time.Duration(-1*time.Minute))
		userLogin := UserLogin{
			Username: "anh",
			Password: "anh123",
		}
		rec := doLogin(t, authHandler, userLogin)
		if rec.Code != http.StatusOK {
			t.Fatalf("expect status %v, get %v", http.StatusOK, rec.Code)
		}
		cookies := rec.Result().Cookies()
		accessToken := ""
		for _, c := range cookies {
			if c.Name == "access_token" {
				accessToken = c.Value
				break
			}
		}
		if accessToken == "" {
			t.Fatalf("expect access_token, but empty")
		}
		// Check if the token is valid or not
		claims := CustomClaims{}
		_, err := jwt.ParseWithClaims(accessToken, &claims, func(t *jwt.Token) (any, error) {
			return []byte("JWT-secret"), nil
		})
		if err == nil || claims.ExpiresAt.Unix() >= time.Now().Unix() {
			t.Fatalf("expect token expires")
		}
	})
}
func TestWhoAmI(t *testing.T) {
	authHandler := NewAuthHandler(nil, nil, time.Duration(15))
	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), userContextKey, UserContext{
		ID:       "u1",
		Username: "anh",
		Role:     "engineer",
	})
	appRes, err := authHandler.WhoAmI(req.WithContext(ctx))
	if err != nil {
		t.Fatalf("expect no error, get %v", err.Error())
	}
	if appRes.Status != http.StatusOK {
		t.Fatalf("expect status %v, get %v", http.StatusOK, appRes.Status)
	}
	userContext := appRes.Body.(UserContext)
	if userContext.ID != "u1" {
		t.Fatalf("expect ID %v, get %v", "u1", userContext.ID)
	}
	if userContext.Username != "anh" {
		t.Fatalf("expect Username %v, get %v", "anh", userContext.Username)
	}
	if userContext.Role != "engineer" {
		t.Fatalf("expect Role %v, get %v", "engineer", userContext.Role)
	}
}
