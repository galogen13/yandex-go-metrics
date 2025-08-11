package router

import (
	"net/http"
	"strings"

	"github.com/galogen13/yandex-go-metrics/internal/handler"
	models "github.com/galogen13/yandex-go-metrics/internal/model"
	"github.com/go-chi/chi/v5"
)

const (
	reqContentTypeTextPlain  = "text/plain"
	respContentTypeTextPlain = "text/plain; charset=utf-8"
)

func Start(storage models.Storage) error {

	r := metricsRouter(storage)
	return http.ListenAndServe("localhost:8080", r)
}

func metricsRouter(storage models.Storage) *chi.Mux {
	r := chi.NewRouter()

	r.NotFound(notFoundHandler())
	r.MethodNotAllowed(methodNotAllowedHandler())

	r.Get("/", handler.GetListHandler(storage))

	r.Route("/update", func(r chi.Router) {
		r.Use(AllowContentType(reqContentTypeTextPlain))
		r.Post("/{mType}/{metrics}/{value}", handler.UpdateHandler(storage))
	})

	r.Get("/value/{mType}/{metrics}", handler.GetValueHandler(storage))

	return r
}

// Заимствована из chi, убрана проверка на ContentLength = 0, скорректирован код ответа, изменен Content-type
func AllowContentType(contentTypes ...string) func(http.Handler) http.Handler {
	allowedContentTypes := make(map[string]struct{}, len(contentTypes))
	for _, ctype := range contentTypes {
		allowedContentTypes[strings.TrimSpace(strings.ToLower(ctype))] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			s, _, _ := strings.Cut(r.Header.Get("Content-Type"), ";")
			s = strings.ToLower(strings.TrimSpace(s))

			if _, ok := allowedContentTypes[s]; ok {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Add("Content-type", respContentTypeTextPlain)
			w.WriteHeader(http.StatusForbidden)
		})
	}
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
