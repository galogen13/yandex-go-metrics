package router

import (
	"net/http"

	"github.com/galogen13/yandex-go-metrics/internal/config"
	"github.com/galogen13/yandex-go-metrics/internal/handler"
	models "github.com/galogen13/yandex-go-metrics/internal/model"
	"github.com/go-chi/chi/v5"
)

const (
	reqContentTypeTextPlain  = "text/plain"
	respContentTypeTextPlain = "text/plain; charset=utf-8"
)

func Start(config config.ServerConfig, storage models.Storage) error {

	r := metricsRouter(storage)
	return http.ListenAndServe(config.Host, r)
}

func metricsRouter(storage models.Storage) *chi.Mux {
	r := chi.NewRouter()

	r.NotFound(notFoundHandler())
	r.MethodNotAllowed(methodNotAllowedHandler())

	r.Get("/", handler.GetListHandler(storage))
	r.Post("/update/{mType}/{metrics}/{value}", handler.UpdateHandler(storage))
	r.Get("/value/{mType}/{metrics}", handler.GetValueHandler(storage))

	return r
}

func notFoundHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", respContentTypeTextPlain)
		w.WriteHeader(http.StatusNotFound)
	})
}

func methodNotAllowedHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", respContentTypeTextPlain)
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
}
