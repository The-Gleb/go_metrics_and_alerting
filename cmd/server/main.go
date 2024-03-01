package main

import (
	// "bufio"
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"time"

	_ "net/http/pprof"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/app"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/filestorage"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/handlers"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repositories"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repositories/database"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repositories/memory"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/retry"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/server"
)

// postgres://metrics:metrics@localhost/metrics?sslmode=disable

// TODO: fix status in logger
func main() {
	config := NewConfigFromFlags()

	if err := logger.Initialize(config.LogLevel); err != nil {
		logger.Log.Fatal(err)
		return
	}
	logger.Log.Info(config)

	var repository repositories.Repositiries
	var fileStorage app.FileStorage

	if config.FileStoragePath != "" {
		repository = memory.New()
		fileStorage = filestorage.NewFileStorage(config.FileStoragePath, config.StoreInterval, config.Restore)
	}

	if config.DatabaseDSN != "" {
		var db *database.DB
		var err error
		err = retry.DefaultRetry(
			context.Background(),
			func(ctx context.Context) error {
				db, err = database.ConnectDB(config.DatabaseDSN)
				return err
			},
		)

		if err != nil {
			logger.Log.Fatal(err)
			return
		}
		repository = db
	}

	app := app.NewApp(repository, fileStorage)
	handlers := handlers.New(app)
	s := server.NewWithProfiler(config.Addres, handlers, []byte(config.SignKey))

	if config.Restore {
		app.LoadDataFromFile(context.Background())
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	if config.StoreInterval > 0 && config.DatabaseDSN == "" {
		saveTicker := time.NewTicker(time.Duration(config.StoreInterval) * time.Second)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-saveTicker.C:
					app.StoreDataToFile(context.Background())
				case <-ctx.Done():
					logger.Log.Debug("stop saving to file")
					return
				}
			}

		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		ServerShutdownSignal := make(chan os.Signal, 1)
		signal.Notify(ServerShutdownSignal, syscall.SIGINT)
		<-ServerShutdownSignal
		s.Shutdown(context.Background())
		cancel()
	}()

	err := server.Run(s)
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}

	wg.Wait()
	logger.Log.Info("server shutdown")
}
