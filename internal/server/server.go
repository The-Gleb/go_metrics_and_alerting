package server

import (
	"context"
	"net/http"
	"os"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
	"github.com/go-chi/chi/v5"
)

type Handlers interface {
	UpdateMetric(rw http.ResponseWriter, r *http.Request)
	GetAllMetrics(rw http.ResponseWriter, r *http.Request)
	GetMetric(rw http.ResponseWriter, r *http.Request)
}

func New(address string, handlers Handlers) *http.Server {
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

func SetupRoutes(r *chi.Mux, h Handlers) {
	r.Post("/update/{mType}/{mName}/{mValue}", logger.LogRequest(h.UpdateMetric))
	r.Get("/", logger.LogRequest(h.GetAllMetrics))
	r.Get("/value/{mType}/{mName}", logger.LogRequest(h.GetMetric))
}

func Run(s *http.Server) error {

	logger.Log.Infow("Running server",
		"address", s.Addr,
	)
	return s.ListenAndServe()
}
