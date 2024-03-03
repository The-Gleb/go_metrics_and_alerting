package service

import (
	"context"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
)

type MetricStorage interface {
	UpdateGauge(ctx context.Context, metric entity.Metric) (entity.Metric, error)
	UpdateCounter(ctx context.Context, metric entity.Metric) (entity.Metric, error)

	UpdateMetricSet(ctx context.Context, metrics []entity.Metric) (int64, error)

	GetGauge(ctx context.Context, metric entity.Metric) (entity.Metric, error)
	GetCounter(ctx context.Context, metric entity.Metric) (entity.Metric, error)
	GetAllMetrics(ctx context.Context) (entity.MetricsMaps, error)

	PingDB() error
}

type metricService struct {
	storage MetricStorage
}

func NewMetricService(s MetricStorage) *metricService {
	return &metricService{s}
}

func (service *metricService) UpdateMetric(ctx context.Context, metric entity.Metric) (entity.Metric, error) {

	var err error
	switch metric.MType {
	case "gauge":
		metric, err = service.storage.UpdateGauge(ctx, metric)

	case "counter":
		metric, err = service.storage.UpdateCounter(ctx, metric)
	}

	if err != nil {
		return entity.Metric{}, err
	}

	return metric, nil

}

func (service *metricService) UpdateMetricSet(ctx context.Context, metrics []entity.Metric) (int64, error) {

	return service.storage.UpdateMetricSet(ctx, metrics)

}

func (service *metricService) GetMetric(ctx context.Context, metric entity.Metric) (entity.Metric, error) {

	var err error
	switch metric.MType {
	case "gauge":
		metric, err = service.storage.GetGauge(ctx, metric)

	case "counter":
		metric, err = service.storage.GetCounter(ctx, metric)
	}

	if err != nil {
		return entity.Metric{}, err
	}

	return metric, nil

}

func (service *metricService) GetAllMetrics(ctx context.Context) (entity.MetricsMaps, error) {

	return service.storage.GetAllMetrics(ctx)

}

// why does metric service ping database???
func (service *metricService) PingDB() error {

	return service.storage.PingDB()

}
