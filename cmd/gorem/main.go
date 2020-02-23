package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/pbaettig/gorem-ipsum/internal/middleware"

	"github.com/gorilla/mux"
	"github.com/pbaettig/gorem-ipsum/internal/handlers"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func main() {
	root := mux.NewRouter()
	config := root.PathPrefix("/config").Subrouter()
	internal := root.PathPrefix("/internal").Subrouter()

	root.Use(middleware.Log)
	root.Handle("/", handlers.HelloWorld)
	root.Handle("/health", handlers.Health)
	root.Handle("/health/history", handlers.HealthHistory)
	root.Handle("/info", handlers.Info)
	root.Handle("/count", handlers.Count)
	root.Handle("/http/get", handlers.HelloWorld)
	root.Handle("/http/post", handlers.HelloWorld)

	config.Handle("/health", handlers.HealthConfig)
	config.Use(middleware.Authenticate)

	internal.Handle("/metrics", promhttp.Handler())

	http.Handle("/", root)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
