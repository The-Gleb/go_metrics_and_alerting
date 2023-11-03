package database

import (
	"context"
	"database/sql"
	// "fmt"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
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

func (db *database) GetAllMetrics() (*sync.Map, *sync.Map) {
	return nil, nil
}
func (db *database) UpdateGauge(name string, value float64) {

}
func (db *database) UpdateCounter(name string, value int64) {

}
func (db *database) GetGauge(name string) (*float64, error) {
	return nil, nil
}
func (db *database) GetCounter(name string) (*int64, error) {
	return nil, nil
}
