package main

import (
	"context"
	"sort"
	"strconv"
	"time"
)

type MemoryStore struct {
	incidents           map[string]Incident
	nextIncidentID      int
	nextEntryTimelineID int
}

func (m *MemoryStore) CreateIncident(ctx context.Context, inc Incident) (Incident, error) {
	m.nextIncidentID++
	inc.ID = incidentIDPrefix + strconv.Itoa(m.nextIncidentID)
	inc.Status = TRIGGERED
	inc.CreatedAt = time.Now()
	inc.UpdatedAt = time.Now()

	m.incidents[inc.ID] = inc
	return inc, nil
}

func (m *MemoryStore) GetIncident(ctx context.Context, id string) (Incident, error) {
	inc, ok := m.incidents[id]
	if ok == false {
		return inc, ErrIncidentNotFound
	}
	return inc, nil
}

func (m *MemoryStore) AddEntry(ctx context.Context, incidentID string, entry TimelineEntry) (TimelineEntry, error) {
	inc, ok := m.incidents[incidentID]
	if ok == false {
		return TimelineEntry{}, ErrIncidentNotFound
	}
	if inc.Status == RESOLVED {
		return TimelineEntry{}, ErrConflict
	}
	m.nextEntryTimelineID++
	entry.ID = entryIDPrefix + strconv.Itoa(m.nextEntryTimelineID)
	entry.Time = time.Now()
	inc.Entries = append(inc.Entries, entry)
	m.incidents[incidentID] = inc
	return entry, nil
}

func (m *MemoryStore) ListIncidents(ctx context.Context, filter IncidentFilter) ([]Incident, error) {
	isServiceMatch := func(incident Incident, filter IncidentFilter) bool {
		return filter.Service == "" || filter.Service == incident.Service
	}
	isStatusMatch := func(incident Incident, filter IncidentFilter) bool {
		return filter.Status == "" ||
			(filter.Status == "active" && incident.Status != RESOLVED) ||
			filter.Status == incident.Status
	}

	array := []Incident{}
	for _, incident := range m.incidents {
		if isServiceMatch(incident, filter) && isStatusMatch(incident, filter) {
			array = append(array, incident)
		}
	}
	sort.Slice(array, func(i, j int) bool {
		return array[i].CreatedAt.Sub(array[j].CreatedAt) < 0
	})
	return array, nil
}

func (m *MemoryStore) UpdateIncident(ctx context.Context, id string, update IncidentUpdate) error {
	incident, ok := m.incidents[id]
	if ok != true {
		return ErrIncidentNotFound
	}
	if update.Status != nil {
		incident.Status = *update.Status
	}
	if update.Severity != nil {
		incident.Severity = *update.Severity
	}
	if update.OnCall != nil {
		incident.OnCall = *update.OnCall
	}
	m.incidents[id] = incident
	return nil
}
