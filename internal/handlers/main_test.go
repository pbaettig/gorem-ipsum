package handlers

import (
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pbaettig/gorem-ipsum/internal/config"
)

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(int(math.Round(num*output))) / output
}

func TestHelloWorldHandler(t *testing.T) {
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	type want struct {
		status int
		body   []byte
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			"test1",
			args{httptest.NewRecorder(), httptest.NewRequest("GET", "http://example.com/foo", nil)},
			want{200, []byte("<p>Hello World</p>")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HelloWorld.ServeHTTP(tt.args.w, tt.args.r)
		})
		resp := tt.args.w.Result()
		if resp.StatusCode != tt.want.status {
			t.Errorf("Test %s failed. Wanted HTTP status %d, got %d", tt.name, tt.want.status, resp.StatusCode)
			t.FailNow()
		}

		b, _ := ioutil.ReadAll(resp.Body)
		if string(b) != string(tt.want.body) {
			t.Errorf("Test %s failed. Wanted HTTP response body '%s', got '%s'", tt.name, tt.want.body, string(b))
			t.FailNow()
		}
	}
}

func invokeHealthHandler() int {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://example.com/health", nil)

	Health.ServeHTTP(w, r)
	return w.Result().StatusCode
}

func mustStatus(t *testing.T, resp *http.Response, status int) {
	as := resp.StatusCode
	if status != http.StatusOK {
		t.Errorf("Request returned HTTP%d, wanted HTTP%d", as, status)
		t.FailNow()
	}
}

func mustConfigRatio(t *testing.T, n float64) {
	wc := httptest.NewRecorder()

	r := httptest.NewRequest("GET", fmt.Sprintf("http://example.com/config/health?failratio=%f", n), nil)
	r.Header = map[string][]string{
		"Authorization": []string{"Authorization: Basic dTpw"},
	}

	HealthConfig.ServeHTTP(wc, r)
	mustStatus(t, wc.Result(), http.StatusOK)
}

func mustConfigEvery(t *testing.T, n int) {
	wc := httptest.NewRecorder()
	r := httptest.NewRequest("GET", fmt.Sprintf("http://example.com/config/health?failevery=%d", n), nil)
	r.Header = map[string][]string{
		"Authorization": []string{"Authorization: Basic dTpw"},
	}
	HealthConfig.ServeHTTP(wc, r)
	mustStatus(t, wc.Result(), http.StatusOK)
}

func mustConfigSeq(t *testing.T, n int) {
	wc := httptest.NewRecorder()
	r := httptest.NewRequest("GET", fmt.Sprintf("http://example.com/config/health?failseq=%d", n), nil)
	r.Header = map[string][]string{
		"Authorization": {"Authorization: Basic dTpw"},
	}
	HealthConfig.ServeHTTP(wc, r)
	mustStatus(t, wc.Result(), http.StatusOK)
}

func TestHealthHandler(t *testing.T) {
	t.Run("ratio", func(t *testing.T) {
		const n = 1000
		for ratio := 0.1; ratio < 1; ratio += 0.1 {
			mustConfigRatio(t, ratio)

			failed := 0
			for i := 0; i < n; i++ {
				if invokeHealthHandler() == http.StatusInternalServerError {
					failed++
				}
			}

			if fr := toFixed(float64(failed)/n, 1); fr != toFixed(ratio, 1) {
				t.Errorf("expected failure ratio: %f, actual failure ratio %f", ratio, fr)
				t.FailNow()

			}

		}
	})

	t.Run("seq", func(t *testing.T) {
		for seq := 1; seq < 50; seq++ {
			mustConfigSeq(t, seq)

			for attempt := 1; attempt <= seq; attempt++ {
				status := invokeHealthHandler()
				if status != http.StatusInternalServerError {
					t.Errorf("expected failed healthcheck on attempt #%d (failseq config: %d), but got HTTP%d", attempt, seq, status)
					t.FailNow()
				}
			}
			if invokeHealthHandler() != http.StatusOK {
				t.Errorf("expected successful healthcheck on attempt #%d (failseq config: %d)", seq+1, seq)
				t.FailNow()
			}
		}
	})

	t.Run("every", func(t *testing.T) {
		for every := 1; every < 200; every++ {
			mustConfigEvery(t, every)

			for attempt := 1; attempt <= every*10; attempt++ {
				if attempt%every == 0 {
					status := invokeHealthHandler()
					if status != http.StatusInternalServerError {
						t.Errorf("expected failed healthcheck on attempt #%d (failevery config: %d), but got HTTP%d", attempt, every, status)
						t.FailNow()
					}
				} else {
					if invokeHealthHandler() != http.StatusOK {
						t.Errorf("expected successful healthcheck on attempt #%d (failevery config: %d)", attempt, every)
						t.FailNow()
					}
				}

			}
		}
	})
}

func TestHealthHistory(t *testing.T) {
	last := func(s []int) int {
		return s[len(s)-1]
	}

	n := 3
	mustConfigEvery(t, n)

	for i := 1; i <= config.HealthHistoryCapacity; i++ {
		_ = invokeHealthHandler()
		if i%n == 0 {
			e := last(hhstore.Get())
			want := http.StatusInternalServerError
			if e != want {
				t.Errorf("last entry in hhstore is %d, should be %d", e, want)
				t.FailNow()
			}
		}
	}

	l := len(hhstore.Get())
	if l != config.HealthHistoryCapacity {
		t.Errorf("hhstore length should be %d, is %d", config.HealthHistoryCapacity, l)
		t.FailNow()
	}
}
