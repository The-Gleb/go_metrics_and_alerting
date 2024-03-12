package v1

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/service"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/usecase"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository/memory"
	"github.com/go-chi/chi/v5"
)

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

	resp, err := ts.Client().Do(req)
	if err != nil {
		fmt.Println("error!: %w", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	b, _ := io.ReadAll(resp.Body)
	fmt.Fprintln(os.Stdout, []any{string(b)}...)

	// Output:
	// 200
	// {"Gauge":[{"id":"Alloc","type":"gauge","value":123.4}],"Counter":[{"id":"PollCount","type":"counter","delta":12}]}

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
		"type": "gauge",
		"id": "Alloc"
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
	// {"id":"Alloc","type":"gauge","value":123.123}

}

func Example_updateMetricSetHandler_ServeHTTP() {
	s := memory.New()
	metricServie := service.NewMetricService(s)
	updateMetricSetUsecase := usecase.NewUpdateMetricSetUsecase(metricServie, nil)
	updateMetricSetHandler := NewUpdateMetricSetHandler(updateMetricSetUsecase)

	router := chi.NewRouter()
	updateMetricSetHandler.AddToRouter(router)
	ts := httptest.NewServer(router)
	defer ts.Close()

	validJSONBody := `[
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

	req, _ := http.NewRequest("POST", ts.URL+"/updates", bytes.NewReader([]byte(validJSONBody)))

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
	// 2 metrics updated

}

func Example_updateMetricHandler_ServeHTTP() {

	s := memory.New()
	metricService := service.NewMetricService(s)
	updateMetricUsecase := usecase.NewUpdateMetricUsecase(metricService, nil)
	updateMetricHandler := NewUpdateMetricHandler(updateMetricUsecase)

	router := chi.NewRouter()
	updateMetricHandler.AddToRouter(router)
	ts := httptest.NewServer(router)
	defer ts.Close()

	req, _ := http.NewRequest("POST", ts.URL+"/update/gauge/Alloc/12.12", nil)

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
	// 12.12

}
