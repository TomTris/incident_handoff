package main

import "context"

type IncidentStore interface {
	CreateIncident(ctx context.Context, inc CreateIncidentRequest) (Incident, error)
	GetIncident(ctx context.Context, id string) (Incident, error)
	ListIncidents(ctx context.Context, filter IncidentFilter) ([]Incident, error)
	UpdateIncident(ctx context.Context, id string, update IncidentUpdate) (Incident, error) // Return incident After
	AddEntry(ctx context.Context, incidentID string, entry TimelineEntry) (TimelineEntry, error)
}
