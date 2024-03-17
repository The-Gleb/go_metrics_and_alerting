package main

import (
	// "flag"
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env"
)

type Config struct {
	Address         string `json:"address" env:"ADDRESS"`
	LogLevel        string
	FileStoragePath string `json:"store_file" env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `json:"database_dsn" env:"DATABASE_DSN"`
	SignKey         string `env:"KEY"`
	StoreInterval   int    `json:"store_interval" env:"STORE_INTERVAL"`
	Restore         bool   `json:"restore" env:"RESTORE"`
	PrivateKeyPath  string `json:"crypto_key" env:"CRYPTO_KEY"`
}

type ConfigBuilder struct {
	config Config
}

func (b ConfigBuilder) SetAddres(address string) ConfigBuilder {
	if address != "" {
		b.config.Address = address
		return b
	}
	if b.config.Address == "" {
		b.config.Address = ":8080"
		return b
	}
	return b
}

func (b ConfigBuilder) SetFileStoragePath(path string) ConfigBuilder {
	if path != "" {
		b.config.FileStoragePath = path
		return b
	}
	return b
}

func (b ConfigBuilder) SetDatabaseDSN(dsn string) ConfigBuilder {
	if dsn != "" {
		b.config.DatabaseDSN = dsn
		return b
	}
	return b
}

func (b ConfigBuilder) SetSignKey(key string) ConfigBuilder {
	if key != "" {
		b.config.SignKey = key
		return b
	}
	return b
}

func (b ConfigBuilder) SetStoreInterval(interval int) ConfigBuilder {
	if interval != 0 {
		b.config.StoreInterval = interval
		return b
	}
	if b.config.StoreInterval == 0 {
		b.config.StoreInterval = 1
		return b
	}
	return b
}

// TODO: it is unclear wether 'false' returned by -r flag is default or it was set by flag.
// So it is unclear if it is necessary to change 'restore' in config
func (b ConfigBuilder) SetRestore(restore bool) ConfigBuilder {
	b.config.Restore = restore
	return b

}

func (b ConfigBuilder) SetPrivateKeyPath(path string) ConfigBuilder {
	if path != "" {
		b.config.PrivateKeyPath = path
		return b
	}
	if b.config.PrivateKeyPath == "" {
		b.config.PrivateKeyPath = "./cmd/server/private.pem"
		return b
	}
	return b
}

func (b ConfigBuilder) SetLogLevel(level string) ConfigBuilder {
	if level != "" {
		b.config.LogLevel = level
		return b
	}
	if b.config.LogLevel == "" {
		b.config.LogLevel = "debug"
		return b
	}
	return b
}

func MustBuildConfig() *Config {
	var builder ConfigBuilder

	configPath := flag.String("config", "server-config.json", "path to config file")
	configFile, err := os.Open(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	err = json.NewDecoder(configFile).Decode(&builder.config)
	if err != nil {
		log.Fatal(err)
	}

	var address string
	flag.StringVar(&address, "a", "", "address and port to run server")

	var loglevel string
	flag.StringVar(&loglevel, "l", "", "address and port to run server")

	var storeInterval int
	flag.IntVar(&storeInterval, "i", 0, "seconds between storing metrics to file")

	var fileStoragePath string
	flag.StringVar(&fileStoragePath, "f", "", "path to file to store metrics")

	var restore bool
	flag.BoolVar(&restore, "r", true, "bool, wether or not restore metrics from file")

	var databaseDSN string
	flag.StringVar(&databaseDSN, "d", "",
		"info to connect to database, host=host port=port user=myuser password=xxxx dbname=mydb sslmode=disable",
	)

	var key string
	flag.StringVar(&key, "k", "", "key for signing")

	var privateKeyPath string
	flag.StringVar(&privateKeyPath, "crypto-key", "", "private key for decryption")

	flag.Parse()

	builder = builder.
		SetAddres(address).
		SetLogLevel(loglevel).
		SetStoreInterval(storeInterval).
		SetFileStoragePath(fileStoragePath).
		SetRestore(restore).
		SetDatabaseDSN(databaseDSN).
		SetSignKey(key).
		SetPrivateKeyPath(privateKeyPath)

	err = env.Parse(&builder.config)
	if err != nil {
		log.Fatal(err)
	}

	return &builder.config
}
