package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"

	"github.com/pbaettig/gorem-ipsum/internal/config"
	"github.com/pbaettig/gorem-ipsum/internal/metrics"
	"github.com/pbaettig/gorem-ipsum/internal/middleware"

	"github.com/gorilla/mux"
	"github.com/pbaettig/gorem-ipsum/internal/handlers"
)

func init() {
	log.SetLevel(config.LogLevel)
}

func mainServer(h http.Handler, errs chan<- error) *http.Server {
	srv := &http.Server{
		Addr:         config.MainServerAddress,
		WriteTimeout: config.MainServerWriteTimeout,
		ReadTimeout:  config.MainServerReadTimeout,
		IdleTimeout:  config.MainServerIdleTimeout,
		Handler:      h,
	}
	go func() {
		log.Debugf("starting main server on %s", config.MainServerAddress)
		if err := srv.ListenAndServe(); err != nil {
			errs <- fmt.Errorf("main server: %w", err)
		}
	}()

	return srv
}

func setupRouter() *mux.Router {
	// Setup Routes
	root := mux.NewRouter()

	root.Use(middleware.Log)
	root.Handle("/", handlers.HelloWorld)
	root.Handle("/health", handlers.Health)
	root.Handle("/health/history", handlers.HealthHistory)
	root.Handle("/info", handlers.Info)
	root.Handle("/count", handlers.Count)
	root.Handle("/http/get", handlers.RequestGet)
	root.Handle("/http/post", handlers.HelloWorld)

	configSubrouter := root.PathPrefix("/config").Subrouter()
	configSubrouter.Handle("/health", handlers.HealthConfig)
	configSubrouter.Use(middleware.Authenticate)

	return root
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	// Start HTTP servers
	srvError := make(chan error)
	mainSrv := mainServer(setupRouter(), srvError)
	metricSrv := metrics.StartServer(srvError)

	// start metrics generators
	metrics.SineGenerator.Run()
	metrics.SawtoothGenerator.Run()
	metrics.TriangleGenerator.Run()

	select {
	case s := <-sigs:
		fmt.Println()
		log.Debugf("Signal '%s' received", s.String())

		metrics.SineGenerator.Stop()
		metrics.SawtoothGenerator.Stop()
		metrics.TriangleGenerator.Stop()

		// create a context to wait for open connections when shutting down servers
		ctx, cancel := context.WithTimeout(context.Background(), config.ServerShutdownGracePeriod)
		defer cancel()

		log.Debug("shutting down servers")
		mainSrv.Shutdown(ctx)
		metricSrv.Shutdown(ctx)

		log.Debug("goodbye")
		os.Exit(0)

	case err := <-srvError:
		// exit with an error if any one of the servers failed to start
		log.Fatal(err.Error())
	}

}
