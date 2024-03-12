package main

import (
	"flag"
	"os"

	"github.com/caarlos0/env"
)

type Config struct {
	Addres         string  `env:"ADDRESS"`
	PollInterval   float64 `env:"POLL_INTERVAL"`
	ReportInterval float64 `env:"REPORT_INTERVAL"`
	SignKey        string  `env:"KEY"`
	RateLimit      int     `env:"RATE_LIMIT"`
}

type ConfigBuilder struct {
	config Config
}

func (b ConfigBuilder) SetAddres(address string) ConfigBuilder {
	b.config.Addres = address
	return b
}

func (b ConfigBuilder) SetPollInterval(interval float64) ConfigBuilder {
	b.config.PollInterval = interval
	return b
}

func (b ConfigBuilder) SetReportInterval(interval float64) ConfigBuilder {
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

	var pollInterval float64
	flag.Float64Var(&pollInterval, "p", 2, "interval between sending metric on server")

	var reportInterval float64
	flag.Float64Var(&reportInterval, "r", 10, "interval between collecting metric from runtime")

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

	return builder.config
}
