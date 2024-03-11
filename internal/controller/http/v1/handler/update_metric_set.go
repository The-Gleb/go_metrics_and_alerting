package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
	"github.com/go-chi/chi/v5"
)

const (
	updateMetricSetURL = "/updates"
)

type UpdateMetricSetUsecase interface {
	UpdateMetricSet(ctx context.Context, metrics []entity.Metric) (int64, error)
}

// updateMetricSetHandler updates metrics' values or creates it if doesn't exist.
// Receives json metric set from request body.
type updateMetricSetHandler struct {
	usecase     UpdateMetricSetUsecase
	middlewares []func(http.Handler) http.Handler
}

func NewUpdateMetricSetHandler(usecase UpdateMetricSetUsecase) *updateMetricSetHandler {
	return &updateMetricSetHandler{
		usecase:     usecase,
		middlewares: make([]func(http.Handler) http.Handler, 0),
	}
}

func (h *updateMetricSetHandler) AddToRouter(r *chi.Mux) {
	r.Route(updateMetricSetURL, func(r chi.Router) {
		r.Use(h.middlewares...)
		r.Post("/", h.ServeHTTP)
	})
}

func (h *updateMetricSetHandler) Middlewares(md ...func(http.Handler) http.Handler) *updateMetricSetHandler {
	h.middlewares = append(h.middlewares, md...)
	return h
}

func (h *updateMetricSetHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	var metrics []entity.Metric
	err := json.NewDecoder(r.Body).Decode(&metrics)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	n, err := h.usecase.UpdateMetricSet(r.Context(), metrics)
	if err != nil {
		err = fmt.Errorf("updateMetricSetHandler: %w", err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	updatedMetricsCount := strconv.FormatInt(int64(n), 10)
	rw.Write([]byte(updatedMetricsCount + " metrics updated"))

}
