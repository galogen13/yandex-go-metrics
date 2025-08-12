package main

import (
	"log"

	"github.com/galogen13/yandex-go-metrics/internal/config"
	models "github.com/galogen13/yandex-go-metrics/internal/model"
	"github.com/galogen13/yandex-go-metrics/internal/router"
	"github.com/galogen13/yandex-go-metrics/internal/storage"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {

	config := config.GetServerConfig()

	var storage models.Storage = storage.NewMemStorage()

	if err := router.Start(config, storage); err != nil {
		return err
	}

	return nil
}
