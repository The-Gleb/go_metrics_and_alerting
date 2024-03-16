package v1

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

const (
	getAllMetricsURL = "/"
)

type GetAllMetricsUsecase interface {
	GetAllMetricsJSON(ctx context.Context) ([]byte, error)
	GetAllMetricsHTML(ctx context.Context) ([]byte, error)
}

// Returns all metrics stored on repository in JSON or HTML format.
type getAllMetricsHandler struct {
	usecase     GetAllMetricsUsecase
	middlewares []func(http.Handler) http.Handler
}

func NewGetAllMetricsHandler(usecase GetAllMetricsUsecase) *getAllMetricsHandler {
	return &getAllMetricsHandler{
		usecase:     usecase,
		middlewares: make([]func(http.Handler) http.Handler, 0),
	}
}

func (h *getAllMetricsHandler) AddToRouter(r *chi.Mux) {
	r.Route(getAllMetricsURL, func(r chi.Router) {
		r.Use(h.middlewares...)
		r.Get("/", h.ServeHTTP)
	})
}

func (h *getAllMetricsHandler) Middlewares(md ...func(http.Handler) http.Handler) *getAllMetricsHandler {
	h.middlewares = append(h.middlewares, md...)
	return h
}

func (h *getAllMetricsHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	body, err := h.usecase.GetAllMetricsJSON(r.Context())
	if err != nil {
		err = fmt.Errorf("getAllMetricsHandler: %w", err) // TODO: handler errors
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.Write(body)
}
