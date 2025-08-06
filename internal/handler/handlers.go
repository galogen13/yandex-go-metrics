package handler

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	models "github.com/galogen13/yandex-go-metrics/internal/model"
	"github.com/galogen13/yandex-go-metrics/internal/service"
)

const (
	indexMType = iota
	indexID
	indexValue
	contentTypeTextPlain = "text/plain"
)

var storage models.Storage = service.MemStorage{Metrics: map[string]models.Metrics{}}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if contentType := r.Header.Get("Content-Type"); !strings.Contains(contentType, contentTypeTextPlain) {
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
		err = storage.UpdateCounter(metricsID, delta)
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
		err = storage.UpdateGauge(metricsID, value)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
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
	return startsWithLetter(id) && isAlphanumeric(id)
}

func startsWithLetter(s string) bool {
	if len(s) == 0 {
		return false
	}
	firstChar := rune(s[0])
	return unicode.IsLetter(firstChar)
}

func isAlphanumeric(s string) bool {
	match, _ := regexp.MatchString("^[a-zA-Z0-9]+$", s)
	return match
}
