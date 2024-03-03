package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"os"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/service"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/usecase"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository/memory"
	"github.com/go-chi/chi/v5"
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

func Test_getAllMetricHandler_ServeHTTP(t *testing.T) {
	var val1 float64 = 3782369280
	var val2 int64 = 123
	metrics := []entity.Metric{
		{
			MType: "gauge",
			ID:    "HeapAlloc",
			Value: &val1,
		},
		{
			MType: "counter",
			ID:    "PollCount",
			Delta: &val2,
		},
	}

	metricMaps := entity.MetricsMaps{
		Gauge:   metrics[:1],
		Counter: metrics[1:],
	}

	jsonMetrics, err := json.Marshal(metricMaps)
	require.NoError(t, err)

	s := memory.New()
	metricService := service.NewMetricService(s)
	metricService.UpdateMetricSet(context.Background(), metrics)
	getAllMetricsUsecase := usecase.NewGetAllMetricsUsecase(metricService)
	getAllMetricsHandler := NewGetAllMetricsHandler(getAllMetricsUsecase)

	router := chi.NewRouter()
	getAllMetricsHandler.AddToRouter(router)
	ts := httptest.NewServer(router)
	defer ts.Close()

	// validJsonBody := `[{"id": "HeapAlloc","type": "gauge","value": 3782369280},{"id": "PollCount","type": "counter","delta": 123}]`

	type want struct {
		code int
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "normal",
			want: want{200},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			resp, b := testRequest(t, ts, "GET", "/", nil)
			defer resp.Body.Close()

			t.Log(b)

			require.Equal(t, tt.want.code, resp.StatusCode)
			if tt.want.code != 200 {
				return
			}

			require.Equal(t, string(jsonMetrics), b)
		})
	}
}

func Example_getAllMetricHandler_ServeHTTP() {

	s := memory.New()
	s.UpdateMetric("gauge", "Alloc", "123.4")
	s.UpdateMetric("counter", "PollCount", "12")
	metricService := service.NewMetricService(s)
	getAllMetricsUsecase := usecase.NewGetAllMetricsUsecase(metricService)
	getAllMetricsHandler := NewGetAllMetricsHandler(getAllMetricsUsecase)

	router := chi.NewRouter()
	getAllMetricsHandler.AddToRouter(router)
	ts := httptest.NewServer(router)
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL+"/", nil)

	resp, _ := ts.Client().Do(req)

	fmt.Println(resp.StatusCode)
	b, _ := io.ReadAll(resp.Body)
	fmt.Fprintln(os.Stdout, []any{string(b)}...)

	// Output:
	// 200
	// {"Gauge":[{"id":"Alloc","type":"gauge","value":123.4}],"Counter":[{"id":"PollCount","type":"counter","delta":12}]}

}
