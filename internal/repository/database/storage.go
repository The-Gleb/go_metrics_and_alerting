package database

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository"
	postgresql "github.com/The-Gleb/go_metrics_and_alerting/pkg/client"
)

type DB struct {
	client postgresql.Client
}

func NewMetricDB(ctx context.Context, dsn string) (*DB, error) {
	client, err := postgresql.NewClient(ctx, dsn)
	if err != nil {
		return nil, err
	}

	err = runMigrations(dsn)
	if err != nil {
		return nil, err
	}

	return &DB{client}, nil
}

//go:embed migration/*.sql
var migrationsDir embed.FS

func runMigrations(dsn string) error {
	d, err := iofs.New(migrationsDir, "migration")
	if err != nil {
		slog.Error(err.Error())
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		slog.Error(err.Error())
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		slog.Error(err.Error())
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}

func (db *DB) PingDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.client.Ping(ctx); err != nil {
		return err
	}
	return nil
}

func (db *DB) UpdateMetricSet(ctx context.Context, metrics []entity.Metric) (int64, error) {

	tx, err := db.client.Begin(ctx)
	if err != nil {
		return 0, checkForConectionErr("UpdateMetricSet", err)
	}

	defer tx.Rollback(ctx)
	var updated int64
	for _, metric := range metrics {
		if metric.MType == "" || metric.ID == "" ||
			(metric.Delta == nil && metric.Value == nil) {
			return 0, fmt.Errorf("%s: %w: ", "UpdateMetricSet", repository.ErrInvalidMetricStruct)
		}

		switch metric.MType {
		case "gauge":
			if metric.Value == nil {
				return 0, fmt.Errorf("%s: %w: ", "UpdateMetricSet", repository.ErrInvalidMetricStruct)
			}

			_, err := tx.Exec(ctx, `
				INSERT INTO gauge_metrics (m_name, m_value)
				VALUES ($1, $2)
				ON CONFLICT (m_name) DO UPDATE
				SET m_value = $2;
			`, metric.ID, metric.Value)
			if err != nil {
				return 0, checkForConectionErr("UpdateMetricSet", err)
			}

			updated++
		case "counter":
			if metric.Delta == nil {
				return 0, fmt.Errorf("%s: %w: ", "UpdateMetricSet", repository.ErrInvalidMetricStruct)
			}

			_, err := tx.Exec(ctx, `
				INSERT INTO counter_metrics (m_name, m_value)
				VALUES ($1, $2)
				ON CONFLICT (m_name) DO UPDATE
				SET m_value = counter_metrics.m_value + EXCLUDED.m_value;
			`, metric.ID, metric.Delta)
			if err != nil {
				return 0, checkForConectionErr("UpdateMetricSet", err)
			}
			updated++
		default:
			return 0, fmt.Errorf("invalid mertic type: %s", metric.MType)
		}
	}
	tx.Commit(ctx)
	return updated, nil
}

func (db *DB) GetAllMetrics(ctx context.Context) (entity.MetricSlices, error) {
	tx, err := db.client.Begin(ctx)
	if err != nil {
		return entity.MetricSlices{}, checkForConectionErr("GetAllMetrics", err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `SELECT m_name, m_value FROM gauge_metrics`)
	if err != nil {
		return entity.MetricSlices{}, checkForConectionErr("GetAllMetrics", err)
	}
	defer rows.Close()

	gaugeMetrics := make([]entity.Metric, 0)
	for rows.Next() {
		var metric entity.Metric
		var value float64
		err := rows.Scan(&metric.ID, &value)
		if err != nil {
			return entity.MetricSlices{}, checkForConectionErr("GetAllMetrics", err)
		}
		metric.Value = &value
		metric.MType = "gauge"
		gaugeMetrics = append(gaugeMetrics, metric)
	}
	if rows.Err() != nil {
		return entity.MetricSlices{}, checkForConectionErr("GetAllMetrics", err)
	}

	rows, err = tx.Query(ctx, `SELECT m_name, m_value FROM counter_metrics`)
	if err != nil {
		return entity.MetricSlices{}, checkForConectionErr("GetAllMetrics", err)
	}
	defer rows.Close()

	counterMetrics := make([]entity.Metric, 0)
	for rows.Next() {
		var metric entity.Metric
		var value int64
		err := rows.Scan(&metric.ID, &value)
		if err != nil {
			return entity.MetricSlices{}, checkForConectionErr("GetAllMetrics", err)
		}
		metric.Delta = &value
		metric.MType = "counter"
		counterMetrics = append(counterMetrics, metric)
	}
	if rows.Err() != nil {
		return entity.MetricSlices{}, checkForConectionErr("GetAllMetrics", err)
	}

	tx.Commit(ctx)

	return entity.MetricSlices{
		Gauge:   gaugeMetrics,
		Counter: counterMetrics,
	}, nil
}

func (db *DB) UpdateGauge(ctx context.Context, metric entity.Metric) (entity.Metric, error) {
	row := db.client.QueryRow(
		ctx,
		`INSERT INTO gauge_metrics (m_name, m_value)
		VALUES ($1, $2)
		ON CONFLICT (m_name) DO UPDATE
		SET m_value = $2
		RETURNING *;`,
		metric.ID, metric.Value,
	)

	err := row.Scan(&metric.ID, &metric.Value)
	if err != nil {
		return entity.Metric{}, checkForConectionErr("UpdateGauge", err)
	}
	return metric, nil
}

func (db *DB) UpdateCounter(ctx context.Context, mertic entity.Metric) (entity.Metric, error) {
	row := db.client.QueryRow(
		ctx,
		`INSERT INTO counter_metrics (m_name, m_value)
		VALUES ($1, $2)
		ON CONFLICT (m_name) DO UPDATE
		SET m_value = counter_metrics.m_value + EXCLUDED.m_value
		RETURNING *;`,
		mertic.ID, mertic.Delta,
	)

	err := row.Scan(&mertic.ID, &mertic.Delta)
	if err != nil {
		return entity.Metric{}, checkForConectionErr("UpdateCounter", err)
	}

	return mertic, nil
}

func (db *DB) GetGauge(ctx context.Context, dto entity.GetMetricDTO) (entity.Metric, error) {
	row := db.client.QueryRow(ctx, "SELECT m_value FROM gauge_metrics WHERE m_name = $1", dto.ID)

	var value float64
	err := row.Scan(&value)
	if err != nil {
		return entity.Metric{}, checkForConectionErr("GetGauge", err)
	}

	return entity.Metric{
		MType: dto.MType,
		ID:    dto.ID,
		Value: &value,
	}, nil
}

func (db *DB) GetCounter(ctx context.Context, dto entity.GetMetricDTO) (entity.Metric, error) {
	row := db.client.QueryRow(ctx, "SELECT m_value FROM counter_metrics WHERE m_name = $1", dto.ID)

	var value int64
	err := row.Scan(&value)
	if err != nil {
		return entity.Metric{}, checkForConectionErr("GetCounter", err)
	}

	return entity.Metric{
		MType: dto.MType,
		ID:    dto.ID,
		Delta: &value,
	}, nil
}

func checkForConectionErr(funcName string, err error) error {
	var pgErr *pgconn.PgError

	switch {
	case errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code):
		err = fmt.Errorf("%s: %w: %w", funcName, repository.ErrConnection, err)
	case errors.Is(err, pgx.ErrNoRows):
		err = fmt.Errorf("%s: %w: %w", funcName, repository.ErrNotFound, err)
	default:
		logger.Log.Debug(err)
		err = fmt.Errorf("%s: %w: ", funcName, err)
	}

	return err
}
