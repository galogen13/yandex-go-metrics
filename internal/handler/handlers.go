package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/galogen13/yandex-go-metrics/internal/logger"
	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
	"github.com/galogen13/yandex-go-metrics/internal/web"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const (
	respContentTypeTextPlain = "text/plain; charset=utf-8"
)

type Server interface {
	UpdateMetric(ctx context.Context, metric *metrics.Metric) error
	GetMetric(ctx context.Context, metric *metrics.Metric) (*metrics.Metric, error)
	GetAllMetricsValues(ctx context.Context) map[string]any
	PingStorage(ctx context.Context) error
}

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

func GetListHandler(serverService Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		metricsValues := serverService.GetAllMetricsValues(ctx)

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
			logger.Log.Error("Error getting metric", zap.Error(err))
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

		err := serverService.UpdateMetric(ctx, metric)
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

func GetValueURLHandler(serverService Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		w.Header().Set("Content-Type", respContentTypeTextPlain)

		mID := chi.URLParam(r, "metrics")
		mType := chi.URLParam(r, "mType")

		metric := metrics.NewMetrics(mID, mType)

		metric, err := serverService.GetMetric(ctx, metric)
		if err != nil {
			logger.Log.Error("Error getting metric value", zap.Error(err))
			w.WriteHeader(resolveHTTPStatus(err))
			return
		}
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, fmt.Sprintf("%v", metric.GetValue()))

	}
}

func UpdateURLHandler(serverService Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		w.Header().Add("Content-type", respContentTypeTextPlain)

		metricType := chi.URLParam(r, "mType")
		metricID := chi.URLParam(r, "metrics")
		value := chi.URLParam(r, "value")

		var (
			valueConverted any
			err            error
		)

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
			logger.Log.Error("Incorrect metric type", zap.String("mType", metricType))
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

		err = serverService.UpdateMetric(ctx, metric)
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
