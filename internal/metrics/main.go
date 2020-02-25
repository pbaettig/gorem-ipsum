package metrics

import (
	"math"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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

	// RequestsHistogram ...
	RequestsHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "gorem_requests_duration_histogram",
		Help: "Histogram of the duration of all processed requests",
	}, []string{"base_url", "status"})
)

type SineGenerator struct {
	done     chan interface{}
	interval time.Duration
	scale    float64
	steps    int
}

func (g *SineGenerator) Run() {
	go func() {
		ticker := time.NewTicker(g.interval)

		sc := 0
		ss := 2 * math.Pi / float64(g.steps)

		log.Debug("SineGenerator started")
		for {
			select {
			case <-ticker.C:
				sineGauge.Set(math.Sin(ss*float64(sc)) * g.scale)
				sc = (sc + 1) % g.steps

			case <-g.done:
				ticker.Stop()
				return
			}
		}

	}()
}

func (g *SineGenerator) Stop() {
	g.done <- struct{}{}
}

func NewSineGenerator(d time.Duration, scale float64, steps int) *SineGenerator {
	return &SineGenerator{
		done:     make(chan interface{}),
		interval: d,
		scale:    scale,
		steps:    steps,
	}
}
