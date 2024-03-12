package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
	"github.com/The-Gleb/go_metrics_and_alerting/pkg/utils/retry"
)

type FileStorage interface {
	WriteData(data []byte) error
	ReadData() ([]byte, error)
}

type backupService struct {
	metricStorage  MetricStorage
	backupStorage  FileStorage
	backupInterval int
	restore        bool
}

func NewBackupService(ms MetricStorage, bs FileStorage, interval int, restore bool) *backupService {
	return &backupService{
		metricStorage:  ms,
		backupStorage:  bs,
		backupInterval: interval,
		restore:        restore,
	}
}

func (service *backupService) Run(ctx context.Context) {

	if service.restore {
		err := service.LoadDataFromFile(ctx)
		if err != nil {
			logger.Log.Errorf("error in backupservice, stopping", "error", err)
			return
		}
	}

	if service.backupInterval <= 0 || service.backupStorage == nil {
		return
	}

	saveTicker := time.NewTicker(time.Duration(service.backupInterval) * time.Second)
	for {
		for {
			select {
			case <-saveTicker.C:
				err := service.StoreDataToFile(ctx)
				if err != nil {
					logger.Log.Errorf("error in backupservice, stopping", "error", err)
					return
				}
			case <-ctx.Done():
				logger.Log.Debug("stop saving to file")
				return
			}
		}
	}

}

func (service *backupService) LoadDataFromFile(ctx context.Context) error {

	data, err := service.backupStorage.ReadData()
	if err != nil {
		return fmt.Errorf("LoadDataFromFile: failed reading data: %w", err)
	}

	var maps entity.MetricSlices

	err = json.Unmarshal(data, &maps)
	if err != nil {
		return fmt.Errorf("LoadDataFromFile: failed unmarshalling: %w", err)
	}

	err = retry.DefaultRetry(ctx, func(ctx context.Context) error {
		_, err = service.metricStorage.UpdateMetricSet(ctx, maps.Gauge)
		return err
	})
	if err != nil {
		return fmt.Errorf("LoadDataFromFile:: %w", err)
	}

	_, err = service.metricStorage.UpdateMetricSet(ctx, maps.Gauge)
	if err != nil {
		return fmt.Errorf("LoadDataFromFile:: %w", err)
	}

	_, err = service.metricStorage.UpdateMetricSet(ctx, maps.Counter)
	if err != nil {
		return fmt.Errorf("LoadDataFromFile: failded updating counter: %w", err)
	}

	return nil
}

func (service *backupService) StoreDataToFile(ctx context.Context) error {

	var MetricSlices entity.MetricSlices
	var err error
	err = retry.DefaultRetry(context.TODO(), func(ctx context.Context) error {
		MetricSlices, err = service.metricStorage.GetAllMetrics(ctx)
		return err
	})
	if err != nil {
		return fmt.Errorf("StoreDataToFile: %w", err)
	}

	data, err := json.Marshal(MetricSlices)
	if err != nil {
		return fmt.Errorf("StoreDataToFile: %w", err)
	}

	err = service.backupStorage.WriteData(data)
	if err != nil {
		return fmt.Errorf("StoreDataToFile: %w", err)
	}

	return nil
}

func (service *backupService) SyncWrite() bool {
	return service.backupInterval == 0
}
