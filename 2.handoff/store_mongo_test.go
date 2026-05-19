package main

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const connectionDBString = "mongodb://127.0.0.1:27017/?directConnection=true"
const DBName = "incident_tracker"

func setupMongoTestEnv(t *testing.T) *MongoStore {
	t.Helper()

	client, err := mongo.Connect(options.Client().ApplyURI(connectionDBString))
	if err != nil {
		t.Fatal(err)
	}
	mongoStore := NewMongoStore(client, "incident_tracker")
	mongoStore.DropAll(context.Background())
	return mongoStore
}
func TestMongoStore_CreateIncident(t *testing.T) {
	m := setupMongoTestEnv(t)

	t.Run("defaults OnCall to OpenedBy", func(t *testing.T) {
		inc, err := m.CreateIncident(context.Background(), CreateIncidentRequest{
			Title:    "outage",
			Service:  "api",
			Severity: "SEV1",
			OpenedBy: "anh",
		})
		if err != nil {
			t.Fatal(err)
		}
		if inc.OnCall != "anh" {
			t.Errorf("Oncall expected `anh`, got `%s`", inc.OnCall)
		}
	})

	t.Run("uses explicit OnCall", func(t *testing.T) {
		onCall := "tom"
		inc, err := m.CreateIncident(context.Background(), CreateIncidentRequest{
			Title:    "outage2",
			Service:  "api",
			Severity: "SEV1",
			OpenedBy: "anh",
			OnCall:   &onCall,
		})
		if err != nil {
			t.Fatal(err)
		}
		if inc.OnCall != onCall {
			t.Errorf("Oncall expected `%s`, got `%s`", onCall, inc.OnCall)
		}
	})

	t.Run("sets correct defaults", func(t *testing.T) {
		inc, _ := m.CreateIncident(context.Background(), CreateIncidentRequest{
			Title:    "outage3",
			Service:  "api",
			Severity: "SEV1",
			OpenedBy: "anh",
		})
		if inc.Status != TRIGGERED {
			t.Errorf("Status expected %s, got %s", TRIGGERED, inc.Status)
		}
		if inc.CreatedAt.IsZero() {
			t.Errorf("CreateAt not set")
		}
		if len(inc.Entries) != 0 {
			t.Errorf("len(inc.Entries) expected 0, got %v", len(inc.Entries))
		}
	})

	t.Run("sets correct properties", func(t *testing.T) {
		inc, _ := m.CreateIncident(context.Background(), CreateIncidentRequest{
			Title:    "outage4",
			Service:  "api",
			Severity: "SEV1",
			OpenedBy: "anh",
		})
		if inc.Title != "outage4" {
			t.Errorf("Title expected %s, got %s", "outage4", inc.Title)
		}
		if inc.Service != "api" {
			t.Errorf("Service expected %s, got %s", "api", inc.Service)
		}
		if inc.Severity != "SEV1" {
			t.Errorf("Severity expected %s, got %s", "SEV", inc.Severity)
		}
		if inc.OpenedBy != "anh" {
			t.Errorf("OpenedBy expected 0, got %v", inc.OpenedBy)
		}
	})

	t.Run("sets correct properties", func(t *testing.T) {
		inc, _ := m.CreateIncident(context.Background(), CreateIncidentRequest{
			Title:    "outage4",
			Service:  "api",
			Severity: "SEV1",
			OpenedBy: "anh",
		})
		if inc.Title != "outage4" {
			t.Errorf("Title expected %s, got %s", "outage4", inc.Title)
		}
		if inc.Service != "api" {
			t.Errorf("Service expected %s, got %s", "api", inc.Service)
		}
		if inc.Severity != "SEV1" {
			t.Errorf("Severity expected %s, got %s", "SEV", inc.Severity)
		}
		if inc.OpenedBy != "anh" {
			t.Errorf("OpenedBy expected 0, got %v", inc.OpenedBy)
		}
	})
	t.Run("sequential IDs", func(t *testing.T) {
		m.DropAll(context.Background())
		inc1, _ := m.CreateIncident(context.Background(), CreateIncidentRequest{
			Title: "a", Service: "s", Severity: "SEV1", OpenedBy: "x",
		})
		inc2, _ := m.CreateIncident(context.Background(), CreateIncidentRequest{
			Title: "b", Service: "s", Severity: "SEV1", OpenedBy: "x",
		})
		if inc1.ID != "INC-1" {
			t.Errorf("expected INC-1, got %s", inc1.ID)
		}
		if inc2.ID != "INC-2" {
			t.Errorf("expected INC-2, got %s", inc2.ID)
		}
	})
}
