package handlers

import (
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
			HelloWorldHandler(tt.args.w, tt.args.r)
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

func TestHealthHandler(t *testing.T) {
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	type want struct {
		status int
		body   []byte
	}
	type conf struct {
		FailSeq   int
		failRatio float64
	}
	tests := []struct {
		name string
		args args
		want want
		conf conf
	}{
		{
			"FailRatio100",
			args{httptest.NewRecorder(), httptest.NewRequest("GET", "http://example.com/foo", nil)},
			want{500, nil},
			conf{FailSeq: 0, failRatio: 1},
		},
		{
			"FailRatio0",
			args{httptest.NewRecorder(), httptest.NewRequest("GET", "http://example.com/foo", nil)},
			want{200, nil},
			conf{FailSeq: 0, failRatio: 0},
		},
		{
			"FailRatio0.5",
			args{httptest.NewRecorder(), httptest.NewRequest("GET", "http://example.com/foo", nil)},
			want{200, nil},
			conf{FailSeq: 0, failRatio: 0.5},
		},
		{
			"FailSeq1",
			args{httptest.NewRecorder(), httptest.NewRequest("GET", "http://example.com/foo", nil)},
			want{500, nil},
			conf{FailSeq: 1, failRatio: 0},
		},
	}
	for _, tt := range tests {
		config.Healthcheck.FailSeq = tt.conf.FailSeq
		config.Healthcheck.FailRatio = tt.conf.failRatio
		t.Run(tt.name, func(t *testing.T) {
			HealthHandler(tt.args.w, tt.args.r)
		})

		resp := tt.args.w.Result()
		if resp.StatusCode != tt.want.status {
			t.Errorf("Test %s failed. Wanted HTTP status %d, got %d", tt.name, tt.want.status, resp.StatusCode)
			t.FailNow()
		}

		b, _ := ioutil.ReadAll(resp.Body)
		if string(b) != string(tt.want.body) {
			t.Errorf("Test %s failed. Wanted HTTP response body '%s', got '%s'", tt.name, "Hello, world", string(b))
			t.FailNow()
		}
	}
}

func TestHealthHandlerRatios(t *testing.T) {
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	const n = 1000
	t.Run("test-ratios", func(t *testing.T) {
		for ratio := 0.1; ratio < 1; ratio += 0.1 {
			config.Healthcheck.FailSeq = 0
			config.Healthcheck.FailRatio = ratio

			failed := 0
			for i := 0; i < n; i++ {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "http://example.com/foo", nil)

				HealthHandler(w, r)
				resp := w.Result()
				if resp.StatusCode == http.StatusInternalServerError {
					failed++
				}
			}

			if fr := toFixed(float64(failed)/n, 1); fr != toFixed(ratio, 1) {
				t.Errorf("expected failure ratio: %f, actual failure ratio %f", ratio, fr)
				t.FailNow()

			}

		}
	})

}
