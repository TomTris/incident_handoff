# PART 2 — Backend Service (Phases 5–9)

This track builds Handoff's backend: a production-grade Go API with database persistence, WebSocket, authentication, observability, and feature flags. By the end, the backend is complete, tested, and instrumented.

---

# PHASE 5 — Production Go HTTP Service

> **Why this phase matters**
> Phases 3 and 4 taught you concurrency and performance. But a real Go service isn't goroutines in a vacuum — it's an HTTP server that receives requests, validates input, routes to handlers, logs what happened, and returns structured responses. `net/http` is the backbone of every Go microservice. The patterns here — middleware chains, structured errors, graceful shutdown, configuration from environment — are the same patterns used in production at companies running Go at scale. This is the phase where your code starts looking like code that ships.

---

## Challenge 5.1 — Build the Handoff Incident API
### `🔴 Intermediate → Advanced`
**🕐 Expected duration: 28–35 hours**

### 1. Context

You're building the backend for Handoff — the on-call incident handoff platform. This API is the foundation. Everything else (database, WebSocket, frontend, metrics) plugs into this. If the API is poorly designed, every future phase suffers. If it's well designed, every future phase is easy.

The API must handle: creating incidents, logging timestamped entries to an incident's timeline, changing incident state (severity, status), generating handoff briefs, and querying incidents with filters. It must also behave like a production service: structured logging, middleware, config from environment, graceful shutdown.

### 2. Goal

Build a complete REST API for Handoff using Go's `net/http` standard library, with middleware, structured error responses, environment-based configuration, and graceful shutdown.

### 3. Scope

**Data model (in-memory for now — Phase 6 adds database persistence):**

```go
type Incident struct {
    ID          string          `json:"id"`
    Title       string          `json:"title"`
    Service     string          `json:"service"`
    Severity    string          `json:"severity"`    // SEV1, SEV2, SEV3
    Status      string          `json:"status"`      // triggered, acknowledged, investigating, mitigated, resolved
    OpenedBy    string          `json:"opened_by"`
    OnCall      string          `json:"on_call"`     // Default: set to OpenedBy on creation
    CreatedAt   time.Time       `json:"created_at"`
    UpdatedAt   time.Time       `json:"updated_at"`
    Entries     []TimelineEntry `json:"entries"`
}

type TimelineEntry struct {
    ID        string    `json:"id"`
    Time      time.Time `json:"time"`
    Author    string    `json:"author"`
    Type      string    `json:"type"`  // observation, action, discovery, open_question, state_change
    Text      string    `json:"text"`
}
```

**Endpoints:**

| Method | Path | Description |
|---|---|---|
| `POST` | `/incidents` | Create a new incident |
| `GET` | `/incidents` | List all incidents (filter: `?status=active&service=order-service`) |
| `GET` | `/incidents/:id` | Get one incident with full timeline |
| `POST` | `/incidents/:id/entries` | Add a timeline entry |
| `PATCH` | `/incidents/:id` | Update severity, status or on call |
| `GET` | `/incidents/:id/handoff` | Auto-generated handoff brief |
| `GET` | `/healthz` | Health check (returns 200 + `{"status":"ok"}`) |

**Handoff brief logic (`GET /incidents/:id/handoff`):**

This endpoint auto-generates the structured summary an incoming on-call engineer reads. It returns:
- Current severity and status
- Affected service
- All entries of type `action` (what was done)
- All entries of type `open_question` (where to start)
- Time elapsed since incident opened
- Number of total entries

This is the core product logic. The frontend will render this directly.

**Middleware chain:**

Build 3 middleware functions that wrap your handlers:
- `LoggingMiddleware` — logs every request: method, path, duration, request_id. Use `log/slog`.
- `CORSMiddleware` — sets `Access-Control-Allow-Origin`, `Allow-Methods`, `Allow-Headers`. Required for Phase 11 when the Vue frontend connects.
- `RequestIDMiddleware` — generates a UUID for each request, adds it to the response header `X-Request-ID`, and includes it in all log lines for that request.

Middleware must be composable:
```go
handler := RequestIDMiddleware(LoggingMiddleware(CORSMiddleware(router)))
```

**Structured error responses:**

Every error returns JSON, never plain text:
```json
{
    "error": {
        "code": "INCIDENT_NOT_FOUND",
        "message": "No incident with ID inc-9999",
        "request_id": "a3f9c012-..."
    }
}
```

Define error codes: `INCIDENT_NOT_FOUND`, `INVALID_SEVERITY`, `INVALID_STATUS`, `MISSING_FIELD`, `INVALID_ENTRY_TYPE`. Return appropriate HTTP status codes:
- `200` — successful GET
- `201` — successful creation (POST)
- `204` — successful update with no body (PATCH)
- `400` — bad request (validation failure)
- `404` — resource not found
- `405` — method not allowed
- `500` — internal server error

**Input validation:**
- Severity must be one of: `SEV1`, `SEV2`, `SEV3`
- Status must be one of: `triggered`, `acknowledged`, `investigating`, `mitigated`, `resolved`
- Entry type must be one of: `observation`, `action`, `discovery`, `open_question`, `state_change`
- Title, service, severity, opened_by are required on creation (return `MISSING_FIELD` if absent)
- Status transitions must be valid: `resolved` incidents cannot accept new entries

**Configuration from environment:**
```go
type Config struct {
    Port        string // default "8080",   env: HANDOFF_PORT
    LogLevel    string // default "info",   env: HANDOFF_LOG_LEVEL
    Environment string // default "development", env: HANDOFF_ENV
}
```
Read from env vars. Fall back to defaults. No config files — this is the 12-factor app approach that all containerized services use.

**Graceful shutdown:**

Listen for `SIGTERM` and `SIGINT`. When received:
1. Stop accepting new connections
2. Wait for in-flight requests to finish (with a 10-second timeout)
3. Log "server shut down gracefully" and exit

This is required for Kubernetes — when a pod is terminated, it sends `SIGTERM` first. Services that don't handle it get hard-killed, dropping in-flight requests.

**Store layer:**
```go
type Store interface {
    CreateIncident(ctx context.Context, inc Incident) (Incident, error)
    GetIncident(ctx context.Context, id string) (Incident, error)
    ListIncidents(ctx context.Context, filter IncidentFilter) ([]Incident, error)
    UpdateIncident(ctx context.Context, id string, update IncidentUpdate) error
    AddEntry(ctx context.Context, incidentID string, entry TimelineEntry) error
}
```
Implement `MemoryStore` using `sync.RWMutex` + `map[string]Incident`. The interface exists so Phase 6 can swap in a database-backed store without changing any handler code. This is the same interface-driven design from Phase 2 — now used for real.

**Project structure:**
```
handoff/
├── main.go                 // wiring, config, server start, shutdown
├── config.go               // Config struct, env reading
├── router.go               // route registration
├── handler_incidents.go     // incident HTTP handlers
├── handler_health.go        // health endpoint
├── middleware.go            // logging, CORS, request ID
├── store.go                 // Store interface
├── store_memory.go          // in-memory implementation
├── models.go               // Incident, TimelineEntry structs
├── errors.go               // structured error types and helpers
└── go.mod
```

You can add more files if needed.

### 4. Expected Output

```bash
$ HANDOFF_PORT=8080 go run .
2026-05-10T08:00:01Z INF server starting port=8080 env=development

$ curl -s -X POST http://localhost:8080/incidents \
  -H "Content-Type: application/json" \
  -d '{"title":"order-service request drop","service":"order-service","severity":"SEV1","opened_by":"Anh Nguyen"}' | jq .
{
  "id": "inc-001",
  "title": "order-service request drop",
  "service": "order-service",
  "severity": "SEV1",
  "status": "triggered",
  "opened_by": "Anh Nguyen",
  "on_call": "Anh Nguyen",
  "created_at": "2026-05-10T08:01:00Z",
  "updated_at": "2026-05-10T08:01:00Z",
  "entries": []
}

$ curl -s -X POST http://localhost:8080/incidents/inc-001/entries \
  -H "Content-Type: application/json" \
  -d '{"author":"Anh Nguyen","type":"observation","text":"Connection pool exhaustion. Pool at 100/100."}' | jq .
{
  "id": "ent-001",
  "time": "2026-05-10T08:02:00Z",
  "author": "Anh Nguyen",
  "type": "observation",
  "text": "Connection pool exhaustion. Pool at 100/100."
}

$ curl -s http://localhost:8080/incidents/inc-999 | jq .
{
  "error": {
    "code": "INCIDENT_NOT_FOUND",
    "message": "No incident with ID inc-999",
    "request_id": "a3f9c012-beef-..."
  }
}

# Server log output (structured):
2026-05-10T08:01:00Z request completed method=POST path=/incidents duration=1.2ms request_id=a3f9c012
2026-05-10T08:02:00Z request completed method=POST path=/incidents/inc-001/entries duration=0.8ms request_id=b4e1d034
2026-05-10T08:03:00Z request completed method=GET path=/incidents/inc-999 duration=0.1ms request_id=c5f2e056
```

### 5. Hints & Knowledge

- `http.NewServeMux()` (Go 1.22+) supports method+path patterns: `mux.HandleFunc("POST /incidents", handler)`.
- `log/slog` — Go's built-in structured logger. `slog.Info("msg", "key", value)` outputs JSON or text depending on handler.
- `json.NewDecoder(r.Body).Decode(&input)` — decode request body. Always check the error.
- `w.WriteHeader(http.StatusCreated)` must be called BEFORE `w.Write()` or `json.Encode` — once bytes are written, the status code is locked at 200.
- `os.Signal`, `signal.Notify`, `context.WithTimeout` — the building blocks of graceful shutdown.
- `http.Server.Shutdown(ctx)` — stops accepting new connections, waits for existing ones to finish.
- `sync.RWMutex` — use `RLock()` for reads, `Lock()` for writes. Multiple readers can hold `RLock` simultaneously. One writer blocks everyone.
- `os.Getenv("KEY")` returns `""` if unset. Write a helper: `envOr(key, fallback string) string`.

### 6. Sources

- `net/http` routing (Go 1.22+): https://go.dev/blog/routing-enhancements
- `log/slog`: https://pkg.go.dev/log/slog
- Graceful shutdown: https://pkg.go.dev/net/http#Server.Shutdown
- `sync.RWMutex`: https://pkg.go.dev/sync#RWMutex
- HTTP status codes: https://developer.mozilla.org/en-US/docs/Web/HTTP/Status
- UUID generation: https://pkg.go.dev/github.com/google/uuid
- 12-factor app config: https://12factor.net/config

### 7. Common Mistakes to Avoid

- Writing `w.WriteHeader(201)` after `json.NewEncoder(w).Encode(data)` — the 201 is silently ignored, response goes out as 200.
- Using `sync.Mutex` instead of `sync.RWMutex` — reads block each other unnecessarily. You learned this in Phase 3, now apply it.
- Not closing `r.Body` — `net/http` handles this for you, but if you wrap the body in a reader, you own the close.
- Putting route registration inside handlers — define all routes in one place (`router.go`).
- Not validating input before storing — bad data in the store causes bugs that surface much later, far from the cause.
- Returning plain text errors — `fmt.Fprintf(w, "not found")` is not an API error response. Always return JSON with an error code.
- Hardcoding `localhost:8080` — use config. Containers and Kubernetes set different ports via environment variables.
- Not passing `context.Context` through the store interface — you'll need it in Phase 6 for database query cancellation.

### 8. Checklist

```
[ ] go run . — server starts, logs "server starting"
[ ] POST /incidents — creates incident, returns 201
[ ] GET /incidents/:id — returns incident with entries
[ ] POST /incidents/:id/entries — adds entry, returns 201
[ ] PATCH /incidents/:id — updates severity/status
[ ] GET /incidents/:id/handoff — returns structured brief
[ ] GET /healthz — returns 200
[ ] GET /incidents/nonexistent — returns 404 with structured error JSON
[ ] POST /incidents with missing title — returns 400 with MISSING_FIELD
[ ] POST /incidents/:id/entries on resolved incident — returns 400
[ ] All requests logged with method, path, duration, request_id
[ ] CORS headers present on all responses
[ ] Server shuts down gracefully on Ctrl+C (SIGINT)
[ ] go vet ./... — zero warnings
```

### 9. Knowledge Gained

```
✅ net/http server — routing, handlers, request/response cycle
✅ HTTP semantics — correct status codes for each operation
✅ Middleware pattern — composable, reusable request processing
✅ Structured logging — log/slog with key-value output
✅ Structured error responses — consistent API error format
✅ Input validation — rejecting bad data at the boundary
✅ Configuration from environment — 12-factor app principle
✅ Graceful shutdown — SIGTERM handling for container orchestration
✅ Interface-driven store — swappable persistence layer
✅ sync.RWMutex — concurrent-safe reads and writes
✅ Project structure — separation of concerns in a Go service
```

---

# PHASE 5.Test — Go Testing Fundamentals

> **Why this phase matters**
> You've just built your first production Go service. It works — you tested it manually with `curl`. But manual testing doesn't scale. When you add a database in Phase 6, WebSocket in Phase 7, and auth in Phase 9, how do you know Phase 5's handlers still work? You don't — unless you have automated tests. This phase introduces Go testing while Phase 5's code is fresh in your memory. From here forward, you're expected to write tests alongside your code in every phase.

---

## Challenge 5.Test — Test the Handoff API
### `🟡 Beginner → Intermediate`
**🕐 Expected duration: 8–12 hours**

### 1. Context

If you've never written a test before: a test is a function that calls your code with known inputs and checks whether the output matches what you expect. If it doesn't, the test fails and tells you exactly what went wrong. That's the entire concept. Everything below is technique for doing it well.

