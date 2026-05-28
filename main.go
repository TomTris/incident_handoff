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

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func NewIncidentStore(conf Config) (*mongo.Client, IncidentStore) {
	var client *mongo.Client = nil
	var store IncidentStore
	if conf.ConnectionString != "" {
		slog.Info("using mongo store", "db", conf.DatabaseName)
		client, err := mongo.Connect(options.Client().ApplyURI(conf.ConnectionString))
		if err != nil {
			log.Fatal(err)
		}
		store = NewMongoIncidentStore(client, conf.DatabaseName)
	} else {
		slog.Info("no connection string, using in-memory store")
		store = NewMemoryIncidentStore()
	}
	return client, store
}

func main() {
	godotenv.Load()
	// init metrics
	promRegistry := prometheus.NewRegistry()
	httpMetrics := NewHttpMetrics(promRegistry)
	registryMetric := NewRegistryMetric(promRegistry)
	incidentStoreMetric := NewIncidentStoreMetric(promRegistry)

	// init Registry (Websocket connection)
	registry := NewRegistry(registryMetric)
	go registry.run()
	defer close(registry.done)

	// init flagHandler
	flagHandler := FlagHandler{store: CreateFlagStore()}

	// Init IncidentHandler and its store
	config := loadConfig()
	client, incidentStore := NewIncidentStore(config)
	instrumentedIncidentStore := InstrumentedIncidentStore{
		inner:   incidentStore,
		metrics: incidentStoreMetric,
	}
	incHandler := IncidentHandler{
		IncidentStore: &instrumentedIncidentStore,
		Registry:      registry,
		FlagEvaluator: &flagHandler.store,
	}

	// init authHandler
	var seedUsers = []User{
		{ID: "u1", Username: "anh", Password: hashPassword("anh123"), Role: "engineer"},
		{ID: "u2", Username: "bernd", Password: hashPassword("bernd123"), Role: "engineer"},
		{ID: "u3", Username: "admin", Password: hashPassword("admin123"), Role: "admin"},
	}
	userStore := NewInMemoryUserStore(seedUsers)
	authHandler := NewAuthHandler(userStore, []byte(config.JWT_SECRET), time.Duration(15))

	// Set router
	router := getRouter(&incHandler, &flagHandler, authHandler, client, promRegistry, httpMetrics)

	// run server
	srv := http.Server{
		Addr:    ":" + config.Port,
		Handler: router,
	}
	go func() {
		slog.Info(fmt.Sprintf("server starting port=%s", srv.Addr))
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
