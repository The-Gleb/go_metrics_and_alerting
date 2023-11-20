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

	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/models"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repositories"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	//go:embed sqls/schema.sql
	schemaQuery string
)

type DB struct {
	db *sql.DB
}

func ConnectDB(dsn string) (*DB, error) {
	// ps := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
	// 	`localhost`, `videos`, `userpassword`, `videos`)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, checkForConectionErr("ConnectDB", err)
	}
	schemaQuery = strings.TrimSpace(schemaQuery)
	logger.Log.Info(schemaQuery)
	_, err = db.ExecContext(context.Background(), schemaQuery)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

type Repositiries interface {
	GetAllMetrics(ctx context.Context) ([]models.Metrics, []models.Metrics)
	UpdateGauge(ctx context.Context, metricObj models.Metrics) error
	UpdateCounter(ctx context.Context, metricObj models.Metrics) error
	GetGauge(ctx context.Context, metricObj models.Metrics) (models.Metrics, error)
	GetCounter(ctx context.Context, metricObj models.Metrics) (models.Metrics, error)
	PingDB() error
}

func (db *DB) PingDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}

func (db *DB) UpdateMetricSet(ctx context.Context, metrics []models.Metrics) (int64, error) {

	tx, err := db.db.Begin()
	if err != nil {
		return 0, checkForConectionErr("UpdateMetricSet", err)
	}

	defer tx.Rollback()
	var updated int64
	for _, metric := range metrics {

		switch metric.MType {
		case "gauge":
			_, err := tx.ExecContext(ctx, `
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
			_, err := tx.ExecContext(ctx, `
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
	tx.Commit()
	return updated, nil
}

func (db *DB) GetAllMetrics(ctx context.Context) ([]models.Metrics, []models.Metrics, error) {
	tx, err := db.db.Begin()
	if err != nil {
		return nil, nil, checkForConectionErr("GetAllMetrics", err)
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `SELECT m_name, m_value FROM gauge_metrics`)
	if err != nil {
		return nil, nil, checkForConectionErr("GetAllMetrics", err)
	}
	if rows.Err() != nil {
		return nil, nil, checkForConectionErr("GetAllMetrics", err)
	}
	defer rows.Close()

	gaugeMetrics := make([]models.Metrics, 0)
	for rows.Next() {
		var metric models.Metrics
		var value float64
		err := rows.Scan(&metric.ID, &value)
		if err != nil {
			return nil, nil, checkForConectionErr("GetAllMetrics", err)
		}
		metric.Value = &value
		metric.MType = "gauge"
		gaugeMetrics = append(gaugeMetrics, metric)
	}

	rows, err = tx.QueryContext(ctx, `SELECT m_name, m_value FROM counter_metrics`)
	if err != nil {
		return nil, nil, checkForConectionErr("GetAllMetrics", err)
	}
	if rows.Err() != nil {
		return nil, nil, checkForConectionErr("GetAllMetrics", err)
	}
	defer rows.Close()

	counterMetrics := make([]models.Metrics, 0)
	for rows.Next() {
		var metric models.Metrics
		var value int64
		err := rows.Scan(&metric.ID, &value)
		if err != nil {
			return nil, nil, checkForConectionErr("GetAllMetrics", err)
		}
		metric.Delta = &value
		metric.MType = "counter"
		counterMetrics = append(counterMetrics, metric)
	}

	tx.Commit()

	return gaugeMetrics, counterMetrics, nil
}

func (db *DB) UpdateGauge(ctx context.Context, metricObj models.Metrics) error {
	_, err := db.db.ExecContext(ctx, `
		INSERT INTO gauge_metrics (m_name, m_value)
		VALUES ($1, $2)
		ON CONFLICT (m_name) DO UPDATE
		SET m_value = $2;
		`, metricObj.ID, metricObj.Value)
	if err != nil {
		return checkForConectionErr("GetAllMetrics", err)
	}
	return nil
}
func (db *DB) UpdateCounter(ctx context.Context, metricObj models.Metrics) error {
	_, err := db.db.ExecContext(ctx, `
		INSERT INTO counter_metrics (m_name, m_value)
		VALUES ($1, $2)
		ON CONFLICT (m_name) DO UPDATE
		SET m_value = counter_metrics.m_value + EXCLUDED.m_value;
		`, metricObj.ID, metricObj.Delta)
	if err != nil {
		return checkForConectionErr("UpdateCounter", err)
	}
	return nil
}
func (db *DB) GetGauge(ctx context.Context, metricObj models.Metrics) (models.Metrics, error) {
	row := db.db.QueryRowContext(ctx, "SELECT m_value FROM gauge_metrics WHERE m_name = $1", metricObj.ID)
	if err := row.Err(); err != nil {
		return metricObj, checkForConectionErr("GetGauge", err)
	}
	var value float64
	err := row.Scan(&value)
	if err != nil {
		return metricObj, checkForConectionErr("GetGauge", err)
	}
	metricObj.Value = &value
	return metricObj, nil
}
func (db *DB) GetCounter(ctx context.Context, metricObj models.Metrics) (models.Metrics, error) {
	row := db.db.QueryRowContext(ctx, "SELECT m_value FROM counter_metrics WHERE m_name = $1", metricObj.ID)
	if err := row.Err(); err != nil {
		return metricObj, checkForConectionErr("GetCounter", err)
	}
	var value int64
	err := row.Scan(&value)
	if err != nil {
		return metricObj, checkForConectionErr("GetCounter", err)
	}
	metricObj.Delta = &value
	return metricObj, nil
}

func checkForConectionErr(funcName string, err error) error {
	var pgErr *pgconn.PgError

	switch {
	case errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code):
		err = fmt.Errorf("%s: %w: %w", funcName, repositories.ErrConnection, err)
	case errors.Is(err, sql.ErrNoRows):
		err = fmt.Errorf("%s: %w: %w", funcName, repositories.ErrNotFound, err)
	default:
		logger.Log.Debug(err)
		err = fmt.Errorf("%s: %w: ", funcName, err)
	}

	return err
}