### 2. Goal

Write unit tests and HTTP handler tests for the Phase 5 API. Learn Go's testing tools: **`testing.T`**, **table-driven** tests, and **`httptest`**.

### 3. Scope

Note: Depending on your design, you may or may not have same functions as below. Be flexible and have similar tests for your own functions. You don't need to have the exact same functions testing exact same thing.

**Write your first test:**

Create `errors_test.go` next to `errors.go`. Write one test:

```go
func TestValidateSeverity(t *testing.T) {
    err := validateSeverity("SEV1")
    if err != nil {
        t.Errorf("validateSeverity(SEV1) returned error: %v", err)
    }
}
```

Run it: `go test ./... -v`. Watch it pass. Now add a case that should fail: `validateSeverity("SEV4")`. Check that it returns an error. You now know the mechanic.

**Table-driven tests:**

One test function, many cases:

```go
func TestValidateSeverity(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid SEV1", "SEV1", false},
        {"valid SEV2", "SEV2", false},
        {"valid SEV3", "SEV3", false},
        {"invalid SEV4", "SEV4", true},
        {"empty string", "", true},
        {"lowercase", "sev1", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateSeverity(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("validateSeverity(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
            }
        })
    }
}
```

Each row is a case. `t.Run` creates a named subtest. If one fails, the name tells you which.

**Tests to write:**

Unit tests (table-driven) for standalone functions, example:
- `validateSeverity` — valid values, invalid values, empty
- `validateStatus` — valid values, invalid values
- `validateEntryType` — valid values, invalid values
- Handoff brief generation — given known entries, assert correct action count, open question count

HTTP handler tests using `httptest`:
- `POST /incidents` with valid input → 201 + incident in response
- `POST /incidents` with missing title → 400 + `MISSING_FIELD` error code
- `GET /incidents/:id` with valid ID → 200 + incident
- `GET /incidents/:id` with nonexistent ID → 404 + `INCIDENT_NOT_FOUND`
- `POST /incidents/:id/entries` on a resolved incident → 400

`httptest` lets you test handlers without starting a real server:
```go
req := httptest.NewRequest("POST", "/incidents", strings.NewReader(body))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, req)
// rec.Code has the status, rec.Body has the response
```

**Run:**
```bash
go test ./... -v          # all tests, verbose
go test ./... -race       # with race detector
go test ./... -cover      # with coverage percentage
```

### 4. Hints & Knowledge

