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

	// "time"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/app"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/database"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/handlers"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/server"
	// "github.com/The-Gleb/go_metrics_and_alerting/internal/storage"
)

// TODO: fix status in logger
func main() {
	config := NewConfigFromFlags()

	if err := logger.Initialize(config.LogLevel); err != nil {
		log.Fatal(err)
		return
	}
	logger.Log.Info(config)

	// storage := storage.New()
	storage, err := database.ConnectDB(config.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
		return
	}
	app := app.NewApp(storage, config.FileStoragePath, config.StoreInterval)
	handlers := handlers.New(app)
	s := server.New(config.Addres, handlers)

	if config.Restore {
		// app.LoadDataFromFile()
	}

	var wg sync.WaitGroup
	_, cancel := context.WithCancel(context.Background())

	// if config.StoreInterval > 0 {
	// 	saveTicker := time.NewTicker(time.Duration(config.StoreInterval) * time.Second)
	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()
	// 		for {
	// 			select {
	// 			case <-saveTicker.C:
	// 				app.StoreDataToFile()
	// 			case <-ctx.Done():
	// 				logger.Log.Debug("stop saving to file")
	// 				return
	// 			}
	// 		}

	// 	}()
	// }

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

	err = server.Run(s)
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}

	wg.Wait()
	log.Printf("server shutdown")
}
