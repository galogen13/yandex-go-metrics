package main

import (
	models "github.com/galogen13/yandex-go-metrics/internal/model"
	"github.com/galogen13/yandex-go-metrics/internal/router"
	"github.com/galogen13/yandex-go-metrics/internal/storage"
)

func main() {

	var storage models.Storage = storage.NewMemStorage()

	if err := router.Start(storage); err != nil {
		panic(err)
	}
}