- Test files live next to the code: `errors.go` → `errors_test.go`, `handler_incidents.go` → `handler_incidents_test.go`.
- `t.Errorf` reports failure but continues running. `t.Fatalf` reports and stops. Use `Fatalf` for setup failures (can't create test data). Use `Errorf` for assertions (check all of them).
- `httptest.NewRecorder()` captures the HTTP response: `rec.Code` for status, `rec.Body.Bytes()` for the body.
- `httptest.NewRequest("GET", "/path", nil)` creates a fake request. No server needed.
- `json.Unmarshal(rec.Body.Bytes(), &result)` — decode the response to check fields.
- `go test -cover` shows coverage percentage. `go test -coverprofile=cover.out && go tool cover -html=cover.out` generates an HTML report showing which lines are tested.

### 5. Sources

- Go testing: https://pkg.go.dev/testing
- Table-driven tests: https://go.dev/wiki/TableDrivenTests
- `httptest`: https://pkg.go.dev/net/http/httptest
- Coverage: https://go.dev/blog/cover

### 6. Common Mistakes to Avoid

- Testing implementation details instead of behavior — test what the function returns, not how it computes it.
- Writing one assertion per test function — use table-driven tests for multiple cases.
- Not testing error paths — the "missing title returns 400" test is as important as "valid input returns 201."
- Forgetting `-race` — concurrent bugs in the store only surface under the race detector.

### 7. Checklist

```
[ ] go test ./... — all tests pass
[ ] go test -race ./... — zero race conditions
[ ] go test -cover ./... — >75% coverage on handler and validation code
[ ] Table-driven tests for all validation functions
[ ] httptest tests for at least 5 handler cases (mix of success and error)
```

### 8. Knowledge Gained

```
✅ Go testing fundamentals — testing.T, test files, go test
✅ Table-driven tests — the idiomatic Go test pattern
✅ httptest — testing HTTP handlers without a real server
✅ Race detection — go test -race
✅ Coverage measurement — go test -cover
```

**From this point forward:** when you build Phases 6–9, write tests alongside your code. You now have the tools. Don't wait — test each new handler and function as you write it. `go test -cover ./...` >75%. Phase 13 will verify your cumulative coverage.

---

# Database Choice

Before Phase 6, you must choose a database. This is a real engineering decision with real tradeoffs. Understand them before you commit.

**Three viable options for Handoff:**

| Database | Model | Strengths for Handoff | Weaknesses for Handoff |
|---|---|---|---|
| **PostgreSQL** | Relational (tables, rows, SQL) | Strongest data integrity — foreign keys enforce relationships between engineers, incidents, and entries. Powerful query language. `LISTEN/NOTIFY` provides built-in pub/sub for real-time updates. Best technical fit for Handoff's relational data. | Requires learning SQL. Requires schema migrations when the data model changes. The Incident struct with embedded entries needs to be split into two tables (incidents + entries) joined by foreign key. |
| **MongoDB** | Document (JSON-like documents, no schema) | The Incident struct from Phase 5 — with its embedded `[]TimelineEntry` — maps directly to a database document without restructuring. No SQL, no migrations, no ORM. Lowest barrier to start. | Relationships between entities (who is on-call for which incident) must be managed in application code, not the database. No referential integrity enforcement. Aggregation pipelines are less expressive than SQL for complex queries. |
| **SQLite** | Relational (embedded, file-based) | Zero infrastructure — no separate database process, no Docker service, just a file on disk. Full SQL support. Good for learning relational concepts without operational overhead. | Single-writer only — concurrent writes serialize through a single lock. Cannot run multiple API instances against the same database. No built-in real-time notification mechanism. Not viable for the multi-replica deployment in Phase 14. |

**Research your choice.** Read the Go driver documentation. Understand how your database handles: creating records, querying with filters, updating specific fields without overwriting the rest, and appending to a list. Phase 6 will ask you to implement these operations — the curriculum describes *what* each operation must accomplish, not *how* your specific database does it.

**This curriculum uses MongoDB.** The Incident struct from Phase 5 drops into MongoDB as-is — no table design, no foreign keys, no SQL. For someone learning Go, Vue, WebSocket, auth, metrics, Docker, and CI/CD simultaneously, removing the database restructuring step reduces cognitive load where it matters least. MongoDB is also a database I need professionally. That said, PostgreSQL is a stronger technical fit for Handoff's data model. The relationships between engineers, incidents, and handoffs are inherently relational. MongoDB pushes referential integrity into application code that PostgreSQL enforces at the database level. I chose convenience and career alignment over optimal data modeling. Know that tradeoff before you follow the same path.

If you choose a different database than this curriculum uses, every phase still works — the `Store` interface from Phase 5 is the abstraction boundary. Your handlers don't know or care what database sits behind it. You will need to translate the database-specific research on your own: find the Go driver, learn its API, map the requirements to your database's operations. That translation work is itself a valuable skill.

---

# PHASE 6 — Database Integration

> **Why this phase matters**
> In-memory stores disappear when the process restarts. Every production service persists data in a database. After this phase, Handoff survives restarts and can scale to multiple API instances sharing the same data — a requirement when Kubernetes runs more than one replica. The `Store` interface from Phase 5 pays off here: you implement a new store backed by a real database, swap it in, and zero handler code changes.

---

## Challenge 6.1 — Replace the In-Memory Store with a Database
### `🟠 Intermediate`
**🕐 Expected duration: 18–25 hours**

### 1. Context

Your Handoff API works — but all data lives in a `map` behind a mutex. Restart the server and everything is gone. In production, incident data is critical. It must survive restarts, be queryable across multiple service instances, and support concurrent access from many clients.

This challenge connects Go to your chosen database using its official Go driver and implements the same `Store` interface from Phase 5 backed by real persistence.

### 2. Goal

Implement a database-backed store satisfying the `Store` interface from Phase 5. Add schema/index management, connection handling, and efficient writes. Zero handler changes.

### 3. Scope

**Data mapping:**

Design how will you map data with your database choice.

**Serialization tags:**

Your Go structs have `json:"..."` tags for HTTP serialization. Your database driver likely needs its own tags to control how fields are named in the database. Without them, field names may not match what you expect.

- MongoDB: add `bson:"..."` tags
- PostgreSQL/SQLite with raw SQL: tags may not be needed if you map fields explicitly in queries
- PostgreSQL with sqlc or similar: follow the tool's conventions

Update your model structs in `models.go` to include whatever tags your database driver requires alongside the existing JSON tags.

**Initialize your database:** Your database needs to be ready before the API starts accepting traffic. Build an initialization function. Run it in `main.go` during startup.

**Store implementation:** Create your database store file (e.g., `store_db.go`) implementing the `Store` interface.

**Connection and configuration:** Add a database connection string to your `Config` struct, reading from an environment variable (e.g., `DATABASE_URL` or `MONGODB_URI`). Most database drivers manage connection pooling internally.

**Docker Compose for local development:**

Add your database as a service in `docker-compose.yml`:
- Use the official Docker image for your database
- Expose the default port
- Persist data with a named volume
- Add a health check so dependent services wait for the database to be ready

Run `docker compose up db` to start your database locally.

**Store swap in main.go:**

The payoff of the interface design from Phase 5: if the database connection string is set, create the database-backed store. If not, fall back to `MemoryStore`. Zero handler changes. One conditional decides the persistence layer.

### 4. Expected Output

```bash
$ docker compose up db -d
$ DATABASE_URL="<your-connection-string>" go run .
2026-05-12T10:00:00Z INF schema/indexes ensured
2026-05-12T10:00:00Z INF server starting port=8080 store=database

# Same curl commands as Phase 5 — identical responses.
# Restart the server. Data survives.

$ curl -s http://localhost:8080/incidents | jq '.[] | .title'
# Data still survives
```

### 5. Hints & Knowledge

**General:**
- Read your database driver's official documentation before writing code. Understand: how to connect, how to insert, how to query with filters, how to update specific fields, and how to handle "not found."
- `context.Context` — pass it through every database call. This enables query cancellation when HTTP requests are cancelled.
- Connection pooling is usually handled by the driver. You don't create a separate pool.

**If you chose MongoDB:**
- Driver: `go get go.mongodb.org/mongo-driver/v2`
- `bson.M` for unordered maps (filters, updates). `bson.D` for ordered maps (index keys where field order matters).
- `mongo.ErrNoDocuments` — the error returned when `FindOne` matches nothing.
- You need at least two update operators: one for setting specific fields, one for appending to an array. Read https://www.mongodb.com/docs/manual/reference/operator/update/
- `cursor.Close(ctx)` — always defer this after `Find()`.
- Configure your local MongoDB as a single-node replica set (add `--replSet rs0` to the Docker command). This is functionally identical to standalone but enables Change Streams and transactions if needed in later phases.

**If you chose PostgreSQL:**
- Driver: `go get github.com/jackc/pgx/v5` (recommended) or `database/sql` with `lib/pq`.
- Consider `sqlc` (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`) — generates type-safe Go code from SQL queries. Eliminates manual `rows.Scan` calls.
- `pgx.ErrNoRows` — the error returned when `QueryRow` matches nothing.
- Use `$1`, `$2` parameterized queries. Never concatenate user input into SQL strings.
- `LISTEN/NOTIFY` — PostgreSQL's built-in pub/sub. Useful in Phase 7 for real-time updates without polling.

**If you chose SQLite:**
- Driver: `go get modernc.org/sqlite` (pure Go, no CGO needed) or `go get github.com/mattn/go-sqlite3` (requires CGO).
- `sql.ErrNoRows` — the error from `QueryRow` when nothing matches.
- Use `WAL` mode for better concurrent read performance: `PRAGMA journal_mode=WAL;`
- SQLite locks the entire database on writes. Acceptable for Handoff's scale, but be aware.

### 6. Sources

**MongoDB:**
- Go driver: https://pkg.go.dev/go.mongodb.org/mongo-driver/v2/mongo
- Quickstart: https://www.mongodb.com/docs/drivers/go/current/quick-start/
- Update operators: https://www.mongodb.com/docs/manual/reference/operator/update/
- Indexes: https://www.mongodb.com/docs/manual/indexes/

**PostgreSQL:**
- pgx driver: https://pkg.go.dev/github.com/jackc/pgx/v5
- sqlc: https://docs.sqlc.dev/en/latest/
- PostgreSQL indexes: https://www.postgresql.org/docs/current/indexes.html
- LISTEN/NOTIFY: https://www.postgresql.org/docs/current/sql-notify.html

**SQLite:**
- modernc.org/sqlite: https://pkg.go.dev/modernc.org/sqlite
- SQLite in Go: https://www.golang.dk/articles/go-and-sqlite-in-the-cloud
- WAL mode: https://www.sqlite.org/wal.html

### 7. Common Mistakes to Avoid

- **Not adding database serialization tags** — fields serialize with wrong names, queries silently match nothing or map incorrectly.
- **Not creating indexes** — `ListIncidents` with a status filter does a full scan on every call. Unacceptable at any real scale.
- **Forgetting to handle the not-found case** — without this check, your handler returns 500 ("decode failed" / "no rows") instead of 404.
- **Initializing entries as nil instead of an empty collection** — downstream code breaks on null vs empty array distinction.
- **Not closing cursors/result sets** — every query that returns multiple results holds a server-side resource. Close it immediately after use.
- **Using the wrong write method** — one write method replaces the entire record, another updates only specified fields. Using the wrong one wipes data. Read your driver's docs carefully.
- **Not using parameterized queries (relational databases)** — string concatenation in queries is always wrong. Use placeholders.

### 8. Checklist

```
[ ] docker compose up db — database starts and is healthy
[ ] go run . with database connection string — schema/indexes ensured, server starts with store=database
[ ] All Phase 5 curl commands produce identical results
[ ] Restart server — data persists
[ ] go run . without connection string — falls back to memory store
[ ] go vet ./... — zero warnings
[ ] Code comment explaining data mapping choice and tradeoffs
```

### 9. Knowledge Gained

```
✅ Database Go driver — connection, queries, writes
✅ Data mapping — how Go structs map to your database's storage model
✅ Connection handling — driver-managed pooling
✅ Efficient writes — updating specific fields and appending to collections without replacing entire records
✅ Schema/index management — ensuring readiness on startup for query performance
✅ Interface swap — memory → database, zero handler changes
✅ Docker Compose for local database dependencies
```

---

# PHASE 7 — WebSocket & Real-Time

> **Why this phase matters**
> Handoff's core value is real-time shared awareness. When Engineer A logs a timeline entry, Engineer B must see it instantly — not on the next page refresh. WebSocket provides a persistent bidirectional connection between client and server. Go's goroutine model makes WebSocket servers natural to build — one goroutine per connection costs ~2KB, so you can hold thousands of connections on a single server. You (should) already understand WebSocket at the protocol level (TCP upgrade, frames, opcodes) from your networking background (Or just look it up to know the overview). Now you implement it in Go.

---

## Challenge 7.1 — Real-Time Incident Timeline
### `🟠 Intermediate`
**🕐 Expected duration: 20–30 hours**

### 1. Context

Right now, if two engineers are viewing the same incident, one adds an entry and the other sees nothing until they manually refresh. That defeats the purpose of a real-time handoff tool. This challenge adds WebSocket support so new timeline entries are broadcast instantly to all connected clients viewing the same incident.

### 2. Goal

Add a WebSocket endpoint to Handoff that broadcasts new timeline entries and state changes to all connected clients in real time, using the hub pattern.

### 3. Scope

**WebSocket endpoint:**

`GET /incidents/:id/ws` — upgrades the HTTP connection to WebSocket. The client subscribes to real-time updates for a specific incident.

**Connection management requirements:**

- Multiple clients can connect to the same incident simultaneously
- When a new timeline entry is added via `POST /incidents/:id/entries`, every connected client viewing that incident receives the entry instantly
- When an incident's severity or status changes, connected clients receive the update
- When a client disconnects, it must be cleaned up — no goroutine leaks, no failed writes to dead connections
- A slow client (one that can't keep up with broadcasts) must not block delivery to other clients — drop it
- The WebSocket handler must be safe for concurrent use — multiple clients connecting and disconnecting simultaneously

Design the data structures yourself. You need to solve three problems:
1. **Registry** — how does the server track which clients are watching which incident?
2. **Write safety** — the WebSocket library forbids concurrent writes to the same connection. How do you ensure only one goroutine writes?
3. **Lifecycle** — each connection needs to detect when the other side disappears (ping/pong). How do you structure the goroutines per client?

**Broadcast payloads:**

New entry:
```json
{
    "type": "new_entry",
    "incident_id": "inc-001",
    "entry": {
        "id": "ent-005",
        "time": "2026-05-10T08:15:00Z",
        "author": "Anh Nguyen",
        "type": "action",
        "text": "Rolled back deployment abc123."
    }
}
```

State change:
```json
{
    "type": "state_change",
    "incident_id": "inc-001",
    "field": "severity",
    "old_value": "SEV1",
    "new_value": "SEV2"
}
```

**Important limitation — single-process broadcasting:**

The hub pattern implemented here is process-local. The hub lives in one Go process's memory. This works when you run a single API instance. In Phase 14, if you deploy multiple replicas, a client connected to replica 1 will not receive broadcasts triggered by writes to replica 2. This is a known limitation. Production solutions include database-level change notifications (e.g., PostgreSQL `LISTEN/NOTIFY`, MongoDB Change Streams) or an external message broker (Redis pub/sub, NATS). This curriculum leaves the single-process hub as-is and documents the limitation. If you want to solve it, research how your database can notify your application of writes made by other instances.

If you want to implement cross-replica broadcasting, look [Bonus](#bonus) part at the end.

### 4. Expected Output

```bash
# Terminal 1: start the server
$ go run .
2026-05-10T10:00:00Z INF server starting port=8080

# Terminal 2: connect via wscat
$ npx wscat -c ws://localhost:8080/incidents/inc-001/ws
Connected

# Terminal 3: add an entry
$ curl -s -X POST http://localhost:8080/incidents/inc-001/entries \
  -H "Content-Type: application/json" \
  -d '{"author":"Anh Nguyen","type":"action","text":"Rolled back deployment."}'

# Terminal 2 receives instantly:
< {"type":"new_entry","incident_id":"inc-001","entry":{"id":"ent-005","time":"2026-05-10T10:01:00Z","author":"Anh Nguyen","type":"action","text":"Rolled back deployment."}}

# Server log:
2026-05-10T10:00:05Z INF websocket client connected incident_id=inc-001 clients=1
2026-05-10T10:01:00Z INF broadcast sent incident_id=inc-001 clients=1 type=new_entry
```

### 5. Hints & Knowledge

- Use `github.com/gorilla/websocket` — still the most widely used Go WebSocket library and recommended by offical `https://pkg.go.dev/golang.org/x/net/websocket?`. Install: `go get github.com/gorilla/websocket`.
- `websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}` — allow all origins during development. Restrict in production.
- Never write to a `*websocket.Conn` from multiple goroutines — that's why you need a design that funnels writes through a single goroutine per connection.
- Ping/pong keepalive: `conn.SetReadDeadline(time.Now().Add(60 * time.Second))` and handle `PongMessage` to detect dead clients. Without this, dead connections stay registered forever.
- `npx wscat -c ws://localhost:8080/incidents/inc-001/ws` — the easiest way to test WebSocket endpoints from the terminal. Install: `npm install -g wscat`.
- The connection manager must be created once in `main()` and passed to handlers. It is shared state.
- CORS middleware must skip WebSocket upgrade requests — check for `r.Header.Get("Upgrade") == "websocket"` and pass through.
- Study the gorilla/websocket chat example before starting: https://github.com/gorilla/websocket/tree/main/examples/chat — it demonstrates a well-known pattern for solving the three problems above. Understand the pattern, then implement your own version for Handoff's per-incident model.

**If stuck after 4+ hours on the design:**
The standard solution is called the "hub pattern." A central struct holds a map of incident_id → set of client connections. Each client gets a `send` channel and two goroutines: one for reading (detecting disconnects) and one for writing (draining the send channel). The hub broadcasts by iterating clients and pushing to their send channels. A buffered send channel (size 16) lets you detect slow clients — if the channel is full, close that client.

### 6. Sources

- gorilla/websocket: https://pkg.go.dev/github.com/gorilla/websocket
- gorilla/websocket chat example: https://github.com/gorilla/websocket/tree/main/examples/chat
- WebSocket protocol (RFC 6455): https://datatracker.ietf.org/doc/html/rfc6455
- wscat: https://github.com/websockets/wscat

### 7. Common Mistakes to Avoid

- Writing to `*websocket.Conn` from multiple goroutines — race condition, corrupted frames, panic. Always funnel writes through a single goroutine via a channel.
- Not removing disconnected clients from the hub — the hub grows indefinitely, broadcasts fail silently on dead connections, goroutines leak.
- Blocking `Broadcast()` on a slow client — if one client can't keep up, the entire broadcast loop stalls. Use a buffered `send` channel and drop clients that overflow.
- Not setting read/write deadlines — a client that disappears without a clean close leaves its goroutines hanging forever.
- Upgrading inside middleware-wrapped handlers without handling the hijacked connection — once the WebSocket upgrade succeeds, the HTTP response writer is gone. CORS headers, logging, etc. don't apply anymore.

### 8. Checklist

```
[ ] GET /incidents/:id/ws — successfully upgrades to WebSocket
[ ] POST /incidents/:id/entries — entry broadcasts to all connected WebSocket clients for that incident
[ ] PATCH /incidents/:id — state change broadcasts to connected clients
[ ] Connect 2 wscat clients to same incident — both receive broadcasts
[ ] Connect wscat to different incidents — messages don't leak between incidents
[ ] Disconnect a client — server logs removal, no errors, no goroutine leak
[ ] Server log shows client count per incident on connect/disconnect
[ ] go run -race . — zero race conditions during WebSocket test
```

### 9. Knowledge Gained

```
✅ WebSocket in Go — upgrade, read, write, close
✅ Hub pattern — connection registry, per-room broadcast, cleanup
✅ Per-client goroutine pairs — readPump + writePump
✅ Concurrent connection management — mutex-protected hub
✅ Slow client handling — buffered channels, drop policy
✅ Ping/pong keepalive — detecting dead connections
✅ Real-time event distribution in a production service
✅ Understanding the single-process limitation and production alternatives
```

---

# PHASE 8 — Observability & Feature Flags

> **Why this phase matters**
> A service you can't measure is a service you can't improve. Observability answers: "Is this service healthy right now? Is it getting worse?" Feature flags answer: "Can I change behavior in production without deploying new code?" Both are expected in teams practicing continuous delivery. Metrics-driven development and A/B testing are how teams make decisions based on data instead of guesses. After this phase, Handoff is instrumented and controllable at runtime.

---

## Challenge 8.1 — Instrument Handoff with Metrics
### `🟠 Intermediate`
**🕐 Expected duration: 11–14 hours**

### 1. Context

Handoff runs. But you have no idea: how many requests per second does it serve? What's the 95th percentile response latency? What percentage of requests fail? Is the database connection pool saturated? These questions are unanswerable without instrumentation. This challenge adds Prometheus-compatible metrics to Handoff — the same monitoring stack used by most Go services in production.

### 2. Goal

Add a `/metrics` endpoint exposing Prometheus-formatted metrics, instrument all HTTP handlers and the database layer, and add a readiness probe.

### 3. Scope

**Metrics to expose:**

| Metric | Type | What it measures |
|---|---|---|
| `handoff_http_requests_total` | Counter | Total requests, labeled by `method`, `path`, `status_code` |
| `handoff_http_request_duration_seconds` | Histogram | Request latency distribution, labeled by `method`, `path` |
| `handoff_incidents_total` | Gauge | Current number of incidents, labeled by `status` |
| `handoff_entries_total` | Counter | Total timeline entries created |
| `handoff_db_query_duration_seconds` | Histogram | Database query latency, labeled by `operation` (e.g., `get_incident`, `list_incidents`) |
| `handoff_websocket_connections` | Gauge | Current number of active WebSocket connections |

**MetricsMiddleware:**

Build a new middleware that wraps every HTTP handler. On each request it:
1. Starts a timer
2. Calls the next handler
3. Records the request count (counter) and duration (histogram), labeled by method, path, and status code.

There's a problem: Go's `http.ResponseWriter` doesn't let you read the status code after `WriteHeader` is called. You need to solve this — figure out how to capture the status code that the inner handler writes. This is a common Go middleware puzzle.

Once you solve it for the metrics middleware, go back and upgrade your Phase 5 `LoggingMiddleware` to use the same wrapper. Your log lines should now include the status code:
```
2026-05-10T08:01:00Z INF request completed method=POST path=/incidents status=201 duration=1.2ms request_id=a3f9c012
```
This is the natural progression: Phase 5 logged what was available. Now you have the tool to capture what wasn't.

**Readiness probe (`GET /readyz`):**

Different from `/healthz`. The health probe says "the process is alive." The readiness probe says "the process can serve traffic." `/readyz` must:
- Ping the database to verify connectivity. Your database driver has a method for this — find it.
- If the ping succeeds: return `200 {"status":"ready"}`
- If the ping fails: return `503 {"status":"not ready","reason":"database unreachable"}`

Kubernetes uses `/healthz` for liveness (should I restart this pod?) and `/readyz` for readiness (should I send traffic to this pod?). They serve different purposes.

**Handoff business metrics:**

Add computed metrics to the `GET /incidents/:id/handoff` response:
```json
{
    "severity": "SEV2",
    "status": "investigating",
    "service": "order-service",
    "elapsed_minutes": 38,
    "total_entries": 7,
    "actions_taken": 4,
    "open_questions": 1,
    "handoff_count": 1
}
```
`handoff_count` counts the number of `state_change` entries that indicate a shift rotation. This is Handoff's own product metric — how many times has context been transferred for this incident?

**Database instrumentation:**

Wrap your database store methods to record query durations:
```go
func (s *InstrumentedStore) GetIncident(ctx context.Context, id string) (Incident, error) {
    timer := prometheus.NewTimer(dbQueryDuration.WithLabelValues("get_incident"))
    defer timer.ObserveDuration()
    return s.inner.GetIncident(ctx, id)
}
```
This uses the decorator pattern — wrapping the real store without modifying it. Same interface, added behavior.

### 4. Expected Output

```bash
$ curl -s http://localhost:8080/metrics | head -20
# HELP handoff_http_requests_total Total HTTP requests
# TYPE handoff_http_requests_total counter
handoff_http_requests_total{method="GET",path="/incidents",status_code="200"} 14
handoff_http_requests_total{method="POST",path="/incidents",status_code="201"} 3
handoff_http_requests_total{method="GET",path="/incidents/inc-001",status_code="200"} 7
# HELP handoff_http_request_duration_seconds HTTP request latency
# TYPE handoff_http_request_duration_seconds histogram
handoff_http_request_duration_seconds_bucket{method="GET",path="/incidents",le="0.005"} 12
...

$ curl -s http://localhost:8080/readyz | jq .
{"status":"ready"}

# Stop the database:
$ docker compose stop db
$ curl -s http://localhost:8080/readyz | jq .
{"status":"not ready","reason":"database unreachable"}

# /healthz still returns 200 — the process is alive, just not ready for traffic.
```

### 5. Hints & Knowledge

- `github.com/prometheus/client_golang/prometheus` — the standard Prometheus client for Go. Install: `go get github.com/prometheus/client_golang`.
- `prometheus.NewCounterVec(opts, []string{"method","path","status_code"})` — counter with labels.
- `prometheus.NewHistogramVec(opts, []string{"method","path"})` — histogram with default buckets covering 5ms to 10s.
- `prometheus.NewTimer(histogram.WithLabelValues(...))` then `defer timer.ObserveDuration()` — one-line latency recording.
- `promhttp.Handler()` serves the `/metrics` endpoint in Prometheus exposition format — you don't write the output format yourself.
- **Database ping:** Every database driver has a ping or health-check method. MongoDB: `client.Ping(ctx, ...)`. PostgreSQL/pgx: `pool.Ping(ctx)`. SQLite: `db.PingContext(ctx)`. Find yours.
- **StatusWriter hint** (if stuck on capturing status codes): the standard approach is to wrap `http.ResponseWriter` with a struct that intercepts `WriteHeader(code)`, stores the code in a field, then forwards the call. Your wrapper struct embeds `http.ResponseWriter` and overrides one method.

### 6. Sources

- Prometheus Go client: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus
- promhttp handler: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/promhttp
- Prometheus metric types: https://prometheus.io/docs/concepts/metric_types/
- Kubernetes probes: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/

### 7. Common Mistakes to Avoid

- Using high-cardinality labels — labeling by `user_id` or `incident_id` creates millions of time series. Label by category (`method`, `path`, `status_code`), never by ID.
- Not registering metrics — `prometheus.MustRegister(myMetric)` must be called at startup, not inside handlers.
- Putting `/metrics` behind authentication — Prometheus needs to scrape it. Keep it unauthenticated (or on a separate internal port in production).
- Confusing `/healthz` and `/readyz` — if you make `/healthz` check the database, Kubernetes restarts the pod every time the DB has a brief hiccup. `/healthz` should always return 200 if the process is running.

### 8. Checklist

```
[ ] GET /metrics — returns Prometheus-formatted metrics
[ ] After several API calls, counters and histograms show accurate numbers
[ ] GET /readyz — returns 200 when DB is up, 503 when DB is down
[ ] GET /healthz — always returns 200 (unchanged from Phase 5)
[ ] Database query durations visible in /metrics
[ ] WebSocket connection gauge increases/decreases with connects/disconnects
[ ] Handoff brief includes elapsed_minutes, actions_taken, open_questions, handoff_count
[ ] go vet ./... — zero warnings
```

### 9. Knowledge Gained

```
✅ Prometheus client library — counters, gauges, histograms, labels
✅ Metrics middleware — request counting and latency recording
✅ Custom ResponseWriter wrapper — capturing status codes
✅ /healthz vs /readyz — liveness vs readiness, different purposes
✅ Database instrumentation — decorator pattern over the store interface
✅ Business metrics — measuring product-level outcomes, not just infrastructure
✅ Observability as a first-class design concern
```

---

## Challenge 8.2 — Build a Feature Flag System
### `🟡 Beginner → Intermediate`
**🕐 Expected duration: 9–12 hours**

### 1. Context

Feature flags control whether a feature is on or off — and for whom. In A/B testing, the flag assigns users to variant A or variant B, and you measure which variant performs better. The team doesn't deploy a new feature to 100% on day one. They roll it out to 5%, measure, then 20%, measure, then 100%. This is how products improve based on data.

This challenge builds a minimal feature flag system inside Handoff so you understand its mechanics.

### 2. Goal

Add an in-memory feature flag system to Handoff with percentage-based rollout and deterministic variant assignment.

### 3. Scope

**Data model:**
```go
type FeatureFlag struct {
    Name     string   `json:"name"`
    Enabled  bool     `json:"enabled"`
    Rollout  int      `json:"rollout"`   // 0–100, percentage of users who see the feature
    Variants []string `json:"variants"`  // e.g., ["control", "variant_a", "variant_b"]
}
```

**Endpoints:**

| Method | Path | Description |
|---|---|---|
| `POST` | `/flags` | Create a feature flag |
| `GET` | `/flags` | List all flags |
| `PATCH` | `/flags/:name` | Update rollout percentage or enabled status |
| `GET` | `/flags/:name/evaluate?user_id=tom` | Evaluate: is this user in the rollout? Which variant? |

**Evaluation logic (`GET /flags/:name/evaluate?user_id=tom`):**

The evaluation must satisfy these constraints:
1. **Deterministic** — the same `user_id` + same `flag_name` must always return the same result. No randomness. A user must not flip between variants on page refresh.
2. **Uniform distribution** — if rollout is 50%, approximately 50% of all distinct user IDs should be in the rollout. Not exactly 50% (that's impossible), but statistically close.
3. **Variant assignment** — if the user is in the rollout and the flag has multiple variants, the user must be assigned to one variant consistently.
4. If `enabled` is `false`, every user is OUT regardless of rollout percentage.
5. Rollout 0 = nobody in. Rollout 100 = everybody in. Both boundaries must work.

How you achieve determinism and uniform distribution is yours to figure out. The `hash` standard library package may be useful.

**Use the flag in Handoff:**

Add a flag called `detailed_handoff_brief`. When the flag evaluates to `in_rollout: true` for the requesting user, the `GET /incidents/:id/handoff` response includes additional computed fields:
- `avg_time_between_entries_seconds` — average gap between consecutive entries
- `longest_gap_seconds` — the longest pause in the timeline (where was the engineer stuck?)

When the flag evaluates to OUT, the handoff brief returns the standard fields only. The handler checks the flag system — no `if/else` on a config file or environment variable.

### 4. Expected Output

```bash
$ curl -s -X POST http://localhost:8080/flags \
  -H "Content-Type: application/json" \
  -d '{"name":"detailed_handoff_brief","enabled":true,"rollout":50,"variants":["control","detailed"]}' | jq .
{
  "name": "detailed_handoff_brief",
  "enabled": true,
  "rollout": 50,
  "variants": ["control", "detailed"]
}

$ curl -s 'http://localhost:8080/flags/detailed_handoff_brief/evaluate?user_id=tom' | jq .
{
  "flag": "detailed_handoff_brief",
  "user_id": "tom",
  "in_rollout": true,
  "variant": "detailed"
}

$ curl -s 'http://localhost:8080/flags/detailed_handoff_brief/evaluate?user_id=marc' | jq .
{
  "flag": "detailed_handoff_brief",
  "user_id": "marc",
  "in_rollout": false,
  "variant": null
}

# Same user, same flag, called 100 times — always same result.
```

### 5. Hints & Knowledge

- The key question is: how do you map a string (user_id + flag_name) to a number between 0 and 99, deterministically?
- Go's `hash/crc32` and `hash/fnv` packages produce deterministic integer outputs from byte inputs.
- "Deterministic" means: no `math/rand`. Same input must always produce same output.
- The evaluation endpoint is a GET (it reads state, doesn't mutate).
- Store flags in a `sync.RWMutex`-protected map, same pattern as the incident memory store.
- Test boundary cases carefully: rollout 0, rollout 100, rollout 50 with many different user IDs.

### 6. Sources

- `hash/crc32`: https://pkg.go.dev/hash/crc32
- `hash/fnv`: https://pkg.go.dev/hash/fnv
- Feature flag concepts: https://martinfowler.com/articles/feature-toggles.html
- LaunchDarkly architecture (reference): https://launchdarkly.com/blog/feature-flag-architecture/

### 7. Common Mistakes to Avoid

- Using `math/rand` for assignment — non-deterministic, user flips between variants.
- Not handling `rollout: 0` and `rollout: 100` as edge cases — test both boundaries.
- Forgetting to check `enabled` before evaluating rollout — a disabled flag should always return OUT.
- Returning variant index instead of variant name — the API consumer needs the name.

### 8. Checklist

```
[ ] POST /flags — creates a flag
[ ] GET /flags — lists all flags
[ ] PATCH /flags/:name — updates rollout/enabled
[ ] GET /flags/:name/evaluate?user_id=X — returns deterministic result
[ ] Same user + same flag → always same variant (test 10+ times)
[ ] Rollout 0 → no user is in rollout
[ ] Rollout 100 → every user is in rollout
[ ] Disabled flag → no user is in rollout regardless of rollout percentage
[ ] Handoff brief includes extra fields when flag evaluates to "detailed" variant
[ ] Handoff brief returns standard fields when flag evaluates to "control" or user is out
```

### 9. Knowledge Gained

```
✅ Feature flag architecture — flags, rollout percentage, variants
✅ Deterministic hashing for user bucketing (crc32/fnv)
✅ A/B testing mechanics — variant assignment and measurement
✅ Conditional behavior without code deployment
✅ The pattern used by LaunchDarkly, Unleash, and internal flag systems
```

---

# PHASE 9 — Authentication & Authorization

> **Why this phase matters**
> Every production service restricts access. Without authentication, anyone can create incidents, log entries, or modify severity. Without authorization, any authenticated user can do anything. JWT (JSON Web Tokens) is the standard for stateless API authentication — the server issues a token on login, the client sends it on every request, and the server verifies it without hitting a database. This is how APIs at scale handle auth.

---

## Challenge 9.1 — Add JWT Auth to Handoff
### `🟠 Intermediate`
**🕐 Expected duration: 18–22 hours**

### 1. Context

Handoff currently has no concept of identity. Anyone who can reach the API can create incidents, log entries, and change severity. In a real on-call system, only authenticated team members should have access. And some operations (like changing severity or resolving an incident) should be restricted to the current on-call engineer.

This challenge adds JWT-based authentication and role-based authorization to Handoff.

### 2. Goal

Add login, token issuance, token verification middleware, and role-based access control to Handoff's API.

### 3. Scope

**User model (in-memory, no database — keep it simple):**
```go
type User struct {
    ID       string `json:"id"`
    Username string `json:"username"`
    Password string `json:"-"`          // never returned in JSON
    Role     string `json:"role"`       // "engineer" or "admin"
}
```

Seed 3 users on startup:
```go
var seedUsers = []User{
    {ID: "u1", Username: "anh", Password: hashPassword("anh123"), Role: "engineer"},
    {ID: "u2", Username: "bernd", Password: hashPassword("bernd123"), Role: "engineer"},
    {ID: "u3", Username: "admin", Password: hashPassword("admin123"), Role: "admin"},
}
```

Use `bcrypt` for password hashing — never store plaintext.

**Endpoints:**

| Method | Path | Auth Required | Description |
|---|---|---|---|
| `POST` | `/auth/login` | No | Accepts `{"username","password"}`, returns JWT |
| `GET` | `/auth/me` | Yes | Returns current user info from token |
| All existing endpoints | Various | Yes | Protected by auth middleware |
| `GET` | `/healthz`, `/readyz`, `/metrics` | No | Always public |

**JWT structure:**
```json
{
  "sub": "u1",
  "username": "anh",
  "role": "engineer",
  "exp": 1716000000,
  "iat": 1715913600
}
```

Token expires after 24 hours. Sign with HMAC-SHA256 using a secret from environment: `JWT_SECRET`.

**Auth middleware:**

A new middleware that:
1. Reads the `Authorization: Bearer <token>` header
2. Verifies the JWT signature and expiration
3. Extracts the user identity and stores it in the request context
4. If the token is missing, invalid, or expired: returns `401 {"error":{"code":"UNAUTHORIZED","message":"..."}}`

```go
// Extracting user from context in a handler:
user := UserFromContext(r.Context())
```

**Authorization rules:**
- Any authenticated user can: list incidents, view incidents, view handoff briefs
- Any authenticated user can: create incidents, add timeline entries
- Only the `on_call` engineer for an incident (or an `admin`) can: change severity, change status, resolve
- `POST /flags` and `PATCH /flags/:name` — admin only

Return `403 {"error":{"code":"FORBIDDEN","message":"..."}}` for unauthorized operations.

**Login flow:**
```bash
$ curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"anh","password":"anh123"}' | jq .
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "u1",
    "username": "anh",
    "role": "engineer"
  }
}

# Use the token:
$ curl -s http://localhost:8080/incidents \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." | jq .
[...]

# Without token:
$ curl -s http://localhost:8080/incidents | jq .
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "missing or invalid authorization header"
  }
}
```

**Timeline entries auto-set author:**

After auth is added, the `author` field in `POST /incidents/:id/entries` is no longer provided by the client. The server sets it from the JWT — `user.Username`. A client cannot impersonate another engineer. Remove `author` from the request body validation; populate it from context.

**Reconnect feature flags to auth:**

Phase 8.2's `GET /flags/:name/evaluate?user_id=tom` takes user_id as a query parameter because auth didn't exist yet. Now it does. Update the endpoint: for authenticated requests, read the user_id from the JWT context. The `?user_id=` query parameter should still work as a fallback for admin testing, but normal evaluation uses the token identity. This means any authenticated endpoint can now call the flag system internally to branch behavior based on the current user — no query parameter needed.

### 4. Expected Output

```bash
# Login
$ TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -d '{"username":"anh","password":"anh123"}' | jq -r '.token')

# Authenticated request
$ curl -s http://localhost:8080/incidents -H "Authorization: Bearer $TOKEN" | jq .
[...]

# Add entry — author auto-set from token
$ curl -s -X POST http://localhost:8080/incidents/inc-001/entries \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"type":"action","text":"Restarted the service."}' | jq .author
"anh"

# Try to resolve someone else's incident
$ curl -s -X PATCH http://localhost:8080/incidents/inc-002 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"resolved"}' | jq .
{
  "error": {
    "code": "FORBIDDEN",
    "message": "only the on-call engineer or admin can change incident status"
  }
}

# Bad password
$ curl -s -X POST http://localhost:8080/auth/login \
  -d '{"username":"anh","password":"wrong"}' | jq .
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "invalid credentials"
  }
}
```

### 5. Hints & Knowledge

- `github.com/golang-jwt/jwt/v5` — the standard JWT library for Go. Install: `go get github.com/golang-jwt/jwt/v5`.
- `jwt.NewWithClaims(jwt.SigningMethodHS256, claims)` — create a token. `token.SignedString([]byte(secret))` — sign it.
- `jwt.Parse(tokenString, keyFunc)` — verify and decode. The `keyFunc` returns the signing key.
- `golang.org/x/crypto/bcrypt` — `bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)` to hash, `bcrypt.CompareHashAndPassword(hash, []byte(password))` to verify.
- Store the user in context: `context.WithValue(r.Context(), userKey, user)`. Retrieve: `r.Context().Value(userKey)`.
- `strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")` — extract the token from the header.
- The auth middleware wraps all routes EXCEPT `/auth/login`, `/healthz`, `/readyz`, `/metrics`.

### 6. Sources

- golang-jwt: https://pkg.go.dev/github.com/golang-jwt/jwt/v5
- bcrypt: https://pkg.go.dev/golang.org/x/crypto/bcrypt
- JWT introduction: https://jwt.io/introduction
- Go context values: https://pkg.go.dev/context#WithValue

### 7. Common Mistakes to Avoid

- Storing plaintext passwords — always bcrypt. Even for a demo project. This is non-negotiable.
- Putting the JWT secret in code — read from `JWT_SECRET` environment variable. Fail at startup if missing in production mode.
- Not checking token expiration — `jwt.Parse` does this automatically if you include `exp` in claims, but verify it's working.
- Using context values without a custom type for the key — `context.WithValue(ctx, "user", user)` uses a string key, which can collide. Define `type contextKey string` and use `const userKey contextKey = "user"`.
- Not skipping auth for health/metrics endpoints — Kubernetes and Prometheus need unauthenticated access to these.
- Trusting client-provided `author` after auth is added — the server must set it from the token.

### 8. Checklist

```
[ ] POST /auth/login — returns JWT on valid credentials
[ ] POST /auth/login — returns 401 on bad credentials
[ ] GET /auth/me — returns user info from token
[ ] All protected endpoints return 401 without token
[ ] All protected endpoints work with valid token
[ ] Expired token returns 401
[ ] POST /incidents/:id/entries — author auto-set from token, not request body
[ ] PATCH /incidents/:id — only on-call engineer or admin can modify
[ ] POST /flags — admin only, engineer gets 403
[ ] GET /flags/:name/evaluate — reads user_id from JWT when authenticated, falls back to query param
[ ] /healthz, /readyz, /metrics — work without token
[ ] Passwords stored as bcrypt hashes
[ ] JWT_SECRET read from environment
[ ] go vet ./... — zero warnings
```

### 9. Knowledge Gained

```
✅ JWT — issuance, signing, verification, expiration
✅ bcrypt — secure password hashing
✅ Auth middleware — token extraction, verification, context injection
✅ Authorization — role-based access control (RBAC)
✅ Context value propagation — carrying user identity through the request lifecycle
✅ Security boundaries — which endpoints are public, which are protected
✅ Separating authentication (who are you?) from authorization (what can you do?)
```

---

# PHASE 10 — Code Review & Pressure Test

> **Why this phase matters**
> Technical skill and performance under pressure are different abilities. This phase trains the second one: reading code under constraint, identifying subtle bugs without running the code, explaining decisions out loud, and writing correct Go quickly. After Phase 9 you have the knowledge — Phase 10 makes sure you can demonstrate it under time pressure.

---

## Challenge 10.1 — Code Review Gauntlet
### `🔴 Advanced`
**🕐 Expected duration: 8–10 hours**

### 1. Context

Every developer will review code written by others. You're shown real-looking Go programs with subtle bugs — goroutine leaks, race conditions, nil panics, interface misuse, security holes — and asked to spot and explain them. No running the code. No compiler. Your eyes and your understanding.

### 2. Goal

Review 5 broken Go programs. For each: identify all issues, explain why each is a problem in production, and write the fix.

### 3. Scope

For each program: identify ALL bugs, explain why each is a problem in production, and write the corrected version.

---

**Program 1 — The Leaking Worker (3 bugs)**

```go
package main

import (
	"fmt"
	"time"
)

func fetchData(id int, results chan string) {
	time.Sleep(time.Duration(id) * 100 * time.Millisecond)
	results <- fmt.Sprintf("data-%d", id)
}

func processAll(ids []int) []string {
	results := make(chan string)
	for _, id := range ids {
		go fetchData(id, results)
	}

	var collected []string
	timeout := time.After(300 * time.Millisecond)
	for {
		select {
		case r := <-results:
			collected = append(collected, r)
			if len(collected) == len(ids) {
				return collected
			}
		case <-timeout:
			fmt.Printf("timeout: got %d/%d results\n", len(collected), len(ids))
			return collected
		}
	}
}

func main() {
	ids := []int{1, 2, 3, 4, 5}
	results := processAll(ids)
	fmt.Println("results:", results)
	select {} // imagine a long-running server
}
```

*Hints: What happens to goroutines that finish after the timeout? The channel is unbuffered — what does that mean for blocked senders? After `processAll` returns, who reads from the channel?*

---

**Program 2 — The Silent Counter (3 bugs)**

```go
package main

import (
	"fmt"
	"net/http"
	"sync"
)

type Stats struct {
	mu       sync.Mutex
	Requests map[string]int
}

func NewStats() *Stats {
	return &Stats{Requests: make(map[string]int)}
}

func (s *Stats) Record(path string) {
	s.mu.Lock()
	s.Requests[path]++
	s.mu.Unlock()
}

func (s *Stats) GetAll() map[string]int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Requests
}

func main() {
	stats := NewStats()

	http.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		stats.Record(r.URL.Path)
		fmt.Fprintln(w, "ok")
	})

	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		all := stats.GetAll()
		for path, count := range all {
			fmt.Fprintf(w, "%s: %d\n", path, count)
		}
	})

	fmt.Println("listening on :8080")
	http.ListenAndServe(":8080", nil)
}
```

*Hints: `GetAll()` returns the internal map reference. What happens when the handler iterates it while `Record()` mutates it concurrently? The mutex protects inside `GetAll()`, but what about after it returns? Is there an error being silently ignored?*

---

**Program 3 — The Nil Surprise (2 bugs)**

```go
package main

import "fmt"

type Logger interface {
	Log(msg string)
}

type FileLogger struct {
	Path string
}

func (f *FileLogger) Log(msg string) {
	if f == nil {
		fmt.Println("[nil logger] " + msg)
		return
	}
	fmt.Printf("[file:%s] %s\n", f.Path, msg)
}

func newLogger(logToFile bool) Logger {
	if logToFile {
		return &FileLogger{Path: "/var/log/app.log"}
	}
	var f *FileLogger
	return f
}

func process(l Logger) {
	if l == nil {
		fmt.Println("no logger, skipping")
		return
	}
	l.Log("processing started")
}

func main() {
	logger := newLogger(false)
	process(logger)
}
```

*Hints: `newLogger(false)` returns a nil `*FileLogger` wrapped in the `Logger` interface. An interface value has two components: (type, value). When is an interface truly `nil`? Will `process` print "no logger, skipping" — or something else?*

---

**Program 4 — The Broken Middleware (5 bugs)**

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		status := w.Header().Get("X-Status-Code")
		log.Printf("%s %s → %s (%s)", r.Method, r.URL.Path, status, time.Since(start))
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH")
	})
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("id")
	if userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIError{Code: "MISSING_ID", Message: "id is required"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Status-Code", fmt.Sprintf("%d", http.StatusOK))
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"id": userID, "name": "Tom"})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/user", handleGetUser)
	handler := loggingMiddleware(corsMiddleware(mux))
	log.Fatal(http.ListenAndServe(":8080", handler))
}
```

*Hints: CORS headers set AFTER response is written. `X-Status-Code` header trick doesn't work — `http.ResponseWriter` doesn't expose status code this way. Error path missing Content-Type. Headers set after `WriteHeader` may be ignored. No CORS preflight (OPTIONS) handling.*

---

**Program 5 — The Insecure Query (6 bugs)**

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var col *mongo.Collection

func init() {
	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}
	col = client.Database("mydb").Collection("users")
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	role := r.URL.Query().Get("role")

	var filter bson.M
	json.Unmarshal([]byte(fmt.Sprintf(`{"name": {"$regex": "%s"}}`, q)), &filter)

	if role != "" {
		filter["role"] = role
	}

	cursor, err := col.Find(context.TODO(), filter)
	if err != nil {
		http.Error(w, "query failed", 500)
		return
	}

	for cursor.Next(context.TODO()) {
		var result bson.M
		cursor.Decode(&result)
		fmt.Fprintf(w, "%s: %s (%s)\n", result["_id"], result["name"], result["email"])
	}
}

func handleAuth(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	token = strings.TrimPrefix(token, "Bearer ")

	if token == "" {
		http.Error(w, "unauthorized", 401)
		return
	}

	filter := bson.M{
		"$where": fmt.Sprintf("this.token == '%s' && this.active == true", token),
	}

	var session bson.M
	err := col.FindOne(context.TODO(), filter).Decode(&session)
	if err != nil {
		http.Error(w, "unauthorized", 401)
		return
	}

	fmt.Fprintf(w, "authenticated as user %s", session["user_id"])
}

func main() {
	http.HandleFunc("/search", handleSearch)
	http.HandleFunc("/auth", handleAuth)
	http.ListenAndServe(":8080", nil)
}
```

