package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/app"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repositories/memory"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string, body []byte) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, bytes.NewReader(body))
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestUpdateMetricSet(t *testing.T) {
	s := memory.New()
	a := app.NewApp(s, nil)
	h := New(a)

	router := chi.NewRouter()
	router.Post("/updates/", h.UpdateMetricSet)
	ts := httptest.NewServer(router)
	defer ts.Close()

	validJsonBody := `[
		{
			"id": "HeapAlloc",
			"type": "gauge",
			"value": 3782369280
		},
		{
			"id": "PollCount",
			"type": "counter",
			"delta": 123
		}]`

	type want struct {
		code int
	}
	tests := []struct {
		name string
		h    *handlers
		body json.RawMessage
		want want
	}{
		{
			name: "normal",
			h:    h,
			body: json.RawMessage(validJsonBody),
			want: want{200},
		},
		{
			name: "request with invalid body",
			h:    h,
			body: json.RawMessage([]byte("some invalid body")),
			want: want{400},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			resp, b := testRequest(t, ts, "POST", "/updates/", tt.body)
			defer resp.Body.Close()

			t.Log(b)

			require.Equal(t, tt.want.code, resp.StatusCode)
			if tt.want.code != 200 {
				return
			}
		})
	}

}

func TestUpdateMetric(t *testing.T) {
	s := memory.New()
	a := app.NewApp(s, nil)
	h := New(a)

	router := chi.NewRouter()
	router.Post("/update/{mType}/{mName}/{mValue}", h.UpdateMetric)
	router.Get("/", h.GetAllMetricsHTML)
	router.Get("/value/{mType}/{mName}", h.GetMetric)
	ts := httptest.NewServer(router)
	defer ts.Close()

	type want struct {
		code int
	}
	tests := []struct {
		name    string
		h       *handlers
		address string
		mType   string
		mName   string
		mValue  string
		want    want
	}{
		{
			name:    "normal gauge test #1",
			address: "/update/gauge/Alloc/23.23",
			mType:   "gauge",
			mName:   "Alloc",
			mValue:  "23.23",
			h:       h,
			want: want{
				code: 200,
			},
		},
		{
			name:    "first add counter test #2",
			address: "/update/counter/counter/23",
			mType:   "counter",
			mName:   "counter",
			mValue:  "23",
			h:       h,
			want: want{
				code: 200,
			},
		},
		{
			name:    "second add counter test #3",
			address: "/update/counter/counter/7",
			mType:   "counter",
			mName:   "counter",
			mValue:  "30",
			h:       h,
			want: want{
				code: 200,
			},
		},
		{
			name:    "name and value not sent - test #4",
			address: "/update/gauge",
			h:       h,
			want: want{
				code: 404,
			},
		},
		{
			name:    "value not sent - test #5",
			address: "/update/gauge/nbhj",
			h:       h,
			want: want{
				code: 404,
			},
		},
		{
			name:    "wrong metric type- test #5",
			address: "/update/gaunjh/efvefv/eefe",
			h:       h,
			want: want{
				code: 400,
			},
		},
		{
			name:    "nothing was sent - test #6",
			address: "",
			h:       h,
			want: want{
				code: 405,
			},
		},
		{
			name:    "incorrect metric value type - test #6",
			address: "/update/gauge/alloc/string",
			h:       h,
			want: want{
				code: 400,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			resp, _ := testRequest(t, ts, "POST", tt.address, nil)
			defer resp.Body.Close()

			require.Equal(t, tt.want.code, resp.StatusCode)
			if tt.want.code != 200 {
				return
			}
			val, _ := h.app.GetMetricFromParams(context.Background(), tt.mType, tt.mName)
			assert.Equal(t, tt.mValue, string(val))

		})
	}
}

func TestGetMetric(t *testing.T) {
	s := memory.New()
	s.UpdateMetric("gauge", "Alloc", "123.4")
	s.UpdateMetric("counter", "Counter", "123")
	a := app.NewApp(s, nil)
	h := New(a)

	router := chi.NewRouter()
	router.Post("/update/{mType}/{mName}/{mValue}", h.UpdateMetric)
	router.Get("/", h.GetAllMetricsHTML)
	router.Get("/value/{mType}/{mName}", h.GetMetric)
	ts := httptest.NewServer(router)
	defer ts.Close()

	type want struct {
		value string
		code  int
	}
	tests := []struct {
		name    string
		address string
		h       *handlers
		want    want
	}{
		{
			name:    "normal gauge test #1",
			address: "/value/gauge/Alloc",
			h:       h,
			want: want{
				value: "123.4",
				code:  200,
			},
		},
		{
			name:    "normal counter test #2",
			address: "/value/counter/Counter",
			h:       h,
			want: want{
				value: "123",
				code:  200,
			},
		},
		{
			name:    "neg counter test #3",
			address: "/value/counter/erff",
			h:       h,
			want: want{
				value: "",
				code:  404,
			},
		},
		{
			name:    "wrong metric type test #4",
			address: "/value/ssds/erff",
			h:       h,
			want: want{
				value: "",
				code:  400,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, val := testRequest(t, ts, "GET", tt.address, nil)
			defer resp.Body.Close()

			require.Equal(t, tt.want.code, resp.StatusCode)
			if tt.want.code != 200 {
				return
			}
			assert.Equal(t, tt.want.value, string(val))

		})
	}
}

