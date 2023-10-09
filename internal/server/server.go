package server

import (
	"context"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/handlers"
	"github.com/go-chi/chi/v5"
	"net/http"
	"os"
)

func New(address string, handlers handlers.Handlers) *http.Server {
	r := chi.NewRouter()
	SetupRoutes(r, handlers)
	return &http.Server{
		Addr:    address,
		Handler: r,
	}
}

func Shutdown(s *http.Server, c chan os.Signal) {
	<-c
	s.Shutdown(context.Background())
}

func SetupRoutes(r *chi.Mux, h handlers.Handlers) {
	r.Post("/update/{mType}/{mName}/{mValue}", h.UpdateMetric)
	r.Get("/", h.GetAllMetrics)
	r.Get("/value/{mType}/{mName}", h.GetMetric)
}

func Run(s *http.Server) error {
	return s.ListenAndServe()
}
