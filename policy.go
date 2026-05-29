package main

import "errors"

type Action string

const (
	ActionCreateIncident Action = "incidents:create"
	ActionViewIncident   Action = "incidents:view"
	ActionListIncident   Action = "incidents:list"
	ActionAddEntry       Action = "incidents:add_entry"

	ActionUpdateIncident Action = "incidents:update"
)

var ErrForbidden = errors.New("forbidden")

func AuthorizeIncidentAction(u UserContext, inc Incident, action Action) error {
	switch action {
	case ActionUpdateIncident, ActionAddEntry:
		if u.Role == "admin" || u.Username == inc.OnCall {
			return nil
		}
		return ErrForbidden
	default:
		return nil
	}
}
