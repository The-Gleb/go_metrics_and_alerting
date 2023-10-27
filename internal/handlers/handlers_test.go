package handlers

// import (
// 	"io"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/The-Gleb/go_metrics_and_alerting/internal/app"
// 	"github.com/The-Gleb/go_metrics_and_alerting/internal/storage"
// 	"github.com/go-chi/chi/v5"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// func testRequest(t *testing.T, ts *httptest.Server, method,
// 	path string) (*http.Response, string) {
// 	req, err := http.NewRequest(method, ts.URL+path, nil)
// 	require.NoError(t, err)

// 	resp, err := ts.Client().Do(req)
// 	require.NoError(t, err)
// 	defer resp.Body.Close()

// 	respBody, err := io.ReadAll(resp.Body)
// 	require.NoError(t, err)

// 	return resp, string(respBody)
// }

// func Test_handlers_UpdateMetric(t *testing.T) {
// 	s := storage.New()
// 	a := app.NewApp(s)
// 	h := New(s, a)
// 	router := chi.NewRouter()
// 	router.Post("/update/{mType}/{mName}/{mValue}", h.UpdateMetric)
// 	router.Get("/", h.GetAllMetricsHTML)
// 	router.Get("/value/{mType}/{mName}", h.GetMetric)
// 	ts := httptest.NewServer(router)
// 	defer ts.Close()

// 	type want struct {
// 		code int
// 	}
// 	tests := []struct {
// 		name    string
// 		h       *handlers
// 		address string
// 		mType   string
// 		mName   string
// 		mValue  string
// 		want    want
// 	}{
// 		{
// 			name:    "normal gauge test #1",
// 			address: "/update/gauge/Alloc/23.23",
// 			mType:   "gauge",
// 			mName:   "Alloc",
// 			mValue:  "23.23",
// 			h:       h,
// 			want: want{
// 				code: 200,
// 			},
// 		},
// 		{
// 			name:    "first add counter test #2",
// 			address: "/update/counter/counter/23",
// 			mType:   "counter",
// 			mName:   "counter",
// 			mValue:  "23",
// 			h:       h,
// 			want: want{
// 				code: 200,
// 			},
// 		},
// 		{
// 			name:    "second add counter test #3",
// 			address: "/update/counter/counter/7",
// 			mType:   "counter",
// 			mName:   "counter",
// 			mValue:  "30",
// 			h:       h,
// 			want: want{
// 				code: 200,
// 			},
// 		},
// 		{
// 			name:    "name and value not sent - test #4",
// 			address: "/update/gauge",
// 			h:       h,
// 			want: want{
// 				code: 404,
// 			},
// 		},
// 		{
// 			name:    "value not sent - test #5",
// 			address: "/update/gauge/nbhj",
// 			h:       h,
// 			want: want{
// 				code: 404,
// 			},
// 		},
// 		{
// 			name:    "wrong metric type- test #5",
// 			address: "/update/gaunjh/efvefv/eefe",
// 			h:       h,
// 			want: want{
// 				code: 400,
// 			},
// 		},
// 		{
// 			name:    "nothing was sent - test #6",
// 			address: "",
// 			h:       h,
// 			want: want{
// 				code: 405,
// 			},
// 		},
// 		{
// 			name:    "incorrect metric value type - test #6",
// 			address: "/update/gauge/alloc/string",
// 			h:       h,
// 			want: want{
// 				code: 400,
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {

// 			resp, _ := testRequest(t, ts, "POST", tt.address)
// 			defer resp.Body.Close()

// 			require.Equal(t, tt.want.code, resp.StatusCode)
// 			if tt.want.code != 200 {
// 				return
// 			}
// 			val, _ := h.storage.GetMetric(tt.mType, tt.mName)
// 			assert.Equal(t, tt.mValue, val)

// 		})
// 	}
// }

// func Test_handlers_GetMetric(t *testing.T) {
// 	s := storage.New()
// 	s.UpdateMetric("gauge", "Alloc", "123.4")
// 	s.UpdateMetric("counter", "Counter", "123")
// 	a := app.NewApp(s)
// 	h := New(s, a)
// 	router := chi.NewRouter()
// 	router.Post("/update/{mType}/{mName}/{mValue}", h.UpdateMetric)
// 	router.Get("/", h.GetAllMetricsHTML)
// 	router.Get("/value/{mType}/{mName}", h.GetMetric)
// 	ts := httptest.NewServer(router)
// 	defer ts.Close()

// 	type want struct {
// 		value string
// 		code  int
// 	}
// 	tests := []struct {
// 		name    string
// 		address string
// 		h       *handlers
// 		want    want
// 	}{
// 		{
// 			name:    "normal gauge test #1",
// 			address: "/value/gauge/Alloc",
// 			h:       h,
// 			want: want{
// 				value: "123.4",
// 				code:  200,
// 			},
// 		},
// 		{
// 			name:    "normal counter test #2",
// 			address: "/value/counter/Counter",
// 			h:       h,
// 			want: want{
// 				value: "123",
// 				code:  200,
// 			},
// 		},
// 		{
// 			name:    "neg counter test #3",
// 			address: "/value/counter/erff",
// 			h:       h,
// 			want: want{
// 				value: "",
// 				code:  404,
// 			},
// 		},
// 		{
// 			name:    "wrong metric type test #4",
// 			address: "/value/ssds/erff",
// 			h:       h,
// 			want: want{
// 				value: "",
// 				code:  400,
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			resp, val := testRequest(t, ts, "GET", tt.address)
// 			defer resp.Body.Close()

// 			require.Equal(t, tt.want.code, resp.StatusCode)
// 			if tt.want.code != 200 {
// 				return
// 			}
// 			assert.Equal(t, tt.want.value, val)

// 		})
// 	}
// }
