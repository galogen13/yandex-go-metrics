package server

import (
	"net/http"

	"github.com/galogen13/yandex-go-metrics/internal/compression"
	"github.com/galogen13/yandex-go-metrics/internal/crypto"
	"github.com/galogen13/yandex-go-metrics/internal/handler"
	"github.com/galogen13/yandex-go-metrics/internal/logger"
	"github.com/galogen13/yandex-go-metrics/internal/validation"
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

	r.Get("/", logger.RequestLogger(
		compression.GzipMiddleware(
			handler.GetListHandler(server))))

	r.Route("/ping", func(r chi.Router) {
		r.Get("/", logger.RequestLogger(
			handler.PingStorageHandler(server)))
	})

	r.Route("/update", func(r chi.Router) {
		updateHandler := handler.UpdateHandler(server)
		updateHandler = compression.GzipMiddleware(updateHandler)
		updateHandler = validation.HashValidation(server.Key(), updateHandler)

		if server.Decryptor() != nil {
			updateHandler = crypto.DecryptMiddleware(server.Decryptor(), updateHandler)
		}

		r.Post("/", logger.RequestLogger(updateHandler))

		r.Post("/{mType}/{metrics}/{value}", logger.RequestLogger(
			handler.UpdateURLHandler(server)))
	})

	r.Route("/updates", func(r chi.Router) {
		updatesHandler := handler.UpdatesHandler(server)
		updatesHandler = compression.GzipMiddleware(updatesHandler)
		updatesHandler = validation.HashValidation(server.Key(), updatesHandler)

		if server.Decryptor() != nil {
			updatesHandler = crypto.DecryptMiddleware(server.Decryptor(), updatesHandler)
		}

		r.Post("/", logger.RequestLogger(updatesHandler))
	})

	r.Route("/value", func(r chi.Router) {
		valueHandler := handler.GetValueHandler(server)
		valueHandler = compression.GzipMiddleware(valueHandler)
		valueHandler = validation.HashValidation(server.Key(), valueHandler)

		if server.Decryptor() != nil {
			valueHandler = crypto.DecryptMiddleware(server.Decryptor(), valueHandler)
		}

		r.Post("/", logger.RequestLogger(valueHandler))

		r.Get("/{mType}/{metrics}", logger.RequestLogger(handler.GetValueURLHandler(server)))
	})

	return r
}

func notFoundHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", respContentTypeTextPlain)
		w.WriteHeader(http.StatusNotFound)
	})
}

func methodNotAllowedHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", respContentTypeTextPlain)
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
}
