package main

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const connectionDBString = "mongodb://127.0.0.1:27017/?directConnection=true"
const DBName = "incident_tracker"

func setupMongoTestEnv(t *testing.T) *MongoIncidentStore {
	t.Helper()

	client, err := mongo.Connect(options.Client().ApplyURI(connectionDBString))
	if err != nil {
		t.Fatal(err)
	}
	mongoIncidentStore := NewMongoIncidentStore(client, "incident_tracker")
	mongoIncidentStore.DropAll(context.Background())
	return mongoIncidentStore
}

func TestMongoStore(t *testing.T) {
	runStoreContractsTests(t, func(t *testing.T) IncidentStore {
		return setupMongoTestEnv(t)
	})
}
