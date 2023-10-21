package main

import (
	"flag"
	"log"
	"os"
	// "github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
)

type Config struct {
	Addres   string
	LogLevel string
}

type ConfigBuilder struct {
	config Config
}

func (b ConfigBuilder) SetAddres(address string) ConfigBuilder {
	b.config.Addres = address
	return b
}

func NewConfigFromFlags() Config {
	flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	var address string
	flag.StringVar(&address, "a", ":8080", "address and port to run server")
	var loglevel string
	flag.StringVar(&loglevel, "l", "info", "address and port to run server")

	flag.Parse()

	var builder ConfigBuilder
	log.Printf("ENV ADDRESS %v", os.Getenv("ADDRESS"))
	builder = builder.SetAddres(address)
	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		builder = builder.SetAddres(envAddress)
	}

	return builder.config
}
