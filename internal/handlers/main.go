package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/pbaettig/gorem-ipsum/internal/metrics"

	"github.com/pbaettig/gorem-ipsum/internal/fifo"

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

	// HealthHistory ...
	HealthHistory handler

	// Count ...
	Count handler

	// HealthConfig configures healthcheck failures ...
	HealthConfig handler

	// RequestGet ...
	RequestGet handler

	// hhstore (health history Store) holds the n last healthcheck results
	hhstore *fifo.Int

	// these predefined headers are used by the handler functions to return response headers
	headerEmpty http.Header
	headerHTML  http.Header
	headerJSON  http.Header

	healthCounter uint64
)

func init() {
	rand.Seed(time.Now().UnixNano())

	Info = newHandler("Info", infoHandler)
	HelloWorld = newHandler("HelloWorld", helloWorldHandler)
	Health = newHandler("Health", healthHandler)
	HealthHistory = newHandler("HealthHistory", healthHistoryHandler)
	Count = newHandler("Count", countHandler)
	HealthConfig = newHandler("HealthConfig", healthConfigHandler)
	RequestGet = newHandler("RequestGet", requestGetHandler)

	// keep a history of the last n healthcheck results
	hhstore = fifo.NewInt(config.HealthHistoryCapacity)

	headerEmpty = make(http.Header)

	headerHTML = make(http.Header)
	headerHTML.Set("Content-Type", "text/html")

	headerJSON = make(http.Header)
	headerJSON.Set("Content-Type", "application/json")

}

type handleFunc func(r *http.Request, h handler) ([]byte, http.Header, int)

// handler ...
type handler struct {
	Name           string
	handleFunc     handleFunc
	RequestCounter *uint64
	lastStatus     *uint64
	log            *log.Entry
}

func newHandler(name string, hf handleFunc) handler {
	rc := new(uint64)
	*rc = 1

	ls := new(uint64)
	*ls = 0

	return handler{
		Name:           name,
		handleFunc:     hf,
		RequestCounter: rc,
		lastStatus:     ls,
		log:            log.WithFields(log.Fields{"handler": name}),
	}
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer atomic.AddUint64(h.RequestCounter, 1)

	lg := log.WithFields(log.Fields{"handler": h.Name})
	start := time.Now()

	buf, header, status := h.handleFunc(r, h)
	defer atomic.StoreUint64(h.lastStatus, uint64(status))

	rh := w.Header()
	for k, vv := range header {
		for _, v := range vv {
			rh.Add(k, v)
		}
	}

	w.WriteHeader(status)
	w.Write(buf)

	took := time.Now().Sub(start)

	metrics.RequestsHistogram.WithLabelValues(r.URL.Path, strconv.Itoa(status)).Observe(took.Seconds())

	lg.Debugf("request counter %d, took %s", *h.RequestCounter, time.Now().Sub(start))
}

// Handlers below

func infoHandler(r *http.Request, h handler) ([]byte, http.Header, int) {
	if err := r.ParseForm(); err != nil {
		h.log.Errorf("cannot parse request params: %s", err.Error())
	}

	d := new(templates.InfoData)

	d.ServerHostname = config.Hostname
	d.FromRequest(r)

	w := new(bytes.Buffer)
	if _, ok := r.Form["pretty"]; ok {
		templates.Info.Render(*d, w)
		return w.Bytes(), headerHTML, 200
	}

	buf, err := json.MarshalIndent(d, "", "   ")
	if err != nil {
		log.Warn(err)
	}

	return buf, headerJSON, 200
}

func healthHandler(r *http.Request, h handler) ([]byte, http.Header, int) {
	atomic.AddUint64(&healthCounter, 1)

	failStatus := http.StatusInternalServerError
	okStatus := http.StatusOK

	if config.Healthcheck.FailSeq > 0 {
		config.Healthcheck.Lock()
		config.Healthcheck.FailSeq--
		config.Healthcheck.Unlock()

		goto Failed

	}

	if config.Healthcheck.FailEvery > 0 {
		if healthCounter%uint64(config.Healthcheck.FailEvery) == 0 {
			goto Failed
		}
	}

	if config.Healthcheck.FailRatio > 0 {
		r := rand.Float64()

		if config.Healthcheck.FailRatio >= r+0.01 {
			goto Failed
		}
	}

	hhstore.Add(okStatus)
	return []byte("healthcheck OK\n"), headerEmpty, okStatus

Failed:
	healthCounter = 0
	hhstore.Add(failStatus)
	return []byte("healthcheck failed\n"), headerEmpty, failStatus
}