*Hints: The search handler builds a regex from raw user input — an attacker can inject arbitrary MongoDB operators via the query string. The auth handler uses `$where` with string formatting — `$where` evaluates server-side JavaScript, and the token value is spliced in unsanitized, enabling NoSQL injection (the attacker can craft a token like `' || true || '` to bypass authentication). The cursor is never closed (resource leak). `cursor.Decode` error is ignored (nil fields cause silent failures). `json.Unmarshal` error is ignored (malformed input proceeds with a nil filter). `ListenAndServe` error is ignored.*

---

### 4. Expected Output

For each program, write:
```
BUG 1: [location] [what's wrong]
WHY:   [why this causes a problem in production]
FIX:   [the corrected code]
```

### 5. Sources

- 100 Go Mistakes: https://100go.co
- Go race detector: https://go.dev/doc/articles/race_detector
- Common Go pitfalls: https://go.dev/doc/faq
- MongoDB NoSQL injection: https://owasp.org/www-project-web-security-testing-guide/latest/4-Web_Application_Security_Testing/07-Input_Validation_Testing/05.6-Testing_for_NoSQL_Injection

### 6. Common Mistakes to Avoid

- Assuming code is correct because it compiles — Go's concurrency bugs are runtime bugs, invisible to the compiler.
- Missing the nil interface subtlety: an interface holding a nil pointer is NOT equal to `nil` itself.
- Not checking if channels are closed — the most common goroutine leak pattern.
- Overlooking NoSQL injection in code that "looks safe" — string formatting in database queries is always wrong, regardless of the database.

