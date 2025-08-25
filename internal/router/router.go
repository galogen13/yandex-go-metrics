package router

import (
	"net/http"

	"github.com/galogen13/yandex-go-metrics/internal/handler"
	"github.com/galogen13/yandex-go-metrics/internal/logger"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const (
	reqContentTypeTextPlain  = "text/plain"
	respContentTypeTextPlain = "text/plain; charset=utf-8"
)

func Start(serverService handler.Server) error {
	r := metricsRouter(serverService)
	logger.Log.Info("Running server", zap.String("address", serverService.Host()))
	return http.ListenAndServe(serverService.Host(), r)
}

func metricsRouter(server handler.Server) *chi.Mux {
	r := chi.NewRouter()

	r.NotFound(notFoundHandler())
	r.MethodNotAllowed(methodNotAllowedHandler())

	r.Get("/", logger.RequestLogger(handler.GetListHandler(server)))
	r.Post("/update/{mType}/{metrics}/{value}", logger.RequestLogger(handler.UpdateURLHandler(server)))
	r.Get("/value/{mType}/{metrics}", logger.RequestLogger(handler.GetValueURLHandler(server)))
	r.Post("/update", logger.RequestLogger(handler.UpdateHandler(server)))
	r.Get("/value", logger.RequestLogger(handler.GetValueHandler(server)))

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
