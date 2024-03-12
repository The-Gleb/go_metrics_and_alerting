package service

import (
	"context"
	"testing"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository/memory"
	"github.com/stretchr/testify/require"
)

func Test_metricService_UpdateMetric(t *testing.T) {
	var gaugeVal float64 = 12345
	var gaugeVal2 float64 = 321
	var counterVal int64 = 123
	var counterVal2 int64 = 246
	// gaugeMetric := entity.Metric{MType: "gauge", ID: "Alloc", Value: &gaugeVal}
	// counterMetric := entity.Metric{MType: "counter", ID: "PollCount", Delta: &counterVal}

	m := memory.New()
	metricService := NewMetricService(m)

	tests := []struct {
		name    string
		metric  entity.Metric
		want    entity.Metric
		wantErr bool
		err     error
	}{
		{
			name:    "positive add gauge",
			metric:  entity.Metric{MType: "gauge", ID: "Alloc", Value: &gaugeVal},
			want:    entity.Metric{MType: "gauge", ID: "Alloc", Value: &gaugeVal},
			wantErr: false,
		},
		{
			name:    "positive update gauge",
			metric:  entity.Metric{MType: "gauge", ID: "Alloc", Value: &gaugeVal2},
			want:    entity.Metric{MType: "gauge", ID: "Alloc", Value: &gaugeVal2},
			wantErr: false,
		},
		{
			name:    "positive add counter",
			metric:  entity.Metric{MType: "counter", ID: "PollCount", Delta: &counterVal},
			want:    entity.Metric{MType: "counter", ID: "PollCount", Delta: &counterVal},
			wantErr: false,
		},
		{
			name:    "positive update counter",
			metric:  entity.Metric{MType: "counter", ID: "PollCount", Delta: &counterVal},
			want:    entity.Metric{MType: "counter", ID: "PollCount", Delta: &counterVal2},
			wantErr: false,
		},
		{
			name:    "negative, gauge, empty metric.Value",
			metric:  entity.Metric{MType: "gauge", ID: "Alloc"},
			wantErr: true,
			err:     repository.ErrInvalidMetricStruct,
		},
		{
			name:    "negative, counter, empty metric.Delta",
			metric:  entity.Metric{MType: "counter", ID: "PollCount"},
			wantErr: true,
			err:     repository.ErrInvalidMetricStruct,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metric, err := metricService.UpdateMetric(context.Background(), tt.metric)
			if tt.wantErr {
				require.ErrorIs(t, err, tt.err)
				return
			}

			require.Equal(t, tt.want, metric)
		})
	}
}

func Test_metricService_UpdateMetricSet(t *testing.T) {
	var validFloat64 float64 = 12345
	var validFloat64two float64 = 321321
	var validInt64 int64 = 5
	var validInt64two int64 = 10

	m := memory.New()
	metricService := NewMetricService(m)

	tests := []struct {
		name    string
		metrics []entity.Metric
		result  []entity.Metric
		wantErr bool
		err     error
	}{
		{
			name: "first insert",
			metrics: []entity.Metric{
				{MType: "gauge", ID: "Alloc", Value: &validFloat64},
				{MType: "counter", ID: "PollCount", Delta: &validInt64},
			},
			result: []entity.Metric{
				{MType: "gauge", ID: "Alloc", Value: &validFloat64},
				{MType: "counter", ID: "PollCount", Delta: &validInt64},
			},
			wantErr: false,
		},
		{
			name: "update metrics",
			metrics: []entity.Metric{
				{MType: "gauge", ID: "Alloc", Value: &validFloat64two},
				{MType: "counter", ID: "PollCount", Delta: &validInt64},
			},
			result: []entity.Metric{
				{MType: "gauge", ID: "Alloc", Value: &validFloat64two},
				{MType: "counter", ID: "PollCount", Delta: &validInt64two},
			},
			wantErr: false,
		},
		{
			name: "invalid metric struct",
			metrics: []entity.Metric{
				{MType: "gauge", ID: "Alloc", Delta: &validInt64},
				{MType: "counter", ID: "PollCount"},
			},
			result: []entity.Metric{
				{MType: "gauge", ID: "Alloc", Value: &validFloat64two},
				{MType: "counter", ID: "PollCount", Delta: &validInt64two},
			},
			wantErr: true,
			err:     repository.ErrInvalidMetricStruct,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			n, err := metricService.UpdateMetricSet(context.Background(), tt.metrics)
			if tt.wantErr {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)

			metrics, err := metricService.GetAllMetrics(context.Background())
			require.NoError(t, err)

			require.Equal(t, int64(len(tt.metrics)), n)
			require.ElementsMatch(t, tt.result, append(metrics.Counter, metrics.Gauge...))

		})
	}
}

func Test_metricService_GetMetric(t *testing.T) {
	var gaugeVal float64 = 12345
	var counterVal int64 = 123

	memory := memory.New()
	memory.UpdateGauge(context.Background(), entity.Metric{MType: "gauge", ID: "Alloc", Value: &gaugeVal})
	memory.UpdateCounter(context.Background(), entity.Metric{MType: "counter", ID: "PollCount", Delta: &counterVal})
	metricService := NewMetricService(memory)

	tests := []struct {
		name    string
		metric  entity.GetMetricDTO
		want    entity.Metric
		wantErr bool
		err     error
	}{
		{
			name:   "pos gauge test #1",
			metric: entity.GetMetricDTO{MType: "gauge", ID: "Alloc"},
			want:   entity.Metric{MType: "gauge", ID: "Alloc", Value: &gaugeVal},
			err:    nil,
		},
		{
			name:   "pos counter test #2",
			metric: entity.GetMetricDTO{MType: "counter", ID: "PollCount"},
			want:   entity.Metric{MType: "counter", ID: "PollCount", Delta: &counterVal},
			err:    nil,
		}, {
			name:    "neg, metric not found",
			metric:  entity.GetMetricDTO{MType: "counter", ID: "asdfas"},
			wantErr: true,
			err:     repository.ErrNotFound,
		},
		{
			name:    "invalid type",
			metric:  entity.GetMetricDTO{MType: "asdf", ID: "PollCount"},
			wantErr: true,
			err:     repository.ErrInvalidMetricStruct,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := metricService.GetMetric(context.Background(), tt.metric)
			if tt.wantErr {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.Equal(t, tt.want, val)
		})
	}
}
