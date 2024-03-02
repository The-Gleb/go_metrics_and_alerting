package server

import (
	"context"
	"net/http"
	"os"

	_ "net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/authentication"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/compressor"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
)

type Handlers interface {
	UpdateMetric(rw http.ResponseWriter, r *http.Request)
	UpdateMetricJSON(rw http.ResponseWriter, r *http.Request)
	UpdateMetricSet(rw http.ResponseWriter, r *http.Request)
	GetMetric(rw http.ResponseWriter, r *http.Request)
	GetMetricJSON(rw http.ResponseWriter, r *http.Request)
	GetAllMetricsHTML(rw http.ResponseWriter, r *http.Request)
	GetAllMetricsJSON(rw http.ResponseWriter, r *http.Request)
	PingDB(rw http.ResponseWriter, r *http.Request)
}

func New(address string, handlers Handlers, signKey []byte) *http.Server {
	r := chi.NewRouter()
	SetupRoutes(r, handlers)
	return &http.Server{
		Addr:    address,
		Handler: logger.LogRequest(compressor.GzipMiddleware(authentication.CheckSignature(signKey, r))),
	}
}

func NewWithProfiler(address string, handlers Handlers, signKey []byte) *http.Server {
	r := chi.NewRouter()

	r.Mount("/debug", middleware.Profiler())
	SetupRoutes(r, handlers)

	return &http.Server{
		Addr:    address,
		Handler: logger.LogRequest(compressor.GzipMiddleware(authentication.CheckSignature(signKey, r))),
	}
}

func Shutdown(s *http.Server, c chan os.Signal) {
	<-c
	s.Shutdown(context.Background())
}

func SetupRoutes(r *chi.Mux, h Handlers) {
	r.Post("/update/{mType}/{mName}/{mValue}", h.UpdateMetric)
	r.Post("/update/", h.UpdateMetricJSON)
	r.Post("/updates/", h.UpdateMetricSet)
	r.Get("/value/{mType}/{mName}", h.GetMetric)
	r.Post("/value/", h.GetMetricJSON)
	r.Get("/", h.GetAllMetricsHTML)
	r.Get("/ping", h.PingDB)
}

func Run(s *http.Server) error {

	logger.Log.Infow("Running server",
		"address", s.Addr,
	)
	return s.ListenAndServe()
}
