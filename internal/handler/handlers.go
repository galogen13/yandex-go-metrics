// Пакет handler предоставляет HTTP-обработчики для сервера сбора метрик.
// Обработчики поддерживают обновление и получение метрик через REST API
// как в формате JSON, так и через URL-параметры.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/galogen13/yandex-go-metrics/internal/crypto"
	"github.com/galogen13/yandex-go-metrics/internal/logger"
	addinfo "github.com/galogen13/yandex-go-metrics/internal/service/additional-info"
	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
	"github.com/galogen13/yandex-go-metrics/internal/web"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const (
	respContentTypeTextPlain = "text/plain; charset=utf-8"
)

// Server определяет интерфейс сервиса для работы с метриками.
// Реализации должны предоставлять методы для обновления, получения
// и проверки состояния метрик.
type Server interface {
	// UpdateMetric обновляет одиночную метрику.
	// Принимает контекст, метрику и дополнительную информацию.
	// Возвращает ошибку в случае неудачи.
	UpdateMetric(ctx context.Context, metric *metrics.Metric, addInfo addinfo.AddInfo) error

	// UpdateMetrics обновляет несколько метрик за один запрос.
	// Принимает контекст, слайс метрик и дополнительную информацию.
	// Возвращает ошибку в случае неудачи.
	UpdateMetrics(ctx context.Context, metrics []*metrics.Metric, addInfo addinfo.AddInfo) error

	// GetMetric возвращает метрику по запросу.
	// Принимает контекст и метрику с идентификатором и типом.
	// Возвращает найденную метрику с значением или ошибку.
	GetMetric(ctx context.Context, metric *metrics.Metric) (*metrics.Metric, error)

	// GetAllMetrics возвращает все доступные метрики.
	// Принимает контекст выполнения.
	// Возвращает слайс метрик или ошибку.
	GetAllMetrics(ctx context.Context) ([]*metrics.Metric, error)

	// PingStorage проверяет доступность хранилища метрик.
	// Принимает контекст выполнения.
	// Возвращает ошибку если хранилище недоступно.
	PingStorage(ctx context.Context) error

	// Key возвращает ключ для подписи метрик.
	Key() string

	// Decryptor возвращает декриптор для расщифровки сообщений
	Decryptor() *crypto.Decryptor
}

