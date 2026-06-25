package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func NewOnCallStore() OnCallStore {
	oc1 := OnCallEntry{
		ID:       "u1",
		Service:  "payment",
		Username: "anh",
		StartsAt: time.Now().Add(-1 * time.Minute),
		EndsAt:   time.Now().Add(999999 * time.Hour),
	}

	oc2 := OnCallEntry{
		ID:       "u2",
		Service:  "database",
		Username: "bernd",
		StartsAt: time.Now().Add(-1 * time.Minute),
		EndsAt:   time.Now().Add(999999 * time.Hour),
	}

	seedOnCalls := map[string]OnCallEntry{
		oc1.Username: oc1,
		oc2.Username: oc2,
	}

	return NewInMemoryOnCallStore(seedOnCalls)
}

func getMongoClient(conf Config) *mongo.Client {
	if conf.ConnectionString == "" {
		slog.Info("HANDOFF_CONNECT_STRING is empty, use Memory store only")
		return nil
	}
	slog.Info("using mongo store", "db", conf.DatabaseName)
	client, err := mongo.Connect(options.Client().ApplyURI(conf.ConnectionString))
	if err != nil {
		log.Fatal("can't connect to db via HANDOFF_CONNECT_STRING")
	}
	return client
}

func main() {
	// init metrics
	promRegistry := prometheus.NewRegistry()
	httpMetrics := NewHttpMetrics(promRegistry)
	metricRegistry := NewMetricRegistry(promRegistry)
	incidentStoreMetric := NewIncidentStoreMetric(promRegistry)

	// init Registry (Websocket connection)
	registry := NewRegistry(metricRegistry)
	go registry.run()
	defer close(registry.done)

	// init flagHandler
	flagHandler := FlagHandler{store: CreateFlagStore()}

	// init onCallHandler
	onCallHandler := &OnCallHandler{Store: NewOnCallStore()}

	// init config
	config := loadConfig()

	// Init IncidentHandler and its store
	client := getMongoClient(config)
	incidentStore := NewIncidentStore(client, config)
	instrumentedIncidentStore := InstrumentedIncidentStore{
		inner:   incidentStore,
		metrics: incidentStoreMetric,
	}
	incHandler := IncidentHandler{
		IncidentStore: &instrumentedIncidentStore,
		Registry:      registry,
		FlagEvaluator: &flagHandler.store,
		CurrentOnCall: onCallHandler.Store,
	}

	// init authHandler and its store
	userStore := NewUsertStore(client, config)
	authHandler := NewAuthHandler(userStore, []byte(config.JWT_SECRET), time.Duration(15*time.Minute))

	// Set router
	router := getRouter(&incHandler, &flagHandler, authHandler, onCallHandler, client, promRegistry, httpMetrics)

	// run server
	srv := http.Server{
		Addr:    ":" + config.Port,
		Handler: router,
	}
	go func() {
		slog.Info(fmt.Sprintf("server starting http://127.0.0.1%s/", srv.Addr))
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// greaceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	slog.Info("server shut down in <= 10 sec")
	srv.Shutdown(ctx)
	slog.Info("server shut down gracefully")
}
