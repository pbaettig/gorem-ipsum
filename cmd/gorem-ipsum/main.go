package main

import (
	"log"
	"net/http"

	"github.com/pbaettig/gorem-ipsum/internal/config"

	"github.com/pbaettig/gorem-ipsum/internal/middleware"

	"github.com/gorilla/mux"
	"github.com/pbaettig/gorem-ipsum/internal/handlers"
)

func main() {
	config.Healthcheck.FailSeq = 0
	config.Healthcheck.FailRatio = 0
	config.Healthcheck.FailEvery = 3

	root := mux.NewRouter()
	config := root.PathPrefix("/config").Subrouter()

	root.Use(middleware.Log)
	root.HandleFunc("/", handlers.HelloWorldHandler)
	root.HandleFunc("/health", handlers.HealthHandler)
	root.HandleFunc("/info", handlers.InfoHandler)

	config.HandleFunc("/fail", handlers.ConfigFailHandler)
	config.Use(middleware.Authenticate)

	http.Handle("/", root)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
