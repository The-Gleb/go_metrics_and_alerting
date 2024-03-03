package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
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

func (service *backupService) LoadDataFromFile(ctx context.Context) error {

	data, err := service.backupStorage.ReadData()
	if err != nil {
		return fmt.Errorf("LoadDataFromFile: failed reading data: %w", err)
	}

	var maps entity.MetricsMaps

	err = json.Unmarshal(data, &maps)
	if err != nil {
		return fmt.Errorf("LoadDataFromFile: failed unmarshalling: %w", err)
	}

	_, err = service.metricStorage.UpdateMetricSet(ctx, maps.Gauge)
	if err != nil {
		return fmt.Errorf("LoadDataFromFile: failded updating gauge: %w", err)
	}

	_, err = service.metricStorage.UpdateMetricSet(ctx, maps.Counter)
	if err != nil {
		return fmt.Errorf("LoadDataFromFile: failded updating counter: %w", err)
	}

	return nil
}

func (service *backupService) StoreDataToFile(ctx context.Context) error {
	metricMaps, err := service.metricStorage.GetAllMetrics(ctx)
	if err != nil {
		return fmt.Errorf("StoreDataToFile: %w", err)
	}

	data, err := json.Marshal(metricMaps)
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