func healthHistoryHandler(r *http.Request, h handler) ([]byte, http.Header, int) {
	buf, err := json.MarshalIndent(hhstore.Get(), "", "    ")
	if err != nil {
		return []byte(err.Error() + "\n"), headerEmpty, http.StatusInternalServerError
	}

	return buf, headerJSON, http.StatusOK
}

func countHandler(r *http.Request, h handler) ([]byte, http.Header, int) {
	return []byte(strconv.FormatUint(*h.RequestCounter, 10)), headerEmpty, 200
}

func helloWorldHandler(r *http.Request, h handler) ([]byte, http.Header, int) {
	w := new(bytes.Buffer)

	templates.Base.Render(templates.BaseData{Body: "Hello World"}, w)

	return w.Bytes(), headerEmpty, 200
}

func goremIpsum(r *http.Request, h handler) ([]byte, http.Header, int) {
	return []byte(""), headerEmpty, 200
}

func healthConfigHandler(r *http.Request, h handler) ([]byte, http.Header, int) {
	config.Healthcheck.Lock()
	defer config.Healthcheck.Unlock()

	if err := r.ParseForm(); err != nil {
		return []byte("cannot parse request params\n"), headerEmpty, http.StatusBadRequest
	}

	clearConfig := func() {
		config.Healthcheck.FailEvery = 0
		config.Healthcheck.FailRatio = 0
		config.Healthcheck.FailSeq = 0
	}

	mustGetConfig := func() []byte {
		buf, err := json.MarshalIndent(config.Healthcheck, "", "    ")
		if err != nil {
			panic(err.Error())
		}
		return buf
	}

	_, clear := r.Form["clear"]
	fsv, fsok := r.Form["failseq"]
	frv, frok := r.Form["failratio"]
	fev, feok := r.Form["failevery"]

	if clear {
		clearConfig()
		return mustGetConfig(), headerJSON, 200
	}

	if !fsok && !frok && !feok {
		// no options specified, return current config
		return mustGetConfig(), headerJSON, http.StatusOK
	}

	if fsok && !(frok || feok) {
		v, err := strconv.Atoi(fsv[0])
		if err != nil {
			return []byte(err.Error() + "\n"), headerEmpty, http.StatusBadRequest
		}
		clearConfig()
		config.Healthcheck.FailSeq = v

		return mustGetConfig(), headerJSON, 200
	}

	if frok && !(fsok || feok) {
		v, err := strconv.ParseFloat(frv[0], 64)
		if err != nil {
			return []byte(err.Error() + "\n"), headerEmpty, http.StatusBadRequest
		}
		clearConfig()
		config.Healthcheck.FailRatio = v

		return mustGetConfig(), headerJSON, 200
	}

	if feok && !(fsok || frok) {
		v, err := strconv.Atoi(fev[0])
		if err != nil {
			return []byte(err.Error() + "\n"), headerEmpty, http.StatusBadRequest
		}
		clearConfig()
		config.Healthcheck.FailEvery = v

		return mustGetConfig(), headerJSON, 200
	}

	return []byte("failratio, failevery and failseq are mutually exclusive\n"), headerEmpty, http.StatusBadRequest
}

func requestGetHandler(r *http.Request, h handler) ([]byte, http.Header, int) {
	type jsonResponse struct {
		Headers http.Header
		Body    string
		Status  int
		TookMs  int64
	}

	if err := r.ParseForm(); err != nil {
		return []byte("cannot parse request params\n"), headerEmpty, http.StatusBadRequest
	}
	url, ok := r.Form["url"]
	if !ok {
		return []byte("url query parameter is required\n"), headerEmpty, http.StatusBadRequest
	}

	// Prepare a request with context for timeout

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url[0], nil)
	if err != nil {
		return []byte(fmt.Sprintf("cannot build request: %s\n", err.Error())), headerEmpty, http.StatusInternalServerError
	}

	// execute the request
	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []byte(fmt.Sprintf("request failed: %s\n", err.Error())), headerEmpty, http.StatusInternalServerError
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte(fmt.Sprintf("cannot read response body: %s\n", err.Error())), headerEmpty, http.StatusInternalServerError
	}

	took := time.Now().Sub(start)

	rj := jsonResponse{Headers: resp.Header, Body: string(body), Status: resp.StatusCode, TookMs: took.Milliseconds()}
	buf, err := json.MarshalIndent(&rj, "", "    ")
	if err != nil {
		return []byte(fmt.Sprintf("cannot marshal response JSON: %s\n", err.Error())), headerEmpty, http.StatusInternalServerError
	}
	return buf, headerEmpty, http.StatusOK
}
