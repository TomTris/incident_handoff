package main

type Config struct {
	Port        string // default "8080",   env: HANDOFF_PORT
	LogLevel    string // default "info",   env: HANDOFF_LOG_LEVEL
	Environment string // default "development", env: HANDOFF_ENV
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
	"":            true,
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

// type ConfigOption func(*Config)

// func WithPort(p string) ConfigOption {
// 	return func(c *Config) { c.Port = p }
// }

// func WithLogLevel(l string) ConfigOption {
// 	return func(c *Config) { c.LogLevel = l }
// }

// func WithEnvironment(e string) ConfigOption {
// 	return func(c *Config) { c.Environment = e }
// }

// func NewConfig(confOpts ...ConfigOption) Config {
// 	newConfig := Config{
// 		Port:        "8080",
// 		LogLevel:    "info",
// 		Environment: "development",
// 	}
// 	for _, configOpt := range confOpts {
// 		configOpt(&newConfig)
// 	}
// 	return newConfig
// }

// curl -s -X POST http://localhost:8080/incidents \
//   -H "Content-Type: application/json" \
//   -d '{"title":"order-service request drop","service":"order-service","severity":"SEV1","opened_by":"Anh Nguyen"}'

// curl -s -X POST http://localhost:8080/incidents/inc-1/entries \
//   -H "Content-Type: application/json" \
//   -d '{"author":"Anh Nguyen","type":"observation","text":"Connection pool exhaustion. Pool at 100/100."}'
