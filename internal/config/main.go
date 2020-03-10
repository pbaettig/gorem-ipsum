package config

import (
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

var (
	Healthcheck *HealthcheckConfig = &HealthcheckConfig{
		FailSeq:   0,
		FailRatio: 0.0,
		FailEvery: 0,
	}
	HealthHistoryCapacity     int
	Hostname                  string
	LogLevel                  logrus.Level
	MainServerAddress         string
	MainServerIdleTimeout     time.Duration
	MainServerReadTimeout     time.Duration
	MainServerWriteTimeout    time.Duration
	MetricsServerAddress      string
	MetricsServerIdleTimeout  time.Duration
	MetricsServerReadTimeout  time.Duration
	MetricsServerWriteTimeout time.Duration
	ServerShutdownGracePeriod time.Duration
	MetricsSineAmplitude      float64       = 100
	MetricsSineUpdateInterval time.Duration = time.Second
	MetricsSinePeriod         time.Duration = 30 * time.Second
)

// HealthcheckConfig ...
type HealthcheckConfig struct {
	sync.RWMutex
	FailSeq   int
	FailRatio float64
	FailEvery int
}

func getEnv(key, dv string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return dv
	}
	return v
}

func fromEnv() {
	var err error
	MainServerAddress = getEnv("GOREM_MAIN_SERVER_ADDRESS", ":8080")
	MetricsServerAddress = getEnv("GOREM_METRICS_SERVER_ADDRESS", ":9100")

	hhc := getEnv("GOREM_HEALTH_HISTORY_CAPACITY", "10")
	i, err := strconv.ParseInt(hhc, 10, 64)
	if err != nil {
		log.Warnf("unable to parse log level: %s", err.Error())
	} else {
		HealthHistoryCapacity = int(i)
	}

	ll := getEnv("GOREM_LOGLEVEL", "debug")
	LogLevel, err = log.ParseLevel(ll)
	if err != nil {
		log.Warnf("unable to parse log level: %s", err.Error())
	}

	ServerShutdownGracePeriod, err = time.ParseDuration(getEnv("GOREM_SERVER_SHUTDOWN_GRACE", "15s"))
	if err != nil {
		log.Warnf("unable to parse server shutdown grace period: %s", err.Error())
	}

	MainServerIdleTimeout, err = time.ParseDuration(getEnv("GOREM_MAIN_SERVER_IDLE_TIMEOUT", "60s"))
	if err != nil {
		log.Warnf("unable to parse main server idle timeout: %s", err.Error())
	}

	MainServerReadTimeout, err = time.ParseDuration(getEnv("GOREM_MAIN_SERVER_READ_TIMEOUT", "15s"))
	if err != nil {
		log.Warnf("unable to parse main server read timeout: %s", err.Error())
	}

	MainServerWriteTimeout, err = time.ParseDuration(getEnv("GOREM_MAIN_SERVER_WRITE_TIMEOUT", "15s"))
	if err != nil {
		log.Warnf("unable to parse main server write timeout: %s", err.Error())
	}

	MetricsServerIdleTimeout, err = time.ParseDuration(getEnv("GOREM_METRICS_SERVER_IDLE_TIMEOUT", "60s"))
	if err != nil {
		log.Warnf("unable to parse metrics server idle timeout: %s", err.Error())
	}

	MetricsServerReadTimeout, err = time.ParseDuration(getEnv("GOREM_METRICS_SERVER_READ_TIMEOUT", "15s"))
	if err != nil {
		log.Warnf("unable to parse metrics server read timeout: %s", err.Error())
	}

	MetricsServerWriteTimeout, err = time.ParseDuration(getEnv("GOREM_METRICS_SERVER_WRITE_TIMEOUT", "15s"))
	if err != nil {
		log.Warnf("unable to parse metrics server write timeout: %s", err.Error())
	}
}

func init() {
	var err error
	Hostname, err = os.Hostname()
	if err != nil {
		log.Panic("cannot determine system hostname")
	}

	fromEnv()
}
