package usecase

import (
	"context"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
)

type updateMetricUsecase struct {
	metricService MetricService
	backupService BackupService
}

func NewUpdateMetricUsecase(ms MetricService, bs BackupService) *updateMetricUsecase {
	return &updateMetricUsecase{
		metricService: ms,
		backupService: bs,
	}
}

func (uc *updateMetricUsecase) UpdateMetric(ctx context.Context, metric entity.Metric) (entity.Metric, error) {
	metric, err := uc.metricService.UpdateMetric(ctx, metric)
	if err != nil {
		return entity.Metric{}, err
	}

	if uc.backupService != nil && uc.backupService.SyncWrite() {
		err := uc.backupService.StoreDataToFile(ctx)
		if err != nil {
			return entity.Metric{}, err
		}
	}

	return metric, nil
}
