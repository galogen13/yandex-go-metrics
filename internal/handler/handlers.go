package handler

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
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
	UpdateMetric(ID string, MType string, Value any) error
	GetMetricValue(ID string, MType string) (any, error)
	GetAllMetricsValues() map[string]any
	Host() string
}

func GetListHandler(serverService Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		metricsValues := serverService.GetAllMetricsValues()

		buf, err := web.MetricsListPage(metricsValues)
		if err != nil {
			logger.Log.Error("Error getting page with list of metrics", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(buf.Bytes())

	}
}

func GetValueHandler(serverService Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", respContentTypeTextPlain)

		mID := chi.URLParam(r, "metrics")

		metricIDIsCorrect, err := checkMetricID(mID)
		if err != nil {
			logger.Log.Error("ID validity analysis error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !metricIDIsCorrect {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		mType := chi.URLParam(r, "mType")
		if !checkMetricType(mType) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		value, err := serverService.GetMetricValue(mID, mType)

		if err != nil {
			logger.Log.Error("Error getting metric value", zap.Error(err))
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, fmt.Sprintf("%v", value))

	}
}

func UpdateHandler(serverService Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add("Content-type", respContentTypeTextPlain)

		metricType := chi.URLParam(r, "mType")

		metricID := chi.URLParam(r, "metrics")

		metricIDIsCorrect, err := checkMetricID(metricID)
		if err != nil {
			logger.Log.Error("ID validity analysis error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !metricIDIsCorrect {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !checkMetricType(metricType) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		value := chi.URLParam(r, "value")

		var valueConverted any

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

		err = serverService.UpdateMetric(metricID, metricType, valueConverted)
		if err != nil {
			logger.Log.Error("Error updating metrics", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func checkMetricType(mType string) bool {
	return mType == metrics.Counter || mType == metrics.Gauge
}

func convertGaugeValue(valueStr string) (float64, error) {
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0, fmt.Errorf("error converting string to gauge value (float64): %w", err)
	}
	return value, err
}

func convertCounterValue(deltaStr string) (int64, error) {
	delta, err := strconv.ParseInt(deltaStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("error converting string to counter value (int64): %w", err)
	}
	return delta, err
}

func checkMetricID(id string) (bool, error) {
	match, err := regexp.MatchString("^[a-zA-Z][a-zA-Z0-9]*$", id)
	if err != nil {
		return false, fmt.Errorf("error executing regular expression: %w", err)
	}
	return match, err
}
