package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/pbaettig/gorem-ipsum/internal/middleware"

	"github.com/gorilla/mux"
	"github.com/pbaettig/gorem-ipsum/internal/handlers"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func mainServer(h http.Handler, errs chan<- error) *http.Server {
	srv := &http.Server{
		Addr:         "0.0.0.0:8080",
		WriteTimeout: time.Second * 5,
		ReadTimeout:  time.Second * 5,
		IdleTimeout:  time.Second * 60,
		Handler:      h,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			errs <- fmt.Errorf("main server: %w", err)
		}
	}()

	return srv
}

func metricsServer(errs chan<- error) *http.Server {
	srv := &http.Server{
		Addr:         "0.0.0.0:9100",
		WriteTimeout: time.Second * 5,
		ReadTimeout:  time.Second * 5,
		IdleTimeout:  time.Second * 60,
		Handler:      promhttp.Handler(),
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			errs <- fmt.Errorf("metrics server: %w", err)
		}
	}()

	return srv
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	root := mux.NewRouter()
	config := root.PathPrefix("/config").Subrouter()
	// internal := root.PathPrefix("/internal").Subrouter()

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

	srvError := make(chan error)
	mainSrv := mainServer(root, srvError)
	metricSrv := metricsServer(srvError)

	for {
		select {
		case s := <-sigs:
			log.Debugf("Singal '%s' received", s.String())

			// create a context to wait for open connections when shutting down servers
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(15*time.Second))
			defer cancel()

			log.Info("shutting down servers")
			mainSrv.Shutdown(ctx)
			metricSrv.Shutdown(ctx)

			log.Info("goodbye")
			os.Exit(0)

		case err := <-srvError:
			// exit with an error if any one of the servers failed to start
			log.Fatal(err.Error())
		}
	}
}
