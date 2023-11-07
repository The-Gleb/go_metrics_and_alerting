package handlers

import (
	"errors"
	"fmt"
	"io"

	"context"
	"log"
	"net/http"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repositories"
	"github.com/go-chi/chi/v5"
)

// type Repositiries interface {
// 	UpdateMetric(mType, mName, mValue string) error
// 	GetMetric(mType, mName string) (string, error)
// 	GetAllMetrics() (*sync.Map, *sync.Map)
// 	UpdateGauge(name string, value float64)
// 	UpdateCounter(name string, value int64)
// }

type App interface {
	UpdateMetricFromJSON(ctx context.Context, body io.Reader) ([]byte, error)
	UpdateMetricFromParams(ctx context.Context, mType, mName, mValue string) ([]byte, error)
	UpdateMetricSet(ctx context.Context, body io.Reader) ([]byte, error)
	GetMetricFromParams(ctx context.Context, mType, mName string) ([]byte, error)
	GetMetricFromJSON(ctx context.Context, body io.Reader) ([]byte, error)
	GetAllMetricsHTML(ctx context.Context) ([]byte, error)
	GetAllMetricsJSON(ctx context.Context) ([]byte, error)
	PingDB() error
	// ParamsToStruct(mType, mName, mValue string) (models.Metrics, error)
}

type handlers struct {
	app App
}

func New(app App) *handlers {
	return &handlers{
		app: app,
	}
}

func (handlers *handlers) PingDB(rw http.ResponseWriter, r *http.Request) {
	err := handlers.app.PingDB()
	if err != nil {
		err = fmt.Errorf("handlers.PingDB: %w", err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (handlers *handlers) UpdateMetricSet(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rw.Header().Set("Content-Type", "application/json")

	body, err := handlers.app.UpdateMetricSet(ctx, r.Body)
	if err != nil {
		err = fmt.Errorf("handlers.UpdateMetricSet: %w", err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}
	rw.Write(body)
	rw.WriteHeader(http.StatusOK)

}

func (handlers *handlers) UpdateMetric(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rw.Header().Set("Content-Type", "application/json")

	mType := chi.URLParam(r, "mType")
	mName := chi.URLParam(r, "mName")
	mValue := chi.URLParam(r, "mValue")
	body, err := handlers.app.UpdateMetricFromParams(ctx, mType, mName, mValue)
	// log.Printf("Body is:\n%s", body)
	if err != nil {
		err = fmt.Errorf("handlers.UpdateMetric: %w", err)
		rw.WriteHeader(http.StatusBadRequest)
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}
	rw.WriteHeader(http.StatusOK)
	rw.Write(body)
}

func (handlers *handlers) UpdateMetricJSON(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rw.Header().Set("Content-Type", "application/json")
	body, err := handlers.app.UpdateMetricFromJSON(ctx, r.Body)
	if err != nil {
		err = fmt.Errorf("handlers.UpdateMetricJSON: %w", err)
		rw.WriteHeader(http.StatusBadRequest)
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}

	rw.Write(body)
	rw.WriteHeader(http.StatusOK)
}

func (handlers *handlers) GetMetric(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rw.Header().Set("Content-Type", "application/json")
	mType := chi.URLParam(r, "mType")
	mName := chi.URLParam(r, "mName")
	resp, err := handlers.app.GetMetricFromParams(ctx, mType, mName)
	// log.Printf("Body is: \n%s\n", resp)
	// log.Printf("Error is: \n%v\n", err)

	if err != nil {
		err = fmt.Errorf("handlers.GetMetric: %w", err)
		logger.Log.Error(err)

		if errors.Is(err, repositories.ErrNotFound) {
			logger.Log.Debug("Yes it IS NOT FOUND ERROR")
			rw.WriteHeader(http.StatusNotFound)
			http.Error(rw, err.Error(), http.StatusNotFound)
		} else {
			rw.WriteHeader(http.StatusBadRequest)
			http.Error(rw, err.Error(), http.StatusBadRequest)
		}
	}

	rw.Write(resp)
}

func (handlers *handlers) GetMetricJSON(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rw.Header().Set("Content-Type", "application/json")

	resp, err := handlers.app.GetMetricFromJSON(ctx, r.Body)
	log.Printf("ОШИБКААа %v", err)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		http.Error(rw, err.Error(), http.StatusNotFound)
	}
	// rw.WriteHeader(http.StatusOK)
	rw.Write(resp)
}

func (handlers *handlers) GetAllMetricsJSON(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	body, err := handlers.app.GetAllMetricsJSON(ctx)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		http.Error(rw, err.Error(), http.StatusNotFound)
	}
	rw.Write(body)
}

func (handlers *handlers) GetAllMetricsHTML(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rw.Header().Set("Content-Type", "text/html")

	body, err := handlers.app.GetAllMetricsHTML(ctx)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
	}
	// log.Println(body)

	// rw.WriteHeader(http.StatusOK)
	rw.Write(body)

}
