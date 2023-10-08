package handlers

import (
// "io"
// "net/http"
// "net/http/httptest"
// "testing"

// "github.com/The-Gleb/go_metrics_and_alerting/internal/storage"
// "github.com/go-chi/chi/v5"
// "github.com/stretchr/testify/assert"
// "github.com/stretchr/testify/require"
)

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
// 	h := New(s)
// 	router := chi.NewRouter()
// 	router.Post("/update/{mType}/{mName}/{mValue}", h.UpdateMetric)
// 	router.Get("/", h.GetAllMetrics)
// 	router.Get("/value/{mType}/{mName}", h.GetMetric)
// 	ts := httptest.NewServer(router)
// 	defer ts.Close()

// 	type want struct {
// 		code int
// 		// err  error
// 	}
// 	tests := []struct {
// 		name         string
// 		h            *handlers
// 		address      string
// 		mType        string
// 		mName        string
// 		gaugeValue   float64
// 		counterValue int64
// 		want         want
// 	}{
// 		{
// 			name:       "normal gauge test #1",
// 			address:    "/update/gauge/alloc/23.23",
// 			mType:      "gauge",
// 			mName:      "alloc",
// 			gaugeValue: 23.23,
// 			h:          h,
// 			want: want{
// 				code: 200,
// 			},
// 		},
// 		{
// 			name:         "first add counter test #1",
// 			address:      "/update/counter/counter/23345",
// 			mType:        "counter",
// 			mName:        "counter",
// 			counterValue: 23345,
// 			h:            h,
// 			want: want{
// 				code: 200,
// 			},
// 		},
// 		{
// 			name:         "second add counter test #1",
// 			address:      "/update/counter/counter/27",
// 			mType:        "counter",
// 			mName:        "counter",
// 			counterValue: 27,
// 			h:            h,
// 			want: want{
// 				code: 200,
// 			},
// 		},
// 		{
// 			name:    "name and value not sent - test #3",
// 			address: "/update/gauge",
// 			h:       h,
// 			want: want{
// 				code: 404,
// 			},
// 		},
// 		{
// 			name:    "value not sent - test #4",
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
// 			var oldCounter int64
// 			if tt.mType == "counter" {
// 				oldCounter, _ = tt.h.storage.GetCounter(tt.mName)
// 			}

// 			resp, _ := testRequest(t, ts, "POST", tt.address)
// 			defer resp.Body.Close()

// 			require.Equal(t, tt.want.code, resp.StatusCode)
// 			if tt.want.code != 200 {
// 				t.Skip()
// 			}

// 			// switch tt.mType {
// 			// case "gauge":
// 			// 	val, _ := tt.h.storage.GetGauge(tt.mName)
// 			// 	assert.Equal(t, tt.gaugeValue, val)
// 			// case "counter":tt,
// 			// 	val, _ := tt.h.storage.GetCounter(tt.mName)
// 			// 	assert.Equal(t, tt.counterValue+oldCounter, val)
// 			// }

// 			val, code, _ := s.GetMetric(tt.mType, tt.mName)
// 			assert.Equal(t, )

// 		})
// 	}
// }
