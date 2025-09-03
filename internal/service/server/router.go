package server

import (
	"net/http"

	"github.com/galogen13/yandex-go-metrics/internal/compression"
	"github.com/galogen13/yandex-go-metrics/internal/handler"
	"github.com/galogen13/yandex-go-metrics/internal/logger"
	"github.com/go-chi/chi/v5"
)

const (
	reqContentTypeTextPlain  = "text/plain"
	respContentTypeTextPlain = "text/plain; charset=utf-8"
)

func metricsRouter(server handler.Server) *chi.Mux {
	r := chi.NewRouter()

	r.NotFound(logger.RequestLogger(notFoundHandler()))
	r.MethodNotAllowed(logger.RequestLogger(methodNotAllowedHandler()))

	r.Get("/", logger.RequestLogger(compression.GzipMiddleware(handler.GetListHandler(server))))
	r.Route("/update", func(r chi.Router) {
		r.Post("/", logger.RequestLogger(compression.GzipMiddleware(handler.UpdateHandler(server))))
		r.Post("/{mType}/{metrics}/{value}", logger.RequestLogger(handler.UpdateURLHandler(server)))
	})

	r.Route("/value", func(r chi.Router) {
		r.Post("/", logger.RequestLogger(compression.GzipMiddleware(handler.GetValueHandler(server))))
		r.Get("/{mType}/{metrics}", logger.RequestLogger(handler.GetValueURLHandler(server)))
	})

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
