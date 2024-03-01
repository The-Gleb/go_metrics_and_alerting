package postgresql

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/The-Gleb/go_metrics_and_alerting/pkg/utils/retry"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Client interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	Ping(ctx context.Context) error
}

func NewClient(ctx context.Context, dsn string) (pool *pgxpool.Pool, err error) {
	// dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", sc.Username, sc.Password, sc.Host, sc.Port, sc.DbName)
	err = retry.DefaultRetry(
		ctx,
		func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			pool, err = pgxpool.New(ctx, dsn)
			if err != nil {
				return err
			}
			return nil
		},
	)

	err = pool.Ping(ctx)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	if err != nil {
		log.Fatal("error do with tries postgresql")
	}

	return pool, nil
}
