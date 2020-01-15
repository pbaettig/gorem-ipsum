package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type healthcheckHandlerConfig struct {
	sync.Mutex
	FailConseq int32
	FailRatio  float32
}

var (
	healthcheckConfig healthcheckHandlerConfig
)

func init() {
	rand.Seed(time.Now().UnixNano())

	healthcheckConfig = healthcheckHandlerConfig{
		FailConseq: 0,
		FailRatio:  0,
	}
}

// HelloWorldHandler says Hello world
func HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, World")
}

// HealthHandler responds to health checks
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if healthcheckConfig.FailConseq > 0 {
		w.WriteHeader(500)

		healthcheckConfig.Lock()
		healthcheckConfig.FailConseq--
		healthcheckConfig.Unlock()
		return
	}

	if healthcheckConfig.FailRatio > 0 {
		if rand.Float32() <= healthcheckConfig.FailRatio {
			w.WriteHeader(500)
			return
		}
	}

	w.WriteHeader(200)
}

// ConfigFailHandler configures healthcheck failures
func ConfigFailHandler(w http.ResponseWriter, r *http.Request) {
	healthcheckConfig.Lock()
	defer healthcheckConfig.Unlock()
	healthcheckConfig.FailConseq = 3

	w.WriteHeader(200)
}
