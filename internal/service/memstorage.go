package service

import (
	models "github.com/galogen13/yandex-go-metrics/internal/model"
)

type MemStorage struct {
	Metrics map[string]models.Metrics `json:"metrics"`
}

func (storage MemStorage) UpdateGauge(ID string, Value float64) error {

	metrics, ok := storage.Metrics[ID]
	if ok {
		if metrics.MType != models.Gauge {
			return models.ErrorIncorrectUse
		}

		*metrics.Value = Value

	} else {
		metrics.ID = ID
		metrics.MType = models.Gauge
		metrics.Value = &Value
		storage.Metrics[ID] = metrics
	}

	return nil
}

func (storage MemStorage) UpdateCounter(ID string, Delta int64) error {

	metrics, ok := storage.Metrics[ID]
	if ok {
		if metrics.MType != models.Counter {
			return models.ErrorIncorrectUse
		}

		*metrics.Delta += Delta

	} else {
		metrics.ID = ID
		metrics.MType = models.Counter
		metrics.Delta = &Delta
		storage.Metrics[ID] = metrics
	}

	return nil
}
