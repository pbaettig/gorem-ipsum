package config

import (
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
)

// HealthcheckConfig ...
type HealthcheckConfig struct {
	sync.Mutex
	FailSeq   int
	FailRatio float64
	FailEvery int
}

var (
	// Hostname ...
	Hostname string

	// Healthcheck ...
	Healthcheck HealthcheckConfig
)

func init() {
	var err error
	Hostname, err = os.Hostname()
	if err != nil {
		log.Panic("cannot determine system hostname")
	}
	Healthcheck = HealthcheckConfig{
		FailSeq:   0,
		FailRatio: 0.0,
		FailEvery: 0,
	}
}
