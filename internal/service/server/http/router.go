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

	baseMiddlewares := []func(http.Handler) http.Handler{
		logger.RequestLogger(),
	}
	if server.trustedSubnet != nil {
		baseMiddlewares = append(baseMiddlewares, trusted.TrustedSubnetMiddleware(server.trustedSubnet))
	}
	r.Use(baseMiddlewares...)

	r.NotFound(notFoundHandler())
	r.MethodNotAllowed(methodNotAllowedHandler())

	r.With(
		compression.GzipMiddleware(),
	).Get("/", handler.GetListHandler(server.serverService))

	r.Route("/ping", func(r chi.Router) {
		r.Get("/", handler.PingStorageHandler(server.serverService))
	})

	r.Route("/update", func(r chi.Router) {
		updateMiddlewares := make([]func(http.Handler) http.Handler, 0, 3)
		if server.decryptor != nil {
			updateMiddlewares = append(updateMiddlewares, crypto.DecryptMiddleware(server.decryptor))
		}
		updateMiddlewares = append(updateMiddlewares, validation.HashValidationMiddleware(server.key))
		updateMiddlewares = append(updateMiddlewares, compression.GzipMiddleware())
		r.With(updateMiddlewares...).Post("/", handler.UpdateHandler(server.serverService))

		r.Post("/{mType}/{metrics}/{value}", handler.UpdateURLHandler(server.serverService))
	})

	r.Route("/updates", func(r chi.Router) {
		updatesMiddlewares := make([]func(http.Handler) http.Handler, 0, 3)
		if server.decryptor != nil {
			updatesMiddlewares = append(updatesMiddlewares, crypto.DecryptMiddleware(server.decryptor))
		}
		updatesMiddlewares = append(updatesMiddlewares, validation.HashValidationMiddleware(server.key))
		updatesMiddlewares = append(updatesMiddlewares, compression.GzipMiddleware())
		r.With(updatesMiddlewares...).Post("/", handler.UpdatesHandler(server.serverService))
	})

	r.Route("/value", func(r chi.Router) {
		valueMiddlewares := make([]func(http.Handler) http.Handler, 0, 3)
		if server.decryptor != nil {
			valueMiddlewares = append(valueMiddlewares, crypto.DecryptMiddleware(server.decryptor))
		}
		valueMiddlewares = append(valueMiddlewares, validation.HashValidationMiddleware(server.key))
		valueMiddlewares = append(valueMiddlewares, compression.GzipMiddleware())
		r.With(valueMiddlewares...).Post("/", handler.GetValueHandler(server.serverService))

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
