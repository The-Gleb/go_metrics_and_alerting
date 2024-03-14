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
	getMetricURL = "/value/{type}/{name}"
)

type GetMetricUsecase interface {
	GetMetric(ctx context.Context, metric entity.GetMetricDTO) (entity.Metric, error)
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
	dto := entity.GetMetricDTO{
		MType: chi.URLParam(r, "type"),
		ID:    chi.URLParam(r, "name"),
	}

	problems := dto.Valid()
	if len(problems) > 0 {
		http.Error(rw, fmt.Sprintf("invalid %T: %d problems", dto, len(problems)), http.StatusBadRequest)
		return
	}

	metric, err := h.usecase.GetMetric(r.Context(), dto)
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
