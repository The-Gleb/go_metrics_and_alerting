package grpcserver

import (
	"context"
	"net"

	v1 "github.com/The-Gleb/go_metrics_and_alerting/internal/controller/http/v1/handler"
	metrics "github.com/The-Gleb/go_metrics_and_alerting/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type grpcServer struct {
	address string
	server  *grpc.Server
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
) (*grpcServer, error) {
	s := grpc.NewServer()

	serverService := serverService{
		updateMetricUsecase:    updateMetricUsecase,
		updateMetricSetUsecase: updateMetricSetUsecase,
		getMetricUsecase:       getMetricUsecase,
		getAllMetricsUsecase:   getAllMetricsUsecase,
		trustedSubnet:          trustedSubnet,
		signKey:                signKey,
		privateKeyPath:         privateKeyPath,
	}

	metrics.RegisterMetricServiceServer(s, serverService)

	return &grpcServer{
		server:  s,
		address: address,
	}, nil

}

func (s *grpcServer) Start() error {
	listen, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}
	return s.server.Serve(listen)
}

func (s *grpcServer) Stop(ctx context.Context) error {
	s.server.GracefulStop()
	return nil
}
