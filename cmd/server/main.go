package main

import (
	"flag"

	models "github.com/galogen13/yandex-go-metrics/internal/model"
	"github.com/galogen13/yandex-go-metrics/internal/router"
	"github.com/galogen13/yandex-go-metrics/internal/storage"
)

func main() {
	hostAddress := flag.String("a", "localhost:8080", "host address")
	flag.Parse()

	var storage models.Storage = storage.NewMemStorage()

	if err := router.Start(*hostAddress, storage); err != nil {
		panic(err)
	}
}
