package v1

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/service"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/usecase"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository/memory"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getMetricHandler_ServeHTTP(t *testing.T) {
	s := memory.New()
	s.UpdateMetric("gauge", "Alloc", "123.4")
	s.UpdateMetric("counter", "Counter", "123")
	metricService := service.NewMetricService(s)
	getMetricUsecase := usecase.NewGetMetricUsecase(metricService)
	getMetricHandler := NewGetMetricHandler(getMetricUsecase)

	router := chi.NewRouter()
	getMetricHandler.AddToRouter(router)
	ts := httptest.NewServer(router)
	defer ts.Close()

	type want struct {
		value string
		code  int
	}
	tests := []struct {
		name    string
		address string
		want    want
	}{
		{
			name:    "normal gauge test #1",
			address: "/value/gauge/Alloc",
			want: want{
				value: "123.4",
				code:  200,
			},
		},
		{
			name:    "normal counter test #2",
			address: "/value/counter/Counter",
			want: want{
				value: "123",
				code:  200,
			},
		},
		{
			name:    "neg counter test #3",
			address: "/value/counter/erff",
			want: want{
				value: "",
				code:  404,
			},
		},
		{
			name:    "wrong metric type test #4",
			address: "/value/ssds/erff",
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

func Example_getMetricHandler_ServeHTTP() {

	s := memory.New()
	s.UpdateMetric("gauge", "Alloc", "123.4")
	metricService := service.NewMetricService(s)
	getMetricUsecase := usecase.NewGetMetricUsecase(metricService)
	getMetricHandler := NewGetMetricHandler(getMetricUsecase)

	router := chi.NewRouter()
	getMetricHandler.AddToRouter(router)
	ts := httptest.NewServer(router)
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL+"/value/gauge/Alloc", nil)

	resp, err := ts.Client().Do(req)
	if err != nil {
		fmt.Println("error!: %w", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	b, _ := io.ReadAll(resp.Body)
	fmt.Println(string(b))

	// Output:
	// 200
	// 123.4

}
