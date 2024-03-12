package service

import (
	"context"
	"fmt"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository"
	"github.com/The-Gleb/go_metrics_and_alerting/pkg/utils/retry"
)

type MetricStorage interface {
	UpdateGauge(ctx context.Context, metric entity.Metric) (entity.Metric, error)
	UpdateCounter(ctx context.Context, metric entity.Metric) (entity.Metric, error)

	UpdateMetricSet(ctx context.Context, metrics []entity.Metric) (int64, error)

	GetGauge(ctx context.Context, metric entity.GetMetricDTO) (entity.Metric, error)
	GetCounter(ctx context.Context, metric entity.GetMetricDTO) (entity.Metric, error)
	GetAllMetrics(ctx context.Context) (entity.MetricSlices, error)

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
		if metric.Value == nil {
			return entity.Metric{}, fmt.Errorf("%s: %w: ", "metricService.UpdateMetric", repository.ErrInvalidMetricStruct)
		}
		err = retry.DefaultRetry(ctx, func(ctx context.Context) error {
			metric, err = service.storage.UpdateGauge(ctx, metric)
			return err
		})

	case "counter":
		if metric.Delta == nil {
			return entity.Metric{}, fmt.Errorf("%s: %w: ", "metricService.UpdateMetric", repository.ErrInvalidMetricStruct)
		}
		err = retry.DefaultRetry(ctx, func(ctx context.Context) error {
			metric, err = service.storage.UpdateCounter(ctx, metric)
			return err
		})
	default:
		return entity.Metric{}, repository.ErrInvalidMetricStruct
	}

	if err != nil {
		return entity.Metric{}, err
	}

	return metric, nil

}

func (service *metricService) UpdateMetricSet(ctx context.Context, metrics []entity.Metric) (int64, error) {
	var n int64
	var err error
	err = retry.DefaultRetry(context.TODO(), func(ctx context.Context) error {
		n, err = service.storage.UpdateMetricSet(ctx, metrics)
		return err
	})
	if err != nil {
		return n, err
	}

	return n, nil

}

func (service *metricService) GetMetric(ctx context.Context, dto entity.GetMetricDTO) (entity.Metric, error) {
	var metric entity.Metric
	var err error
	switch metric.MType {
	case "gauge":
		err = retry.DefaultRetry(ctx, func(ctx context.Context) error {
			metric, err = service.storage.GetGauge(ctx, dto)
			return err
		})

	case "counter":
		err = retry.DefaultRetry(ctx, func(ctx context.Context) error {
			metric, err = service.storage.GetCounter(ctx, dto)
			return err
		})
	default:
		return entity.Metric{}, repository.ErrInvalidMetricStruct
	}

	if err != nil {
		return entity.Metric{}, err
	}

	return metric, nil

}

func (service *metricService) GetAllMetrics(ctx context.Context) (entity.MetricSlices, error) {

	var MetricSlices entity.MetricSlices
	var err error
	err = retry.DefaultRetry(context.TODO(), func(ctx context.Context) error {
		MetricSlices, err = service.storage.GetAllMetrics(ctx)
		return err
	})
	if err != nil {
		return entity.MetricSlices{}, err
	}

	return MetricSlices, nil

}

// why does metric service ping database???
func (service *metricService) PingDB() error {

	return service.storage.PingDB()

}
