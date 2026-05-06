package main

import (
	"log"
	"net/http"
)

func main() {

	memoryStore := MemoryStore{incidents: make(map[string]Incident)}
	incHandler := IncidentHandler{Store: &memoryStore}
	router := getRouter(incHandler)

	var srv http.Server
	srv.Addr = ":8080"
	srv.Handler = router

	log.Fatal(srv.ListenAndServe())
}