// PingStorageHandler возвращает HTTP-обработчик для проверки доступности хранилища.
// Обработчик проверяет соединение с базой данных или другим хранилищем.
// В случае успеха возвращает статус 200 OK, при ошибке - 500 Internal Server Error.
//
// Пример запроса:
//
//	GET /ping HTTP/1.1
//
// Пример успешного ответа:
//
//	HTTP/1.1 200 OK
func PingStorageHandler(serverService Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		if err := serverService.PingStorage(ctx); err != nil {
			logger.Log.Error("Error ping storage", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

	}
}

// GetListHandler возвращает HTTP-обработчик для получения списка всех метрик.
// Обработчик возвращает HTML-страницу с таблицей всех метрик и их значений.
// В случае ошибки возвращает статус 500 Internal Server Error.
//
// Пример запроса:
//
//	GET / HTTP/1.1
//
// Пример ответа (HTML):
//
//	HTTP/1.1 200 OK
//	Content-Type: text/html; charset=utf-8
//
//	<html>...список метрик...</html>
func GetListHandler(serverService Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		metricsValues, err := serverService.GetAllMetrics(ctx)
		if err != nil {
			logger.Log.Error("Error getting list of metrics", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		buf, err := web.MetricsListPage(metricsValues)
		if err != nil {
			logger.Log.Error("Error getting page with list of metrics", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(buf.Bytes())

	}
}

// GetValueHandler возвращает HTTP-обработчик для получения значения метрики в формате JSON.
// Обработчик принимает метрику в формате JSON, находит её и возвращает с значением.
// Поддерживает метрики типа gauge и counter.
//
// Пример запроса:
//
//	POST /value HTTP/1.1
//	Content-Type: application/json
//
//	{
//	    "id": "Alloc",
//	    "type": "gauge"
//	}
//
// Пример успешного ответа:
//
//	HTTP/1.1 200 OK
//	Content-Type: application/json
//
//	{
//	    "id": "Alloc",
//	    "type": "gauge",
//	    "value": 123.45
//	}
//
// В случае ошибки возвращает:
//   - 400 Bad Request - некорректный запрос
//   - 404 Not Found - метрика не найдена
//   - 500 Internal Server Error - внутренняя ошибка сервера
func GetValueHandler(serverService Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		w.Header().Set("Content-Type", respContentTypeTextPlain)

		metric := &metrics.Metric{}
		if err := json.NewDecoder(r.Body).Decode(metric); err != nil {
			logger.Log.Error("JSON decoding error", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		metric, err := serverService.GetMetric(ctx, metric)
		if err != nil {
			if errors.Is(err, metrics.ErrMetricNotFound) {
				logger.Log.Info("Error getting metric", zap.Error(err))
			} else {
				logger.Log.Error("Error getting metric", zap.Error(err))
			}

			w.WriteHeader(resolveHTTPStatus(err))
			return
		}

		resp, err := json.Marshal(metric)
		if err != nil {
			logger.Log.Error("Error marshaling metric", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}

// UpdateHandler возвращает HTTP-обработчик для обновления метрики в формате JSON.
// Обработчик принимает метрику с новым значением и сохраняет её.
// Поддерживает обновление одиночных метрик типа gauge и counter.
//
// Пример запроса:
//
//	POST /update HTTP/1.1
//	Content-Type: application/json
//
//	{
//	    "id": "Alloc",
//	    "type": "gauge",
//	    "value": 123.45
//	}
//
// Пример успешного ответа:
//
//	HTTP/1.1 200 OK
//
// В случае ошибки возвращает:
//   - 400 Bad Request - некорректный запрос или валидация
//   - 500 Internal Server Error - внутренняя ошибка сервера
func UpdateHandler(serverService Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		w.Header().Set("Content-type", respContentTypeTextPlain)

		metric := &metrics.Metric{}
		if err := json.NewDecoder(r.Body).Decode(metric); err != nil {
			logger.Log.Error("JSON decoding error", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err := serverService.UpdateMetric(ctx, metric, addinfo.AddInfo{RemoteAddr: r.RemoteAddr})
		if err != nil {
			logger.Log.Error("Error updating metrics", zap.Error(err))
			w.WriteHeader(resolveHTTPStatus(err))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, err = http.NoBody.WriteTo(w)
		if err != nil {
			logger.Log.Error("Error writing body", zap.Error(err))
			w.WriteHeader(resolveHTTPStatus(err))
			return
		}
	}
}

// UpdatesHandler возвращает HTTP-обработчик для массового обновления метрик в формате JSON.
// Обработчик принимает массив метрик и сохраняет их все за один запрос.
//
// Пример запроса:
//
//	POST /updates HTTP/1.1
//	Content-Type: application/json
//
//	[
//	    {
//	        "id": "Alloc",
//	        "type": "gauge",
//	        "value": 123.45
//	    },
//	    {
//	        "id": "PollCount",
//	        "type": "counter",
//	        "delta": 1
//	    }
//	]
//
// Пример успешного ответа:
//
//	HTTP/1.1 200 OK
//
// В случае ошибки возвращает:
//   - 400 Bad Request - некорректный запрос или валидация
//   - 500 Internal Server Error - внутренняя ошибка сервера
func UpdatesHandler(serverService Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		w.Header().Set("Content-type", respContentTypeTextPlain)

		metrics := []*metrics.Metric{}
		if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
			logger.Log.Error("JSON decoding error", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err := serverService.UpdateMetrics(ctx, metrics, addinfo.AddInfo{RemoteAddr: r.RemoteAddr})
		if err != nil {
			logger.Log.Error("Error updating metrics", zap.Error(err))
			w.WriteHeader(resolveHTTPStatus(err))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, err = http.NoBody.WriteTo(w)
		if err != nil {
			logger.Log.Error("Error writing body", zap.Error(err))
			w.WriteHeader(resolveHTTPStatus(err))
			return
		}
	}
}

// GetValueURLHandler возвращает HTTP-обработчик для получения значения метрики через URL.
// Обработчик извлекает параметры из URL и возвращает значение метрики в текстовом формате.
//
// Пример запроса:
//
//	GET /value/gauge/Alloc HTTP/1.1
//
// Пример успешного ответа:
//
//	HTTP/1.1 200 OK
//	Content-Type: text/plain; charset=utf-8
//
//	123.45
//
// В случае ошибки возвращает:
//   - 404 Not Found - метрика не найдена
//   - 500 Internal Server Error - внутренняя ошибка сервера
func GetValueURLHandler(serverService Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		w.Header().Set("Content-Type", respContentTypeTextPlain)

		mID := chi.URLParam(r, "metrics")
		mTypeParam := chi.URLParam(r, "mType")

		mType := metrics.MetricType(mTypeParam)

		metric := metrics.NewMetrics(mID, mType)

		metric, err := serverService.GetMetric(ctx, metric)
		if err != nil {
			if errors.Is(err, metrics.ErrMetricNotFound) {
				logger.Log.Info("Error getting by URL metric", zap.Error(err))
			} else {
				logger.Log.Error("Error getting by URL metric", zap.Error(err))
			}
			w.WriteHeader(resolveHTTPStatus(err))
			return
		}
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, metric.ValueStr)

	}
}

// UpdateURLHandler возвращает HTTP-обработчик для обновления метрики через URL.
// Обработчик извлекает параметры из URL, преобразует значения и сохраняет метрику.
// Поддерживает URL формата: /update/{type}/{name}/{value}
//
// Пример запроса:
//
//	POST /update/gauge/Alloc/123.45 HTTP/1.1
//
// Пример успешного ответа:
//
//	HTTP/1.1 200 OK
//
// В случае ошибки возвращает:
//   - 400 Bad Request - некорректный тип, имя или значение метрики
//   - 500 Internal Server Error - внутренняя ошибка сервера
func UpdateURLHandler(serverService Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		w.Header().Add("Content-type", respContentTypeTextPlain)

		metricTypeParam := chi.URLParam(r, "mType")
		metricID := chi.URLParam(r, "metrics")
		value := chi.URLParam(r, "value")

		var (
			valueConverted any
			err            error
		)

		metricType := metrics.MetricType(metricTypeParam)

		switch metricType {
		case metrics.Counter:
			valueConverted, err = convertCounterValue(value)
			if err != nil {
				logger.Log.Error("Incorrect counter value", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		case metrics.Gauge:
			valueConverted, err = convertGaugeValue(value)
			if err != nil {
				logger.Log.Error("Incorrect gauge value", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		default:
			logger.Log.Error("Incorrect metric type", zap.String("mType", metricTypeParam))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		metric := metrics.NewMetrics(metricID, metricType)
		if err = metric.UpdateValue(valueConverted); err != nil {
			logger.Log.Error("Incorrect metric value",
				zap.Any("type", metric.MType),
				zap.Any("value", valueConverted))
			w.WriteHeader(http.StatusBadRequest)
		}

		err = serverService.UpdateMetric(ctx, metric, addinfo.AddInfo{RemoteAddr: r.RemoteAddr})
		if err != nil {
			logger.Log.Error("Error updating metrics", zap.Error(err))
			w.WriteHeader(resolveHTTPStatus(err))
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func convertGaugeValue(valueStr string) (float64, error) {
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0, fmt.Errorf("error converting string to gauge value (float64): %w", err)
	}
	return value, nil
}

func convertCounterValue(deltaStr string) (int64, error) {
	delta, err := strconv.ParseInt(deltaStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("error converting string to counter value (int64): %w", err)
	}
	return delta, nil
}

func resolveHTTPStatus(err error) int {
	if errors.Is(err, metrics.ErrMetricValidation) {
		return http.StatusBadRequest
	}

	if errors.Is(err, metrics.ErrMetricNotFound) {
		return http.StatusNotFound
	}

	return http.StatusInternalServerError
}
