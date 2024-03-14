package service

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
	filestorage "github.com/The-Gleb/go_metrics_and_alerting/internal/repository/file_storage"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository/memory"
	"github.com/stretchr/testify/require"
)

func Test_backupService_StoreDataToFile(t *testing.T) {
	var gaugeVal float64 = 12345
	var counterVal int64 = 123

	m := memory.New()
	backupMemory := filestorage.NewFileStorage("/tmp/backup.json")
	defer os.Remove("/tmp/backup.json")
	backupService := NewBackupService(m, backupMemory, 1, false)

	// upload metrics to srorage
	_, err := m.UpdateMetricSet(
		context.Background(),
		[]entity.Metric{
			{MType: "gauge", ID: "Alloc", Value: &gaugeVal},
			{MType: "counter", ID: "PollCount", Delta: &counterVal},
		},
	)
	require.NoError(t, err)

	// make backup
	err = backupService.StoreDataToFile(context.Background())
	require.NoError(t, err)

	// get metrics from storage
	metrics, err := m.GetAllMetrics(context.Background())
	require.NoError(t, err)
	jsonMetrics, err := json.Marshal(metrics)
	require.NoError(t, err)

	// get metrics from backup
	data, err := os.ReadFile("/tmp/backup.json")
	require.NoError(t, err)

	require.Equal(t, jsonMetrics, data)
}

func Test_backupService_LoadDataFromFile(t *testing.T) {

	var gaugeVal float64 = 12345
	var counterVal int64 = 123

	m := memory.New()
	backupMemory := filestorage.NewFileStorage("/tmp/backup.json")
	defer os.Remove("/tmp/backup.json")
	backupService := NewBackupService(m, backupMemory, 1, false)

	metricSlices := entity.MetricSlices{
		Gauge:   []entity.Metric{{MType: "gauge", ID: "Alloc", Value: &gaugeVal}},
		Counter: []entity.Metric{{MType: "counter", ID: "PollCount", Delta: &counterVal}},
	}

	jsonMetrics, err := json.Marshal(metricSlices)
	require.NoError(t, err)

	err = backupMemory.WriteData(jsonMetrics)
	require.NoError(t, err)

	err = backupService.LoadDataFromFile(context.Background())
	require.NoError(t, err)

	metrics, err := m.GetAllMetrics(context.Background())
	require.NoError(t, err)

	require.Equal(t, metricSlices, metrics)

}

func Test_backupService_Run(t *testing.T) {
	var gaugeVal float64 = 12345
	var gaugeVal2 float64 = 321
	var counterVal int64 = 123

	m := memory.New()
	backupMemory := filestorage.NewFileStorage("/tmp/backup.json")
	defer os.Remove("/tmp/backup.json")
	backupService := NewBackupService(m, backupMemory, 2, true)

	metricSlices := entity.MetricSlices{
		Gauge:   []entity.Metric{{MType: "gauge", ID: "Alloc", Value: &gaugeVal}},
		Counter: []entity.Metric{{MType: "counter", ID: "PollCount", Delta: &counterVal}},
	}

	jsonMetrics, err := json.Marshal(metricSlices)
	require.NoError(t, err)

	// write data to backup
	err = backupMemory.WriteData(jsonMetrics)
	require.NoError(t, err)

	// start backup service
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go backupService.Run(ctx)

	time.Sleep(1 * time.Second)

	// check restored data
	metrics, err := m.GetAllMetrics(context.Background())
	require.NoError(t, err)

	require.Equal(t, metricSlices, metrics)

	// update metrics
	_, err = m.UpdateMetricSet(
		context.Background(),
		[]entity.Metric{
			{MType: "gauge", ID: "Malloc", Value: &gaugeVal2},
			{MType: "counter", ID: "PollCount", Delta: &counterVal},
		},
	)
	require.NoError(t, err)

	// wait for backup
	time.Sleep(2 * time.Second)

	metrics, err = m.GetAllMetrics(context.Background())
	require.NoError(t, err)

	jsonMetrics, err = json.Marshal(metrics)
	require.NoError(t, err)

	data, err := backupMemory.ReadData()
	require.NoError(t, err)

	// check if backup was made
	require.Equal(t, string(jsonMetrics), string(data))

}
