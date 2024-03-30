package grpcserver

import (
	"context"
	"errors"
	"strings"

	v1 "github.com/The-Gleb/go_metrics_and_alerting/internal/controller/http/v1/handler"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
	metrics "github.com/The-Gleb/go_metrics_and_alerting/internal/proto"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverService struct {
	metrics.UnimplementedMetricServiceServer
	updateMetricUsecase    v1.UpdateMetricUsecase
	updateMetricSetUsecase v1.UpdateMetricSetUsecase
	getMetricUsecase       v1.GetMetricUsecase
	getAllMetricsUsecase   v1.GetAllMetricsUsecase
	signKey                []byte
	// logger *zap.SugaredLogger
	privateKeyPath string
	trustedSubnet  string
}

// func (s serverService) GetAllMetrics(context.Context, *metrics.GetAllMetricsRequest) (*metrics.GetAllMetricsResponse, error) {
// 	return nil, nil
// }

// func (s serverService) GetMetric(context.Context, *metrics.GetMetricRequest) (*metrics.GetMetricResponse, error) {
// 	return nil, nil
// }

// func (s serverService) UpdateMetric(context.Context, *metrics.UpdateMetricRequest) (*metrics.UpdateMetricResponse, error) {
// 	return nil, nil
// }

func (s serverService) UpdateMetricSet(ctx context.Context, updateMetricSetRequest *metrics.UpdateMetricSetRequest) (*metrics.UpdateMetricSetResponse, error) {

	grpcMetrics := updateMetricSetRequest.Metrics

	metricSlice := make([]entity.Metric, len(grpcMetrics))

	for i, m := range grpcMetrics {
		metricSlice[i] = entity.Metric{
			MType: strings.ToLower(metrics.MetricType_name[int32(m.Type)]),
			ID:    m.Name,
			Value: &m.Value,
			Delta: &m.Delta,
		}
	}

	n, err := s.updateMetricSetUsecase.UpdateMetricSet(ctx, metricSlice)
	if err != nil {
		if errors.Is(err, repository.ErrConnection) {
			return nil, status.Error(codes.Internal, err.Error())
		} else {
			return nil, status.Error(codes.InvalidArgument, "")
		}
	}

	return &metrics.UpdateMetricSetResponse{UpdatedNum: int32(n)}, nil

}
