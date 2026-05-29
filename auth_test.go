package main

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestPasswordRoundtrip(t *testing.T) {
	hashed := hashPassword("correct-password")
	if err := VerifyPassword(hashed, "correct-password"); err != nil {
		t.Fatalf("expected no error, get %v", err.Error())
	}
	if err := VerifyPassword(hashed, "wrong-password"); err == nil {
		t.Fatalf("expected error, get no error")
	}
	hashed2 := hashPassword("correct-password")
	if hashed == hashed2 {
		t.Fatalf("expected non deterministic hash result")
	}
}

func TestIssueToken(t *testing.T) {
	user := User{
		ID:       "123",
		Username: "anh",
		Role:     "engineer",
	}
	secret := "random-jwt-secret"
	now := time.Now()
	ttl := time.Duration(15 * time.Minute)

	t.Run("normal behavior", func(t *testing.T) {
		tokenSigned, err := IssueToken(user, []byte(secret), ttl, now)
		if err != nil {
			t.Fatalf("expect no error, get error %v", err.Error())
		}

		var claims CustomClaims
		token, err := jwt.ParseWithClaims(tokenSigned, &claims, func(t *jwt.Token) (any, error) {
			if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		})

		if err != nil || token.Valid == false {
			msg := "jwt invalid"
			if err != nil {
				msg = "error: " + err.Error()
			}
			t.Fatalf("expect normal jwt decode, but %v", msg)
		}

		if claims.Subject != user.ID {
			t.Errorf("not the same ID")
		}
		if claims.Username != user.Username {
			t.Errorf("not the same Username")
		}
		if claims.Role != user.Role {
			t.Errorf("not the same Role")
		}
		if claims.IssuedAt.Unix() != now.Unix() {
			t.Errorf("IssuedAt %v", claims.IssuedAt)
		}
		if claims.ExpiresAt.Unix() != (now.Add(15 * time.Minute)).Unix() {
			t.Errorf("ExpiresAt %v", claims.ExpiresAt)
		}
		if claims.ID == "" {
			t.Errorf("JWT ID empty")
		}
	})
	t.Run("not correct token", func(t *testing.T) {
		tokenSigned, err := IssueToken(user, []byte(secret), ttl, now)
		if err != nil {
			t.Fatalf("expect no error, get error %v", err.Error())
		}

		var claims CustomClaims
		_, err = jwt.ParseWithClaims(tokenSigned, &claims, func(t *jwt.Token) (any, error) {
			if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, errors.New("unexpected signing method")
			}
			return []byte("wrong-jwt-secret"), nil
		})

		if err == nil {
			t.Errorf("wrong secret should fail")
		}
	})
	t.Run("test using Algorithm", func(t *testing.T) {
		tokenSigned, _ := IssueToken(user, []byte(secret), ttl, now)
		parsed, _ := jwt.Parse(tokenSigned, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if parsed.Method.Alg() != "HS256" {
			t.Errorf("alg expect HS256, get %v", parsed.Method.Alg())
		}
	})
}
