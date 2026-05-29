package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoIncidentStore struct {
	db *mongo.Database
}

func (m *MongoIncidentStore) DropAll(ctx context.Context) error {
	return m.db.Drop(ctx)
}

func NewMongoIncidentStore(client *mongo.Client, DBName string) *MongoIncidentStore {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := client.Ping(ctx, nil)
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
	return &MongoIncidentStore{db: db}
}

func (m *MongoIncidentStore) nextID(ctx context.Context, name string, prefix string) (string, error) {
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

func (m *MongoIncidentStore) CreateIncident(ctx context.Context, req CreateIncidentRequest) (Incident, error) {
	id, err := m.nextID(ctx, "incident", incidentIDPrefix)
	if err != nil {
		return Incident{}, errors.New("Failed to get next incident Id: " + err.Error())
	}
	inc := Incident{
		ID:        id,
		Title:     req.Title,
		Service:   req.Service,
		Severity:  req.Severity,
		OpenedBy:  req.OpenedBy,
		OnCall:    req.OnCall,
		Status:    TRIGGERED,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Entries:   []TimelineEntry{},
	}

	col := m.db.Collection(CollectionIncidents)
	_, err = col.InsertOne(ctx, inc)
	if err != nil {
		return Incident{}, errors.New("Failed to insert Incident: " + err.Error())
	}
	return inc, nil
}

func (m *MongoIncidentStore) GetIncident(ctx context.Context, id string) (Incident, error) {
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

func (m *MongoIncidentStore) AddEntry(ctx context.Context, incidentID string, entry TimelineEntry) (TimelineEntry, error) {
	id, err := m.nextID(ctx, "timeline_entry", entryIDPrefix)
	if err != nil {
		return entry, err
	}

	now := time.Now()
	entry.ID = id
	entry.Time = now

	isActive := bson.M{"$ne": bson.A{"$status", RESOLVED}}
	appendEntry := bson.M{"$concatArrays": bson.A{"$entries", bson.A{entry}}}

	pipeline := bson.A{
		bson.M{"$set": bson.M{
			"entries":    bson.M{"$cond": bson.A{isActive, appendEntry, "$entries"}},
			"updated_at": bson.M{"$cond": bson.A{isActive, now, "$updated_at"}},
		}},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.Before)

	var prev struct {
		Status string `bson:"status"`
	}
	err = m.db.Collection(CollectionIncidents).
		FindOneAndUpdate(ctx, bson.M{"_id": incidentID}, pipeline, opts).
		Decode(&prev)

	switch {
	case errors.Is(err, mongo.ErrNoDocuments):
		return entry, ErrIncidentNotFound
	case err != nil:
		return entry, err
	case prev.Status == RESOLVED:
		return entry, ErrIncidentConflict
	default:
		return entry, nil
	}
}

func (m *MongoIncidentStore) ListIncidents(ctx context.Context, incFilter IncidentFilter) ([]Incident, error) {
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

func (m *MongoIncidentStore) UpdateIncident(ctx context.Context, incidentId string, update IncidentUpdate) (Incident, error) {
	fields := bson.M{"updated_at": time.Now()}

	if update.Status != nil {
		fields["status"] = *update.Status
	}
	if update.Severity != nil {
		fields["severity"] = *update.Severity
	}
	if update.OnCall != nil {
		fields["on_call"] = *update.OnCall
	}

	col := m.db.Collection(CollectionIncidents)
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var incAfter Incident
	err := col.FindOneAndUpdate(
		ctx,
		bson.M{"_id": incidentId},
		bson.M{"$set": fields},
		opts,
	).Decode(&incAfter)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return incAfter, ErrIncidentNotFound
		}
		return incAfter, fmt.Errorf("Update Incident %s: %v", incidentId, err)
	}
	return incAfter, nil
}
