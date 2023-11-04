package main

import (
	// "bufio"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"time"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/app"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/database"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/filestorage"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/handlers"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/server"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/storage"
)

// postgres://metrics:metrics@localhost/metrics?sslmode=disable

// TODO: fix status in logger
func main() {
	config := NewConfigFromFlags()

	if err := logger.Initialize(config.LogLevel); err != nil {
		log.Fatal(err)
		return
	}
	logger.Log.Info(config)

	var repository app.Repositiries
	var fileStorage app.FileStorage

	if config.DatabaseDSN != "" {
		db, err := database.ConnectDB(config.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
			return
		}
		repository = db
	} else {
		repository = storage.New()
		fileStorage = filestorage.NewFileStorage(config.FileStoragePath, config.StoreInterval, config.Restore)
	}

	app := app.NewApp(repository, fileStorage)
	handlers := handlers.New(app)
	s := server.New(config.Addres, handlers)

	if config.Restore && config.DatabaseDSN == "" {
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
		logger.Log.Debug("stop saving to file")
		cancel()
	}()

	err := server.Run(s)
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}

	wg.Wait()
	log.Printf("server shutdown")
}
