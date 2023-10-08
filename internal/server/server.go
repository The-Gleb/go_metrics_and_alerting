package server

import (
	"context"
	"net/http"
	"os"
	"sync"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/handlers"
	"github.com/go-chi/chi/v5"
)

type server struct {
	address  string
	handlers handlers.Handlers
	router   chi.Router
}

func New1(address string, handlers handlers.Handlers) *http.Server {
	r := chi.NewRouter()
	SetupRoutes1(r, handlers)
	return &http.Server{
		Addr:    address,
		Handler: r,
	}
}

func Shutdown(s *http.Server, c chan os.Signal, wg *sync.WaitGroup) {
	<-c
	s.Shutdown(context.Background())
	wg.Done()
}

func SetupRoutes1(r *chi.Mux, h handlers.Handlers) {
	r.Post("/update/{mType}/{mName}/{mValue}", h.UpdateMetric)
	r.Get("/", h.GetAllMetrics)
	r.Get("/value/{mType}/{mName}", h.GetMetric)
}

func Run1(s *http.Server) error {
	return s.ListenAndServe()
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