### 7. Knowledge Gained

```
✅ Critical code reading skills — spotting bugs without running code
✅ Goroutine leak patterns and fixes
✅ Race condition identification
✅ Nil interface gotchas
✅ Security bug recognition (including NoSQL injection)
```

---

## Challenge 10.2 — Build Under Pressure
### `🔴 Advanced`
**🕐 Expected duration: 6–8 hours**

### 1. Context

The final Go challenge. 3 timed problems, 45 minutes each, no hints. Designed to simulate real pressure — the kind you face during interviews, incidents, tight deadlines, or live debugging.

### 2. Goal

Solve 3 problems under time pressure. After each attempt: review what you got right, what you missed, and what you'd do differently.

### 3. Scope

---

**Problem 1 — Rate Limiter (Interfaces + Design) ⏱ 45 minutes**

Build a rate limiter that controls how many operations a user can perform in a time window.

Requirements:
1. Define a `RateLimiter` interface with methods `Allow(userID string) bool` and `Reset(userID string)`.
2. Implement two strategies:
   - `FixedWindowLimiter` — allows N requests per minute. Counter resets at the start of each new minute.
   - `SlidingWindowLimiter` — allows N requests in any rolling 60-second window. Uses a timestamp slice per user.
3. Both must be safe for concurrent use.
4. Write `main()` that creates both with limit=5, simulates 10 rapid calls from "user-1" on each, prints which calls were allowed/rejected, demonstrates that `Reset` works.

Expected:
```
=== Fixed Window ===
call 1: allowed
...
call 5: allowed
call 6: REJECTED
...
After reset:
call 1: allowed

=== Sliding Window ===
call 1: allowed
...
call 6: REJECTED
```

Evaluation: interface defined correctly, both satisfy it, thread-safe, `main()` uses only the interface.

---

**Problem 2 — Pipeline Processor (Concurrency) ⏱ 45 minutes**

Build a concurrent 3-stage pipeline where data flows through channels between stages.

