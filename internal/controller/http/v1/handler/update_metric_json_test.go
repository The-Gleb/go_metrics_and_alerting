package v1

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/service"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/usecase"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository/memory"
	"github.com/go-chi/chi/v5"
)

func Example_updateMetricJSONHandler_ServeHTTP() {

	s := memory.New()
	metricService := service.NewMetricService(s)
	updateMetricUsecase := usecase.NewUpdateMetricUsecase(metricService, nil)
	updateMetricJSONHandler := NewUpdateMetricJSONHandler(updateMetricUsecase)

	router := chi.NewRouter()
	updateMetricJSONHandler.AddToRouter(router)
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
