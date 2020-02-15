package main

import (
	"net/http"

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

	root.Use(middleware.Log)
	root.Handle("/", handlers.HelloWorld)
	root.Handle("/health", handlers.Health)
	root.Handle("/info", handlers.Info)
	root.Handle("/count", handlers.Count)

	config.Handle("/fail", handlers.FailConfig)
	config.Use(middleware.Authenticate)

	http.Handle("/", root)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
