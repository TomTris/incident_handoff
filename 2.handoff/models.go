package main

import (
	"errors"
	"strings"
	"time"
)

type Incident struct {
	ID        string          `json:"id"`
	Title     string          `json:"title"`
	Service   string          `json:"service"`
	Severity  string          `json:"severity"` // SEV1, SEV2, SEV3
	Status    string          `json:"status"`   // triggered, acknowledged, investigating, mitigated, resolved
	OpenedBy  string          `json:"opened_by"`
	OnCall    string          `json:"on_call"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	Entries   []TimelineEntry `json:"entries"`
}

type TimelineEntry struct {
	ID     string    `json:"id"`
	Time   time.Time `json:"time"`
	Author string    `json:"author"`
	Type   string    `json:"type"` // observation, action, discovery, open_question, state_change
	Text   string    `json:"text"`
}

func (c *TimelineEntry) Validate() error {
	var validEntryTypes = map[string]bool{
		OBSERVATION:   true,
		ACTION:        true,
		DISCOVERY:     true,
		OPEN_QUESTION: true,
		STATE_CHANGE:  true,
	}

	if strings.Trim(c.Author, " ") == "" {
		return errors.New("Request doesn't contain Author")
	}
	if validEntryTypes[strings.Trim(c.Type, " ")] == false {
		return errors.New("Request doesn't have valid entry type")
	}
	if strings.Trim(c.Text, " ") == "" {
		return errors.New("Request doesn't contain text")
	}
	return nil
}

// type IncidentFilter struct {
// }

// type IncidentUpdate struct {
// }

type CreateIncidentRequest struct {
	Title    string `json:"title"`
	Service  string `json:"service"`
	Severity string `json:"severity"` // SEV1, SEV2, SEV3
	OpenedBy string `json:"opened_by"`
}

func (c *CreateIncidentRequest) Validate() error {
	if strings.Trim(c.Title, " ") == "" {
		return errors.New("Request doesn't contain title")
	}
	if strings.Trim(c.Service, " ") == "" {
		return errors.New("Request doesn't contain service")
	}
	if strings.Trim(c.Severity, " ") == "" {
		return errors.New("Request doesn't contain severity")
	}
	if strings.Trim(c.OpenedBy, " ") == "" {
		return errors.New("Request doesn't contain opened_by")
	}
	return nil
}
