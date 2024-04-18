package main

import (
	"encoding/json"
	"flag"
	"os"
	"strings"

	"github.com/caarlos0/env"
)

type Config struct {
	Address         string `json:"address" env:"ADDRESS"`
	LogLevel        string `json:"log_level" env:"LOG_LEVEL"`
	FileStoragePath string `json:"store_file" env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `json:"database_dsn" env:"DATABASE_DSN"`
	SignKey         string `env:"KEY"`
	StoreInterval   int    `json:"store_interval" env:"STORE_INTERVAL"`
	Restore         bool   `json:"restore" env:"RESTORE"`
	PrivateKeyPath  string `json:"crypto_key" env:"CRYPTO_KEY"`
	TrustedSubnet   string `json:"trusted_subnet" env:"TRUSTED_SUBNET"`
}

type ConfigBuilder struct {
	config Config
}

func (b ConfigBuilder) setAddress(address string) ConfigBuilder {
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

func (b ConfigBuilder) setFileStoragePath(path string) ConfigBuilder {
	if path != "" {
		b.config.FileStoragePath = path
		return b
	}
	return b
}

func (b ConfigBuilder) setDatabaseDSN(dsn string) ConfigBuilder {
	if dsn != "" {
		b.config.DatabaseDSN = dsn
		return b
	}
	return b
}

func (b ConfigBuilder) setSignKey(key string) ConfigBuilder {
	if key != "" {
		b.config.SignKey = key
		return b
	}
	return b
}

func (b ConfigBuilder) setStoreInterval(interval int) ConfigBuilder {
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
// So it is unclear if it is necessary to change 'restore' in config.
func (b ConfigBuilder) setRestore(restore string) ConfigBuilder {
	if restore != "" {
		restore = strings.ToLower(restore)
		if restore == "true" || restore == "t" || restore == "1" {
			b.config.Restore = true
			return b
		} else if restore == "false" || restore == "f" || restore == "0" {
			b.config.Restore = false
			return b
		}

	}

	return b
}

func (b ConfigBuilder) setPrivateKeyPath(path string) ConfigBuilder {
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

func (b ConfigBuilder) setLogLevel(level string) ConfigBuilder {
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

func (b ConfigBuilder) setTrustedSubnet(net string) ConfigBuilder {
	if net != "" {
		b.config.TrustedSubnet = net
	}
	return b

}

func BuildConfig() (*Config, error) {
	var builder ConfigBuilder

	configPath := flag.String("config", "server-config.json", "path to config file")
	configFile, err := os.Open(*configPath)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(configFile).Decode(&builder.config)
	if err != nil {
		return nil, err
	}

	var address string
	flag.StringVar(&address, "a", "", "address and port to run server")

	var loglevel string
	flag.StringVar(&loglevel, "l", "", "address and port to run server")

	var storeInterval int
	flag.IntVar(&storeInterval, "i", 0, "seconds between storing metrics to file")

	var fileStoragePath string
	flag.StringVar(&fileStoragePath, "f", "", "path to file to store metrics")

	var restore string
	flag.StringVar(&restore, "r", "", "bool, wether or not restore metrics from file")

	var databaseDSN string
	flag.StringVar(&databaseDSN, "d", "",
		"info to connect to database, host=host port=port user=myuser password=xxxx dbname=mydb sslmode=disable",
	)

	var key string
	flag.StringVar(&key, "k", "", "key for signing")

	var privateKeyPath string
	flag.StringVar(&privateKeyPath, "crypto-key", "", "private key for decryption")

	var trustedSubnet string
	flag.StringVar(&trustedSubnet, "t", "", "trusted_subnet")

	flag.Parse()

	builder = builder.
		setAddress(address).
		setLogLevel(loglevel).
		setStoreInterval(storeInterval).
		setFileStoragePath(fileStoragePath).
		setRestore(restore).
		setDatabaseDSN(databaseDSN).
		setSignKey(key).
		setPrivateKeyPath(privateKeyPath).
		setTrustedSubnet(trustedSubnet)

	err = env.Parse(&builder.config)
	if err != nil {
		return nil, err
	}

	return &builder.config, nil
}
