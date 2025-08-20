package main

import (
	"log"

	"github.com/galogen13/yandex-go-metrics/internal/agent"
	"github.com/galogen13/yandex-go-metrics/internal/config"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	config := config.GetAgentConfig()
	agent.Start(config)

	return nil
}
