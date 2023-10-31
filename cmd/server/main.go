package main

import (
	// "bufio"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	// "time"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/app"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/handlers"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/server"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/storage"
)

// TODO: fix status in logger
func main() {
	config := NewConfigFromFlags()

	if err := logger.Initialize(config.LogLevel); err != nil {
		log.Fatal(err)
		return
	}
	logger.Log.Info(config)

	storage := storage.New()
	app := app.NewApp(storage, config.FileStoragePath, config.StoreInterval)
	handlers := handlers.New(app)
	s := server.New(config.Addres, handlers)

	if config.Restore {
		log.Println("try to load")
		app.LoadDataFromFile()
	}

	var wg sync.WaitGroup

	if config.StoreInterval > 0 {
		fTicker := time.NewTicker(time.Duration(config.StoreInterval) * time.Second)
		fSignal := make(chan os.Signal, 1)
		signal.Notify(fSignal, syscall.SIGINT)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-fTicker.C:
					app.StoreDataToFile()
				case <-fSignal:
					return
				}
			}

		}()
	}

	ServerShutdownSignal := make(chan os.Signal, 1)
	signal.Notify(ServerShutdownSignal, syscall.SIGINT)
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.Shutdown(s, ServerShutdownSignal)
	}()

	err := server.Run(s)
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}

	wg.Wait()
	log.Printf("server shutdown")
}
