package router

import (
	"net/http"

	"github.com/galogen13/yandex-go-metrics/internal/handler"
	"github.com/go-chi/chi/v5"
)

const (
	reqContentTypeTextPlain  = "text/plain"
	respContentTypeTextPlain = "text/plain; charset=utf-8"
)

func Start(serverService handler.Server) error {
	r := metricsRouter(serverService)
	return http.ListenAndServe(serverService.Host(), r)
}

func metricsRouter(server handler.Server) *chi.Mux {
	r := chi.NewRouter()

	r.NotFound(notFoundHandler())
	r.MethodNotAllowed(methodNotAllowedHandler())

	r.Get("/", handler.GetListHandler(server))
	r.Post("/update/{mType}/{metrics}/{value}", handler.UpdateHandler(server))
	r.Get("/value/{mType}/{metrics}", handler.GetValueHandler(server))

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
