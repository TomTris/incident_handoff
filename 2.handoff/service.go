package main

import (
	"time"
)

func buildHandoffBrief(inc Incident, flagStore *FlagStore, userID string) HandoffBrief {
	actionsList := []TimelineEntry{}
	openQuestionsList := []TimelineEntry{}
	author := ""
	handoffCount := 0
	for _, entry := range inc.Entries {
		if author != entry.Author {
			author = entry.Author
			handoffCount++
		}
		switch entry.Type {
		case ACTION:
			actionsList = append(actionsList, entry)
		case OPEN_QUESTION:
			openQuestionsList = append(openQuestionsList, entry)
		}
	}

	if handoffCount != 0 {
		handoffCount--
	}

	brief := HandoffBrief{
		Severity:      inc.Severity,
		Status:        inc.Status,
		Service:       inc.Service,
		ElapsedMinute: int(time.Since(inc.CreatedAt).Minutes()),
		TotalEntry:    len(inc.Entries),
		TakenActions:  len(actionsList),
		OpenQuestion:  len(openQuestionsList),
		HandoffCount:  handoffCount,
		CreatedAt:     inc.CreatedAt,
	}

	flagAnswer, err := flagStore.Evaluate("detailed_handoff_brief", userID)
	if err == nil && flagAnswer.InRollout == true && *flagAnswer.Variant == "detailed" {
		brief.OpenQuestionList = &openQuestionsList
		brief.TakenActionsList = &actionsList
	}
	return brief
}
