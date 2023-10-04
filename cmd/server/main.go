package main

import (
	"fmt"
	"log"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/handlers"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/server"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/storage"
)

func main() {
	parseFlags()
	storage := storage.New()
	handlers := handlers.New(storage)
	baseURL := fmt.Sprintf("http://%s", flagRunAddr)

	server := server.New(flagRunAddr, handlers)
	log.Println(flagRunAddr)
	log.Println(baseURL)

	err := server.Run()
	if err != nil {
		panic(err)
	}
	log.Printf("server started")
}
