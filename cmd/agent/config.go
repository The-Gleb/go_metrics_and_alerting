package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env"
)

type Config struct {
	Address        string  `json:"address" env:"ADDRESS"`
	LogLevel       string  `env:"LOGLEVEL"`
	SignKey        string  `env:"KEY"`
	PollInterval   float64 `json:"poll_interval" env:"POLL_INTERVAL"`
	ReportInterval float64 `json:"report_interval" env:"REPORT_INTERVAL"`
	PublicKeyPath  string  `json:"crypto_key" env:"CRYPTO_KEY"`
	RateLimit      int     `env:"RATE_LIMIT"`
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

func (b ConfigBuilder) SetPollInterval(interval float64) ConfigBuilder {
	if interval != 0.0 {
		b.config.PollInterval = interval
		return b
	}
	if b.config.PollInterval == 0.0 {
		b.config.PollInterval = 1
		return b
	}
	return b
}

func (b ConfigBuilder) SetReportInterval(interval float64) ConfigBuilder {
	if interval != 0.0 {
		b.config.ReportInterval = interval
		return b
	}
	if b.config.ReportInterval == 0.0 && interval == 0.0 {
		b.config.ReportInterval = 2
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

func (b ConfigBuilder) SetRateLimit(limit int) ConfigBuilder {
	if limit != 0 {
		b.config.RateLimit = limit
		return b
	}
	if b.config.RateLimit == 0 {
		b.config.RateLimit = 1
		return b
	}
	return b
}

func (b ConfigBuilder) SetPublicKeyPath(path string) ConfigBuilder {
	if path != "" {
		b.config.PublicKeyPath = path
		return b
	}
	if b.config.PublicKeyPath == "" {
		b.config.PublicKeyPath = "./cmd/server/public.pem"
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

	configPath := flag.String("config", "agent-config.json", "path to config file")
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

	var pollInterval float64
	flag.Float64Var(&pollInterval, "p", 0.0, "interval between sending metric on server")

	var reportInterval float64
	flag.Float64Var(&reportInterval, "r", 0.0, "interval between collecting metric from runtime")

	var key string
	flag.StringVar(&key, "k", "", "key for signing")

	var rateLimit int
	flag.IntVar(&rateLimit, "l", 0, "number of requests that can be sent simultaniously")

	var logLevel string
	flag.StringVar(&logLevel, "log", "", "address and port to run server")

	var publicKeyPath string
	flag.StringVar(&publicKeyPath, "crypto-key", "", "public key path")

	flag.Parse()

	builder = builder.
		SetAddres(address).
		SetPollInterval(pollInterval).
		SetReportInterval(reportInterval).
		SetSignKey(key).
		SetRateLimit(rateLimit).
		SetLogLevel(logLevel).
		SetPublicKeyPath(publicKeyPath)

	env.Parse(&builder.config)

	return &builder.config
}
