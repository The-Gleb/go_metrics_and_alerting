package usecase

import (
	"context"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
)

type MetricService interface {
	UpdateMetric(ctx context.Context, metric entity.Metric) (entity.Metric, error)
	UpdateMetricSet(ctx context.Context, metrics []entity.Metric) (int64, error)
	GetMetric(ctx context.Context, metric entity.Metric) (entity.Metric, error)
	GetAllMetrics(ctx context.Context) (entity.MetricSlices, error)

	PingDB() error
}

type BackupService interface {
	LoadDataFromFile(ctx context.Context) error
	StoreDataToFile(ctx context.Context) error
	SyncWrite() bool
}
