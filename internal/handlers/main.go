package handlers

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/pbaettig/gorem-ipsum/internal/config"
	"github.com/pbaettig/gorem-ipsum/internal/templates"
	log "github.com/sirupsen/logrus"
)

var (
	healthHandlerRequestCount int64 = 0

	// Info ...
	Info handler

	// HelloWorld says Hello world
	HelloWorld handler

	// Health responds to health checks
	Health handler

	// Count ...
	Count handler

	// FailConfig configures healthcheck failures ...
	FailConfig handler
)

type handleFunc func(r *http.Request, h handler) ([]byte, int)

// handler ...
type handler struct {
	Name           string
	handleFunc     handleFunc
	RequestCounter *uint64
	log            *log.Entry
}

func newHandler(name string, hf handleFunc) handler {
	rc := new(uint64)
	*rc = 1

	return handler{
		Name:           name,
		handleFunc:     hf,
		RequestCounter: rc,
		log:            log.WithFields(log.Fields{"handler": name}),
	}
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer atomic.AddUint64(h.RequestCounter, 1)
	lg := log.WithFields(log.Fields{"handler": h.Name})
	start := time.Now()

	buf, status := h.handleFunc(r, h)

	w.WriteHeader(status)
	w.Write(buf)

	lg.Debugf("request counter %d, took %s", *h.RequestCounter, time.Now().Sub(start))
}

func init() {
	rand.Seed(time.Now().UnixNano())

	Info = newHandler("Info", infoHandler)
	HelloWorld = newHandler("HelloWorld", helloWorldHandler)
	Health = newHandler("Health", healthHandler)
	Count = newHandler("Count", countHandler)
	FailConfig = newHandler("FailConfig", failConfigHandler)
}

func infoHandler(r *http.Request, h handler) ([]byte, int) {
	if err := r.ParseForm(); err != nil {
		h.log.Errorf("cannot parse request params: %s", err.Error())
	}

	d := new(templates.InfoData)

	d.ServerHostname = config.Hostname
	d.FromRequest(r)

	w := new(bytes.Buffer)
	if _, ok := r.Form["pretty"]; ok {
		templates.Info.Render(*d, w)
		return w.Bytes(), 200
	}

	buf, err := json.MarshalIndent(d, "", "   ")
	if err != nil {
		log.Warn(err)
	}

	return buf, 200
}

func healthHandler(r *http.Request, h handler) ([]byte, int) {
	if config.Healthcheck.FailSeq > 0 {
		config.Healthcheck.Lock()
		config.Healthcheck.FailSeq--
		config.Healthcheck.Unlock()

		return []byte("healthcheck failed (FailSeq)\n"), http.StatusInternalServerError
	}

	if config.Healthcheck.FailEvery > 0 {
		if *h.RequestCounter%uint64(config.Healthcheck.FailEvery) == 0 {
			return []byte("healthcheck failed (FailEvery)\n"), http.StatusInternalServerError
		}
	}

	if config.Healthcheck.FailRatio > 0 {
		r := rand.Float64()

		if config.Healthcheck.FailRatio >= r+0.01 {
			return []byte("healthcheck failed (FailRatio)\n"), http.StatusInternalServerError
		}
	}

	return []byte("healthcheck OK\n"), http.StatusOK
}

func countHandler(r *http.Request, h handler) ([]byte, int) {
	return []byte(strconv.FormatUint(*h.RequestCounter, 10)), 200
}

func helloWorldHandler(r *http.Request, h handler) ([]byte, int) {
	w := new(bytes.Buffer)

	templates.Base.Render(templates.BaseData{Body: "Hello World"}, w)

	return w.Bytes(), 200
}

func goremIpsum(r *http.Request, h handler) ([]byte, int) {
	return []byte(""), 200
}

func failConfigHandler(r *http.Request, h handler) ([]byte, int) {
	config.Healthcheck.Lock()
	defer config.Healthcheck.Unlock()

	if err := r.ParseForm(); err != nil {
		return []byte("cannot parse request params"), http.StatusInternalServerError
	}

	fsv, fsok := r.Form["failseq"]
	frv, frok := r.Form["failratio"]
	fev, feok := r.Form["failevery"]

	clearConfig := func() {
		config.Healthcheck.FailEvery = 0
		config.Healthcheck.FailRatio = 0
		config.Healthcheck.FailSeq = 0
	}

	if fsok && !(frok || feok) {
		v, err := strconv.Atoi(fsv[0])
		if err != nil {
			return []byte(err.Error()), http.StatusBadRequest
		}
		clearConfig()
		config.Healthcheck.FailSeq = v

		return []byte("failseq"), 200
	}

	if frok && !(fsok || feok) {
		v, err := strconv.ParseFloat(frv[0], 64)
		if err != nil {
			return []byte(err.Error()), http.StatusBadRequest
		}
		clearConfig()
		config.Healthcheck.FailRatio = v

		return []byte("failratio"), 200
	}

	if feok && !(fsok || frok) {
		v, err := strconv.Atoi(fev[0])
		if err != nil {
			return []byte(err.Error()), http.StatusBadRequest
		}
		clearConfig()
		config.Healthcheck.FailEvery = v

		return []byte("failevery"), 200
	}

	return []byte("failratio, failevery and failseq are mutually exclusive"), http.StatusBadRequest
}
