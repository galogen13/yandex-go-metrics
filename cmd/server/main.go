package main

import (
	"log"

	"github.com/galogen13/yandex-go-metrics/internal/config"
	"github.com/galogen13/yandex-go-metrics/internal/logger"
	storage "github.com/galogen13/yandex-go-metrics/internal/repository"
	"github.com/galogen13/yandex-go-metrics/internal/router"
	"github.com/galogen13/yandex-go-metrics/internal/service/server"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {

	config := config.GetServerConfig()

	if err := logger.Initialize(config.LogLevel); err != nil {
		return err
	}
	defer logger.Log.Sync()

	storage := storage.NewMemStorage()
	serverService := server.NewServerService(config, storage)

	if err := router.Start(serverService); err != nil {
		return err
	}

	return nil
}
