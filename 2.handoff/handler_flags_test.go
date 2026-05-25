package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateFlag(t *testing.T) {
	validFeatureFlag := func() FeatureFlag {
		return FeatureFlag{
			Name:     "test-feature-flag",
			Enabled:  false,
			Rollout:  50,
			Variants: []string{"controlled", "detailed"},
		}
	}

	validFeatureFlagUpdate := func() FeatureFlagUpdate {
		return FeatureFlagUpdate{Name: "test-feature-flag", Enabled: new(true)}
	}

	flagHandler := FlagHandler{store: CreateFlagStore()}
	t.Run("test CreateFlag func", func(t *testing.T) {

		t.Run("Create Flag Normally", func(t *testing.T) {
			body, err := json.Marshal(validFeatureFlag())

			if err != nil {
				t.Fatal("can't Marshal FeatureFlag")
			}

			req := httptest.NewRequest("POST", "/flags", bytes.NewReader(body))
			appRes, err := flagHandler.CreateFlag(req)

			if err != nil {
				t.Fatalf("expected no error, got error %v", err)
			}
			if appRes.Status != http.StatusCreated {
				t.Fatalf("expected status %v, got  %v", http.StatusCreated, appRes.Status)
			}
			f := appRes.Body.(FeatureFlag)
			if f.Name != "test-feature-flag" {
				t.Fatalf("expected FeatureFlag name %v, got %v", "test-feature-flag", f.Name)
			}
		})

		t.Run("Create Flag Conflict", func(t *testing.T) {
			body, err := json.Marshal(validFeatureFlag())

			if err != nil {
				t.Fatal("can't Marshal FeatureFlag")
			}

			req := httptest.NewRequest("POST", "/flags", bytes.NewReader(body))
			_, err = flagHandler.CreateFlag(req)

			if err == nil {
				t.Fatal("expected error conflict, got no error")
			}

			var appErr *AppError
			errors.As(err, &appErr)
			if appErr.Status != http.StatusConflict {
				t.Fatalf("Expected %v, got %v", http.StatusConflict, appErr.Status)
			}
		})
	})

	t.Run("test UpdateFlag func", func(t *testing.T) {
		t.Run("Update Flag Notfound", func(t *testing.T) {
			body, err := json.Marshal(FeatureFlagUpdate{Name: "not-exist-feature-flag", Rollout: new(60)})
			if err != nil {
				t.Fatal("can't marshal FeatureFlagUpdate")
			}
			req := httptest.NewRequest("POST", "/flags/not-exist-feature-flag", bytes.NewReader(body))
			req.SetPathValue("name", "not-exist-feature-flag")

			_, err = flagHandler.UpdateFlag(req)
			if err == nil {
				t.Fatalf("expected error, got no error")
			}
			var appErr *AppError
			errors.As(err, &appErr)
			if appErr.Status != http.StatusNotFound {
				t.Fatalf("expected status %v, got %v", http.StatusNotFound, appErr.Status)
			}
		})

		t.Run("Update Flag", func(t *testing.T) {
			update := validFeatureFlagUpdate()
			body, err := json.Marshal(update)
			if err != nil {
				t.Fatal("can't marshal FeatureFlagUpdate")
			}
			req := httptest.NewRequest("POST", fmt.Sprintf("/flag/%v", update.Name), bytes.NewReader(body))
			req.SetPathValue("name", update.Name)

			appRes, err := flagHandler.UpdateFlag(req)
			if err != nil {
				t.Fatalf("expected no error, got error %v", appRes)
			}
			if appRes.Status != http.StatusNoContent {
				t.Fatalf("expected status %v, got %v", http.StatusNoContent, appRes.Status)
			}
		})
	})

}
