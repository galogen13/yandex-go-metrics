package main

import (
	router "github.com/galogen13/yandex-go-metrics/internal/router"
)

func main() {
	if err := router.Start(); err != nil {
		panic(err)
	}
}
