package server

import (
	"github.com/The-Gleb/go_metrics_and_alerting/internal/handlers"
	"net/http"
)

type server struct {
	address  string
	handlers handlers.Handlers
	mux      *http.ServeMux
}

func New(address string, handlers handlers.Handlers) *server {
	s := &server{
		address:  address,
		handlers: handlers,
		mux:      http.NewServeMux(),
	}
	s.SetupRoutes()
	return s
}

func (s *server) SetupRoutes() {
	// s.mux = http.NewServeMux()
	s.mux.HandleFunc("/update/", s.handlers.UpdateMetric)
}

func (s *server) Run() error {
	return http.ListenAndServe(s.address, s.mux)
}
