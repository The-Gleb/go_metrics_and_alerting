package database

import (
	"context"
	_ "embed"
	"fmt"
	"testing"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository"
	postgresql "github.com/The-Gleb/go_metrics_and_alerting/pkg/client"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
)

func getTestClient(t *testing.T) postgresql.Client {
	ctx := context.Background()
	client, err := postgresql.NewClient(
		ctx,
		"postgres://metric_db:metric_db@localhost:5434/metric_db?sslmode=disable",
	)
	require.NoError(t, err)
	return client

}

func cleanTables(t *testing.T, client postgresql.Client, tableNames ...string) {
	for _, name := range tableNames {
		query := fmt.Sprintf("TRUNCATE TABLE \"%s\" CASCADE", name)
		_, err := client.Exec(
			context.Background(),
			query,
		)
		require.NoError(t, err)
	}

}

func TestDB_UpdateMetricSet(t *testing.T) {
	var validFloat64 float64 = 123.123
	var validFloat64two float64 = 321321
	var validInt64 int64 = 5
	var validInt64two int64 = 10

	client := getTestClient(t)
	cleanTables(
		t, client,
		"gauge_metrics", "counter_metrics",
	)
	storage, err := NewMetricDB(client)
	require.NoError(t, err)

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

			n, err := storage.UpdateMetricSet(context.Background(), tt.metrics)
			if tt.wantErr {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)

			metrics, err := storage.GetAllMetrics(context.Background())
			require.NoError(t, err)

			require.Equal(t, int64(len(tt.metrics)), n)
			require.ElementsMatch(t, tt.result, append(metrics.Counter, metrics.Gauge...))

		})
	}
}

func TestDB_GetAllMetrics(t *testing.T) {
	var validFloat64 float64 = 123.123
	var validInt64 int64 = 5

	client := getTestClient(t)
	cleanTables(
		t, client,
		"gauge_metrics", "counter_metrics",
	)
	storage, err := NewMetricDB(client)
	require.NoError(t, err)

	_, err = client.Exec(
		context.Background(),
		`INSERT INTO gauge_metrics (m_name, m_value)
		VALUES ('Alloc', 123.123);
		INSERT INTO counter_metrics (m_name, m_value)
		VALUES ('PollCount', 5);`,
	)
	require.NoError(t, err)
	tests := []struct {
		name    string
		result  entity.MetricSlices
		wantErr bool
		err     error
	}{
		{
			name: "positive",
			result: entity.MetricSlices{
				Gauge: []entity.Metric{
					{MType: "gauge", ID: "Alloc", Value: &validFloat64},
				},
				Counter: []entity.Metric{
					{MType: "counter", ID: "PollCount", Delta: &validInt64},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics, err := storage.GetAllMetrics(context.Background())
			require.NoError(t, err)

			require.ElementsMatch(t, tt.result.Gauge, metrics.Gauge)
			require.ElementsMatch(t, tt.result.Counter, metrics.Counter)
		})
	}
}

func TestDB_UpdateGauge(t *testing.T) {
	var validFloat64 float64 = 123.123
	var validFloat64two float64 = 321321

	client := getTestClient(t)
	cleanTables(
		t, client,
		"gauge_metrics", "counter_metrics",
	)
	storage, err := NewMetricDB(client)
	require.NoError(t, err)

	tests := []struct {
		name    string
		metric  entity.Metric
		result  entity.Metric
		wantErr bool
		err     error
	}{
		{
			name:    "first insert",
			metric:  entity.Metric{MType: "gauge", ID: "Alloc", Value: &validFloat64},
			result:  entity.Metric{MType: "gauge", ID: "Alloc", Value: &validFloat64},
			wantErr: false,
		},
		{
			name:    "update metrics",
			metric:  entity.Metric{MType: "gauge", ID: "Alloc", Value: &validFloat64two},
			result:  entity.Metric{MType: "gauge", ID: "Alloc", Value: &validFloat64two},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			m, err := storage.UpdateGauge(context.Background(), tt.metric)
			if tt.wantErr {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.result, m)

		})
	}
}

func TestDB_UpdateCounter(t *testing.T) {
	var validInt64 int64 = 5
	var validInt64two int64 = 10

	client := getTestClient(t)
	cleanTables(
		t, client,
		"gauge_metrics", "counter_metrics",
	)
	storage, err := NewMetricDB(client)
	require.NoError(t, err)

	tests := []struct {
		name    string
		metric  entity.Metric
		result  entity.Metric
		wantErr bool
		err     error
	}{
		{
			name:    "first insert",
			metric:  entity.Metric{MType: "counter", ID: "PollCount", Delta: &validInt64},
			result:  entity.Metric{MType: "counter", ID: "PollCount", Delta: &validInt64},
			wantErr: false,
		},
		{
			name:    "update metrics",
			metric:  entity.Metric{MType: "counter", ID: "PollCount", Delta: &validInt64},
			result:  entity.Metric{MType: "counter", ID: "PollCount", Delta: &validInt64two},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			m, err := storage.UpdateCounter(context.Background(), tt.metric)
			if tt.wantErr {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)

			// metrics, err := storage.GetAllMetrics(context.Background())
			// require.NoError(t, err)

			// require.Contains(t, metrics.Gauge, tt.result)

			require.Equal(t, tt.result, m)

		})
	}
}

func TestDB_GetGauge(t *testing.T) {
	var validFloat64 float64 = 123.123

	client := getTestClient(t)
	cleanTables(
		t, client,
		"gauge_metrics", "counter_metrics",
	)
	storage, err := NewMetricDB(client)
	require.NoError(t, err)

	_, err = client.Exec(
		context.Background(),
		`INSERT INTO gauge_metrics (m_name, m_value)
		VALUES ('Alloc', 123.123);`,
	)
	require.NoError(t, err)
	tests := []struct {
		name    string
		metric  entity.Metric
		result  entity.Metric
		wantErr bool
		err     error
	}{
		{
			name:    "positive",
			metric:  entity.Metric{MType: "gauge", ID: "Alloc"},
			result:  entity.Metric{MType: "gauge", ID: "Alloc", Value: &validFloat64},
			wantErr: false,
		},
		{
			name:    "metric doesn`t exists",
			metric:  entity.Metric{MType: "gauge", ID: "notfound"},
			wantErr: true,
			err:     repository.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metric, err := storage.GetGauge(context.Background(), tt.metric)
			if tt.wantErr {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.result, metric)
		})
	}
}

func TestDB_GetCounter(t *testing.T) {
	var validInt64 int64 = 123

	client := getTestClient(t)
	cleanTables(
		t, client,
		"gauge_metrics", "counter_metrics",
	)
	storage, err := NewMetricDB(client)
	require.NoError(t, err)

	_, err = client.Exec(
		context.Background(),
		`INSERT INTO counter_metrics (m_name, m_value)
		VALUES ('PollCount', 123);`,
	)
	require.NoError(t, err)
	tests := []struct {
		name    string
		metric  entity.Metric
		result  entity.Metric
		wantErr bool
		err     error
	}{
		{
			name:    "positive",
			metric:  entity.Metric{MType: "counter", ID: "PollCount"},
			result:  entity.Metric{MType: "counter", ID: "PollCount", Delta: &validInt64},
			wantErr: false,
		},
		{
			name:    "metric doesn`t exists",
			metric:  entity.Metric{MType: "counter", ID: "notfound"},
			wantErr: true,
			err:     repository.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metric, err := storage.GetCounter(context.Background(), tt.metric)
			if tt.wantErr {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.result, metric)
		})
	}
}
