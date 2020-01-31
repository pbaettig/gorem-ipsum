package handlers

import (
	"log"
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/pbaettig/gorem-ipsum/internal/config"
	"github.com/pbaettig/gorem-ipsum/internal/templates"
)

var (
	healthHandlerRequestCount int64 = 0
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// HelloWorldHandler says Hello world
func HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	templates.Base.Render(templates.BaseData{Body: "Hello World"}, w)
}

// HealthHandler responds to health checks
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		atomic.AddInt64(&healthHandlerRequestCount, 1)
	}()

	const (
		failureStatusCode = http.StatusInternalServerError
		successStatusCode = http.StatusOK
	)

	if config.Healthcheck.FailSeq > 0 {
		config.Healthcheck.Lock()
		config.Healthcheck.FailSeq--
		config.Healthcheck.Unlock()

		log.Println("fail consec")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if config.Healthcheck.FailEvery > 0 {
		if healthHandlerRequestCount == int64(config.Healthcheck.FailEvery-1) {
			// TODO: is this concurrency safe or do I need to introduce a lock?
			// healthHandlerRequestCount is increased immediately after returning
			// so we set it to -1 here
			healthHandlerRequestCount = -1

			log.Println("fail every")
			w.WriteHeader(failureStatusCode)
			return
		}
	}

	if config.Healthcheck.FailRatio > 0 {
		r := rand.Float64()
		log.Printf("%f >= %f", config.Healthcheck.FailRatio, r)

		if config.Healthcheck.FailRatio >= r+0.01 {
			log.Println("fail ratio")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(successStatusCode)
}

// GoremIpsumHandler ...
func GoremIpsumHandler(w http.ResponseWriter, r *http.Request) {}

// InfoHandler ...
func InfoHandler(w http.ResponseWriter, r *http.Request) {
	d := new(templates.InfoData)
	d.FromRequest(r)

	templates.Info.Render(*d, w)
}

// ConfigFailHandler configures healthcheck failures
func ConfigFailHandler(w http.ResponseWriter, r *http.Request) {
	config.Healthcheck.Lock()
	defer config.Healthcheck.Unlock()
	config.Healthcheck.FailSeq = 3

	w.WriteHeader(http.StatusOK)
}
