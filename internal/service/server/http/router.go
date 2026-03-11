package httpserver

import (
	"net/http"

	"github.com/galogen13/yandex-go-metrics/internal/compression"
	"github.com/galogen13/yandex-go-metrics/internal/crypto"
	"github.com/galogen13/yandex-go-metrics/internal/handler"
	"github.com/galogen13/yandex-go-metrics/internal/logger"
	"github.com/galogen13/yandex-go-metrics/internal/trusted"
	"github.com/galogen13/yandex-go-metrics/internal/validation"
	"github.com/go-chi/chi/v5"
)

const (
	reqContentTypeTextPlain  = "text/plain"
	respContentTypeTextPlain = "text/plain; charset=utf-8"
)

func metricsRouter(server *MetricsServer) *chi.Mux {
	r := chi.NewRouter()

	r.Use(
		logger.RequestLogger(),
		trusted.TrustedSubnetMiddleware(server.trustedSubnet),
	)

	r.NotFound(notFoundHandler())
	r.MethodNotAllowed(methodNotAllowedHandler())

	listHandler := handler.GetListHandler(server.serverService)
	listHandler = compression.GzipMiddleware(listHandler)
	r.Get("/", listHandler)

	r.Route("/ping", func(r chi.Router) {
		r.Get("/", handler.PingStorageHandler(server.serverService))
	})

	r.Route("/update", func(r chi.Router) {
		updateHandler := handler.UpdateHandler(server.serverService)
		updateHandler = compression.GzipMiddleware(updateHandler)
		updateHandler = validation.HashValidation(server.key, updateHandler)
		if server.decryptor != nil {
			updateHandler = crypto.DecryptMiddleware(server.decryptor, updateHandler)
		}
		r.Post("/", updateHandler)

		r.Post("/{mType}/{metrics}/{value}", handler.UpdateURLHandler(server.serverService))
	})

	r.Route("/updates", func(r chi.Router) {
		updatesHandler := handler.UpdatesHandler(server.serverService)
		updatesHandler = compression.GzipMiddleware(updatesHandler)
		updatesHandler = validation.HashValidation(server.key, updatesHandler)
		if server.decryptor != nil {
			updatesHandler = crypto.DecryptMiddleware(server.decryptor, updatesHandler)
		}
		r.Post("/", updatesHandler)
	})

	r.Route("/value", func(r chi.Router) {
		valueHandler := handler.GetValueHandler(server.serverService)
		valueHandler = compression.GzipMiddleware(valueHandler)
		valueHandler = validation.HashValidation(server.key, valueHandler)
		if server.decryptor != nil {
			valueHandler = crypto.DecryptMiddleware(server.decryptor, valueHandler)
		}
		r.Post("/", valueHandler)

		r.Get("/{mType}/{metrics}", handler.GetValueURLHandler(server.serverService))
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
