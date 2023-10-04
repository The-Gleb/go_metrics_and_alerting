package main

import (
	"flag"
	// "github.com/caarlos0/env/v6"
	// "log"
	"os"
	"strconv"
)

var (
	flagRunAddr    string
	reportInterval int
	pollInterval   int
)

// type Config struct {
// 	flagRunAddr    string `env:"ADDRES"`
// 	reportInterval int    `env:"REPORT_INTERVAL"`
// 	// required требует, чтобы переменная TASK_DURATION была определена
// 	polltInterval int `env:"POLL_INTERVAL"`
// }

// var config Config

func parseFlags() {

	flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.IntVar(&reportInterval, "r", 10, "interval between sending metric on server")
	flag.IntVar(&pollInterval, "p", 2, "interval between collecting metric from runtime")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRES"); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		reportInterval, _ = strconv.Atoi(envReportInterval)
	}
	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		pollInterval, _ = strconv.Atoi(envPollInterval)
	}
}
