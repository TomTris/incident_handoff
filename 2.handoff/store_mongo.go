package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoStore struct {
	db *mongo.Database
}

func (m *MongoStore) DropAll(ctx context.Context) error {
	return m.db.Drop(ctx)
}

func NewMongoStore(uri string, DBName string) *MongoStore {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	db := client.Database(DBName)

	// Sofar, it's the best, even at scale.
	// Becaus most incidents should be resolved overtime
	// And the heaviest case that index can help is list status = "active"
	db.Collection(CollectionIncidents).Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{Keys: bson.D{
			{Key: "status", Value: 1},
			{Key: "service", Value: 1},
			{Key: "created_at", Value: 1},
		}},
		{Keys: bson.D{
			{Key: "service", Value: 1},
			{Key: "created_at", Value: 1},
		}},
	})
	slog.Info("schema/indexes ensured")
	return &MongoStore{db: db}
}

func (m *MongoStore) nextID(ctx context.Context, name string, prefix string) (string, error) {
	col := m.db.Collection(CollectionCounters)
	filter := bson.M{"_id": name}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var result struct {
		Seq int `bson:"seq"`
	}
	err := col.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return "", err
	}
	return prefix + strconv.Itoa(result.Seq), nil
}

func (m *MongoStore) CreateIncident(ctx context.Context, inc Incident) (Incident, error) {
	id, err := m.nextID(ctx, "incident", incidentIDPrefix)
	if err != nil {
		return inc, err
	}
	inc.ID = id
	inc.Status = TRIGGERED
	inc.CreatedAt = time.Now()
	inc.UpdatedAt = time.Now()
	inc.Entries = []TimelineEntry{}
	if inc.OnCall == "" {
		inc.OnCall = inc.OpenedBy
	}

	col := m.db.Collection(CollectionIncidents)
	_, err = col.InsertOne(ctx, inc)
	return inc, err
}

func (m *MongoStore) GetIncident(ctx context.Context, id string) (Incident, error) {
	col := m.db.Collection(CollectionIncidents)
	filter := bson.M{"_id": id}
	var inc Incident
	err := col.FindOne(ctx, filter).Decode(&inc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return inc, ErrIncidentNotFound
		}
		return inc, err
	}
	return inc, nil
}

// TODO: This might cause race, find another method in critical cases (banking, ...)
// Consequence of current method: might send wrong error code, but at least data is safe, not so critical
func (m *MongoStore) AddEntry(ctx context.Context, incidentID string, entry TimelineEntry) (TimelineEntry, error) {
	id, err := m.nextID(ctx, "timeline_entry", entryIDPrefix)
	if err != nil {
		return entry, err
	}

	entry.ID = id
	entry.Time = time.Now()

	filter := bson.M{
		"_id":    incidentID,
		"status": bson.M{"$ne": RESOLVED},
	}
	update := bson.M{
		"$push": bson.M{"entries": entry},
		"$set":  bson.M{"updated_at": time.Now()},
	}
	col := m.db.Collection(CollectionIncidents)
	result, err := col.UpdateOne(ctx, filter, update)
	if err != nil {
		return entry, err
	}
	if result.MatchedCount == 0 {
		err = col.FindOne(ctx, bson.M{"_id": incidentID}).Err()
		if errors.Is(err, mongo.ErrNoDocuments) {
			return entry, ErrIncidentNotFound
		}
		return entry, ErrIncidentConflict
	}
	return entry, nil
}

func (m *MongoStore) ListIncidents(ctx context.Context, incFilter IncidentFilter) ([]Incident, error) {
	dbFilter := bson.M{}
	if incFilter.Service != "" {
		dbFilter["service"] = incFilter.Service
	}

	switch incFilter.Status {
	case "":
		break
	case "active":
		dbFilter["status"] = bson.M{"$ne": RESOLVED}
	default:
		dbFilter["status"] = incFilter.Status
	}

	col := m.db.Collection(CollectionIncidents)
	opts := options.Find().SetSort(bson.M{"created_at": 1})
	cursor, err := col.Find(ctx, dbFilter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var incidents []Incident
	err = cursor.All(ctx, &incidents)

	if incidents == nil {
		incidents = []Incident{}
	}
	return incidents, err
}

func (m *MongoStore) UpdateIncident(ctx context.Context, incidentId string, update IncidentUpdate) error {
	DBUpdate := bson.M{}

	switch {
	case update.Status != nil:
		DBUpdate["status"] = *update.Status
	case update.Severity != nil:
		DBUpdate["severity"] = *update.Severity
	case update.OnCall != nil:
		DBUpdate["on_call"] = *update.OnCall
	}

	DBUpdate["updated_at"] = time.Now()

	col := m.db.Collection(CollectionIncidents)
	result, err := col.UpdateByID(ctx, incidentId, bson.M{"$set": DBUpdate})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrIncidentNotFound
	}
	return nil
}
