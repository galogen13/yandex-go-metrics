package handler

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"

	models "github.com/galogen13/yandex-go-metrics/internal/model"
	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
	"github.com/galogen13/yandex-go-metrics/internal/web"
	"github.com/go-chi/chi/v5"
)

const (
	respContentTypeTextPlain = "text/plain; charset=utf-8"
)

func GetListHandler(serverService models.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		metricsValues := serverService.GetAllMetricsValues()

		err := web.MetricsListPage(w, metricsValues)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	}
}

func GetValueHandler(serverService models.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", respContentTypeTextPlain)

		mID := chi.URLParam(r, "metrics")
		if !checkMetricID(mID) {
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
			w.WriteHeader(http.StatusNotFound)
			return
		}

		io.WriteString(w, fmt.Sprintf("%v", value))

	}
}

func UpdateHandler(serverService models.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add("Content-type", respContentTypeTextPlain)

		metricType := chi.URLParam(r, "mType")

		metricID := chi.URLParam(r, "metrics")
		if !checkMetricID(metricID) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !checkMetricType(metricType) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		value := chi.URLParam(r, "value")

		var (
			valueConverted any
			err            error
		)

		switch metricType {
		case metrics.Counter:
			valueConverted, err = convertCounterValue(value)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		case metrics.Gauge:
			valueConverted, err = convertGaugeValue(value)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = serverService.UpdateMetric(metricID, metricType, valueConverted)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}

func checkMetricType(mType string) bool {
	return mType == metrics.Counter || mType == metrics.Gauge
}

func convertGaugeValue(valueStr string) (float64, error) {
	value, err := strconv.ParseFloat(valueStr, 64)
	return value, err
}

func convertCounterValue(deltaStr string) (int64, error) {
	delta, err := strconv.ParseInt(deltaStr, 10, 64)
	return delta, err
}

func checkMetricID(id string) bool {
	match, _ := regexp.MatchString("^[a-zA-Z][a-zA-Z0-9]*$", id)
	return match
}
