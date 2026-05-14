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

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func NewStore(conf Config) (*mongo.Client, Store) {
	if conf.ConnectionString != "" {
		slog.Info("using mongo store", "db", conf.DatabaseName)
		client := connectMongoDB(conf.ConnectionString)
		mongoStore := NewMongoStore(client, conf.DatabaseName)
		return client, mongoStore
	}
	slog.Info("no connection string, using in-memory store")
	return nil, NewMemoryStore()
}

func main() {
	config := loadConfig()
	client, store := NewStore(config)
	registry := NewRegistry()
	incHandler := IncidentHandler{Store: store, Registry: registry}
	router := getRouter(&incHandler, client)

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
