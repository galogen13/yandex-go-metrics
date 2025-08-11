package router

import (
	"net/http"

	"github.com/galogen13/yandex-go-metrics/internal/handler"
	models "github.com/galogen13/yandex-go-metrics/internal/model"
)

func Start(storage models.Storage) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", handler.UpdateHandler(storage))
	return http.ListenAndServe("localhost:8080", mux)
}