Requirements:
1. **Stage 1 (Generator):** Produces numbers 1–20, sends to stage 2.
2. **Stage 2 (Filter + Transform):** Exactly 3 worker goroutines read from input. Discard even numbers. Square odd numbers. Send to stage 3.
3. **Stage 3 (Aggregator):** Collect all results, sort them, print.
4. All channels closed properly. No goroutine leaks. `go run -race` passes.
5. Aggregator waits for ALL stage 2 workers before printing.

Expected:
```
Pipeline starting...
[stage1] Generated 20 numbers
[stage2-worker1] 1 → 1
[stage2-worker3] 3 → 9
...
[stage3] Received 10 results
[stage3] Sorted: [1, 9, 25, 49, 81, 121, 169, 225, 289, 361]
Pipeline complete.
```

Worker lines may appear in any order. Sorted output is deterministic.

Evaluation: correct channel wiring, exactly 3 workers, WaitGroup correct, channels closed at right time, race-free.

---

**Problem 3 — Key-Value Store API (HTTP Service) ⏱ 45 minutes**

Build an in-memory key-value store as an HTTP API.

Endpoints:

| Method | Path | Behavior |
|---|---|---|
| `PUT` | `/store/:key` | Set value (plain text body). Return 201. |
| `GET` | `/store/:key` | Get value. 200 + value, or 404. |
| `DELETE` | `/store/:key` | Delete key. 204, or 404. |
| `GET` | `/store` | List all keys as JSON array. 200. |
| `GET` | `/stats` | JSON: `{"total_keys":N,"total_gets":N,"total_sets":N,"total_deletes":N}` |

Additional requirements:
- Thread-safe
- Every request logged: `METHOD /path → STATUS (Xms)`
- Graceful shutdown on SIGTERM
- Stats counters atomically incremented

Evaluation: all endpoints correct, thread-safe, logging middleware, graceful shutdown, correct status codes.

### 4. Rules

- Set a 45-minute timer for each problem. When it rings, stop.
- No looking at previous code, documentation, or hints during the timer.
- Write your solution, then review it yourself after the timer.
- Note: what did you get right? What did you miss? What took the longest?

### 5. Tips for Performing Well

- Read the problem twice — misunderstanding costs more time than reading slowly.
- Define your data structures before writing logic.
- Write the happy path first, then add error handling.
- Name things clearly — `workerCount` not `wc`, `incidentStore` not `is`.
- Say what you're doing out loud as you type — this is how technical interviews work.

### 6. Knowledge Gained

```
✅ Performing under time pressure
✅ Structuring solutions quickly
✅ Prioritizing correctness over completeness
✅ Technical communication while coding
```

---

# PART 3 — Frontend (Phases 11–12)

This track builds Handoff's frontend: a typed Vue.js application with state management, real-time WebSocket updates, authentication UI, and proper error handling. By the end, the full-stack application is usable.

---

# PHASE 11 — TypeScript + Vue.js

> **Why this phase matters**
> You've built a complete Go backend — API, database, WebSocket, metrics, feature flags, auth. Now it needs a frontend. Vue.js with TypeScript is one of the most widely used full-stack combinations. You know JavaScript. You have some exposure to reactive UI from Flutter (or you don't — this phase assumes nothing). Vue's model is: you declare state, you declare how state maps to UI, and Vue keeps them in sync. When state changes, the UI updates automatically. That's the entire idea. This phase teaches it from scratch.

---

## Challenge 11.1 — Vue Fundamentals + Component Library
### `🟡 Beginner → Intermediate`
**🕐 Expected duration: 35–50 hours**

### 1. Context

Before building the full Handoff frontend, you need two things: a TypeScript type layer that mirrors your Go models, and a set of reusable Vue components that you'll assemble into pages in Phase 12. This is how professional frontend development works — build components in isolation, test them visually, then compose them into views.

But first, if you've never written Vue, you need to understand reactivity.

### 2. Goal

Learn Vue's reactivity model, create the TypeScript type layer, and build 8 standalone Vue components for Handoff.

### 3. Scope

**Setup:**
```bash
npm create vue@latest handoff-ui -- --typescript --router --pinia
cd handoff-ui
npm install
npm run dev
```

This gives you a running dev server at `http://localhost:5173`.

---

**Part A — Reactivity warmup (do this first if Vue is new to you):**

Read the Vue 3 tutorial first: https://vuejs.org/tutorial/ (takes ~1 hour). Then build this:

Create `src/views/Warmup.vue` and mount it at `/warmup`. Build a single page that does ALL of the following:

1. **Counter**: A number displayed on screen. Two buttons: one increments it, one resets to 0. A line below shows the number doubled (computed from the counter, not stored separately).
2. **Item list**: A text input and an "Add" button. Typing text and pressing Enter (or clicking Add) adds the text to a visible list. Each item in the list has a "×" button that removes it. When the list is empty, show "No items yet." When it has items, show the count.
3. All state must be reactive — changing it updates the UI instantly without page refresh.

You need to discover and use these 6 Vue concepts (look them up in the Vue docs as you go):
- `ref()` — reactive state
- `computed()` — derived state
- `v-model` — two-way input binding
- `v-for` — list rendering
- `v-if` / `v-else` — conditional rendering
- `@click` / `@keyup.enter` — event handling

These 6 concepts are 90% of Vue. Everything else builds on them. Don't move to Part B until all three features work.

---

**Part B — TypeScript types (`src/types/index.ts`):**

Open your Go `models.go` from Phase 5. For every struct (`Incident`, `TimelineEntry`, `HandoffBrief`), write the TypeScript equivalent in `src/types/index.ts`. Also add types for `ApiError` and `AuthResponse` based on the JSON your API returns.

Rules:
- Use `interface` for object types (TypeScript equivalent of Go struct)
- Use union types for constrained string fields — `severity` should be `'SEV1' | 'SEV2' | 'SEV3'`, not `string`. Same for status and entry type.
- Every field that is a `time.Time` in Go becomes `string` in TypeScript (JSON serializes it as ISO string)
- Use the Go → TypeScript comparison table in the Hints section to map types

Verify: run `npm run type-check`. Zero errors.

---

**Part C — Components:**

Build these 8 components in `src/components/`:

| Component | Props / Emits | What it renders |
|---|---|---|
| `SeverityBadge.vue` | `severity: Severity` | Colored badge: red for SEV1, orange for SEV2, blue for SEV3 |
| `StatusIndicator.vue` | `status: Status` | Status text + pulsing dot for active statuses, static dot for resolved |
| `IncidentCard.vue` | `incident: Incident` | Summary card: ID, title, severity badge, status indicator, service, entry count. Emits `click`. |
| `TimelineEntry.vue` | `entry: TimelineEntry` | Single entry with colored left border by type, timestamp, author badge, text |
| `Timeline.vue` | `entries: TimelineEntry[]` | Vertical list of TimelineEntry components with a connecting line |
| `EntryInput.vue` | emits `submit(payload)` | Type selector buttons + text input + submit button. Validates non-empty text before emitting. |
| `HandoffBrief.vue` | `brief: HandoffBrief, entries: TimelineEntry[]` | Structured summary: severity, status, actions list, open questions highlighted in red |
| `MetricCard.vue` | `label: string, value: string or number, color?: string` | Simple stat card for dashboard counters |

**CSS requirements:**

You need to learn enough CSS to build a usable interface. For every component:

- Dark theme: background `#0d0d14`, text `#e0e0e8`, borders `#1a1a2e`
- Use CSS variables in a global stylesheet (`src/assets/main.css`):
  ```css
  :root {
    --bg-primary: #0d0d14;
    --bg-card: #12121e;
    --border: #1a1a2e;
    --text-primary: #e0e0e8;
    --text-secondary: #666;
    --color-sev1: #ff3b5c;
    --color-sev2: #f5a623;
    --color-sev3: #4a6cf7;
    --color-success: #0cce6b;
    --color-purple: #8b5cf6;
    --font-mono: 'JetBrains Mono', 'Fira Code', monospace;
  }
  ```
- Use `flexbox` for all layouts: `display: flex; gap: 12px; align-items: center;`
- Use `<style scoped>` in every component — styles only apply to that component
- Every component must look acceptable at both 1200px and 375px width (basic responsive). Use `flex-wrap: wrap` and percentage-based or `min-width` sizing.
- Import `JetBrains Mono` from Google Fonts in `index.html`.

**CSS concepts you need (and only these):**
- `display: flex` — lay items in a row or column. `flex-direction: column` for vertical.
- `gap` — space between flex items. Replaces margin hacks.
- `align-items: center` — vertically center items in a row.
- `justify-content: space-between` — spread items to edges.
- `padding`, `margin` — internal and external spacing.
- `border-radius` — rounded corners.
- `border-left: 3px solid var(--color-sev1)` — the colored left border on timeline entries.
- `@media (max-width: 768px) { ... }` — change layout on small screens.
- `transition: border-color 0.15s` — smooth hover effects.
- CSS variables: `var(--color-sev1)` — define once, use everywhere.

That's it. No grid, no float, no position absolute, no CSS frameworks, no Tailwind. Flexbox + variables covers everything in this project.

---

**Part D — Visual sandbox:**

Create `src/views/Sandbox.vue` at route `/sandbox` that renders every component with hardcoded mock data. This is your visual test harness — you verify each component here before assembling real pages.

```vue
<script setup lang="ts">
import SeverityBadge from '@/components/SeverityBadge.vue'
import StatusIndicator from '@/components/StatusIndicator.vue'
// ... import all components

const mockIncident: Incident = {
  id: 'inc-001',
  title: 'order-service request drop',
  // ... full mock data
}
</script>

<template>
  <div class="sandbox">
    <h2>SeverityBadge</h2>
    <SeverityBadge severity="SEV1" />
    <SeverityBadge severity="SEV2" />
    <SeverityBadge severity="SEV3" />

    <h2>StatusIndicator</h2>
    <StatusIndicator status="investigating" />
    <StatusIndicator status="resolved" />

    <!-- ... every component with every meaningful prop variation -->
  </div>
</template>
```

### 4. Expected Output

`http://localhost:5173/sandbox` — all 8 components rendered with mock data, styled in the dark theme, readable at both desktop and mobile widths.

### 5. Hints & Knowledge

- `defineProps<{ severity: Severity }>()` — typed props in `<script setup lang="ts">`.
- `defineEmits<{ submit: [payload: { type: EntryType; text: string }] }>()` — typed event emission.
- `ref<Incident | null>(null)` — reactive variable with explicit type.
- `computed(() => entries.filter(e => e.type === 'action'))` — derived state, recalculates when `entries` changes.
- `v-for="entry in entries" :key="entry.id"` — always use `:key` with a unique identifier. Without it, Vue reuses DOM nodes incorrectly.
- `v-if` removes elements from the DOM. `v-show` hides them with CSS (`display: none`). Use `v-if` for things that rarely toggle, `v-show` for things that toggle often.
- `<style scoped>` — styles only apply to this component's template. Always use it.

### 6. Go → Vue Comparison

| Go | Vue / TypeScript |
|---|---|
| `struct` fields | `interface` properties |
| Function parameters + types | `defineProps<{ ... }>()` |
| Return values | `defineEmits<{ ... }>()` / template output |
| `for _, item := range items` | `v-for="item in items"` |
| `if condition { }` | `v-if="condition"` |
| Package-level `var` | `ref()` or `reactive()` |
| Derived value from other vars | `computed()` |
| `fmt.Sprintf("hello %s", name)` | `{{ name }}` in template |

### 7. Sources

- Vue 3 tutorial (start here): https://vuejs.org/tutorial/
- Vue + TypeScript: https://vuejs.org/guide/typescript/overview
- Vue single-file components: https://vuejs.org/guide/scaling-up/sfc
- Flexbox guide: https://css-tricks.com/snippets/css/a-guide-to-flexbox/
- CSS variables: https://developer.mozilla.org/en-US/docs/Web/CSS/Using_CSS_custom_properties
- Google Fonts — JetBrains Mono: https://fonts.google.com/specimen/JetBrains+Mono

### 8. Common Mistakes to Avoid

- Using `any` for props — defeats TypeScript. Always use the interfaces from `types/index.ts`.
- Mutating props directly — Vue warns against this. If you need to modify prop data, copy it to a local `ref` or emit an event to the parent.
- Forgetting `:key` in `v-for` — causes subtle rendering bugs when items are added or removed.
- Writing all CSS in `App.vue` — scope styles to each component.
- Over-engineering: no Tailwind, no utility libraries, no CSS-in-JS. Raw CSS with scoped styles and variables is enough.
- Not checking mobile width — open DevTools, toggle device toolbar (Ctrl+Shift+M), test at 375px width.

### 9. Checklist

```
[ ] npm run dev — Vue app starts at localhost:5173
[ ] /warmup route — counter, list, conditional rendering all work
[ ] src/types/index.ts — all types defined with union types, not string
[ ] All 8 components render correctly in /sandbox
[ ] Dark theme applied consistently via CSS variables
[ ] Components look acceptable at 1200px and 375px width
[ ] No TypeScript errors (npm run type-check)
[ ] No 'any' types in component props
```

### 10. Knowledge Gained

```
✅ Vue 3 reactivity — ref, computed, reactive
✅ Template directives — v-for, v-if, v-model, v-bind, @event
✅ Composition API — <script setup lang="ts">
✅ defineProps, defineEmits — typed component API contracts
✅ TypeScript interfaces mirroring Go structs
✅ Union types for constrained values
✅ CSS fundamentals — flexbox, variables, scoped styles, responsive
✅ Component-first development — build parts, then assemble
```

---

# PHASE 12 — Full Handoff Frontend

