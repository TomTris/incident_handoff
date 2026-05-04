package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

const (
	numFiles  = 7
	numLines  = 300
	outputDir = "logs"
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

func getTimestamp(previousTs time.Time) time.Time {
	return previousTs.Add(time.Duration(1000+rand.Intn(14001)) * time.Millisecond)
}

func getLevel() string {
	// Level distribution: 70% INFO, 20% WARN, 10% ERROR
	levelValue := rand.Intn(100)
	switch {
	case levelValue < 70:
		return "INFO"
	case levelValue < 90:
		return "WARN"
	default:
		return "ERROR"
	}
}

func getTraceID() string {
	return fmt.Sprintf("%08x", rand.Uint32())
}

func getMessage(level string) string {
	switch {
	case level == "INFO":
		return infoMessages[rand.Intn(len(infoMessages))]
	case level == "WARN":
		return warnMessages[rand.Intn(len(warnMessages))]
	default:
		return errorMessages[rand.Intn(len(errorMessages))]
	}
}

func logGenerate(idx int, serviceName string, baseTs time.Time) {
	fileName := outputDir + "/" + serviceName + "-" + strconv.Itoa(idx+1) + ".log"
	f, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	currentTime := baseTs
	for i := 0; i < numLines; i++ {
		level := getLevel()
		traceID := getTraceID()
		message := getMessage(level)
		currentTime = getTimestamp(currentTime)
		line := fmt.Sprintf("%s [%-5s] [%s] [trace=%s] %s\n",
			currentTime.Format("2026-05-01T08:00:03.000Z"), level, serviceName, traceID, message)
		_, err := f.WriteString(line)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf(" ✔ %s (%d lines)\n", fileName, numLines)
}

func main() {
	err := os.MkdirAll(outputDir, 0750)
	if err != nil {
		log.Fatal(err)
	}
	// -24 so we have log in the past
	baseTs := time.Now().Add(-24 * time.Hour)
	fmt.Println("Generating 7 log files (300 lines each)...\n")
	for idx, service := range services {
		logGenerate(idx, service, baseTs)
	}
	fmt.Println("Done!")
}
