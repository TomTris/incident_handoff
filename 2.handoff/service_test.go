package main

import (
	"testing"
	"time"
)

func TestBuildHandoffBrief(t *testing.T) {
	t.Run("filters entries and maps fields", func(t *testing.T) {
		inc := Incident{
			Severity:  SEV1,
			Status:    INVESTIGATING,
			Service:   "payments",
			CreatedAt: time.Now().Add(-30 * time.Minute),
			Entries: []TimelineEntry{
				{ID: entryIDPrefix + "-1", Type: ACTION, Text: "restarted pod", Author: "alice"},
				{ID: entryIDPrefix + "-2", Type: OPEN_QUESTION, Text: "why did latency spike?", Author: "bob"},
				{ID: entryIDPrefix + "-3", Type: OBSERVATION, Text: "error rate at 5%", Author: "alice"},
				{ID: entryIDPrefix + "-4", Type: DISCOVERY, Text: "found memory leak", Author: "carol"},
				{ID: entryIDPrefix + "-5", Type: ACTION, Text: "rolled back deploy", Author: "bob"},
			},
		}

		brief := buildHandoffBrief(inc)
		if brief.Severity != inc.Severity {
			t.Errorf("Severity = %q, want %q", brief.Severity, SEV1)
		}
		if brief.Status != INVESTIGATING {
			t.Errorf("Status = %q, want %q", brief.Status, INVESTIGATING)
		}
		if brief.Service != "payments" {
			t.Errorf("Service = %q, want %q", brief.Service, "payments")
		}
		if brief.TotalEntry != 5 {
			t.Errorf("TotalEntry = %d, want 5", brief.TotalEntry)
		}
		if len(brief.TakenActions) != 2 {
			t.Errorf("TakenActions count = %d, want 2", len(brief.TakenActions))
		}
		if len(brief.OpenQuestion) != 1 {
			t.Errorf("OpenQuestion count = %d, want 1", len(brief.OpenQuestion))
		}
		if brief.ElapsedMinute < 29 || brief.ElapsedMinute > 31 {
			t.Errorf("ElapsedMinute = %d, want ~30", brief.ElapsedMinute)
		}
		if !brief.CreatedAt.Equal(inc.CreatedAt) {
			t.Errorf("CreatedAt = %v, want %v", brief.CreatedAt, inc.CreatedAt)
		}
	})
	t.Run("empty entry", func(t *testing.T) {
		inc := Incident{
			Severity:  SEV2,
			Status:    ACKNOWLEDGED,
			Service:   "search",
			CreatedAt: time.Now(),
			Entries:   []TimelineEntry{},
		}

		brief := buildHandoffBrief(inc)

		if brief.TotalEntry != 0 {
			t.Errorf("TotalEntry = %d, want 0", brief.TotalEntry)
		}
		if len(brief.TakenActions) != 0 {
			t.Errorf("TakenActions count = %d, want 0", len(brief.TakenActions))
		}
		if len(brief.OpenQuestion) != 0 {
			t.Errorf("OpenQuestion count = %d, want 0", len(brief.OpenQuestion))
		}
	})

}