> **Why this phase matters**
> Phase 11 built the parts. This phase assembles them into a working application connected to the Go backend. This is where full-stack engineering happens: the frontend makes an HTTP request, the Go API responds with JSON, TypeScript validates the shape, Vue renders it reactively, and the user interacts. When something breaks, the bug could be in any layer — Go handler, database query, JSON serialization, TypeScript type, Vue reactivity, CSS layout. Knowing how to trace a problem across the full stack is the skill that separates a frontend developer from a full-stack engineer.

---

## Challenge 12.1 — Build the Handoff Dashboard
### `🟠 Intermediate → Advanced`
**🕐 Expected duration: 30–40 hours**

### 1. Context

You have 8 components and a running Go API with auth, WebSocket, metrics, and feature flags. Time to wire everything together into a real application.

### 2. Goal

Build the complete Handoff frontend: login, dashboard, incident detail with live timeline, handoff brief, incident creation form, and proper error handling on every view.

### 3. Scope

**API client (`src/api/client.ts`):**

Build a centralized API client that all components use. Requirements:
- A single `request<T>(path, options)` function that handles every API call
- Automatically reads the JWT token from `localStorage` and attaches it as an `Authorization: Bearer` header
- Reads the API base URL from `import.meta.env.VITE_API_URL` (falls back to `http://localhost:8080`)
- On non-2xx responses: parses the structured error JSON from your Go API and throws it as an Error
- Returns typed data (`Promise<T>`)
- No component should ever call `fetch` directly — all calls go through this client

**State management with Pinia:**

Build two Pinia stores: `useAuthStore` and `useIncidentStore`.

First, understand why: open your DashboardPage and IncidentPage in your head. Both need the incident list. Without a shared store, each page fetches independently — duplicate requests, inconsistent state. If you prop-drill from a parent, the parent holds data it doesn't use. A Pinia store solves this: one source of truth, any component can read from it.

`useAuthStore` must manage: token (persisted to localStorage), current user, login action, logout action, and an `isAuthenticated` computed property.

`useIncidentStore` must manage: incidents array, loading state, error state, and actions for fetching, creating, and adding entries. Components read state reactively. Components call actions to trigger API calls.

**Pages and routing (`src/router/index.ts`):**

| Route | Page | Auth | Description |
|---|---|---|---|
| `/login` | `LoginPage.vue` | No | Login form |
| `/` | `DashboardPage.vue` | Yes | Metric cards + incident list |
| `/incidents/new` | `NewIncidentPage.vue` | Yes | Create incident form |
| `/incidents/:id` | `IncidentPage.vue` | Yes | Full timeline + entry input + live updates |
| `/incidents/:id/handoff` | `HandoffPage.vue` | Yes | Handoff brief view |
| `/:pathMatch(.*)*` | `NotFoundPage.vue` | No | 404 page for unmatched routes |

**Route guard:**

Implement a Vue Router navigation guard (`router.beforeEach`) that checks if the target route requires auth (use `meta: { requiresAuth: true }` on protected routes). If the user isn't authenticated, redirect to `/login`.

**Login page (`LoginPage.vue`):**
- Username + password fields
- Submit button (disabled while request is in-flight)
- Show error message from API on invalid credentials
- On success: store token, redirect to `/`

**Dashboard page (`DashboardPage.vue`):**
- MetricCard row: active incidents, total incidents, total open questions
- List of IncidentCards, sorted by most recently updated
- Clicking an IncidentCard navigates to `/incidents/:id`
- Auto-refresh every 30 seconds

**Incident page (`IncidentPage.vue`):**
- Full Timeline component showing all entries
- EntryInput at the bottom for adding new entries
- Severity and status displayed (with SeverityBadge + StatusIndicator)
- Link to handoff brief: `/incidents/:id/handoff`
- WebSocket connection for real-time updates

**WebSocket integration (`src/composables/useIncidentSocket.ts`):**

Build a Vue composable that manages a WebSocket connection to your Go backend. Requirements:
- Connect to `ws://host/incidents/:id/ws` when the component mounts
- On message: parse JSON, call a callback provided by the component
- On disconnect: automatically reconnect with exponential backoff (start at 1s, double each attempt, cap at 30s). Reset the delay after a successful message.
- On component unmount: close the WebSocket cleanly (no leaked connections)

Use in `IncidentPage.vue` — when a `new_entry` message arrives, append it to the timeline. When a `state_change` arrives, update the incident's severity or status. The timeline updates without a page refresh or API refetch.

**New incident form (`NewIncidentPage.vue`):**
- Fields: title (text), service (text), severity (dropdown), initial observation (textarea)
- Client-side validation: all fields required, severity must be one of the three values
- Show inline validation errors (red text below each invalid field)
- Disable submit button while request is in-flight
- On success: navigate to `/incidents/:id`
- On API error: display the error message from the backend

**Entry input (`IncidentPage.vue` using `EntryInput.vue`):**
- Select entry type via buttons (observation, action, discovery, open_question, state_change)
- Type text, press Enter or click Log
- On submit: call `addEntry()` in the store. The WebSocket broadcast handles updating the timeline — don't refetch.

**Four states every page must handle:**
1. **Loading** — show a text indicator or skeleton while data is being fetched
2. **Error** — show the error message with a retry button
3. **Empty** — "No incidents yet" on dashboard, "No entries yet" on timeline
4. **Data** — render the content

A page that shows a blank white screen while loading is a bug. A page that silently fails on API error is a bug. Handle all four states.

**Browser DevTools — learn these now:**

Before debugging any frontend issue, check these three browser panels:
- **Network tab** (F12 → Network): see every HTTP request, its status code, response body, and timing. If the API returns an error, you'll see it here before the UI shows anything.
- **Console tab** (F12 → Console): JavaScript errors, unhandled promise rejections, and your `console.log` output.
- **Vue DevTools** (browser extension): inspect component tree, see prop values, reactive state, Pinia store contents. Install from: https://devtools.vuejs.org/

These three tools answer 90% of frontend debugging questions.

### 4. Expected Output

```
http://localhost:5173/login            → Login form, redirects to / on success
http://localhost:5173/                 → Dashboard with incident cards + metrics
http://localhost:5173/incidents/new    → Create incident form with validation
http://localhost:5173/incidents/inc-001  → Live timeline with entry input
http://localhost:5173/incidents/inc-001/handoff → Structured handoff brief
http://localhost:5173/anything-else    → 404 page

# Open two browser tabs on the same incident.
# Add an entry in one tab.
# It appears instantly in the other tab via WebSocket.

# Log out. Try to navigate to /. Redirected to /login.
```

### 5. Hints & Knowledge

