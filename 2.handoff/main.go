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
)

func main() {
	memoryStore := MemoryStore{incidents: make(map[string]Incident)}
	incHandler := IncidentHandler{Store: &memoryStore}
	router := getRouter(incHandler)

	config := loadConfig()
	srv := http.Server{
		Addr:    ":" + config.Port,
		Handler: router,
	}
	fmt.Println(srv.Addr)

	go func() {
		slog.Info("server running")
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
	slog.Info("server shut down in 10 sec")
	srv.Shutdown(ctx)
	slog.Info("server shut down gracefully")
}
