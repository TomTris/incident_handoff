package main

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
)

func setup_testInstrumented_Env(t *testing.T) *InstrumentedIncidentStore {
	t.Helper()
	reg := prometheus.NewRegistry()
	metrics := NewIncidentStoreMetric(reg)
	instrumented := InstrumentedIncidentStore{
		inner:   NewMemoryIncidentStore(),
		metrics: metrics,
	}
	return &instrumented
}

func validCreateIncidentRequest() CreateIncidentRequest {
	return CreateIncidentRequest{
		Title:    "Title",
		Service:  "payment",
		Severity: "SEV1",
		OpenedBy: "OpenedBy",
	}
}

func historyQueryDurationSampleCount(instrumented *InstrumentedIncidentStore, lvs ...string) uint64 {
	observer := instrumented.metrics.DbQueryDurationSeconds.WithLabelValues(lvs...)
	m := &dto.Metric{}
	observer.(prometheus.Metric).Write(m)
	count := m.GetHistogram().GetSampleCount()
	return count
}

func TestInstrumented(t *testing.T) {
	instrumented := setup_testInstrumented_Env(t)

	t.Run("Create Incident", func(t *testing.T) {
		_, err := instrumented.CreateIncident(context.Background(), validCreateIncidentRequest())
		if err != nil {
			t.Fatal(err)
		}
		_, err = instrumented.CreateIncident(context.Background(), validCreateIncidentRequest())
		if err != nil {
			t.Fatal(err)
		}

		t.Run("Increments gauge", func(t *testing.T) {
			val := testutil.ToFloat64(instrumented.metrics.IncidentTotal.WithLabelValues("triggered"))
			if val != 2 {
				t.Errorf(`Incident["triggered"] expected 2, got %v`, val)
			}
		})
		t.Run("Observe Duration Histogram", func(t *testing.T) {
			count := historyQueryDurationSampleCount(instrumented, "create_incident")
			if count != 2 {
				t.Errorf("DbQueryDurationSeconds[create_incident] expected 2 inputs, got %v", count)
			}
		})
	})
	t.Run("Get Incident", func(t *testing.T) {
		_, err := instrumented.GetIncident(context.Background(), "INC-1")
		if err != nil {
			t.Fatal(err)
		}
		count := historyQueryDurationSampleCount(instrumented, "get_incident")
		if count != 1 {
			t.Errorf("DbQueryDurationSeconds[get_incident] expected 1 inputs, got %v", count)
		}
	})
	t.Run("Add Entry", func(t *testing.T) {
		_, err := instrumented.AddEntry(context.Background(), "INC-2", TimelineEntry{
			ID:     "TLE-1",
			Time:   time.Now(),
			Author: "me",
			Type:   "observation",
			Text:   "test",
		})
		if err != nil {
			t.Fatal(err)
		}
		val := testutil.ToFloat64(instrumented.metrics.TotalEntries)
		if val != 1 {
			t.Errorf("expected 1, got %v", val)
		}
		count := historyQueryDurationSampleCount(instrumented, "add_entry")
		if count != 1 {
			t.Errorf("DbQueryDurationSeconds[add_entry] expected 1, got %v", count)
		}
	})
	t.Run("List Incident", func(t *testing.T) {
		_, err := instrumented.ListIncidents(context.Background(), IncidentFilter{})
		if err != nil {
			t.Fatal(err)
		}
		count := historyQueryDurationSampleCount(instrumented, "list_incident")
		if count != 1 {
			t.Errorf("DbQueryDurationSeconds[list_incident] expected 1, got %v", count)
		}
	})
	t.Run("Update Incident", func(t *testing.T) {
		_, err := instrumented.UpdateIncident(context.Background(), "INC-1", IncidentUpdate{Severity: new(SEV3)})
		if err != nil {
			t.Fatal(err)
		}
		count := historyQueryDurationSampleCount(instrumented, "update_incident")
		if count != 1 {
			t.Errorf("DbQueryDurationSeconds[update_incident] expected 1, got %v", count)
		}
	})
	t.Run("Update Incident by changing status", func(t *testing.T) {
		_, err := instrumented.UpdateIncident(context.Background(), "INC-1", IncidentUpdate{Status: new(INVESTIGATING)})
		if err != nil {
			t.Fatal(err)
		}
		count := historyQueryDurationSampleCount(instrumented, "update_incident")
		if count != 2 {
			t.Errorf("DbQueryDurationSeconds[update_incident] expected 2, got %v", count)
		}
	})
}
