package v1

import (
	"fmt"
	"net/http"

	v1 "github.com/The-Gleb/go_metrics_and_alerting/internal/controller/http/v1"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
	"github.com/go-chi/chi/v5"
)

const (
	updateMetricJSONURL = "/update"
)

// updateMetricJSONHandler receives metric type, name and value in json from request body.
// It updates value of one metric or creates it if doesn't exist.
type updateMetricJSONHandler struct {
	usecase     UpdateMetricUsecase
	middlewares []func(http.Handler) http.Handler
}

func NewUpdateMetricJSONHandler(usecase UpdateMetricUsecase) *updateMetricJSONHandler {
	return &updateMetricJSONHandler{
		usecase:     usecase,
		middlewares: make([]func(http.Handler) http.Handler, 0),
	}
}

func (h *updateMetricJSONHandler) AddToRouter(r *chi.Mux) {
	r.Route(updateMetricJSONURL, func(r chi.Router) {
		r.Use(h.middlewares...)
		r.Post("/", h.ServeHTTP)
	})
}

func (h *updateMetricJSONHandler) Middlewares(md ...func(http.Handler) http.Handler) *updateMetricJSONHandler {
	h.middlewares = append(h.middlewares, md...)
	return h
}

func (h *updateMetricJSONHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	dto, _, err := v1.DecodeValid[entity.Metric](r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	metric, err := h.usecase.UpdateMetric(r.Context(), dto)
	if err != nil {
		err = fmt.Errorf("updateMetricJSONHandler: %w", err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	err = v1.Encode(rw, r, 200, metric)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}
