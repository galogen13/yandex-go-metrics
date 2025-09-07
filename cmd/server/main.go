package main

import (
	"log"

	"github.com/galogen13/yandex-go-metrics/internal/config"
	"github.com/galogen13/yandex-go-metrics/internal/logger"
	storage "github.com/galogen13/yandex-go-metrics/internal/repository"
	"github.com/galogen13/yandex-go-metrics/internal/service/server"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {

	config, err := config.GetServerConfig()
	if err != nil {
		return err
	}

	if err := logger.Initialize(config.LogLevel); err != nil {
		return err
	}
	defer logger.Log.Sync()

	var mStorage server.Storage

	if config.DatabaseDSN == "" {
		mStorage = storage.NewMemStorage()
	} else {
		mStorage, err = storage.NewPGStorage(config.DatabaseDSN)
		if err != nil {
			return err
		}
		defer mStorage.Close()
	}

	serverService := server.NewServerService(config, mStorage)

	if err := serverService.Start(); err != nil {
		return err
	}

	return nil
}
