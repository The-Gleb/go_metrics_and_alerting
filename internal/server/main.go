package server

import (
	"github.com/The-Gleb/go_metrics_and_alerting/internal/handlers"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type server struct {
	address  string
	handlers handlers.Handlers
	router   chi.Router
}

func New(address string, handlers handlers.Handlers) *server {
	s := &server{
		address:  address,
		handlers: handlers,
		router:   chi.NewRouter(),
	}
	s.SetupRoutes()
	return s
}

func (s *server) SetupRoutes() {
	s.router.Post("/update/{mType}/{mName}/{mValue}", s.handlers.UpdateMetric)
	s.router.Get("/", s.handlers.GetAllMetrics)
	s.router.Get("/value/{mType}/{mName}", s.handlers.GetMetric)
}

func (s *server) Run() error {
	return http.ListenAndServe(s.address, s.router)
}
