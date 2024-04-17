package v1

import (
	"errors"
	"fmt"
	"net/http"

	v1 "github.com/The-Gleb/go_metrics_and_alerting/internal/controller/http/v1"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository"
	"github.com/go-chi/chi/v5"
)

const (
	getMetricJSONURL = "/value"
)

// getMetricJSONHandler receives metric type, name in json from request body.
// Returns metric value.
type getMetricJSONHandler struct {
	usecase     GetMetricUsecase
	middlewares []func(http.Handler) http.Handler
}

func NewGetMetricJSONHandler(usecase GetMetricUsecase) *getMetricJSONHandler {
	return &getMetricJSONHandler{
		usecase:     usecase,
		middlewares: make([]func(http.Handler) http.Handler, 0),
	}
}

func (h *getMetricJSONHandler) AddToRouter(r *chi.Mux) {
	r.Route(getMetricJSONURL, func(r chi.Router) {
		r.Use(h.middlewares...)
		r.Post("/", h.ServeHTTP)
	})
}

func (h *getMetricJSONHandler) Middlewares(md ...func(http.Handler) http.Handler) *getMetricJSONHandler {
	h.middlewares = append(h.middlewares, md...)
	return h
}

func (h *getMetricJSONHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	dto, _, err := v1.DecodeValid[entity.GetMetricDTO](r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	metric, err := h.usecase.GetMetric(r.Context(), dto)
	if err != nil {
		err = fmt.Errorf("handlers.GetMetricJSON: %w", err)
		logger.Log.Error(err)

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

	err = v1.Encode(rw, r, 200, metric)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}
