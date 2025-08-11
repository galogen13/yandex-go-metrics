package handler

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"

	models "github.com/galogen13/yandex-go-metrics/internal/model"
	"github.com/galogen13/yandex-go-metrics/internal/service"
	"github.com/galogen13/yandex-go-metrics/internal/web"
	"github.com/go-chi/chi/v5"
)

const (
	respContentTypeTextPlain = "text/plain; charset=utf-8"
	respContentTypeTextHTML  = "text/html; charset=utf-8"
)

func GetListHandler(storage models.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		metrics := storage.GetAll()
		metricsValues := service.GetMetricsValues(metrics)

		err := web.MetricsListPage(w, metricsValues)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	}
}

func GetValueHandler(storage models.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", respContentTypeTextPlain)

		mID := chi.URLParam(r, "metrics")
		if !checkMetricsID(mID) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		mType := chi.URLParam(r, "mType")
		if !checkMetricsType(mType) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		metrics, err := storage.Get(mID, mType)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		value := service.GetMetricsValue(metrics)
		io.WriteString(w, fmt.Sprintf("%v", value))

	}
}

func UpdateHandler(storage models.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add("Content-type", respContentTypeTextPlain)

		metricsType := chi.URLParam(r, "mType")

		metricsID := chi.URLParam(r, "metrics")
		if !checkMetricsID(metricsID) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !checkMetricsType(metricsType) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		value := chi.URLParam(r, "value")

		var (
			valueConverted any
			err            error
		)

		switch metricsType {
		case models.Counter:
			valueConverted, err = convertCounterValue(value)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		case models.Gauge:
			valueConverted, err = convertGaugeValue(value)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = storage.Update(metricsID, metricsType, valueConverted)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}

func checkMetricsType(mType string) bool {
	return mType == models.Counter || mType == models.Gauge
}

func convertGaugeValue(valueStr string) (float64, error) {
	value, err := strconv.ParseFloat(valueStr, 64)
	return value, err
}

func convertCounterValue(deltaStr string) (int64, error) {
	delta, err := strconv.ParseInt(deltaStr, 10, 64)
	return delta, err
}

func checkMetricsID(id string) bool {
	match, _ := regexp.MatchString("^[a-zA-Z][a-zA-Z0-9]*$", id)
	return match
}
