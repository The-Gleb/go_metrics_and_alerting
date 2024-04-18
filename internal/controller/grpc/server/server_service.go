package grpcserver

import (
	"context"
	"errors"
	"strings"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
	metrics "github.com/The-Gleb/go_metrics_and_alerting/internal/proto"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverService struct {
	metrics.UnimplementedMetricServiceServer
	updateMetricUsecase    UpdateMetricUsecase
	updateMetricSetUsecase UpdateMetricSetUsecase
	getMetricUsecase       GetMetricUsecase
	getAllMetricsUsecase   GetAllMetricsUsecase
	signKey                []byte
	// logger *zap.SugaredLogger
	privateKeyPath string
	trustedSubnet  string
}

func (s serverService) GetAllMetrics(context.Context, *metrics.GetAllMetricsRequest) (*metrics.GetAllMetricsResponse, error) {

	metricSlices, err := s.getAllMetricsUsecase.GetAllMetrics(context.Background())
	if err != nil {
		return nil, err
	}

	grpcMetrics := make([]*metrics.Metric, len(metricSlices.Gauge)+1)

	for _, m := range metricSlices.Gauge {
		metric := &metrics.Metric{
			Type:       metrics.MetricType_GAUGE,
			Name:       m.ID,
			GaugeValue: *m.Value,
		}

		grpcMetrics = append(grpcMetrics, metric)

	}

	pollCountMetric := metricSlices.Counter[0]

	grpcMetrics = append(grpcMetrics, &metrics.Metric{
		Type:         metrics.MetricType_COUNTER,
		Name:         pollCountMetric.ID,
		CounterValue: *pollCountMetric.Delta,
	})

	return &metrics.GetAllMetricsResponse{
		Metrics: grpcMetrics,
	}, nil
}

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

	logger.Log.Debug(metricSlice)

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
