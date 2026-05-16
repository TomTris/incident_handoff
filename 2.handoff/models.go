package main

import (
	"errors"
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
	f.Service = strings.TrimSpace(f.Service)
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
	OnCall   *string `json:"on_call,omitempty" bson:"on_call"`
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
	Severity         string           `json:"severity"`
	Status           string           `json:"status"`
	Service          string           `json:"service"`
	TotalEntry       int              `json:"total_entry"`
	ElapsedMinute    int              `json:"elapsed_minute"`
	TakenActions     int              `json:"taken_actions"`
	OpenQuestion     int              `json:"open_question"`
	HandoffCount     int              `json:"handoff_count"`
	TakenActionsList *[]TimelineEntry `json:"taken_actions_list,omitempty"`
	OpenQuestionList *[]TimelineEntry `json:"open_question_list,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
}

type FeatureFlag struct {
	Name     string   `json:"name"`
	Enabled  bool     `json:"enabled"`
	Rollout  int      `json:"rollout"`  // 0–100, percentage of users who see the feature
	Variants []string `json:"variants"` // e.g., ["control", "variant_a", "variant_b"]
}

func (f *FeatureFlag) Validate() error {
	if strings.TrimSpace(f.Name) == "" {
		return ErrBadRequest
	}
	if f.Rollout < 0 && 100 > f.Rollout {
		return ErrBadRequest
	}
	if f.Variants == nil {
		return ErrBadRequest
	}
	variants := make(map[string]bool)
	for _, variant := range f.Variants {
		if strings.TrimSpace(variant) == "" || variants[variant] == true {
			return ErrBadRequest
		}
		variants[variant] = true
	}
	return nil
}

type FeatureFlagUpdate struct {
	Name    string `json:"name"`
	Enabled *bool  `json:"enabled"`
	Rollout *int   `json:"rollout"` // 0–100, percentage of users who see the feature
}

func (u *FeatureFlagUpdate) Validate() error {
	if strings.TrimSpace(u.Name) == "" {
		return errors.New("Bad Flag Name")
	}
	if u.Enabled == nil && u.Rollout == nil {
		return errors.New("both enabled and rollout are empty")
	}
	if u.Rollout != nil && *u.Rollout < 0 && 100 > *u.Rollout {
		return errors.New("invalid rollout")
	}
	return nil
}
