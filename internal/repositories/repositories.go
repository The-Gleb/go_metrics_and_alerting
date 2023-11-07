package repositories

import (
	"context"
	"errors"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/models"
)

var (
	ErrConnection error = errors.New("failed to connect to db")
	ErrNotFound   error = errors.New("metric name not found")
)

type Repositiries interface {
	GetAllMetrics(ctx context.Context) ([]models.Metrics, []models.Metrics, error)
	UpdateGauge(ctx context.Context, metricObj models.Metrics) error
	UpdateCounter(ctx context.Context, metricObj models.Metrics) error
	UpdateMetricSet(ctx context.Context, metrics []models.Metrics) (int64, error)
	GetGauge(ctx context.Context, metricObj models.Metrics) (models.Metrics, error)
	GetCounter(ctx context.Context, metricObj models.Metrics) (models.Metrics, error)
	PingDB() error
}
