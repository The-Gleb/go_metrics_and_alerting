package handlers

import (
	"errors"
	"fmt"
	"io"
	"log/slog"

	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repositories"
)

// App interface with all buisness logic.
type App interface {
	UpdateMetricFromJSON(ctx context.Context, body io.Reader) ([]byte, error)
	UpdateMetricFromParams(ctx context.Context, mType, mName, mValue string) ([]byte, error)
	UpdateMetricSet(ctx context.Context, body io.Reader) ([]byte, error)
	GetMetricFromParams(ctx context.Context, mType, mName string) ([]byte, error)
	GetMetricFromJSON(ctx context.Context, body io.Reader) ([]byte, error)
	GetAllMetricsHTML(ctx context.Context) ([]byte, error)
	GetAllMetricsJSON(ctx context.Context) ([]byte, error)
	PingDB() error
}

type handlers struct {
	app App
}

func New(app App) *handlers {
	return &handlers{
		app: app,
	}
}

// PingDB checks DB connection.
func (handlers *handlers) PingDB(rw http.ResponseWriter, r *http.Request) {
	err := handlers.app.PingDB()
	if err != nil {
		err = fmt.Errorf("handlers.PingDB: %w", err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

// UpdateMetricSet updates metrics' values or creates it if doesn't exist.
// Receives json metric set from request body.
func (handlers *handlers) UpdateMetricSet(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rw.Header().Set("Content-Type", "application/json")

	body, err := handlers.app.UpdateMetricSet(ctx, r.Body)
	if err != nil {
		err = fmt.Errorf("handlers.UpdateMetricSet: %w", err)
		logger.Log.Error(err)
		slog.Error(err.Error())
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}
	rw.Write(body)

}

// UpdateMetric receives metric type, name and value from url params.
// It updates value of one metric or creates it if doesn't exist.
func (handlers *handlers) UpdateMetric(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rw.Header().Set("Content-Type", "application/json")

	mType := chi.URLParam(r, "mType")
	mName := chi.URLParam(r, "mName")
	mValue := chi.URLParam(r, "mValue")
	body, err := handlers.app.UpdateMetricFromParams(ctx, mType, mName, mValue)

	if err != nil {
		err = fmt.Errorf("handlers.UpdateMetric: %w", err)
		rw.WriteHeader(http.StatusBadRequest)
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}
	rw.Write(body)
}

// UpdateMetric receives metric type, name and value in json from request body.
// It updates value of one metric or creates it if doesn't exist.
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
}

// GetMetric receives metric type, name in url params.
// Returns metric value.
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
			logger.Log.Debug("Yes it is NOT FOUND ERROR")
			rw.WriteHeader(http.StatusNotFound)
			http.Error(rw, err.Error(), http.StatusNotFound)
		} else {
			rw.WriteHeader(http.StatusBadRequest)
			http.Error(rw, err.Error(), http.StatusBadRequest)
		}
	}

	rw.Write(resp)
}

// GetMetricJSON receives metric type, name in json from request body.
// Returns metric value.
func (handlers *handlers) GetMetricJSON(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rw.Header().Set("Content-Type", "application/json")

	resp, err := handlers.app.GetMetricFromJSON(ctx, r.Body)
	if err != nil {
		err = fmt.Errorf("handlers.GetMetricJSON: %w", err)
		logger.Log.Error(err)

		if !errors.Is(err, repositories.ErrNotFound) {
			rw.WriteHeader(http.StatusBadRequest)
			http.Error(rw, err.Error(), http.StatusBadRequest)
		} else {
			rw.WriteHeader(http.StatusNotFound)
			http.Error(rw, err.Error(), http.StatusNotFound)
		}
	}

	rw.Write(resp)
}

// Returns all metrics stored on repository in JSON format.
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

// Returns all metrics stored on repository in HTML format.
func (handlers *handlers) GetAllMetricsHTML(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rw.Header().Set("Content-Type", "text/html")

	body, err := handlers.app.GetAllMetricsHTML(ctx)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
	}
	// log.Println(body)

	rw.Write(body)

}
