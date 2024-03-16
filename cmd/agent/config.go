package main

import "github.com/num30/config"

type Config struct {
	Address        string  `default:":8080" flag:"a" envvar:"ADDRESS"`
	LogLevel       string  `default:"info" flag:"log" envvar:"LOGLEVEL"`
	SignKey        string  `default:"secret" flag:"k" envvar:"KEY"`
	PollInterval   float64 `default:"2" flag:"p" envvar:"POLL_INTERVAL"`
	ReportInterval float64 `default:"10" flag:"r" envvar:"REPORT_INTERVAL"`
	PublicKeyPath  string  `default:"/mnt/d/Programming/Go/src/Study/Practicum/go_metrics_and_alerting/cmd/server/public.pem" flag:"crypto-key" envvar:"CRYPTO_KEY"`
	RateLimit      int     `default:"1" flag:"l" envvar:"RATE_LIMIT"`
}

func MustBuildConfig(cfgFile string) *Config {
	var conf Config
	err := config.NewConfReader(cfgFile).Read(&conf)
	if err != nil {
		panic(err)
	}
	return &conf
}
