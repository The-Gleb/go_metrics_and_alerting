package usecase

import (
	"context"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
)

type updateMetricSetUsecase struct {
	metricService MetricService
	backupService BackupService
}

func NewUpdateMetricSetUsecase(ms MetricService, bs BackupService) *updateMetricSetUsecase {
	return &updateMetricSetUsecase{
		metricService: ms,
		backupService: bs,
	}
}

func (uc *updateMetricSetUsecase) UpdateMetricSet(ctx context.Context, metrics []entity.Metric) (int64, error) {
	n, err := uc.metricService.UpdateMetricSet(ctx, metrics)
	if err != nil {
		return n, err
	}

	if uc.backupService != nil && uc.backupService.SyncWrite() {
		err := uc.backupService.StoreDataToFile(ctx)
		if err != nil {
			return n, err
		}
	}

	return n, nil
}
