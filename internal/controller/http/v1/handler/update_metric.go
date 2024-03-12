package v1

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository"
	"github.com/go-chi/chi/v5"
)

const (
	updateMetricURL = "/update/{type}/{name}/{value}"
)

type UpdateMetricUsecase interface {
	UpdateMetric(ctx context.Context, metrics entity.Metric) (entity.Metric, error)
}

// updateMetricHandler receives metric type, name and value from url params.
// It updates value of one metric or creates it if doesn't exist.
type updateMetricHandler struct {
	usecase     UpdateMetricUsecase
	middlewares []func(http.Handler) http.Handler
}

func NewUpdateMetricHandler(usecase UpdateMetricUsecase) *updateMetricHandler {
	return &updateMetricHandler{
		usecase:     usecase,
		middlewares: make([]func(http.Handler) http.Handler, 0),
	}
}

func (h *updateMetricHandler) AddToRouter(r *chi.Mux) {
	r.Route(updateMetricURL, func(r chi.Router) {
		r.Use(h.middlewares...)
		r.Post("/", h.ServeHTTP)
	})
}

func (h *updateMetricHandler) Middlewares(md ...func(http.Handler) http.Handler) *updateMetricHandler {
	h.middlewares = append(h.middlewares, md...)
	return h
}

func (h *updateMetricHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	dto := entity.Metric{
		MType: chi.URLParam(r, "type"),
		ID:    chi.URLParam(r, "name"),
	}

	switch dto.MType {
	case "gauge":
		val, err := strconv.ParseFloat(chi.URLParam(r, "value"), 64)
		if err != nil {
			err = fmt.Errorf("getMetricHandler: %w", err)
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		dto.Value = &val

	case "counter":
		delta, err := strconv.ParseInt(chi.URLParam(r, "value"), 10, 64)
		if err != nil {
			err = fmt.Errorf("getMetricHandler: %w", err)
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		dto.Delta = &delta
	}

	problems := dto.Valid()
	if len(problems) > 0 {
		http.Error(rw, fmt.Sprintf("invalid %T: %d problems", dto, len(problems)), http.StatusBadRequest)
		return
	}

	metric, err := h.usecase.UpdateMetric(r.Context(), dto)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		} else {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
	}

	switch metric.MType {
	case "gauge":
		rw.Write([]byte(fmt.Sprint(metric.Value)))
		return
	case "counter":
		rw.Write([]byte(fmt.Sprint(metric.Delta)))
		return
	default:
		http.Error(rw, "metric with invalid type returned", http.StatusInternalServerError)
		return
	}

}
