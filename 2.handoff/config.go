package main

import "os"

type Config struct {
	Port        string // default "8080",   env: HANDOFF_PORT
	LogLevel    string // default "info",   env: HANDOFF_LOG_LEVEL
	Environment string // default "development", env: HANDOFF_ENV
}

func loadConfig() Config {
	return Config{
		Port:        envOr("HANDOFF_PORT", "8080"),
		LogLevel:    envOr("HANDOFF_LOG_LEVEL", "info"),
		Environment: envOr("HANDOFF_ENV", "development"),
	}
}

func envOr(envKey string, defaultValue string) string {
	envValue := os.Getenv(envKey)
	if envValue == "" {
		return defaultValue
	}
	return envValue
}

// Incident Severity
const (
	SEV1 = "SEV1"
	SEV2 = "SEV2"
	SEV3 = "SEV3"
)

var IncidentSeverity = map[string]bool{
	SEV1: true,
	SEV2: true,
	SEV3: true,
}

// Incident status
const (
	TRIGGERED     = "triggered"
	ACKNOWLEDGED  = "acknowledged"
	INVESTIGATING = "investigating"
	MITIGATED     = "mitigated"
	RESOLVED      = "resolved"
)

var IncidentStatus = map[string]bool{
	TRIGGERED:     true,
	ACKNOWLEDGED:  true,
	INVESTIGATING: true,
	MITIGATED:     true,
	RESOLVED:      true,
	"active":      true,
}

// Entry type
const (
	OBSERVATION   = "observation"
	ACTION        = "action"
	DISCOVERY     = "discovery"
	OPEN_QUESTION = "open_question"
	STATE_CHANGE  = "state_change"
)

var validEntryTypes = map[string]bool{
	OBSERVATION:   true,
	ACTION:        true,
	DISCOVERY:     true,
	OPEN_QUESTION: true,
	STATE_CHANGE:  true,
}

const requestIDKey = "request_id"
const incidentIDPrefix = "inc-"
const entryIDPrefix = "ent-"
