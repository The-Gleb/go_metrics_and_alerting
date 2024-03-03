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

	mType := chi.URLParam(r, "type")
	mName := chi.URLParam(r, "name")
	metric := entity.Metric{
		MType: mType,
		ID:    mName,
	}

	switch mType {
	case "gauge":
		val, err := strconv.ParseFloat(chi.URLParam(r, "value"), 64)
		if err != nil {
			err = fmt.Errorf("getMetricHandler: %w", err)
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		metric.Value = &val

		metric, err = h.usecase.UpdateMetric(r.Context(), metric)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				http.Error(rw, err.Error(), http.StatusNotFound)
				return
			} else {
				http.Error(rw, err.Error(), http.StatusBadRequest)
				return
			}
		}

		// strVal := strconv.FormatFloat(*metric.Value, 'g', -1, 64)
		// rw.Write([]byte(strVal))

		rw.Write([]byte(chi.URLParam(r, "value")))

	case "counter":
		delta, err := strconv.ParseInt(chi.URLParam(r, "value"), 10, 64)
		if err != nil {
			err = fmt.Errorf("getMetricHandler: %w", err)
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		metric.Delta = &delta

		metric, err = h.usecase.UpdateMetric(r.Context(), metric)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				rw.WriteHeader(http.StatusNotFound)
				http.Error(rw, err.Error(), http.StatusNotFound)
				return
			} else {
				rw.WriteHeader(http.StatusBadRequest)
				http.Error(rw, err.Error(), http.StatusBadRequest)
				return
			}
		}

		strVal := strconv.FormatInt(*metric.Delta, 10)
		rw.Write([]byte(strVal))

		// rw.Write([]byte(chi.URLParam(r, "value")))

	default:
		http.Error(rw, "invalid metric type", http.StatusBadRequest)
		return
	}

}
