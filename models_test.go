package main

import (
	"errors"
	"testing"
)

func TestCreateIncidentRequest_Validate(t *testing.T) {
	valid := func() CreateIncidentRequest {
		return CreateIncidentRequest{
			Title: "outage", Service: "api", Severity: "SEV1", OpenedBy: "anh",
		}
	}

	t.Run("valid request", func(t *testing.T) {
		r := valid()
		if err := r.Validate(); err != nil {
			t.Errorf("expected no error, got %v", err.Error())
		}
	})

	t.Run("empty title", func(t *testing.T) {
		r := valid()
		r.Title = ""
		if !errors.Is(r.Validate(), ErrNoTitle) {
			t.Error("expected ErrNoTitle")
		}
	})

	t.Run("whitespace title", func(t *testing.T) {
		r := valid()
		r.Title = "   "
		if !errors.Is(r.Validate(), ErrNoTitle) {
			t.Error("expected ErrNoTitle")
		}
	})

	t.Run("empty service", func(t *testing.T) {
		r := valid()
		r.Service = ""
		if !errors.Is(r.Validate(), ErrNoService) {
			t.Error("expected ErrNoService")
		}
	})

	t.Run("invalid severity", func(t *testing.T) {
		r := valid()
		r.Severity = "SEV4"
		if !errors.Is(r.Validate(), ErrInvalidSeverity) {
			t.Error("expected ErrInvalidSeverity")
		}
	})

	t.Run("empty opened_by", func(t *testing.T) {
		r := valid()
		r.OpenedBy = ""
		if !errors.Is(r.Validate(), ErrOpenedBy) {
			t.Error("expected ErrOpenedBy")
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		r := CreateIncidentRequest{
			Title: "  outage  ", Service: "  api  ", Severity: "  SEV1  ", OpenedBy: "  anh  ",
		}
		r.Validate()
		if r.Title != "outage" {
			t.Errorf("Title not trimmed: %q", r.Title)
		}
		if r.Service != "api" {
			t.Errorf("Service not trimmed: %q", r.Service)
		}
		if r.OpenedBy != "anh" {
			t.Errorf("OpenedBy not trimmed: %q", r.OpenedBy)
		}
	})
}

func TestTimelineEntry_Validate(t *testing.T) {
	valid := func() TimelineEntry {
		return TimelineEntry{Author: "anh", Type: OBSERVATION, Text: "cpu high"}
	}

	t.Run("valid entry", func(t *testing.T) {
		e := valid()
		if err := e.Validate(); err != nil {
			t.Errorf("expected no error, got %v", err.Error())
		}
	})

	t.Run("empty author", func(t *testing.T) {
		e := valid()
		e.Author = "   "
		if !errors.Is(e.Validate(), ErrNoAuthor) {
			t.Error("expected ErrNoAuthor")
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		e := valid()
		e.Type = "active"
		if !errors.Is(e.Validate(), ErrBadEntryType) {
			t.Error("expected ErrBadEntryType")
		}

		e = valid()
		e.Type = "invalid"
		if !errors.Is(e.Validate(), ErrBadEntryType) {
			t.Error("expected ErrBadEntryType")
		}
	})

	t.Run("empty text", func(t *testing.T) {
		e := valid()
		e.Text = ""
		if !errors.Is(e.Validate(), ErrNoText) {
			t.Error("expected ErrNoText")
		}
	})

	t.Run("all valid entry types", func(t *testing.T) {
		for _, typ := range []string{OBSERVATION, ACTION, DISCOVERY, OPEN_QUESTION, STATE_CHANGE} {
			e := valid()
			e.Type = typ
			if err := e.Validate(); err != nil {
				t.Errorf("type %s should be valid, got %v", typ, err)
			}
		}
	})
}

func TestIncidentFilter_Validate(t *testing.T) {
	t.Run("empty filter valid", func(t *testing.T) {
		f := IncidentFilter{}
		if err := f.Validate(); err != nil {
			t.Errorf("expected no error, got %v", err.Error())
		}
	})

	t.Run("valid status", func(t *testing.T) {
		for _, s := range []string{TRIGGERED, ACKNOWLEDGED, INVESTIGATING, MITIGATED, RESOLVED, "active"} {
			f := IncidentFilter{Status: s}
			if err := f.Validate(); err != nil {
				t.Errorf("status %s should be valid, got %v", s, err)
			}
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		f := IncidentFilter{Status: "abc"}
		if !errors.Is(f.Validate(), ErrBadIncidentStatus) {
			t.Error("expected ErrBadIncidentStatus")
		}
	})

	t.Run("service passes through", func(t *testing.T) {
		f := IncidentFilter{Service: "api"}
		if err := f.Validate(); err != nil {
			t.Errorf("expected no error, got %v", err.Error())
		}
	})

	t.Run("empty service not fail", func(t *testing.T) {
		f := IncidentFilter{Service: "  "}
		if err := f.Validate(); err != nil {
			t.Errorf("expected no error, got %v", err.Error())
		}
	})
}

func TestIncidentUpdate_Validate(t *testing.T) {
	t.Run("valid status", func(t *testing.T) {
		u := IncidentUpdate{Status: new(RESOLVED)}
		if err := u.Validate(); err != nil {
			t.Errorf("expected no error, got %v", err.Error())
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		u := IncidentUpdate{Status: new("active")}
		if !errors.Is(u.Validate(), ErrBadIncidentStatus) {
			t.Error("expected ErrBadIncidentStatus")
		}
	})

	t.Run("valid severity", func(t *testing.T) {
		u := IncidentUpdate{Severity: new(SEV2)}
		if err := u.Validate(); err != nil {
			t.Errorf("expected no error, got %v", err.Error())
		}
	})

	t.Run("invalid severity", func(t *testing.T) {
		u := IncidentUpdate{Severity: new("SEV9")}
		if !errors.Is(u.Validate(), ErrInvalidSeverity) {
			t.Error("expected ErrInvalidSeverity")
		}
	})

	t.Run("valid on_call", func(t *testing.T) {
		u := IncidentUpdate{OnCall: new("bernd")}
		if err := u.Validate(); err != nil {
			t.Errorf("expected no error, got %v", err.Error())
		}
	})

	t.Run("empty on_call", func(t *testing.T) {
		u := IncidentUpdate{OnCall: new("")}
		if !errors.Is(u.Validate(), ErrOnCall) {
			t.Error("expected ErrOnCall")
		}
	})

	t.Run("all fields nil", func(t *testing.T) {
		u := IncidentUpdate{}
		if !errors.Is(u.Validate(), ErrBadRequest) {
			t.Error("expected ErrBadRequest")
		}
	})

	t.Run("set many fields", func(t *testing.T) {
		u := IncidentUpdate{
			Status:   new(RESOLVED),
			Severity: new(SEV2),
			OnCall:   new("bernd1111"),
		}
		u.Validate()
		if *u.Status != RESOLVED {
			t.Errorf("expected RESOLVED, got %v", u.Status)
		}
		if *u.Severity != SEV2 {
			t.Errorf("expected SEV, got %v", u.Severity)
		}
		if *u.OnCall != "bernd1111" {
			t.Errorf("expected bernd1111, get %v", u.OnCall)
		}
	})
}

func TestFeatureFlag_Validate(t *testing.T) {
	valid := func() FeatureFlag {
		return FeatureFlag{
			Name:     "dark_mode",
			Enabled:  true,
			Rollout:  50,
			Variants: []string{"control", "variant_a"},
		}
	}

	t.Run("valid flag", func(t *testing.T) {
		f := valid()
		if err := f.Validate(); err != nil {
			t.Errorf("expected no error, got %v", err.Error())
		}
	})

	t.Run("empty name", func(t *testing.T) {
		f := valid()
		f.Name = ""
		if err := f.Validate(); err == nil {
			t.Error("expected error")
		}
	})

	t.Run("rollout negative", func(t *testing.T) {
		f := valid()
		f.Rollout = -1
		if err := f.Validate(); err == nil {
			t.Error("expected error")
		}
	})

	t.Run("rollout over 100", func(t *testing.T) {
		f := valid()
		f.Rollout = 101
		if err := f.Validate(); err == nil {
			t.Error("expected error")
		}
	})

	t.Run("rollout boundary 0", func(t *testing.T) {
		f := valid()
		f.Rollout = 0
		if err := f.Validate(); err != nil {
			t.Errorf("expected no error, got %v", err.Error())
		}
	})

	t.Run("rollout boundary 100", func(t *testing.T) {
		f := valid()
		f.Rollout = 100
		if err := f.Validate(); err != nil {
			t.Errorf("expected no error, got %v", err.Error())
		}
	})

	t.Run("nil variants", func(t *testing.T) {
		f := valid()
		f.Variants = nil
		if err := f.Validate(); err == nil {
			t.Error("expected error")
		}
	})

	t.Run("0 variants", func(t *testing.T) {
		f := valid()
		f.Variants = []string{}
		if err := f.Validate(); err == nil {
			t.Error("expected error")
		}
	})

	t.Run("empty variant string", func(t *testing.T) {
		f := valid()
		f.Variants = []string{"control", ""}
		if err := f.Validate(); err == nil {
			t.Error("expected error")
		}
	})

	t.Run("duplicate variants", func(t *testing.T) {
		f := valid()
		f.Variants = []string{"control", "control"}
		if err := f.Validate(); err == nil {
			t.Error("expected error")
		}
	})
}

func TestFeatureFlagUpdate_Validate(t *testing.T) {
	t.Run("valid enabled only", func(t *testing.T) {
		u := FeatureFlagUpdate{Name: "flag1", Enabled: new(true)}
		if err := u.Validate(); err != nil {
			t.Errorf("expected no error, got %v", err.Error())
		}
	})

	t.Run("valid rollout only", func(t *testing.T) {
		u := FeatureFlagUpdate{Name: "flag1", Rollout: new(50)}
		if err := u.Validate(); err != nil {
			t.Errorf("expected no error, got %v", err.Error())
		}
	})

	t.Run("empty name", func(t *testing.T) {
		u := FeatureFlagUpdate{Name: "", Enabled: new(true)}
		if err := u.Validate(); err == nil {
			t.Error("expected error")
		}
	})

	t.Run("both nil", func(t *testing.T) {
		u := FeatureFlagUpdate{Name: "flag1"}
		if err := u.Validate(); err == nil {
			t.Error("expected error")
		}
	})

	t.Run("rollout negative", func(t *testing.T) {
		u := FeatureFlagUpdate{Name: "flag1", Rollout: new(-1)}
		if err := u.Validate(); err == nil {
			t.Error("expected error")
		}
	})

	t.Run("rollout over 100", func(t *testing.T) {
		u := FeatureFlagUpdate{Name: "flag1", Rollout: new(101)}
		if err := u.Validate(); err == nil {
			t.Error("expected error")
		}
	})

	t.Run("rollout boundary 0", func(t *testing.T) {
		u := FeatureFlagUpdate{Name: "flag1", Rollout: new(0)}
		if err := u.Validate(); err != nil {
			t.Errorf("expected no error, got %v", err.Error())
		}
	})

	t.Run("rollout boundary 100", func(t *testing.T) {
		u := FeatureFlagUpdate{Name: "flag1", Rollout: new(100)}
		if err := u.Validate(); err != nil {
			t.Errorf("expected no error, got %v", err.Error())
		}
	})
}
