package main

import (
	"log"

	"github.com/galogen13/yandex-go-metrics/internal/config"
	"github.com/galogen13/yandex-go-metrics/internal/router"
	"github.com/galogen13/yandex-go-metrics/internal/service/server"
	"github.com/galogen13/yandex-go-metrics/internal/storage"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {

	config := config.GetServerConfig()

	storage := storage.NewMemStorage()
	serverService := server.NewServerService(config, storage)

	if err := router.Start(serverService); err != nil {
		return err
	}

	return nil
}
