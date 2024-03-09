package usecase

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/service"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository/memory"
	"github.com/stretchr/testify/require"
)

func Test_getAllMetricsUsecase_GetAllMetricsJSON(t *testing.T) {

	var gaugeVal float64 = 12345
	var counterVal int64 = 123
	gaugeMetric := entity.Metric{MType: "gauge", ID: "Alloc", Value: &gaugeVal}
	counterMetric := entity.Metric{MType: "counter", ID: "PollCount", Delta: &counterVal}

	m := memory.New()
	metricService := service.NewMetricService(m)
	usecase := NewGetAllMetricsUsecase(metricService)
	_, err := metricService.UpdateMetricSet(
		context.Background(),
		[]entity.Metric{gaugeMetric, counterMetric},
	)
	require.NoError(t, err)

	wantBody, err := json.Marshal(entity.MetricSlices{
		Gauge:   []entity.Metric{gaugeMetric},
		Counter: []entity.Metric{counterMetric},
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		want    []byte
		wantErr bool
		err     error
	}{
		{
			name:    "positive",
			want:    wantBody,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := usecase.GetAllMetricsJSON(context.Background())
			if (err != nil) != tt.wantErr {
				require.ErrorIs(t, err, tt.err)
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
