package v1

import (
	"encoding/json"
	"fmt"
	"net/http"

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

	var metric entity.Metric
	err := json.NewDecoder(r.Body).Decode(&metric)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if metric.MType == "" || metric.ID == "" ||
		(metric.Delta == nil && metric.Value == nil) {
		http.Error(rw, "invalid request body,some fields are empty, but they shouldn`t", http.StatusBadRequest)
		return
	}

	metric, err = h.usecase.UpdateMetric(r.Context(), metric)
	if err != nil {
		err = fmt.Errorf("updateMetricJSONHandler: %w", err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	b, err := json.Marshal(metric)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Write(b)

}