- `onMounted(() => { store.fetchIncidents() })` — fetch data when the component mounts (appears on screen).
- `watch(() => route.params.id, (newId) => { ... })` — react when the route parameter changes without remounting.
- `router.push('/incidents/' + id)` — navigate programmatically after form submission.
- `v-model` on a `<select>` element binds to a ref — the selected option's value is automatically stored.
- `<input :class="{ 'input-error': !isValid }" />` — conditional CSS class based on state.
- `event.preventDefault()` — stop form from causing a page reload (Vue's `@submit.prevent` does this automatically).
- `localStorage.getItem('token')` / `localStorage.setItem('token', value)` — persist the JWT across page refreshes.

### 6. Common Mistakes to Avoid

- Calling `fetch` directly inside components — use the Pinia store. Components render, stores manage data.
- Not cleaning up WebSocket on unmount — `onUnmounted(() => ws?.close())`. Without this, connections accumulate and the server runs out of file descriptors.
- Not handling the API-unreachable case — if the Go server isn't running, the user should see "Cannot connect to server," not a blank page or console error.
- Forgetting the empty state — "No incidents yet. Create one." is better than a blank white screen.
- Not testing with auth — after adding auth to the API, every fetch call needs the token. The centralized API client handles this, but verify it.
- Hardcoding the API URL — use `import.meta.env.VITE_API_URL`. In development it points to `localhost:8080`. In production (Phase 14) it points to the deployed API.

### 7. Checklist

```
[ ] Login works — valid credentials → dashboard, invalid → error message
[ ] Logout works — token cleared, redirected to /login
[ ] Route guard — unauthenticated access to / redirects to /login
[ ] Dashboard — shows metric cards + incident list
[ ] Dashboard — auto-refreshes every 30 seconds
[ ] Incident page — shows full timeline + entry input
[ ] WebSocket — new entry in one tab appears in another tab instantly
[ ] WebSocket — reconnects after disconnect (kill and restart Go server to test)
[ ] New incident form — validates all fields, shows inline errors
[ ] New incident form — redirects to incident page on success
[ ] Handoff brief — renders structured summary correctly
[ ] 404 page — shown for unmatched routes
[ ] All pages handle: loading, error, empty, and data states
[ ] No TypeScript errors (npm run type-check)
[ ] Works at both 1200px and 375px width
```

### 8. Knowledge Gained

```
✅ Pinia — centralized state management (why it exists, how to use it)
✅ API client — typed, centralized, with auth token injection
✅ Vue Router — dynamic routes, navigation guards, 404 handling
✅ WebSocket client — connect, receive, reconnect with exponential backoff, cleanup
✅ Form handling — validation, submission, error display, disabled state
✅ Four-state UI — loading, error, empty, data
✅ Authentication flow — login, token storage, logout, route protection
✅ Browser DevTools — network tab, console, Vue DevTools
✅ Full frontend ↔ backend integration
```

---

# PHASE 12.Test — Vue Testing Fundamentals

> **Why this phase matters**
> Same principle as Phase 5.Test. You've just built 8 components, a Pinia store, an API client, and a WebSocket composable. Test them now while the code is fresh. A button that doesn't emit, a store that doesn't set loading state, a component that renders the wrong severity color — these are caught by component tests. Vitest + Vue Test Utils is the standard Vue testing stack.

---

## Challenge 12.Test — Test the Handoff Frontend
### `🟡 Beginner → Intermediate`
**🕐 Expected duration: 8–10 hours**

### 1. Context

If you've written Go tests in Phase 5.Test, the concept is identical. A test mounts a component with known props and checks whether the rendered output matches expectations. Instead of `testing.T`, you use `describe` / `it` / `expect`. Instead of `httptest`, you mock `fetch`.

### 2. Goal

Add a Vitest + Vue Test Utils test suite covering component rendering and store logic.

### 3. Scope

**Setup:**
```bash
npm install -D vitest @vue/test-utils jsdom @pinia/testing
```

Add to `vite.config.ts`:
```typescript
test: {
  environment: 'jsdom',
}
```

**Your first test:**

```typescript
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import SeverityBadge from '@/components/SeverityBadge.vue'

describe('SeverityBadge', () => {
  it('renders SEV1 text', () => {
    const wrapper = mount(SeverityBadge, { props: { severity: 'SEV1' } })
    expect(wrapper.text()).toContain('SEV1')
  })
})
```

Run it: `npx vitest run`. Same mechanic as Go tests — known input, expected output.

**Component tests to write:**
- `SeverityBadge` — renders correct text for each severity level
- `StatusIndicator` — renders correct text for each status
- `IncidentCard` — renders title, severity badge, status, service name
- `EntryInput` — emits `submit` event with correct payload on Enter
- `EntryInput` — does NOT emit when text field is empty
- `Timeline` — renders correct number of TimelineEntry components
- `HandoffBrief` — renders open questions section only when open questions exist
- `HandoffBrief` — hides open questions section when there are none

**Store tests (mock `fetch`):**

```typescript
import { vi } from 'vitest'

vi.stubGlobal('fetch', vi.fn())
```

- Test `fetchIncidents` — sets `loading=true` during fetch, populates `incidents` on success
- Test `fetchIncidents` — sets `error` on failure
- Test `createIncident` — calls POST with correct body

**Run:**
```bash
npx vitest run --coverage
# Target: >60% component coverage
```

### 4. Hints & Knowledge

- `mount(Component, { props: {...} })` — mount a component with props.
- `wrapper.text()` — all rendered text content.
- `wrapper.find('.class')` — find a DOM element by CSS selector.
- `await wrapper.find('input').setValue('hello')` — simulate typing.
- `await wrapper.find('button').trigger('click')` — simulate click.
- `wrapper.emitted('submit')` — returns array of emitted events with payloads.
- `vi.fn()` — create a mock function. `vi.stubGlobal('fetch', vi.fn())` — mock `fetch` globally.
- `vi.mocked(fetch).mockResolvedValue({ ok: true, json: async () => data } as Response)` — make mock return data.
- `beforeEach(() => vi.mocked(fetch).mockReset())` — reset mocks between tests.

### 5. Sources

- Vitest: https://vitest.dev/
- Vue Test Utils: https://test-utils.vuejs.org/
- Testing Pinia stores: https://pinia.vuejs.org/cookbook/testing

### 6. Common Mistakes to Avoid

- Testing Vue internals (raw ref values) instead of rendered output — test what the user sees.
- Not mocking `fetch` — tests should never make real HTTP calls.
- Snapshot testing everything — snapshots break on any UI change. Test specific behaviors.
- Not resetting mocks between tests — `beforeEach(() => vi.mocked(fetch).mockReset())`.

### 7. Checklist

```
[ ] npx vitest run — all tests pass
[ ] npx vitest run --coverage — >60% coverage
[ ] Component tests cover: props rendering, event emission, conditional rendering
[ ] Store tests cover: loading state, success state, error state
[ ] All tests run without network access (fetch is mocked)
```

### 8. Knowledge Gained

```
✅ Vitest — describe, it, expect
✅ Vue Test Utils — mount, props, find, trigger, emitted
✅ Mocking fetch — vi.stubGlobal, vi.fn, mockResolvedValue
✅ Store testing — Pinia with mock fetch
✅ Frontend coverage measurement
```

**From this point forward:** as with Go, write Vue tests alongside new code.

---

# PART 4 — Test Completion & Delivery (Phases 13–14)

This track completes test coverage, containerizes the stack, automates the pipeline, and deploys.

---

# PHASE 13 — Complete Test Coverage

> **Why this phase matters**
> You've been writing tests since Phase 5.Test (Go) and Phase 12.Test (Vue). Some code is well-tested; some has gaps — especially code from Phases 6–9 if you didn't write tests as you went. This phase is the systematic sweep: identify what's untested, fill the gaps, and hit the coverage targets that make CI meaningful. After this, every push to `main` is verified by automated tests.

---

## Challenge 13.1 — Complete the Go Test Suite
### `🟠 Intermediate`
**🕐 Expected duration: 6–12 hours**

### 1. Context

Phase 5.Test covered validation and basic handler tests. Since then you've added database persistence, WebSocket, metrics, feature flags, and auth. If you wrote tests as you built (as instructed), some of this is covered. For the rest, this is where you catch up.

### 2. Goal

Achieve >80% test coverage across the Go backend with zero race conditions.

### 3. Scope

Run coverage and identify gaps:
```bash
go test -coverprofile=cover.out ./... && go tool cover -html=cover.out
```

Open the HTML report. Red lines are untested. Prioritize:

**Auth tests:**
- Login with valid credentials → token
- Login with invalid credentials → 401
- Protected endpoint without token → 401
- Expired token → 401
- RBAC: non-on-call engineer tries to resolve → 403
- RBAC: admin overrides → 200

**Feature flag tests:**
- Deterministic: same user + flag → same result across 10 calls
- Rollout 0% → nobody in
- Rollout 100% → everybody in
- Disabled flag → nobody in regardless of rollout
- Multiple variants assigned consistently

**Store concurrency:**
- 10 goroutines: 5 writing, 5 reading simultaneously. Pass `go test -race`.

**Handoff brief:**
- Known entries → correct action count, open question count, handoff count, elapsed time

**Config:**
- Defaults when env vars unset
- Overrides when set

**Run:**
```bash
go test ./... -race -cover -v
# Target: >80% coverage, 0 race conditions
```

---

## Challenge 13.2 — Complete the Vue Test Suite
### `🟡 Beginner → Intermediate`
**🕐 Expected duration: 4–7 hours**

### 1. Context

Phase 12.Test covered component rendering and store basics. This challenge fills the remaining gaps.

### 2. Goal

Achieve >75% component coverage across the Vue frontend.

### 3. Scope

Run coverage and identify gaps:
```bash
npx vitest run --coverage
```

Prioritize:
- Login form: submits credentials, stores token on success, shows error on failure
- Route guard: unauthenticated navigation to `/` redirects to `/login`
- New incident form: validates required fields, shows inline errors, disables button during submission
- WebSocket composable: reconnects after disconnect (mock WebSocket)
- Empty states: dashboard with no incidents renders "No incidents yet"
- Error states: API unreachable renders error message with retry button

**Run:**
```bash
npx vitest run --coverage
# Target: >75% coverage
```

### 4. Checklist (both challenges)

```
[ ] go test ./... -race -cover — >80% Go coverage, 0 race conditions
[ ] npx vitest run --coverage — >75% Vue coverage
[ ] Auth flow tested: login, token, protected endpoints, RBAC
[ ] Feature flag evaluation tested: deterministic, boundaries, disabled
[ ] Store concurrency tested with -race
[ ] Vue forms, route guards, and error states tested
[ ] All tests pass without network access
```

### 5. Knowledge Gained

```
✅ Systematic coverage analysis — HTML reports, identifying gaps
✅ Testing auth flows — tokens, expiry, RBAC
✅ Testing concurrent code — race detection
✅ Testing frontend edge cases — empty states, error states, form validation
✅ Coverage as a CI gate
```

---

# PHASE 14 — Ship It

> **Why this phase matters**
> Code on your laptop is not a product. Code running at a live URL with automated tests on every push — that's a product. This phase containerizes Handoff, automates the build pipeline with GitHub Actions, writes Kubernetes manifests for production deployment, and deploys to a free hosting platform. After this, anyone with a browser can use Handoff.

---

## Challenge 14.1 — Containerize, Automate, Deploy
### `🟠 Intermediate → Advanced`
**🕐 Expected duration: 25–30 hours**

### 1. Context

Handoff has a Go backend, a Vue frontend, and a database. They run locally via manual commands. This challenge packages everything into Docker containers, wires them with Docker Compose, automates testing and building with GitHub Actions, writes Kubernetes manifests, and deploys the result to a live URL.

### 2. Goal

Ship Handoff as a containerized, CI-gated, publicly accessible product.

### 3. Scope

**Go API Dockerfile:**

Write a multi-stage Dockerfile for the Go API. Requirements:
- Stage 1 (builder): use a Go base image, copy source, build the binary with `CGO_ENABLED=0` (produces a static binary with no C dependencies — required for minimal runtime images)
- Stage 2 (runtime): use a minimal base image (alpine), copy only the binary from the builder stage, run as a non-root user, expose port 8080
- The final image must be under 20MB. Without multi-stage, it would be ~300MB (includes the Go compiler). Verify with `docker images`.
- Add a `.dockerignore` to exclude `.git`, test files, and build artifacts

**Vue frontend Dockerfile:**

Write a multi-stage Dockerfile for the Vue frontend. Requirements:
- Stage 1: use a Node base image, install dependencies, build the Vue app (`npm run build` produces a `dist/` directory of static files)
- Stage 2: use an nginx base image, copy the `dist/` from the builder, copy a custom `nginx.conf`
- The `VITE_API_URL` must be configurable at build time (use a Docker build arg)

**nginx.conf:**

Write an nginx config that does two things:
- Serves Vue static files from `/usr/share/nginx/html`
- Handles SPA routing: when a user refreshes `/incidents/inc-001`, nginx must serve `index.html` instead of returning 404 (because client-side routing handles the path). Research `try_files` to solve this.

**Docker Compose (`docker-compose.yml`):**

Wire all three services (Go API, Vue frontend, database) into one Docker Compose file. Requirements:
- API depends on the database (database must be healthy before API starts)
- API reads the database connection string and `JWT_SECRET` from environment
- Database data persists via a named volume
- Health checks on API (`/healthz`) and database (use the database's native health-check mechanism)
- Frontend depends on API
- One command runs everything: `docker compose up --build`

Adapt the database service to match your choice from Phase 6. The structure is the same regardless of database — image, port, volume, health check — only the specifics change.

**GitHub Actions (`.github/workflows/ci.yml`):**

Write a CI pipeline with three jobs:
- `test-go`: checkout, setup Go, run `go vet` and `go test -race -cover`
- `test-vue`: checkout, setup Node, install, run Vitest with coverage
- `build`: runs only after both test jobs pass, only on `main` branch. Builds both Docker images and pushes to `ghcr.io` (GitHub Container Registry). Use `secrets.GITHUB_TOKEN` for auth (automatically available, no setup needed).

Read the GitHub Actions documentation to understand the YAML structure: triggers (`on`), jobs, steps, `needs` (job dependencies), and `if` conditionals.

**Kubernetes manifests (`k8s/`):**

```
k8s/
├── api-deployment.yaml      # 1 replica, resource limits, liveness + readiness probes
├── api-service.yaml          # ClusterIP service
├── frontend-deployment.yaml  # 2 replicas
├── frontend-service.yaml     # LoadBalancer service
├── db-configmap.yaml         # Database connection string
└── db-secret.yaml            # JWT_SECRET (base64 encoded)
```

API deployment includes:
- `replicas: 1` (single replica — the WebSocket hub pattern from Phase 7 is process-local; multiple replicas would break real-time broadcasting. See Phase 7's limitation note.)
- `resources: { requests: { cpu: 100m, memory: 64Mi }, limits: { cpu: 200m, memory: 128Mi } }`
- `livenessProbe: { httpGet: { path: /healthz, port: 8080 } }`
- `readinessProbe: { httpGet: { path: /readyz, port: 8080 } }`
- `envFrom: configMapRef + secretRef`

These manifests exist in the repo and demonstrate understanding. Running them on Minikube is optional.

**Deploy:**
- Go API → Render Web Service (Docker deploy from GitHub)
- Database → managed cloud service (e.g., MongoDB Atlas free tier, Neon PostgreSQL free tier, or Turso for SQLite — choose the managed service that matches your database)
- Vue frontend → Render Static Site (build: `npm run build`, publish: `dist`)
- Set `VITE_API_URL` to the deployed API URL during build
- Set the database connection string and `JWT_SECRET` as environment variables in Render

**README.md:**

The README is a deliverable. It must contain:
- What Handoff is (2–3 sentences)
- Architecture diagram (text-based or Mermaid)
- How to run locally (`docker compose up`)
- How to run in development mode (Go + Vue separately)
- API endpoint table
- Tech stack list
- Screenshot or GIF of the dashboard
- CI badge: `![CI](https://github.com/yourname/handoff/actions/workflows/ci.yml/badge.svg)`
- Link to live deployment

### 4. Expected Output

```bash
# Anyone can run:
git clone github.com/yourname/handoff
docker compose up --build
# → http://localhost:3000 — Vue dashboard
# → http://localhost:8080 — Go API

# Live at:
# https://handoff-api.onrender.com
# https://handoff.onrender.com
# GitHub CI badge: green
```

### 5. Hints & Knowledge

- `docker build -t myapp .` — build an image. `docker run -p 8080:8080 myapp` — run it.
- `docker compose up --build` — build and start all services. `docker compose logs -f api` — stream one service's logs.
- `CGO_ENABLED=0` — produces a static binary with no C dependencies. Required for `alpine` runtime images.
- `.dockerignore` — list files Docker should NOT copy into the build context (`node_modules`, `.git`, `*.test.go`). Without this, builds are slow and images are large.
- `secrets.GITHUB_TOKEN` — automatically available in GitHub Actions, no setup needed for `ghcr.io`.
- `actions/cache@v4` — cache Go modules and npm packages between CI runs for faster builds.
- `kubectl apply -f k8s/` — deploy all manifests. `kubectl get pods -w` — watch pods start.

### 6. Sources

- Docker multi-stage builds: https://docs.docker.com/build/building/multi-stage/
- Docker Compose: https://docs.docker.com/compose/
- Go Docker best practices: https://docs.docker.com/language/golang/
- GitHub Actions: https://docs.github.com/en/actions
- Kubernetes concepts: https://kubernetes.io/docs/concepts/
- Render deployment: https://render.com/docs

### 7. Common Mistakes to Avoid

- Single-stage Docker builds — includes the Go compiler or Node runtime in the final image. Always use multi-stage.
- Not using `.dockerignore` — copies `node_modules` (200MB+) and `.git` into the build context.
- Running containers as root — always add `USER nonroot` or equivalent. Security baseline.
- Using `:latest` tag in Kubernetes manifests — always pin a specific version or commit SHA.
- Hardcoding secrets in `docker-compose.yml` or Kubernetes manifests — use environment variables and GitHub Secrets.
- Forgetting `try_files ... /index.html` in nginx — Vue Router URLs return 404 on page refresh.
- Not caching dependencies in CI — npm and Go module downloads on every run waste minutes.
- Cloud database free tiers often sleep after inactivity — first request after sleep takes several seconds. This is normal, not a bug.

### 8. Checklist

```
[ ] docker compose up --build — entire stack starts, frontend connects to API
[ ] Go image < 20MB (check with docker images)
[ ] Push to GitHub — CI runs, tests pass, badge is green
[ ] Push to main — Docker images pushed to ghcr.io
[ ] k8s/ manifests present with deployments, services, configmap, secret, probes
[ ] Deployed to Render + managed database — live URL works
[ ] README contains: description, architecture, run instructions, API docs, CI badge, live link
[ ] git log shows meaningful commit history across all phases
```

### 9. Knowledge Gained

```
✅ Docker multi-stage builds — minimal production images
✅ Docker Compose — multi-service local orchestration
✅ nginx — static file serving, SPA routing, reverse proxy
✅ GitHub Actions — automated test, build, push pipeline
✅ Container registry — ghcr.io image publishing
✅ Kubernetes manifests — Deployment, Service, ConfigMap, Secret, probes
✅ Free-tier deployment — shipping to a live URL
✅ README as documentation — the first thing anyone reads
```

---

*Complete phases in order. Don't skip. Each phase builds on the previous one.*

> # Bonus
> 1. **Multi-replica WebSocket fan-out** — Solve the single-process hub limitation from Phase 7. Use MongoDB Change Streams or PostgreSQL `LISTEN/NOTIFY` to broadcast across replicas. Then set `replicas: 3` in Kubernetes and verify real-time updates work across all instances. This is the most important production gap in the current architecture.
>
> 2. **Token refresh** — The current JWT expires after 24 hours with no renewal. Add a `/auth/refresh` endpoint that issues a new token from a valid (non-expired) existing token. Update the Vue API client to detect 401 responses, attempt a refresh, and retry the original request transparently.
>
> 3. **Persistent feature flags** — Phase 8.2's flags live in memory and reset on restart. Move them to the database. Add a simple admin UI page for toggling flags without curl.
>
> 4. **Grafana dashboard** — Add Grafana to Docker Compose. Configure Prometheus to scrape `/metrics`. Build a dashboard showing request rate, latency percentiles, error rate, and active WebSocket connections. This turns Phase 8.1's instrumentation into something visual and operational.
>
> None of these are required. The curriculum is complete at Phase 14. These exist for continued growth after the core is shipped.