func Example_handlers_UpdateMetricSet() {
	s := memory.New()
	a := app.NewApp(s, nil)
	h := New(a)

	router := chi.NewRouter()
	router.Post("/updates/", h.UpdateMetricSet)
	ts := httptest.NewServer(router)
	defer ts.Close()

	validJsonBody := `[
		{
			"id": "HeapAlloc",
			"type": "gauge",
			"value": 3782369280
		},
		{
			"id": "PollCount",
			"type": "counter",
			"delta": 123
		}
	]`

	req, _ := http.NewRequest("POST", ts.URL+"/updates/", bytes.NewReader([]byte(validJsonBody)))

	resp, _ := ts.Client().Do(req)
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)

	// Output:
	// 200

}

func Example_handlers_UpdateMetric() {

	s := memory.New()
	a := app.NewApp(s, nil)
	h := New(a)

	router := chi.NewRouter()
	router.Post("/update/{mType}/{mName}/{mValue}", h.UpdateMetric)
	ts := httptest.NewServer(router)
	defer ts.Close()

	req, _ := http.NewRequest("POST", ts.URL+"/update/gauge/Alloc/12.12", nil)

	resp, _ := ts.Client().Do(req)
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)

	// Output:
	// 200

}

func Example_handlers_GetMetric() {

	s := memory.New()
	s.UpdateMetric("gauge", "Alloc", "123.4")
	a := app.NewApp(s, nil)
	h := New(a)

	router := chi.NewRouter()
	router.Get("/value/{mType}/{mName}", h.GetMetric)
	ts := httptest.NewServer(router)
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL+"/value/gauge/Alloc", nil)

	resp, _ := ts.Client().Do(req)

	fmt.Println(resp.StatusCode)

	// Output:
	// 200

}

func Example_handlers_UpdateMetricJSON() {

	s := memory.New()
	s.UpdateMetric("gauge", "Alloc", "123.123")
	a := app.NewApp(s, nil)
	h := New(a)

	router := chi.NewRouter()
	router.Post("/update", h.UpdateMetricJSON)
	ts := httptest.NewServer(router)
	defer ts.Close()

	validBody := `{
		"id": "Alloc",
		"type": "gauge",
		"value": 123.123
	}`

	req, _ := http.NewRequest("POST", ts.URL+"/update", bytes.NewReader([]byte(validBody)))

	resp, _ := ts.Client().Do(req)

	fmt.Println(resp.StatusCode)
	b, _ := io.ReadAll(resp.Body)
	fmt.Println(string(b))

	// Output:
	// 200
	// {"id":"Alloc","type":"gauge","value":123.123}

}

func Example_handlers_GetMetricJSON() {

	s := memory.New()
	s.UpdateMetric("gauge", "Alloc", "3782369280")
	a := app.NewApp(s, nil)
	h := New(a)

	router := chi.NewRouter()
	router.Get("/value", h.GetMetricJSON)
	ts := httptest.NewServer(router)
	defer ts.Close()

	validBody := `{
		"id": "Alloc",
		"type": "gauge"
	}`

	req, _ := http.NewRequest("GET", ts.URL+"/value", bytes.NewReader([]byte(validBody)))

	resp, _ := ts.Client().Do(req)

	fmt.Println(resp.StatusCode)
	b, _ := io.ReadAll(resp.Body)
	fmt.Println(string(b))

	// Output:
	// 200
	// {"id":"Alloc","type":"gauge","value":3782369280}

}

func Example_handlers_GetAllMetricJSON() {

	s := memory.New()
	s.UpdateMetric("gauge", "Alloc", "123.4")
	s.UpdateMetric("counter", "PollCount", "12")
	a := app.NewApp(s, nil)
	h := New(a)

	router := chi.NewRouter()
	router.Get("/values", h.GetAllMetricsJSON)
	ts := httptest.NewServer(router)
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL+"/values", nil)

	resp, _ := ts.Client().Do(req)

	fmt.Println(resp.StatusCode)
	b, _ := io.ReadAll(resp.Body)
	fmt.Println(string(b))

	// Output:
	// 200
	// {"Gauge":[{"id":"Alloc","type":"gauge","value":123.4}],"Counter":[{"id":"PollCount","type":"counter","delta":12}]}

}

func Example_handlers_GetAllMetricHTML() {

	s := memory.New()
	s.UpdateMetric("gauge", "Alloc", "123.4")
	s.UpdateMetric("counter", "PollCount", "12")
	a := app.NewApp(s, nil)
	h := New(a)

	router := chi.NewRouter()
	router.Get("/", h.GetAllMetricsHTML)
	ts := httptest.NewServer(router)
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL+"/", nil)

	resp, _ := ts.Client().Do(req)

	fmt.Println(resp.StatusCode)

	b, _ := io.ReadAll(resp.Body)
	fmt.Println(string(b))
	// Output:
	// 200
	//
	//	<html>
	//		<head>
	//			<meta charset="utf-8">
	//			<title>list-style-type</title>
	//			<style>
	//				ul {
	//					list-style-type: none;
	//				}
	//			</style>
	//		</head>
	//		<body>
	//		<ul><li>Alloc = 123.400000</li><li>PollCount = 12</li></ul></body></body>

}
