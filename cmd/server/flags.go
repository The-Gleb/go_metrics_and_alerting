package main

import (
	"flag"
	"log"
	"os"

	"fmt"
	"github.com/caarlos0/env"
	// "github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
)

type Config struct {
	Addres          string `env:"ADDRESS"`
	LogLevel        string
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

type ConfigBuilder struct {
	config Config
}

func (b ConfigBuilder) SetAddres(address string) ConfigBuilder {
	b.config.Addres = address
	return b
}
func (b ConfigBuilder) SetLogLevel(level string) ConfigBuilder {
	b.config.LogLevel = level
	return b
}
func (b ConfigBuilder) SetStoreInterval(interval int) ConfigBuilder {
	b.config.StoreInterval = interval
	return b
}
func (b ConfigBuilder) SetFileStoragePath(path string) ConfigBuilder {
	b.config.FileStoragePath = path
	return b
}
func (b ConfigBuilder) SetRestore(restore bool) ConfigBuilder {
	b.config.Restore = restore
	return b
}

func (b ConfigBuilder) SetDatabaseDSN(dsn string) ConfigBuilder {
	b.config.DatabaseDSN = dsn
	return b
}

func NewConfigFromFlags() Config {
	flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	var address string
	flag.StringVar(&address, "a", ":8080", "address and port to run server")
	var loglevel string
	flag.StringVar(&loglevel, "l", "info", "address and port to run server")
	var storeInterval int
	flag.IntVar(&storeInterval, "i", 300, "seconds between storing metrics to file")
	var fileStoragePath string
	flag.StringVar(&fileStoragePath, "f", "/tmp/metrics-db.json", "path to file to store metrics")
	var restore bool
	flag.BoolVar(&restore, "r", true, "bool, wether or not store metrics to file")
	var databaseDSN string
	flag.StringVar(&databaseDSN, "d", fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		`localhost`, `videos`, `userpassword`, `videos`),
		"info to connect to database, host=host port=port user=myuser password=xxxx dbname=mydb sslmode=disable")

	flag.Parse()

	var builder ConfigBuilder
	log.Printf("ENV ADDRESS %v", os.Getenv("ADDRESS"))

	builder = builder.SetAddres(address).
		SetLogLevel(loglevel).
		SetStoreInterval(storeInterval).
		SetFileStoragePath(fileStoragePath).
		SetRestore(restore).
		SetDatabaseDSN(databaseDSN)

	env.Parse(&builder.config)
	// if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
	// 	builder = builder.SetAddres(envAddress)
	// }
	// if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
	// 	builder = builder.SetStoreInterval(envStoreInterval)
	// }
	// if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
	// 	builder = builder.SetFileStoragePath(envFileStoragePath)
	// }
	// if envRestore := os.Getenv("RESTORE"); envRestore != "" {
	// 	builder = builder.SetRestore(envRestore)
	// }

	return builder.config
}
