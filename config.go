package main

import (
	"log"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

type Config struct {
	Port             string // default "8080",   env: HANDOFF_PORT
	LogLevel         string // default "info",   env: HANDOFF_LOG_LEVEL
	Environment      string // default "development", env: HANDOFF_ENV
	ConnectionString string // default "", env HANDOFF_CONNECT_STRING
	DatabaseName     string
	JWT_SECRET       string
}

func loadConfig() Config {
	godotenv.Load()
	config := Config{
		Port:             envOr("HANDOFF_PORT", "8080"),
		LogLevel:         envOr("HANDOFF_LOG_LEVEL", "info"),
		Environment:      envOr("HANDOFF_ENV", "development"),
		ConnectionString: envOr("HANDOFF_CONNECT_STRING", ""),
		DatabaseName:     envOr("HANDOFF_DB", "incident_tracker"),
		JWT_SECRET:       envOr("JWT_SECRET", ""),
	}
	if len(config.JWT_SECRET) == 0 {
		log.Fatalln("JWT_SECRET empty")
	}
	return config
}

func envOr(envKey string, defaultValue string) string {
	envValue := os.Getenv(envKey)
	if envValue == "" {
		return defaultValue
	}
	return envValue
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

const (
	timeout = time.Duration(5 * time.Second)
)

const (
	CollectionIncidents = "incidents"
	CollectionCounters  = "counters"
)

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
const userContextKey = "user"
const incidentIDPrefix = "INC-"
const entryIDPrefix = "TLE-"
const OnCallEntryPrefix = "ONc-"
const UserPrefix = "Usr-"
