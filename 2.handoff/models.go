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
	c.Author = strings.TrimSpace(c.Author)
	if c.Author == "" {
		return ErrNoAuthor
	}
	c.Type = strings.TrimSpace(c.Type)
	if validEntryTypes[c.Type] == false {
		return ErrBadEntryType
	}
	c.Text = strings.TrimSpace(c.Text)
	if c.Text == "" {
		return ErrNoText
	}
	return nil
}

type IncidentFilter struct {
	Status  string `json:"status" bson:"status"`
	Service string `json:"service" bson:"service"`
}

func (f *IncidentFilter) Validate() error {
	f.Status = strings.TrimSpace(f.Status)
	f.Service = strings.TrimSpace(f.Status)
	if f.Status != "" && !IncidentStatus[f.Status] {
		return ErrBadIncidentStatus
	}
	return nil
}

type IncidentUpdate struct {
	Status   *string `json:"status,omitempty" bson:"status"`
	Severity *string `json:"severity,omitempty" bson:"severity"`
	OnCall   *string `json:"on_call,omitempty" bson:"on_call"`
}

func (f *IncidentUpdate) Validate() error {
	switch {
	case f.Status != nil:
		trimmed := strings.TrimSpace(*f.Status)
		if IncidentStatus[trimmed] == false {
			return ErrBadIncidentStatus
		}
		*f = IncidentUpdate{Status: &trimmed}
		return nil
	case f.Severity != nil:
		trimmed := strings.TrimSpace(*f.Severity)
		if IncidentSeverity[trimmed] == false {
			return ErrInvalidSeverity
		}
		*f = IncidentUpdate{Severity: &trimmed}
		return nil
	case f.OnCall != nil:
		trimmed := strings.TrimSpace(*f.OnCall)
		if trimmed == "" {
			return ErrOnCall
		}
		*f = IncidentUpdate{OnCall: &trimmed}
		return nil
	default:
		return ErrBadRequest
	}
}

type CreateIncidentRequest struct {
	Title    string  `json:"title" bson:"title"`
	Service  string  `json:"service" bson:"service"`
	Severity string  `json:"severity" bson:"severity"` // SEV1, SEV2, SEV3
	OpenedBy string  `json:"opened_by" bson:"opened_by"`
	OnCall   *string `json:"on_call" bson:"on_call"`
}

func (c *CreateIncidentRequest) Validate() error {
	c.Title = strings.TrimSpace(c.Title)
	if c.Title == "" {
		return ErrNoTitle
	}

	c.Service = strings.TrimSpace(c.Service)
	if c.Service == "" {
		return ErrNoService
	}

	c.Severity = strings.TrimSpace(c.Severity)
	if IncidentSeverity[c.Severity] == false {
		return ErrInvalidSeverity
	}

	c.OpenedBy = strings.TrimSpace(c.OpenedBy)
	if c.OpenedBy == "" {
		return ErrOpenedBy
	}

	if c.OnCall != nil {
		*c.OnCall = strings.TrimSpace(*c.OnCall)
		if *c.OnCall == "" {
			return ErrOnCall
		}
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
