package main

import (
	"github.com/The-Gleb/go_metrics_and_alerting/cmd/agent"
	"github.com/The-Gleb/go_metrics_and_alerting/cmd/server"
)

func main() {
	server.StartServer()
	agent.StartAgent()

}
