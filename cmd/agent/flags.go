package main

import (
	"flag"
	"github.com/caarlos0/env"
	"os"
	// "strconv"
)

type Config struct {
	Addres         string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	SignKey        string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
}

type ConfigBuilder struct {
	config Config
}

func (b ConfigBuilder) SetAddres(address string) ConfigBuilder {
	b.config.Addres = address
	return b
}

func (b ConfigBuilder) SetPollInterval(interval int) ConfigBuilder {
	b.config.PollInterval = interval
	return b
}

func (b ConfigBuilder) SetReportInterval(interval int) ConfigBuilder {
	b.config.ReportInterval = interval
	return b
}

func (b ConfigBuilder) SetSignKey(key string) ConfigBuilder {
	b.config.SignKey = key
	return b
}

func (b ConfigBuilder) SetRateLimit(limit int) ConfigBuilder {
	b.config.RateLimit = limit
	return b
}

func NewConfigFromFlags() Config {
	flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	var address string
	flag.StringVar(&address, "a", ":8080", "address and port to run server")

	var pollInterval int
	flag.IntVar(&pollInterval, "p", 2, "interval between sending metric on server")

	var reportInterval int
	flag.IntVar(&reportInterval, "r", 10, "interval between collecting metric from runtime")

	var key string
	flag.StringVar(&key, "k", "", "key for signing")

	var rateLimit int
	flag.IntVar(&rateLimit, "l", 1, "number of requests that can be sent simultaniously")

	flag.Parse()

	var builder ConfigBuilder

	builder = builder.SetAddres(address).
		SetPollInterval(pollInterval).
		SetReportInterval(reportInterval).
		SetSignKey(key).
		SetRateLimit(rateLimit)

	env.Parse(&builder.config)

	// if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
	// 	builder = builder.SetAddres(envAddress)
	// }
	// if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
	// 	builder = builder.SetPollInterval(pollInterval)
	// }
	// if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
	// 	builder = builder.SetReportInterval(reportInterval)
	// }
	// if envSignKey := os.Getenv("KEY"); envSignKey != "" {
	// 	builder = builder.SetSignKey(envSignKey)
	// }

	return builder.config
}
