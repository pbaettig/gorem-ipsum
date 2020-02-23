package config

import (
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
)

var (
	// Hostname ...
	Hostname string

	// Healthcheck ...
	Healthcheck *HealthcheckConfig
)

const (
	// HealthHistoryCapacity ...
	HealthHistoryCapacity = 10
)

// HealthcheckConfig ...
type HealthcheckConfig struct {
	sync.RWMutex
	FailSeq   int
	FailRatio float64
	FailEvery int
}

func init() {
	var err error
	Hostname, err = os.Hostname()
	if err != nil {
		log.Panic("cannot determine system hostname")
	}
	Healthcheck = &HealthcheckConfig{
		FailSeq:   0,
		FailRatio: 0.0,
		FailEvery: 0,
	}
}
