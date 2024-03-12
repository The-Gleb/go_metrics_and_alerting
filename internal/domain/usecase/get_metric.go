package usecase

import (
	"context"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
)

type getMetricUsecase struct {
	metricService MetricService
}

func NewGetMetricUsecase(ms MetricService) *getMetricUsecase {
	return &getMetricUsecase{
		metricService: ms,
	}
}

func (uc *getMetricUsecase) GetMetric(ctx context.Context, metric entity.Metric) (entity.Metric, error) {

	metric, err := uc.metricService.GetMetric(ctx, metric)
	if err != nil {
		return entity.Metric{}, err
	}

	return metric, nil

}
