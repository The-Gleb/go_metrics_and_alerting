package database

import (
	"context"
	"database/sql"
	"strings"

	// "fmt"
	// "sync"
	_ "embed"
	"time"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
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

func (db *database) PingDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}

func (db *database) GetAllMetrics() (map[string]float64, map[string]int64) {
	return nil, nil
}
func (db *database) UpdateGauge(name string, value float64) {
	_, err := db.db.ExecContext(context.TODO(), `
		INSERT INTO gauge_metrics (m_name, m_value)
		VALUES ($1, $2)
		ON CONFLICT (m_name) DO UPDATE
		SET m_value = $2;
		`, name, value)
	if err != nil {
		logger.Log.Error(err)
	}
}
func (db *database) UpdateCounter(name string, value int64) {
	_, err := db.db.ExecContext(context.TODO(), `
		INSERT INTO counter_metrics (m_name, m_value)
		VALUES ($1, $2)
		ON CONFLICT (m_name) DO UPDATE
		SET m_value = counter_metrics.m_value + EXCLUDED.m_value;
		`, name, value)
	if err != nil {
		logger.Log.Error(err)
	}
}
func (db *database) GetGauge(name string) (*float64, error) {
	row := db.db.QueryRowContext(context.TODO(), "SELECT m_value FROM gauge_metrics WHERE m_name = $1", name)

	var value float64
	err := row.Scan(&value)
	if err != nil {
		return nil, err
	}
	return &value, nil
}
func (db *database) GetCounter(name string) (*int64, error) {
	row := db.db.QueryRowContext(context.TODO(), "SELECT m_value FROM counter_metrics WHERE m_name = $1", name)

	var value int64
	err := row.Scan(&value)
	if err != nil {
		return nil, err
	}
	return &value, nil
}
