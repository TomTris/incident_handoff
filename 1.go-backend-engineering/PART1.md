
# PART 1 — Go Engineering (Phases 1–4)

This track builds a deep, production-ready understanding of Go as a systems and backend language. You will work through Go's type system and interface-driven design, master its concurrency model using goroutines and channels, profile and optimize memory allocation under realistic workloads. Every challenge is modeled on patterns used in real Go codebases.

---

# PHASE 1 — Foundations

> **Why this phase matters**
> Before you can write Go that solves real problems, you need fluency in the basics: how Go organizes code into packages, how data structures work (slices, maps, structs), how errors are handled explicitly (no exceptions), and how to read/write files. These are the building blocks every future phase depends on.

Start by completing the Go Tour (https://go.dev/tour/). Then do the challenge below to verify you're ready for Phase 2.

---

## Challenge 1.1 — Build a Word Frequency Counter
### `🟢 Beginner`
**🕐 Expected duration: 6–8 hours**

### 1. Context
Every language has a "prove you can use the basics" project. This is Go's. You'll read a file, process strings, count things, sort results, and write output. Every Go fundamental appears at least once.

### 2. Goal
Build a CLI tool that reads a text file, counts how often each word appears, and writes a sorted frequency report to an output file.

### 3. Scope
- Read a text file path from command-line arguments (`os.Args`)
- Read the file contents, split into words, normalize to lowercase
- Count word frequencies using a `map[string]int`
- Define a `WordCount` struct: `Word string`, `Count int`
- Convert the map to a `[]WordCount` slice and sort by count descending
- Print the top 10 words to the terminal
- Write the full sorted report to `report.txt`
- Handle errors: file not found, empty file, no arguments provided — print a clear message and exit with code 1

### 4. Expected Output
```bash
$ go run main.go sample.txt
Top 10 words in sample.txt:
  1. the     — 42
  2. of      — 28
  3. and     — 25
  4. to      — 19
  5. a       — 17
  ...

Full report written to report.txt (347 unique words)
```

### 5. Hints
- `os.ReadFile(path)` returns `[]byte` and an `error` — handle both.
- `strings.Fields(text)` splits on whitespace (better than `Split` for this).
- `strings.ToLower(word)` normalizes case.
- `sort.Slice(slice, func(i, j int) bool { return slice[i].Count > slice[j].Count })` — sort by count descending.
- `os.Create("report.txt")` + `fmt.Fprintf(file, ...)` — write to a file.
- `os.Exit(1)` — exit with error code.

### 6. Checklist
```
[ ] go run main.go sample.txt — prints top 10 words
[ ] report.txt is created with all words sorted by frequency
[ ] go run main.go nonexistent.txt — prints error, exits with code 1
[ ] go run main.go (no args) — prints usage message, exits with code 1
[ ] Words are case-insensitive ("The" and "the" count as one)
[ ] go vet ./... — zero warnings
```

### 7. Knowledge Gained
```
✅ File I/O — os.ReadFile, os.Create, fmt.Fprintf
✅ Strings — Fields, ToLower
✅ Maps — counting with map[string]int
✅ Structs — defining and using custom types
✅ Slices — building, sorting
✅ Error handling — checking errors, printing messages, os.Exit
✅ Command-line arguments — os.Args
```

---

# PHASE 2 — Interfaces & Type System

> **Why this phase matters**
> Interfaces are the backbone of Go's entire standard library. `http.Handler`, `io.Reader`, `io.Writer`, `error` — all interfaces. If you don't understand interfaces deeply, you'll struggle to read Go code written by others, and you'll write brittle code yourself. This phase teaches you to think the Go way: *program to behavior, not to concrete types.*

---

## Challenge 2.1 — Build a Multi-Format Logger
### `🟡 Beginner → Intermediate`
**🕐 Expected duration: 4–7 hours**

### 1. Context
Every production system logs events. But *where* those logs go changes depending on the environment: console during local development, structured files in staging, JSON for log aggregation tools like Datadog, Grafana Loki, or AWS CloudWatch in production.

A well-designed logging system should let you swap the destination without changing the code that *uses* the logger. This is exactly how Go's `io.Writer`, `log/slog`, and popular libraries like `uber-go/zap` work internally.

### 2. Goal
Build a logging system that can write to multiple destinations (console, file, JSON) using a common interface. The logger must be swappable — the rest of the code should not care where logs go.

### 3. Scope
- Define a `Logger` interface with at least one method: `Log(level, message string)`
- Implement 3 concrete loggers, all satisfying the `Logger` interface:
  - `ConsoleLogger` — prints to terminal with timestamp
  - `FileLogger` — writes to a `.log` file
  - `JSONLogger` — writes structured JSON lines to a file
- Write a function `RunApp(l Logger)` that takes any logger and logs 3 events (startup, a warning, a shutdown)
- In `main()`, call `RunApp` three times — once with each logger type
- No `if/else` based on logger type anywhere in `RunApp` — it must work purely through the interface

### 4. Expected Output
Console:
```
[2026-03-22 10:00:01] INFO  app started
[2026-03-22 10:00:01] WARN  high memory usage
[2026-03-22 10:00:01] INFO  app shutdown
```
File (`app.log`):
```
2026-03-22 10:00:01 | INFO  | app started
2026-03-22 10:00:01 | WARN  | high memory usage
2026-03-22 10:00:01 | INFO  | app shutdown
```
JSON (`app.json`):
```json
{"time":"2026-03-22T10:00:01","level":"INFO","message":"app started"}
{"time":"2026-03-22T10:00:01","level":"WARN","message":"high memory usage"}
{"time":"2026-03-22T10:00:01","level":"INFO","message":"app shutdown"}
```

### 5. Hints & Knowledge
- In Go, interfaces are implemented **implicitly** — no `implements` keyword. If your struct has the right methods, it satisfies the interface automatically.
- `io.Writer` is Go's most important interface: `Write(p []byte) (n int, err error)`. `os.Stdout` and `os.File` both implement it — that's why you can write to both the same way.
- `time.Now().Format("2006-01-02 15:04:05")` — Go uses a reference time for formatting (Jan 2, 2006 = Go's birthday).
- `encoding/json` — use `json.Marshal(struct)` to convert a struct to JSON bytes.
- `os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)` — open a file for appending.

### 6. Sources
- Go interfaces explained: https://go.dev/tour/methods/9
- `io.Writer`: https://pkg.go.dev/io#Writer
- `encoding/json`: https://pkg.go.dev/encoding/json
- `time.Format`: https://pkg.go.dev/time#Time.Format
- `os.OpenFile`: https://pkg.go.dev/os#OpenFile

### 7. Knowledge Gained
```
✅ How Go interfaces work (implicit implementation)
✅ Writing to io.Writer — the foundation of all Go I/O
✅ encoding/json for structured data
✅ Dependency injection via interfaces (pass behavior, not implementation)
✅ The design pattern used by net/http, os, bufio, and most Go packages
```

---

## Challenge 2.2 — Fix the Shape Calculator
### `🟢 Beginner`
**🕐 Expected duration: 1–2 hours**

### 1. Context
A junior developer tried to build a geometry calculator that computes the area of different shapes using Go interfaces. The code compiles in some places and panics in others. Your job is to fix it and make it robust.

### 2. Goal
Fix all bugs in the provided broken code, understand why each bug exists, and add one defensive improvement using a **type switch**.

### 3. Scope
Here is the broken code:

```go
package main

import (
    "fmt"
    "math"
)

type Shape interface {
    Area() float64
    Describe() string
}

type Circle struct {
    Radius float64
}

type Rectangle struct {
    Width, Height float64
}

type Triangle struct {
    Base, Height float64
}

func (c Circle) Area() float64 {
    return math.Pi * c.Radius * c.Radius
}

func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

func (r Rectangle) Describe() string {
    return fmt.Sprintf("Rectangle %.1f x %.1f", r.Width, r.Height)
}

func (c Circle) Describe() string {
    return fmt.Sprintf("Circle r=%.1f", c.Radius)
}

func printArea(s Shape) {
    fmt.Printf("%s → area: %.2f\n", s.Describe, s.Area)
}

func totalArea(shapes []Shape) float64 {
    total := 0
    for _, s := range shapes {
        total += s.Area()
    }
    return total
}

func main() {
    shapes := []Shape{
        Circle{Radius: 5},
        Rectangle{Width: 3, Height: 4},
        Triangle{Base: 6, Height: 8},
    }
    for _, s := range shapes {
        printArea(s)
    }
    fmt.Printf("Total area: %.2f\n", totalArea(shapes))
}
```

Find ALL bugs (there are 5), fix them, then add:
- A `type switch` inside `printArea` that prints `"(is a circle)"` if the shape is a `Circle`

### 4. Expected Output
```
Circle r=5.0 (is a circle) → area: 78.54
Rectangle 3.0 x 4.0 → area: 12.00
Triangle b=6.0 h=8.0 → area: 24.00
Total area: 114.54
```

### 5. Hints & Knowledge
- Missing method on a type = does NOT satisfy the interface → compile error
- `s.Describe` vs `s.Describe()` — calling a method needs `()`
- `total := 0` makes `total` an `int` — can't add `float64` to it
- Type switch syntax: `switch v := s.(type) { case Circle: ... }`
- A `Triangle` must implement ALL methods of `Shape` to be used as one

### 6. Sources
- Type assertions: https://go.dev/tour/methods/15
- Type switches: https://go.dev/tour/methods/16

### 7. Knowledge Gained
```
✅ Interface satisfaction rules — ALL methods must be implemented
✅ Type assertion and type switch
✅ Common interface bugs and how to spot them
✅ Zero values and type mismatches
```

---

# PHASE 3 — Goroutines & Channels

> **Why this phase matters**
> Concurrency is Go's killer feature. Goroutines are why Go is chosen for backend services. A goroutine costs ~2KB of memory vs ~8MB for a thread — you can run hundreds of thousands of them. Channels replace shared memory with message passing, eliminating entire classes of bugs. If you can write correct concurrent Go, you have what makes Go valuable.

---

## Challenge 3.1A — Log File Generator
### `🟢 Beginner`
**🕐 Expected duration: 2–3 hours**

### 1. Context
Before you can build a concurrent log scanner, you need log files to scan. This challenge generates realistic fake logs that Challenge 3.1B will consume.

### 2. Goal
Build `log_generator.go` that creates 7 fake `.log` files in `./logs/`, each with 300 lines of random log entries.

### 3. Scope
- Create `./logs/` directory safely (no crash if it already exists)
- Generate files named `gateway_1.log` through `inventory-sync_7.log`, using service names
- Each line follows this format:
```
2026-05-01T08:00:03.000Z [INFO ] [gateway] [trace=a3f9c012] Request processed in 42ms
```
- Level distribution: 70% INFO, 20% WARN, 10% ERROR
- Each file uses a different service name
- Timestamps must strictly increase within each file, spaced 1–15 seconds apart
- Print progress as each file is created

### 4. Given code
```go
const (
	numFiles    = 7
	numLines    = 300
	outputDir   = "logs"
)
 
var (
	infoMessages = []string{
		"Service started successfully",
		"Configuration loaded from /etc/app/config.yaml",
		"Database connection pool initialized (size=10)",
		"Health check endpoint responding on :8080/health",
		"Cache warmed up with 1482 entries",
		"Scheduled job 'cleanup' registered (interval=5m)",
		"TLS certificate valid until 2027-03-15",
		"Request processed in 42ms",
		"User session created",
		"Metrics exported to Prometheus",
		"Batch import completed: 350 records",
		"Webhook delivered to https://hooks.example.com/notify",
		"Rate limiter reset for tenant acme-corp",
		"Graceful reload triggered by SIGHUP",
		"New worker spawned (pool=4/8)",
	}
 
	warnMessages = []string{
		"Response time exceeded 500ms threshold (actual=623ms)",
		"Disk usage at 82%% on /var/lib/data",
		"Retry attempt 2/5 for upstream service payment-api",
		"Deprecated header X-Request-ID used by client 10.0.3.44",
		"Connection pool nearing capacity (8/10)",
		"Clock skew detected: 1.3s drift from NTP server",
		"Certificate expires in 30 days",
		"Memory usage above 75%% (current=78%%)",
		"Slow query detected: SELECT * FROM orders (1.8s)",
		"Rate limit approaching for API key sk-...a3f9",
		"Fallback to secondary DNS resolver",
		"Stale cache entry served for key user:9281",
		"Config key 'legacy_mode' is deprecated, migrate to v2 schema",
		"Unrecognized query parameter 'debug' ignored",
		"Partial response returned: 3 of 5 shards responded",
	}
 
	errorMessages = []string{
		"Failed to connect to postgres://db:5432/app — connection refused",
		"Panic recovered in handler /api/v1/orders: index out of range [5]",
		"TLS handshake failed: certificate signed by unknown authority",
		"Out of memory: cannot allocate 256MB for image processing",
		"Deadlock detected between goroutine 47 and goroutine 52",
		"Kafka consumer lag exceeded 10000 messages on topic=events",
		"Disk write failed: /var/log/app.log — no space left on device",
		"Authentication failed for user admin@example.com (attempt 5/5)",
		"Circuit breaker OPEN for service inventory-api (failures=12)",
		"Unhandled exception in middleware chain: nil pointer dereference",
		"DNS resolution failed for api.partner.io",
		"S3 upload failed: AccessDenied on bucket prod-assets",
		"Request timeout after 30s: POST /api/v1/reports/generate",
		"Invalid JWT token: signature verification failed",
		"Migration 0042_add_index.sql failed: duplicate column name",
	}
 
	services = []string{
		"gateway", "auth-service", "order-engine",
		"payment-api", "notification-worker", "scheduler", "inventory-sync",
	}
)
```

### 5. Expected Output
```
Generating 7 log files (300 lines each)...
  ✔ logs/gateway-1.log  (300 lines)
  ✔ logs/auth-service-2.log  (300 lines)
  ✔ logs/order-engine-3.log  (300 lines)
  ✔ logs/payment-api-4.log  (300 lines)
  ✔ logs/notification-worker-5.log  (300 lines)
  ✔ logs/scheduler-6.log  (300 lines)
  ✔ logs/inventory-sync-7.log  (300 lines)
Done!
```

### 6. Hints
- `os.MkdirAll`, `os.Create`, `fmt.Sprintf("%08x", rand.Uint32())`
- Accumulate timestamps, don't calculate from base + index

### 7. Knowledge Gained
```
✅ File I/O — os.Create, WriteString
✅ time.Time formatting and arithmetic
✅ Weighted random generation
```

---

## Challenge 3.1B — Concurrent Log Scanner
### `🟡 Beginner → Intermediate`
**🕐 Expected duration: 8–12 hours**

### 1. Context
You have 7 log files from Challenge 3.1A. An incident just happened — scan all files simultaneously, count levels per file, capture every ERROR line, and print a report. Doing it sequentially is too slow.

### 2. Goal
Build `scanner.go` that processes all log files in parallel using goroutines, collects results through a channel AND a mutex-protected shared list, and prints a sorted report.

### 3. Scope
- Read all `.log` files from `./logs/`
- Launch **one goroutine per file** — fan-out
- Each worker scans its file line by line, counts INFO/WARN/ERROR
- Workers send per-file statistics through a **channel** — fan-in
- Workers append individual ERROR lines to a **shared slice protected by mutex**
- Use `sync.WaitGroup` to know when all workers are done
- Close the channel after all goroutines finish
- Print a sorted report: table of counts per file, totals, worst file, and first 5 ERROR lines
- Must pass `go run -race .` with zero race conditions

### 4. Given code
```go
type FileResult struct {
    Filename   string
    InfoCount  int
    WarnCount  int
    ErrorCount int
}

type ErrorLog struct {
    Filename string
    Line     string
}

var (
    allErrors []ErrorLog
    mu        sync.Mutex
)
```

One struct goes through a channel. The other goes into a shared slice. Which is which — and why — is yours to figure out.

### 5. Must Use
```
goroutine (1 per file)       — fan-out
channel (for per-file stats) — fan-in
sync.WaitGroup               — coordinate completion
sync.Mutex                   — protect shared error list
bufio.Scanner                — read files line by line
```

### 6. Design Constraint — Why Both Channel AND Mutex?
You must use **both** mechanisms in this challenge. Think about what kind of data each worker produces:
- One type is sent **exactly once** per worker
- The other is sent **zero to many times**, unpredictably

Which mechanism fits which? Figure this out — it's the core design decision.

### 7. Expected Output
```
Scanning 7 files concurrently...

[worker] gateway-1.log → INFO:138  WARN:41  ERROR:21
[worker] order-engine-3.log → INFO:140  WARN:39  ERROR:21
[worker] auth-service-2.log → INFO:142  WARN:38  ERROR:20
[worker] notification-worker-5.log → INFO:137  WARN:42  ERROR:21
[worker] payment-api-4.log → INFO:141  WARN:40  ERROR:19
[worker] scheduler-6.log → INFO:139  WARN:41  ERROR:20
[worker] inventory-sync-7.log → INFO:140  WARN:40  ERROR:20

══════════════════════════════════════════════════════
                    ERROR REPORT
══════════════════════════════════════════════════════
 File                          INFO    WARN    ERROR
──────────────────────────────────────────────────────
 gateway-1.log                  138      41       21
 auth-service-2.log             142      38       20
 order-engine-3.log             140      39       21
 payment-api-4.log              141      40       19
 notification-worker-5.log      137      42       21
 scheduler-6.log                139      41       20
 inventory-sync-7.log           140      40       20
──────────────────────────────────────────────────────
 TOTAL                          977     281      142
══════════════════════════════════════════════════════

Worst file: notification-worker-5.log (21 errors)

First 5 ERROR lines:
  [gateway-1.log] 2026-05-01T08:02:11.000Z [ERROR] Failed to connect...
  [order-engine-3.log] 2026-05-01T08:01:02.000Z [ERROR] Out of memory...
```

Worker lines appear in random order (proves concurrency). Report table is sorted by filename.

### 8. Common Mistakes to Avoid
- `wg.Add(1)` in the wrong place — silent data loss, no crash, no error, just missing results
- `wg.Wait(); close(ch)` in the wrong goroutine — deadlock
- Forgetting `close(ch)` entirely — deadlock
- Appending to a shared slice without protection — race condition
- Passing `wg` instead of `&wg` — `Done()` on a copy does nothing

### 9. Sources
- Goroutines: https://go.dev/tour/concurrency/1
- Channels: https://go.dev/tour/concurrency/2
- `sync.WaitGroup`: https://pkg.go.dev/sync#WaitGroup
- `sync.Mutex`: https://pkg.go.dev/sync#Mutex
- Race detector: https://go.dev/doc/articles/race_detector

### 10. Checklist
```
[ ] go run scanner.go          — runs without errors
[ ] go run -race scanner.go   — ZERO race conditions
[ ] go vet ./...               — zero warnings
[ ] Worker lines appear in random order
[ ] Report table sorted by filename
[ ] ERROR lines are captured and printed
[ ] Both channel AND mutex are used correctly
```

### 11. Knowledge Gained
```
✅ Goroutine — launch and manage concurrent work
✅ Channel — buffered, send/receive, close, range
✅ Directional chan — chan<- (send-only) in function signatures
✅ sync.WaitGroup — coordinate goroutine completion
✅ sync.Mutex — protect shared state
✅ Fan-out / Fan-in — most important Go concurrency pattern
✅ Channel vs Mutex — when to use which
✅ Race detection — go run -race
✅ bufio.Scanner — efficient line-by-line file reading
```

---

## Challenge 3.2 — Build a Worker Pool URL Checker
### `🟠 Intermediate`
**🕐 Expected duration: 15–20 hours**

### 1. Context
In production, you never spawn unlimited goroutines. If 10,000 requests come in and you launch 10,000 goroutines, your server runs out of memory. The solution: a **worker pool** — a fixed number of goroutines that pick jobs from a queue, process them, and stay alive for the next job. Worker pools are how Go HTTP servers, job queues, and data processors work under the hood.

### 2. Goal
Your team runs 20 internal services. Build a URL health checker using a fixed worker pool of the **minimum number of goroutines possible**. All 20 services must be checked within **7 seconds**, even when 6 services time out (each taking 4 seconds to timeout). Under normal conditions, each check should finish well under 500ms.

### 3. Scope
- Define **minimum worker goroutines** needed to meet the 7-second constraint
- Feed 20 URLs into a jobs channel (mix of valid/invalid/timeout URLs)
- Each worker performs an HTTP GET with a **4-second timeout** using `context`
- Results sent to a results channel, printed as they arrive
- Graceful handling: timeouts, unreachable hosts, invalid URLs
- Print final summary: total success vs failed
- Workers stop cleanly when there are no more jobs

### 4. Given Variables
```go

var urls = []string{
	"https://google.com",
	"https://github.com",
	"https://go.dev",
	"https://pkg.go.dev",
	"https://cloudflare.com",
	"https://fastly.com",
	"https://stackoverflow.com",
	"https://reddit.com",
	"https://news.ycombinator.com",
	"https://gitlab.com",
	"https://bitbucket.org",
	"https://hub.docker.com",
	"https://kubernetes.io",
	"https://prometheus.io",
	// 6 URLs that will timeout
	"https://httpstat.us/200?sleep=10000",
	"https://httpstat.us/200?sleep=15000",
	"https://httpstat.us/200?sleep=20000",
	"https://10.255.255.1",             // non-routable IP, hangs
	"https://192.0.2.1",               // TEST-NET, hangs
	"https://198.51.100.1",            // TEST-NET-2, hangs
}

type Job struct {
    ID  int
    URL string
}

type Result struct {
	Job        Job
	WorkerID   int
	StatusCode int
	Duration   time.Duration
	Err        error
}
```

### 5. Expected Output
```
[worker 4] 1  ✅ https://github.com           → 200 (130ms)
[worker 2] 5  ✅ https://fastly.com           → 200 (191ms)
[worker 6] 3  ✅ https://pkg.go.dev           → 200 (209ms)
[worker 4] 6 ❌ https://stackoverflow.com    → 403 (80ms)
[worker 3] 2  ✅ https://go.dev               → 200 (226ms)
[worker 1] 0  ✅ https://google.com           → 200 (249ms)
[worker 5] 4  ✅ https://cloudflare.com       → 200 (364ms)
[worker 3] 10 ✅ https://bitbucket.org        → 200 (144ms)
[worker 1] 11 ✅ https://hub.docker.com       → 200 (189ms)
[worker 5] 12 ✅ https://kubernetes.io        → 200 (101ms)
[worker 3] 13 ✅ https://prometheus.io        → 200 (99ms)
[worker 2] 7  ✅ https://reddit.com           → 200 (332ms)
[worker 4] 9  ✅ https://gitlab.com           → 200 (467ms)
[worker 6] 8  ✅ https://news.ycombinator.com → 200 (666ms)
[worker 1] 14 ❌ https://10.255.255.1         → Get "https://10.255.255.1": context deadline exceeded (4001ms)
[worker 5] 15 ❌ https://10.255.255.2         → Get "https://10.255.255.2": context deadline exceeded (4001ms)
[worker 3] 16 ❌ https://10.255.255.3         → Get "https://10.255.255.3": context deadline exceeded (4000ms)
[worker 2] 17 ❌ https://10.255.255.1         → Get "https://10.255.255.1": context deadline exceeded (4001ms)
[worker 4] 18 ❌ https://192.0.2.1            → Get "https://192.0.2.1": context deadline exceeded (4001ms)
[worker 6] 19 ❌ https://198.51.100.1         → Get "https://198.51.100.1": context deadline exceeded (4001ms)

══════════════════════════════════════════════════
                      SUMMARY
══════════════════════════════════════════════════
 ✅  Healthy  (2xx)   :  13
 ❌  Unreachable      :  7
──────────────────────────────────────────────────
 Total                :  20
 Fastest              :  https://stackoverflow.com (80ms)
 Slowest (healthy)    :  https://news.ycombinator.com (666ms)
 Total runtime        :  4877ms
```

### 6. Common Mistakes to Avoid
- Not closing the jobs channel — workers loop forever (goroutine leak)
- Closing results channel from a worker — if multiple workers do this, panic
- Not using `context` for timeouts — HTTP requests hang forever without it
- Using `time.Sleep` instead of `context.WithTimeout` — never do this
- Not closing response body after successful request

### 7. Hints & Knowledge
- `context.WithTimeout(context.Background(), 3*time.Second)` — cancels after 3s
- `http.NewRequestWithContext(ctx, "GET", url, nil)` — attaches context to request
- `close(jobs)` — workers reading `for job := range jobs` stop automatically
- Channel direction: `jobs <-chan Job` (receive-only), `results chan<- Result` (send-only)
- `time.Since(start)` — measure elapsed time

### 8. Sources
- Worker pools: https://gobyexample.com/worker-pools
- `context` package: https://pkg.go.dev/context
- `net/http` client: https://pkg.go.dev/net/http
- `select` statement: https://go.dev/tour/concurrency/5

### 9. Knowledge Gained
```
✅ Worker pool — fixed concurrency pattern
✅ context — timeout and cancellation (essential for all network code)
✅ net/http client — making HTTP requests in Go
✅ Channel directionality — enforcing send/receive contracts
✅ Graceful goroutine shutdown
```

---

# PHASE 4 — Memory & Performance

> **Why this phase matters**
> Go is used for high-performance systems because it gives you control over memory — without the danger of C. Companies running Go at scale care deeply about allocations per request, GC pauses, and heap pressure. Knowing how to measure, profile, and reduce allocations is what separates surface-level Go from deep Go.

---

## Challenge 4 — The Benchmark Battle
### `🟠 Intermediate → Advanced`
**🕐 Expected duration: 15–20 hours**

### 1. Context
A data pipeline processes millions of log lines per day. Each line is parsed into a key-value map. The current implementation is correct but slow — it allocates a new map on every call, causing GC pressure. Your job: measure it, understand why it's slow, and fix it.

This is a real scenario. Datadog, Cloudflare, and similar companies do this kind of optimization routinely on their log ingestion pipelines.

### 2. Goal
Benchmark two existing implementations, analyze their memory behavior using Go's built-in tooling, then write a faster third version that wins on both time and allocations.

### 3. Scope
- Write proper benchmark tests for Version A and Version B (provided). For each round, the parser needs to parse the whole file.
- Run `go test -bench=. -benchmem` and record: `ns/op`, `B/op`, `allocs/op`
- Run `go build -gcflags="-m" .` to see escape analysis output
- Write Version C using `sync.Pool` to reduce heap allocations
- Version C must beat both A and B in `ns/op`, `B/op`, and `allocs/op`
- Write comments explaining what each optimization does and why

### 4. Given
#### Code
```go
func ParseA(line string) map[string]string {
    result := map[string]string{}
    parts := strings.Split(line, "|")
    for _, p := range parts {
        kv := strings.Split(strings.TrimSpace(p), "=")
        if len(kv) == 2 {
            result[kv[0]] = kv[1]
        }
    }
    return result
}

func ParseB(line string) map[string]string {
    pairs := strings.Split(line, "|")
    result := make(map[string]string, len(pairs))
    for _, p := range pairs {
        p = strings.TrimSpace(p)
        if idx := strings.Index(p, "="); idx != -1 {
            result[p[:idx]] = p[idx+1:]
        }
    }
    return result
}
```

#### Required structure
```
your_folder/
├── parser.go
├── parser_test.go
└── testdata/
    └── logs.txt       ← provided (10,000 sample log lines)
```

### 5. Expected Output
```
BenchmarkParseA    268    4329510 ns/op    8315234 B/op    116476 allocs/op
BenchmarkParseB    504    2427512 ns/op    5220761 B/op     33408 allocs/op
BenchmarkParseC    586    2037913 ns/op    1302817 B/op     10002 allocs/op
```

### 6. Tools

| Command | What it does |
|---|---|
| `go test -bench=. -benchmem` | Run benchmarks with memory stats |
| `go test -bench=. -benchmem -count=5` | Run 5x for stable numbers |
| `go build -gcflags="-m" .` | Show escape analysis |
| `go test -cpuprofile cpu.out -bench=.` | Generate CPU profile (optional) |
| `go tool pprof cpu.out` | Explore profile (optional) |

### 7. Common Mistakes to Avoid
- Optimizing without measuring first — "premature optimization is the root of all evil"
- Not calling `mapPool.Put(result)` after use — defeats the purpose of sync.Pool
- Not clearing the map after Get() — stale data leaks between uses
- Confusing `b.N` — Go determines the right N automatically, never hardcode it

### 8. Sources
- Go benchmarks: https://pkg.go.dev/testing#hdr-Benchmarks
- `sync.Pool`: https://pkg.go.dev/sync#Pool
- Escape analysis: https://go.dev/doc/faq#stack_or_heap
- pprof tutorial: https://go.dev/blog/pprof

### 9. Knowledge Gained
```
✅ Writing and interpreting Go benchmark tests
✅ benchmem — reading allocation output (B/op, allocs/op)
✅ Escape analysis — understanding stack vs heap
✅ sync.Pool — object reuse pattern
✅ How to approach performance work: measure → profile → optimize → re-measure
```
