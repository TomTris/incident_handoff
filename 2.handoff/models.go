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

type IncidentFilter struct {
	Status  string
	Service string
}

func (f *IncidentFilter) Validate() error {
	if IncidentStatus[strings.Trim(f.Status, " ")] == false {
		return errors.New("Invalid Incident status")
	}
	return nil
}

type IncidentUpdate struct {
	Status   *string
	Severity *string
	OnCall   *string
}

func (f *IncidentUpdate) Validate() error {
	if f.Status != nil && IncidentStatus[strings.Trim(*f.Status, " ")] == false {
		return errors.New("Invalid Incident status")
	}
	if f.Severity != nil && IncidentSeverity[strings.Trim(*f.Severity, " ")] == false {
		return errors.New("Invalid Incident Severity")
	}
	if f.OnCall != nil && strings.Trim(*f.OnCall, " ") == "" {
		return errors.New("On Call can't be empty")
	}
	return nil
}

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
