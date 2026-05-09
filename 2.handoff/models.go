package main

import (
	"strings"
	"time"
)

type Incident struct {
	ID        string          `json:"id" bson:"_id,omitempty"`
	Title     string          `json:"title" bson:"title"`
	Service   string          `json:"service" bson:"service"`
	Severity  string          `json:"severity" bson:"severity"` // SEV1, SEV2, SEV3
	Status    string          `json:"status" bson:"status"`     // triggered, acknowledged, investigating, mitigated, resolved
	OpenedBy  string          `json:"opened_by" bson:"opened_by"`
	OnCall    string          `json:"on_call" bson:"on_call"`
	CreatedAt time.Time       `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" bson:"updated_at"`
	Entries   []TimelineEntry `json:"entries" bson:"entries"`
}

type TimelineEntry struct {
	ID     string    `json:"id" bson:"id"`
	Time   time.Time `json:"time" bson:"time"`
	Author string    `json:"author" bson:"author"`
	Type   string    `json:"type" bson:"type"` // observation, action, discovery, open_question, state_change
	Text   string    `json:"text" bson:"text"`
}

func (c *TimelineEntry) Validate() error {
	if strings.TrimSpace(c.Author) == "" {
		return ErrNoAuthor
	}
	if validEntryTypes[strings.TrimSpace(c.Type)] == false {
		return ErrBadEntryType
	}
	if strings.TrimSpace(c.Text) == "" {
		return ErrNoText
	}
	return nil
}

type IncidentFilter struct {
	Status  string `json:"status" bson:"status"`
	Service string `json:"service" bson:"service"`
}

func (f *IncidentFilter) Validate() error {
	if f.Status != "" && !IncidentStatus[strings.TrimSpace(f.Status)] {
		return ErrBadIncidentStatus
	}
	return nil
}

type IncidentUpdate struct {
	Status   *string `json:"status" bson:"status"`
	Severity *string `json:"severity" bson:"severity"`
	OnCall   *string `json:"on_call" bson:"on_call"`
}

func (f *IncidentUpdate) Validate() error {
	if f.Status != nil && IncidentStatus[strings.TrimSpace(*f.Status)] == false {
		return ErrBadIncidentStatus
	}
	if f.Severity != nil && IncidentSeverity[strings.TrimSpace(*f.Severity)] == false {
		return ErrInvalidSeverity
	}
	if f.OnCall != nil && strings.TrimSpace(*f.OnCall) == "" {
		return ErrOnCall
	}
	return nil
}

type CreateIncidentRequest struct {
	Title    string  `json:"title" bson:"title"`
	Service  string  `json:"service" bson:"service"`
	Severity string  `json:"severity" bson:"severity"` // SEV1, SEV2, SEV3
	OpenedBy string  `json:"opened_by" bson:"opened_by"`
	OnCall   *string `json:"on_call" bson:"on_call"`
}

func (c *CreateIncidentRequest) Validate() error {
	if strings.TrimSpace(c.Title) == "" {
		return ErrNoTitle
	}
	if strings.TrimSpace(c.Service) == "" {
		return ErrNoService
	}
	if IncidentSeverity[strings.TrimSpace(c.Severity)] == false {
		return ErrInvalidSeverity
	}
	if strings.TrimSpace(c.OpenedBy) == "" {
		return ErrOpenedBy
	}
	if c.OnCall != nil && strings.TrimSpace(*c.OnCall) == "" {
		return ErrOnCall
	}
	return nil
}

type HandoffBrief struct {
	Severity      string          `json:"severity" bson:"severity"`
	Status        string          `json:"status" bson:"status"`
	Service       string          `json:"service" bson:"service"`
	TotalEntry    int             `json:"total_entry" bson:"total_entry"`
	ElapsedMinute int             `json:"elapsed_minute" bson:"elapsed_minute"`
	TakenActions  []TimelineEntry `json:"taken_actions" bson:"taken_actions"`
	OpenQuestion  []TimelineEntry `json:"open_question" bson:"open_question"`
	CreatedAt     time.Time       `json:"created_at" bson:"created_at"`
}
