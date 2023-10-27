package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

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
	storage := storage.New()
	app := app.NewApp(storage)
	handlers := handlers.New(storage, app)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	s := server.New(config.Addres, handlers)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.Shutdown(s, c)
	}()
	err := server.Run(s)
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
	wg.Wait()
	log.Printf("server shutdown")
}
