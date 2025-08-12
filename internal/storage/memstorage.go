package storage

import (
	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
)

type MemStorage struct {
	Metrics map[string]metrics.Metric
}

func NewMemStorage() *MemStorage {
	newStorage := MemStorage{}
	newStorage.Metrics = map[string]metrics.Metric{}
	return &newStorage
}

func (storage *MemStorage) Update(metrics metrics.Metric) {
	storage.Metrics[metrics.ID] = metrics
}

func (storage MemStorage) Get(ID string) (bool, metrics.Metric) {

	metrics, ok := storage.Metrics[ID]
	return ok, metrics
}

func (storage MemStorage) GetAll() []metrics.Metric {
	list := make([]metrics.Metric, 0, len(storage.Metrics))
	for _, metrics := range storage.Metrics {
		list = append(list, metrics)
	}
	return list
}
