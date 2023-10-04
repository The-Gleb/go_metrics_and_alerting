package main

import (
	"flag"
	"os"
)

var (
	flagRunAddr    string
	reportInterval int
	pollInterval   int
)

func parseFlags() {
	flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.IntVar(&reportInterval, "r", 10, "interval between sending metric on server")
	flag.IntVar(&pollInterval, "p", 2, "interval between collecting metric from runtime")
	flag.Parse()
}
