package handler

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	models "github.com/galogen13/yandex-go-metrics/internal/model"
)

const (
	indexMType = iota
	indexID
	indexValue
	reqContentTypeTextPlain  = "text/plain"
	respContentTypeTextPlain = "text/plain; charset=utf-8"
)

func UpdateHandler(storage models.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", respContentTypeTextPlain)
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if contentType := r.Header.Get("Content-Type"); !strings.Contains(contentType, reqContentTypeTextPlain) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		URIParts := strings.Split(strings.TrimPrefix(r.RequestURI, r.Pattern), "/")

		if len(URIParts) != 3 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		metricsType := URIParts[indexMType]
		if metricsType != models.Counter && metricsType != models.Gauge {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		metricsID := URIParts[indexID]
		if !checkMetricsID(metricsID) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		switch metricsType {
		case models.Counter:
			delta, err := convertCounterValue(URIParts[indexValue])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			err = storage.Update(metricsID, metricsType, delta)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		case models.Gauge:
			value, err := convertGaugeValue(URIParts[indexValue])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			err = storage.Update(metricsID, metricsType, value)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
	}
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
