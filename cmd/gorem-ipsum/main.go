package main

import (
	"log"
	"net/http"

	"github.com/pbaettig/gorem-ipsum/internal/middleware"

	"github.com/gorilla/mux"
	"github.com/pbaettig/gorem-ipsum/internal/handlers"
)

func main() {
	root := mux.NewRouter()
	config := root.PathPrefix("/config").Subrouter()

	root.Use(middleware.Log)
	root.HandleFunc("/", handlers.HelloWorldHandler)
	root.HandleFunc("/health", handlers.HealthHandler)

	config.HandleFunc("/fail", handlers.ConfigFailHandler)
	config.Use(middleware.Authenticate)

	http.Handle("/", root)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
