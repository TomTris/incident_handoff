package main

import (
	"context"
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

// func listAllIncidents(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "allication/json")
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(incidents)
// }
