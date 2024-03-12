package database

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"log/slog"
	"os"
	"testing"
	"time"

	// "time"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository"
	postgresql "github.com/The-Gleb/go_metrics_and_alerting/pkg/client"
	"github.com/jackc/pgx/v4"

	// "github.com/jackc/pgx/v4"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	code, err := createTestDBContainer(m)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(code)

}

var dsn string = "postgres://test_db:test_db@:5434/tcp/test_db?sslmode=disable"

func cleanTables(t *testing.T, dsn string, tableNames ...string) {
	client, err := postgresql.NewClient(context.Background(), dsn)
	require.NoError(t, err)
	for _, name := range tableNames {
		query := fmt.Sprintf("TRUNCATE TABLE \"%s\" CASCADE", name)
		_, err := client.Exec(
			context.Background(),
			query,
		)
		require.NoError(t, err)
	}
}

func createTestDBContainer(m *testing.M) (int, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return 0, err
	}

	pg, err := pool.RunWithOptions(

		&dockertest.RunOptions{
			Repository: "postgres",
			Tag:        "alpine",
			Name:       "migrations-integration-tests",
			Env: []string{
				"POSTGRES_USER=postgres",
				"POSTGRES_PASSWORD=postgres",
			},
			ExposedPorts: []string{"5432"},
		},
		func(config *docker.HostConfig) {
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{Name: "no"}
		},
	)
	if err != nil {
		return 0, err
	}

	defer func() {
		if err := pool.Purge(pg); err != nil {
			log.Printf("failed to purge the postgres container: %v", err)
		}
	}()

	dsn = fmt.Sprintf("postgres://postgres:postgres@%s/postgres?sslmode=disable", pg.GetHostPort("5432/tcp"))
	slog.Info(dsn)

	pool.MaxWait = 2 * time.Second
	var conn *pgx.Conn
	err = pool.Retry(func() error {
		conn, err = pgx.Connect(context.Background(), dsn)
		if err != nil {
			return fmt.Errorf("failed to connect to the DB: %w", err)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	defer func() {
		if err := conn.Close(context.Background()); err != nil {
			log.Printf("failed to correctly close the connection: %v", err)
		}
	}()

	code := m.Run()

	return code, nil
}

func TestDB_UpdateMetricSet(t *testing.T) {

	var validFloat64 float64 = 12345
	var validFloat64two float64 = 321321
	var validInt64 int64 = 5
	var validInt64two int64 = 10

	slog.Info(dsn)
	storage, err := NewMetricDB(context.Background(), dsn)
	require.NoError(t, err)

	cleanTables(
		t, dsn,
		"gauge_metrics", "counter_metrics",
	)

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
	var validFloat64 float64 = 12345
	var validInt64 int64 = 5

	storage, err := NewMetricDB(context.Background(), dsn)
	require.NoError(t, err)

	cleanTables(
		t, dsn,
		"gauge_metrics", "counter_metrics",
	)

	client, err := postgresql.NewClient(context.Background(), dsn)
	require.NoError(t, err)
	_, err = client.Exec(
		context.Background(),
		`INSERT INTO gauge_metrics (m_name, m_value)
		VALUES ('Alloc', 12345);
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
	var validFloat64 float64 = 12345
	var validFloat64two float64 = 321321

	storage, err := NewMetricDB(context.Background(), dsn)
	require.NoError(t, err)

	cleanTables(
		t, dsn,
		"gauge_metrics", "counter_metrics",
	)

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

	storage, err := NewMetricDB(context.Background(), dsn)
	require.NoError(t, err)

	cleanTables(
		t, dsn,
		"gauge_metrics", "counter_metrics",
	)

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
	var validFloat64 float64 = 12345

	storage, err := NewMetricDB(context.Background(), dsn)
	require.NoError(t, err)

	cleanTables(
		t, dsn,
		"gauge_metrics", "counter_metrics",
	)

	client, err := postgresql.NewClient(context.Background(), dsn)
	require.NoError(t, err)

	_, err = client.Exec(
		context.Background(),
		`INSERT INTO gauge_metrics (m_name, m_value)
		VALUES ('Alloc', 12345);`,
	)
	require.NoError(t, err)
	tests := []struct {
		name    string
		metric  entity.GetMetricDTO
		result  entity.Metric
		wantErr bool
		err     error
	}{
		{
			name:    "positive",
			metric:  entity.GetMetricDTO{MType: "gauge", ID: "Alloc"},
			result:  entity.Metric{MType: "gauge", ID: "Alloc", Value: &validFloat64},
			wantErr: false,
		},
		{
			name:    "metric doesn`t exists",
			metric:  entity.GetMetricDTO{MType: "gauge", ID: "notfound"},
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

	storage, err := NewMetricDB(context.Background(), dsn)
	require.NoError(t, err)

	cleanTables(
		t, dsn,
		"gauge_metrics", "counter_metrics",
	)

	client, err := postgresql.NewClient(context.Background(), dsn)
	require.NoError(t, err)

	_, err = client.Exec(
		context.Background(),
		`INSERT INTO counter_metrics (m_name, m_value)
		VALUES ('PollCount', 123);`,
	)
	require.NoError(t, err)
	tests := []struct {
		name    string
		metric  entity.GetMetricDTO
		result  entity.Metric
		wantErr bool
		err     error
	}{
		{
			name:    "positive",
			metric:  entity.GetMetricDTO{MType: "counter", ID: "PollCount"},
			result:  entity.Metric{MType: "counter", ID: "PollCount", Delta: &validInt64},
			wantErr: false,
		},
		{
			name:    "metric doesn`t exists",
			metric:  entity.GetMetricDTO{MType: "counter", ID: "notfound"},
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
