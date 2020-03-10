package metrics

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/pbaettig/gorem-ipsum/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	promRequestsCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "sample_processed_requests_total",
		Help: "The total number of processed requests",
	})

	sineGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gorem_sine",
		Help: "A sine wave",
	})

	sawtoothGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gorem_sawtooth",
		Help: "A saw wave",
	})

	triangleGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gorem_triangle",
		Help: "A triangle wave",
	})

	// SineGenerator creates a Sine wave on sineGauge
	SineGenerator = newMetricsGenerator(
		"SineGenerator",
		time.Second,
		240,
		2*math.Pi/240,
		func(mg *MetricsGenerator) float64 {
			return math.Sin(mg.stepSize * float64(mg.stepCount))
		},
		10,
		sineGauge,
	)

	// SawtoothGenerator creates a sawtooth pattern on sawtoothGauge
	SawtoothGenerator = newMetricsGenerator(
		"SawtoothGenerator",
		time.Second,
		120,
		1,
		func(mg *MetricsGenerator) float64 {
			return float64(mg.stepCount) * mg.stepSize
		},
		1,
		sawtoothGauge,
	)

	// TriangleGenerator creates a triangle wave on triangleGauge
	TriangleGenerator = newMetricsGenerator(
		"TriangleGenerator",
		time.Second,
		60,
		1,
		func(mg *MetricsGenerator) float64 {
			if mg.cycleCount%2 == 0 {
				// go down
				maxValue := float64(mg.steps) * mg.stepSize
				return maxValue - (float64(mg.stepCount) * mg.stepSize)
			}
			// go up
			return float64(mg.stepCount) * mg.stepSize
		},
		1,
		triangleGauge,
	)

	// RequestsHistogram ...
	RequestsHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "gorem_requests_duration_histogram",
		Help: "Histogram of the duration of all processed requests",
	}, []string{"base_url", "status"})
)

type MetricsGenerator struct {
	// name to identify the MetricsGenerator in log messages etc.
	name string

	// channel to stop the goroutine started by Run()
	done chan interface{}

	// how often a value will be emitted to `gauge`
	interval time.Duration

	// number of steps before the step counter resets
	steps int

	// value increase inbetween steps
	stepSize float64

	// counter for completed steps
	stepCount int

	// counter for completed cycles (when step counter was reset)
	cycleCount int

	// function that will produce a value based on stepcount (sc), cycles (cs) and stepsize (ss)
	valueFunc func(*MetricsGenerator) float64

	// scaling factor applied to the return value of `valueFunc`
	scale float64

	// gauge that will receive the value produced by `valueFunc`
	gauge prometheus.Gauge
}

func (mg *MetricsGenerator) update() {
	return
}

func newMetricsGenerator(
	name string,
	interval time.Duration,
	steps int,
	stepSize float64,
	valueFunc func(mg *MetricsGenerator) float64,
	scale float64,
	gauge prometheus.Gauge) *MetricsGenerator {

	return &MetricsGenerator{
		name:       name,
		gauge:      gauge,
		done:       make(chan interface{}),
		interval:   interval,
		scale:      scale,
		steps:      steps,
		stepSize:   stepSize,
		cycleCount: 1,
		valueFunc:  valueFunc,
	}
}

func (mg *MetricsGenerator) Run() {
	go func() {
		incrementSc := func() {
			if mg.steps >= 1 {
				if mg.stepCount == mg.steps-1 {
					mg.stepCount = 0
					mg.cycleCount++
				} else {
					mg.stepCount++
				}

			} else {
				mg.stepCount++
			}
		}
		ticker := time.NewTicker(mg.interval)

		log.Debugf("MetricsGenerator \"%s\" started", mg.name)
		for {
			select {
			case <-ticker.C:
				v := mg.valueFunc(mg) * mg.scale
				mg.gauge.Set(v)

				incrementSc()

			case <-mg.done:
				log.Debugf("MetricsGenerator \"%s\" stopped", mg.name)

				ticker.Stop()
				return
			}
		}

	}()
}

func (g *MetricsGenerator) Stop() {
	g.done <- struct{}{}
}

func StartServer(errs chan<- error) *http.Server {
	srv := &http.Server{
		Addr:         config.MetricsServerAddress,
		WriteTimeout: config.MetricsServerWriteTimeout,
		ReadTimeout:  config.MetricsServerReadTimeout,
		IdleTimeout:  config.MetricsServerIdleTimeout,
		Handler:      promhttp.Handler(),
	}
	go func() {
		log.Debugf("starting metrics server on %s", config.MetricsServerAddress)
		if err := srv.ListenAndServe(); err != nil {
			errs <- fmt.Errorf("metrics server: %w", err)
		}
	}()

	return srv
}
