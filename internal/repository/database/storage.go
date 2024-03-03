package database

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	// "github.com/jackc/pgerrcode"
	// "github.com/jackc/pgx/v5/pgconn"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository"
	postgresql "github.com/The-Gleb/go_metrics_and_alerting/pkg/client"
)

var (
	//go:embed sqls/schema.sql
	schemaQuery string
)

type DB struct {
	client postgresql.Client
}

func ConnectDB(dsn string) (*DB, error) {
	// ps := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
	// 	`localhost`, `videos`, `userpassword`, `videos`)

	client, err := postgresql.NewClient(context.Background(), dsn)
	if err != nil {
		return nil, checkForConectionErr("ConnectDB", err)
	}

	// migrate.
	schemaQuery = strings.TrimSpace(schemaQuery)

	_, err = client.Exec(context.Background(), schemaQuery)
	if err != nil {
		return nil, err
	}

	return &DB{client}, nil
}

type Repositiries interface {
	GetAllMetrics(ctx context.Context) ([]entity.Metric, []entity.Metric)
	UpdateGauge(ctx context.Context, metric entity.Metric) error
	UpdateCounter(ctx context.Context, metric entity.Metric) error
	GetGauge(ctx context.Context, metric entity.Metric) (entity.Metric, error)
	GetCounter(ctx context.Context, metric entity.Metric) (entity.Metric, error)
	PingDB() error
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

		switch metric.MType {
		case "gauge":
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

func (db *DB) GetAllMetrics(ctx context.Context) (entity.MetricsMaps, error) {
	tx, err := db.client.Begin(ctx)
	if err != nil {
		return entity.MetricsMaps{}, checkForConectionErr("GetAllMetrics", err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `SELECT m_name, m_value FROM gauge_metrics`)
	if err != nil {
		return entity.MetricsMaps{}, checkForConectionErr("GetAllMetrics", err)
	}
	if rows.Err() != nil {
		return entity.MetricsMaps{}, checkForConectionErr("GetAllMetrics", err)
	}
	defer rows.Close()

	gaugeMetrics := make([]entity.Metric, 0)
	for rows.Next() {
		var metric entity.Metric
		var value float64
		err := rows.Scan(&metric.ID, &value)
		if err != nil {
			return entity.MetricsMaps{}, checkForConectionErr("GetAllMetrics", err)
		}
		metric.Value = &value
		metric.MType = "gauge"
		gaugeMetrics = append(gaugeMetrics, metric)
	}

	rows, err = tx.Query(ctx, `SELECT m_name, m_value FROM counter_metrics`)
	if err != nil {
		return entity.MetricsMaps{}, checkForConectionErr("GetAllMetrics", err)
	}
	if rows.Err() != nil {
		return entity.MetricsMaps{}, checkForConectionErr("GetAllMetrics", err)
	}
	defer rows.Close()

	counterMetrics := make([]entity.Metric, 0)
	for rows.Next() {
		var metric entity.Metric
		var value int64
		err := rows.Scan(&metric.ID, &value)
		if err != nil {
			return entity.MetricsMaps{}, checkForConectionErr("GetAllMetrics", err)
		}
		metric.Delta = &value
		metric.MType = "counter"
		counterMetrics = append(counterMetrics, metric)
	}

	tx.Commit(ctx)

	return entity.MetricsMaps{
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

	return entity.Metric{}, nil
}
func (db *DB) GetGauge(ctx context.Context, metric entity.Metric) (entity.Metric, error) {
	row := db.client.QueryRow(ctx, "SELECT m_value FROM gauge_metrics WHERE m_name = $1", metric.ID)

	var value float64
	err := row.Scan(&value)
	if err != nil {
		return metric, checkForConectionErr("GetGauge", err)
	}
	metric.Value = &value
	return metric, nil
}
func (db *DB) GetCounter(ctx context.Context, metric entity.Metric) (entity.Metric, error) {
	row := db.client.QueryRow(ctx, "SELECT m_value FROM counter_metrics WHERE m_name = $1", metric.ID)

	var value int64
	err := row.Scan(&value)
	if err != nil {
		return metric, checkForConectionErr("GetCounter", err)
	}
	metric.Delta = &value
	return metric, nil
}

func checkForConectionErr(funcName string, err error) error {
	var pgErr *pgconn.PgError

	switch {
	case errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code):
		err = fmt.Errorf("%s: %w: %w", funcName, repository.ErrConnection, err)
	case errors.Is(err, sql.ErrNoRows):
		err = fmt.Errorf("%s: %w: %w", funcName, repository.ErrNotFound, err)
	default:
		logger.Log.Debug(err)
		err = fmt.Errorf("%s: %w: ", funcName, err)
	}

	return err
}
