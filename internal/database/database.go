package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	// "fmt"
	// "sync"
	_ "embed"
	"time"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/models"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	//go:embed sqls/schema.sql
	schemaQuery string
)

// type Repositiries interface {
// 	GetAllMetrics() (*sync.Map, *sync.Map)
// 	UpdateGauge(name string, value float64)
// 	UpdateCounter(name string, value int64)
// 	GetGauge(name string) (*float64, error)
// 	GetCounter(name string) (*int64, error)
// }

type database struct {
	db *sql.DB
}

func ConnectDB(dsn string) (*database, error) {
	// ps := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
	// 	`localhost`, `videos`, `userpassword`, `videos`)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	schemaQuery = strings.TrimSpace(schemaQuery)
	logger.Log.Info(schemaQuery)
	_, err = db.ExecContext(context.Background(), schemaQuery)
	if err != nil {
		return nil, err
	}
	return &database{db}, nil
}

type Repositiries interface {
	GetAllMetrics(ctx context.Context) ([]models.Metrics, []models.Metrics)
	UpdateGauge(ctx context.Context, metricObj models.Metrics) error
	UpdateCounter(ctx context.Context, metricObj models.Metrics) error
	GetGauge(ctx context.Context, metricObj models.Metrics) (models.Metrics, error)
	GetCounter(ctx context.Context, metricObj models.Metrics) (models.Metrics, error)
	PingDB() error
}

func (db *database) PingDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}

func (db *database) UpdateMetricSet(ctx context.Context, metrics []models.Metrics) (int64, error) {
	tx, err := db.db.Begin()
	if err != nil {
		return 0, err
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
				return 0, err
			}
			updated++
		case "counter":
			_, err := tx.ExecContext(ctx, `
				INSERT INTO counter_metrics (m_name, m_value)
				VALUES ($1, $2)
				ON CONFLICT (m_name) DO UPDATE
				SET m_value = $2;
			`, metric.ID, metric.Delta)
			if err != nil {
				return 0, err
			}
			updated++
		default:
			return 0, fmt.Errorf("invalid mertic type: %s", metric.MType)
		}
	}
	tx.Commit()
	return updated, nil
}

func (db *database) GetAllMetrics(ctx context.Context) ([]models.Metrics, []models.Metrics, error) {
	tx, err := db.db.Begin()
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `SELECT m_name, m_value FROM gauge_metrics`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	gaugeMetrics := make([]models.Metrics, 0)
	for rows.Next() {
		var metric models.Metrics
		var value float64
		err := rows.Scan(&metric.ID, &value)
		if err != nil {
			return nil, nil, err
		}
		metric.Value = &value
		metric.MType = "gauge"
		gaugeMetrics = append(gaugeMetrics, metric)
	}

	rows, err = tx.QueryContext(ctx, `SELECT m_name, m_value FROM counter_metrics`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	counterMetrics := make([]models.Metrics, 0)
	for rows.Next() {
		var metric models.Metrics
		var value int64
		err := rows.Scan(&metric.ID, &value)
		if err != nil {
			return nil, nil, err
		}
		metric.Delta = &value
		metric.MType = "counter"
		counterMetrics = append(counterMetrics, metric)
	}

	tx.Commit()

	return gaugeMetrics, counterMetrics, nil
}
func (db *database) UpdateGauge(ctx context.Context, metricObj models.Metrics) error {
	_, err := db.db.ExecContext(ctx, `
		INSERT INTO gauge_metrics (m_name, m_value)
		VALUES ($1, $2)
		ON CONFLICT (m_name) DO UPDATE
		SET m_value = $2;
		`, metricObj.ID, metricObj.Value)
	if err != nil {
		return err
	}
	return nil
}
func (db *database) UpdateCounter(ctx context.Context, metricObj models.Metrics) error {
	_, err := db.db.ExecContext(ctx, `
		INSERT INTO counter_metrics (m_name, m_value)
		VALUES ($1, $2)
		ON CONFLICT (m_name) DO UPDATE
		SET m_value = counter_metrics.m_value + EXCLUDED.m_value;
		`, metricObj.ID, metricObj.Delta)
	if err != nil {
		return err
	}
	return nil
}
func (db *database) GetGauge(ctx context.Context, metricObj models.Metrics) (models.Metrics, error) {
	row := db.db.QueryRowContext(ctx, "SELECT m_value FROM gauge_metrics WHERE m_name = $1", metricObj.ID)

	var value float64
	err := row.Scan(&value)
	if err != nil {
		return metricObj, err
	}
	metricObj.Value = &value
	return metricObj, nil
}
func (db *database) GetCounter(ctx context.Context, metricObj models.Metrics) (models.Metrics, error) {
	row := db.db.QueryRowContext(ctx, "SELECT m_value FROM counter_metrics WHERE m_name = $1", metricObj.ID)

	var value int64
	err := row.Scan(&value)
	if err != nil {
		return metricObj, err
	}
	metricObj.Delta = &value
	return metricObj, nil
}
