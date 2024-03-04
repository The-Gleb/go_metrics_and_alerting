package v1

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/service"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/usecase"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository/memory"
	"github.com/go-chi/chi/v5"
)

func Test_getMetricJSONHandler_ServeHTTP(t *testing.T) {
	type args struct {
		rw http.ResponseWriter
		r  *http.Request
	}
	tests := []struct {
		name string
		h    *getMetricJSONHandler
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.h.ServeHTTP(tt.args.rw, tt.args.r)
		})
	}
}

func Example_getMetricJSONHandler_ServeHTTP() {

	s := memory.New()
	s.UpdateMetric("gauge", "Alloc", "3782369280")
	metricService := service.NewMetricService(s)
	getMetricUsecase := usecase.NewGetMetricUsecase(metricService)
	getMetricJSONHandler := NewGetMetricJSONHandler(getMetricUsecase)

	router := chi.NewRouter()
	getMetricJSONHandler.AddToRouter(router)
	ts := httptest.NewServer(router)
	defer ts.Close()

	validBody := `{
		"id": "Alloc",
		"type": "gauge"
	}`

	req, _ := http.NewRequest("POST", ts.URL+"/value", bytes.NewReader([]byte(validBody)))

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
	// {"id":"Alloc","type":"gauge","value":3782369280}

}
