package grpcserver

import (
	"context"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
)

type GetAllMetricsUsecase interface {
	GetAllMetricsJSON(ctx context.Context) ([]byte, error)
	GetAllMetricsHTML(ctx context.Context) ([]byte, error)
	GetAllMetrics(ctx context.Context) (entity.MetricSlices, error)
}

type UpdateMetricSetUsecase interface {
	UpdateMetricSet(ctx context.Context, metrics []entity.Metric) (int64, error)
}

type UpdateMetricUsecase interface {
	UpdateMetric(ctx context.Context, metrics entity.Metric) (entity.Metric, error)
}

type GetMetricUsecase interface {
	GetMetric(ctx context.Context, metric entity.GetMetricDTO) (entity.Metric, error)
}
