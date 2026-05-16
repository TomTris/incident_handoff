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

func NewStore(conf Config) (*mongo.Client, Store) {
	var client *mongo.Client = nil
	var store InstrumentedStore

	if conf.ConnectionString != "" {
		slog.Info("using mongo store", "db", conf.DatabaseName)
		client, err := mongo.Connect(options.Client().ApplyURI(conf.ConnectionString))
		if err != nil {
			log.Fatal(err)
		}
		mongoStore := NewMongoStore(client, conf.DatabaseName)
		store = InstrumentedStore{s: mongoStore}
	} else {
		slog.Info("no connection string, using in-memory store")
		store = InstrumentedStore{NewMemoryStore()}
	}
	return client, &store
}

func main() {
	config := loadConfig()
	client, store := NewStore(config)
	registry := NewRegistry()
	promRegistry := prometheus.NewRegistry()
	NewMetrics(promRegistry)
	incHandler := IncidentHandler{Store: store, Registry: registry, FlagStore: CreateFlagStore()}
	router := getRouter(&incHandler, client, promRegistry)

	srv := http.Server{
		Addr:    ":" + config.Port,
		Handler: router,
	}
	go incHandler.Registry.run()
	defer close(incHandler.Registry.done)

	go func() {
		slog.Info(fmt.Sprintf("server starting port=%s", srv.Addr))
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	slog.Info("server shut down in <= 10 sec")
	srv.Shutdown(ctx)
	slog.Info("server shut down gracefully")
}
