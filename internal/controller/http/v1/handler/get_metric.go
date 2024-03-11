package v1

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository"
	"github.com/go-chi/chi/v5"
)

const (
	getMetricURL = "/value/{type}/{name}"
)

type GetMetricUsecase interface {
	GetMetric(ctx context.Context, metrics entity.Metric) (entity.Metric, error)
}

// getMetricHandler receives metric type, name in url params.
// Returns metric value.
type getMetricHandler struct {
	usecase     GetMetricUsecase
	middlewares []func(http.Handler) http.Handler
}

func NewGetMetricHandler(usecase GetMetricUsecase) *getMetricHandler {
	return &getMetricHandler{
		usecase:     usecase,
		middlewares: make([]func(http.Handler) http.Handler, 0),
	}
}

func (h *getMetricHandler) AddToRouter(r *chi.Mux) {
	r.Route(getMetricURL, func(r chi.Router) {
		r.Use(h.middlewares...)
		r.Get("/", h.ServeHTTP)
	})
}

func (h *getMetricHandler) Middlewares(md ...func(http.Handler) http.Handler) *getMetricHandler {
	h.middlewares = append(h.middlewares, md...)
	return h
}

func (h *getMetricHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	mType := chi.URLParam(r, "type")
	mName := chi.URLParam(r, "name")
	metric := entity.Metric{
		MType: mType,
		ID:    mName,
	}

	if mType == "" || mName == "" {
		http.Error(rw, "invalid request, metric type or metric name is empty", http.StatusBadRequest)
		return
	}

	metric, err := h.usecase.GetMetric(r.Context(), metric)
	if err != nil {

		if errors.Is(err, repository.ErrNotFound) {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		} else {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
	}

	switch mType {
	case "gauge":

		strVal := strconv.FormatFloat(*metric.Value, 'g', -1, 64)
		rw.Write([]byte(strVal))

	case "counter":

		strVal := strconv.FormatInt(*metric.Delta, 10)
		rw.Write([]byte(strVal))

	default:

		http.Error(rw, "invalid metric type", http.StatusInternalServerError)
		return

	}

}
