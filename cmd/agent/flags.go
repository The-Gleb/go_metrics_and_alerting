package main

import (
	"flag"
	"os"
	// "strconv"
)

type Config struct {
	Addres         string
	PollInterval   int
	ReportInterval int
	SignKey        []byte
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
	b.config.SignKey = []byte(key)
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

	flag.Parse()

	var builder ConfigBuilder

	builder = builder.SetAddres(address).
		SetPollInterval(pollInterval).
		SetReportInterval(reportInterval).
		SetSignKey(key)

	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		builder = builder.SetAddres(envAddress)
	}
	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		builder = builder.SetPollInterval(pollInterval)
	}
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		builder = builder.SetReportInterval(reportInterval)
	}
	if envSignKey := os.Getenv("KEY"); envSignKey != "" {
		builder = builder.SetSignKey(envSignKey)
	}

	return builder.config
}
