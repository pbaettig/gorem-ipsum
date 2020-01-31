package config

import (
	"sync"
)

type HealthcheckConfig struct {
	sync.Mutex
	FailSeq   int
	FailRatio float64
	FailEvery int
}

var Healthcheck HealthcheckConfig

func init() {
	Healthcheck = HealthcheckConfig{
		FailSeq:   0,
		FailRatio: 0.0,
		FailEvery: 0,
	}
}
