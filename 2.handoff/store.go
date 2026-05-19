package main

import "context"

type Store interface {
	CreateIncident(ctx context.Context, inc CreateIncidentRequest) (Incident, error)
	GetIncident(ctx context.Context, id string) (Incident, error)
	ListIncidents(ctx context.Context, filter IncidentFilter) ([]Incident, error)
	UpdateIncident(ctx context.Context, id string, update IncidentUpdate) (Incident, error)
	AddEntry(ctx context.Context, incidentID string, entry TimelineEntry) (TimelineEntry, error)
}
