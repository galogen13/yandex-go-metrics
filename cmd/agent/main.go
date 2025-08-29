package main

import (
	"log"

	"github.com/galogen13/yandex-go-metrics/internal/agent"
	"github.com/galogen13/yandex-go-metrics/internal/config"
	"github.com/galogen13/yandex-go-metrics/internal/logger"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {

	if err := logger.Initialize("info"); err != nil {
		return err
	}
	defer logger.Log.Sync()

	config, err := config.GetAgentConfig()
	if err != nil {
		return err
	}
	agent.Start(config)

	return nil
}
