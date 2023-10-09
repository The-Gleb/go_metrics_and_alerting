package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/handlers"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/server"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/storage"
)

func main() {
	parseFlags()
	storage := storage.New()
	handlers := handlers.New(storage)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	s := server.New(flagRunAddr, handlers)
	log.Println(flagRunAddr)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		go server.Shutdown(s, c)
	}()
	err := server.Run(s)
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
	log.Printf("server shutdown")
}
