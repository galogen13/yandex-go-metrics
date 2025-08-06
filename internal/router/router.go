package router

import (
	"net/http"

	handler "github.com/galogen13/yandex-go-metrics/internal/handler"
)

func Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", handler.UpdateHandler)
	return http.ListenAndServe("localhost:8080", mux)
}
