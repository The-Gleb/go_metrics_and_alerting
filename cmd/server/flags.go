package main

import (
	"flag"
	"os"
)

var flagRunAddr string

func parseFlags() {
	flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}
}
