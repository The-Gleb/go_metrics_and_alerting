package httpServer

import (
	"context"
	"net/http"

	v1 "github.com/The-Gleb/go_metrics_and_alerting/internal/controller/http/v1/handler"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/controller/http/v1/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type httpServer struct {
	server *http.Server
}

func NewServer(
	address string,
	signKey []byte,
	logger *zap.SugaredLogger,
	privateKeyPath string,
	trustedSubnet string,
	updateMetricUsecase v1.UpdateMetricUsecase,
	updateMetricSetUsecase v1.UpdateMetricSetUsecase,
	getMetricUsecase v1.GetMetricUsecase,
	getAllMetricsUsecase v1.GetAllMetricsUsecase,

) (*httpServer, error) {
	updateMetricHandler := v1.NewUpdateMetricHandler(updateMetricUsecase)
	updateMetricJSONHandler := v1.NewUpdateMetricJSONHandler(updateMetricUsecase)
	getMetricHandler := v1.NewGetMetricHandler(getMetricUsecase)
	getMetricJSONHandler := v1.NewGetMetricJSONHandler(getMetricUsecase)
	updateMetricSetHandler := v1.NewUpdateMetricSetHandler(updateMetricSetUsecase)
	getAllMetricsHandler := v1.NewGetAllMetricsHandler(getAllMetricsUsecase)

	gzipMiddleware := middleware.NewGzipMiddleware()
	checkSignatureMiddleware := middleware.NewCheckSignatureMiddleware(signKey)
	loggerMidleware := middleware.NewLoggerMiddleware(logger)
	decryptionMiddleware := middleware.NewDecryptionMiddleware(privateKeyPath)
	checkSubnetMiddleware, err := middleware.NewCheckSubnetMiddleware(trustedSubnet)
	if err != nil {
		return nil, err
	}

	r := chi.NewMux()
	r.Use(
		checkSubnetMiddleware.Do,
		loggerMidleware.Do,
		decryptionMiddleware.Do,
		gzipMiddleware.Do,
		checkSignatureMiddleware.Do,
	)

	updateMetricHandler.AddToRouter(r)
	updateMetricJSONHandler.AddToRouter(r)
	getMetricHandler.AddToRouter(r)
	getMetricJSONHandler.AddToRouter(r)
	updateMetricSetHandler.AddToRouter(r)
	getAllMetricsHandler.AddToRouter(r)

	server := &http.Server{
		Addr:    address,
		Handler: r,
	}

	return &httpServer{server: server}, nil
}

func (s *httpServer) Start() error {
	return s.server.ListenAndServe()
}

func (s *httpServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
