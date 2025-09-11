package storage

import (
	"context"
	"sync"

	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
)

type MemStorage struct {
	mu      sync.RWMutex
	Metrics map[string]*metrics.Metric
}

func NewMemStorage() *MemStorage {
	newStorage := MemStorage{Metrics: map[string]*metrics.Metric{}}
	return &newStorage
}

func (storage *MemStorage) Ping(ctx context.Context) error {
	return nil
}

func (storage *MemStorage) Close() error {
	return nil
}

func (storage *MemStorage) Insert(ctx context.Context, metrics []*metrics.Metric) error {
	return storage.Update(ctx, metrics)
}

func (storage *MemStorage) Update(ctx context.Context, metrics []*metrics.Metric) error {

	storage.mu.Lock()
	defer storage.mu.Unlock()
	for _, metric := range metrics {
		storage.Metrics[metric.ID] = metric
	}

	return nil
}

func (storage *MemStorage) Get(ctx context.Context, incomingMetric *metrics.Metric) (bool, *metrics.Metric, error) {
	storage.mu.RLock()
	defer storage.mu.RUnlock()
	metric, ok := storage.Metrics[incomingMetric.ID]
	if !ok || metric.MType != incomingMetric.MType {
		return false, nil, nil
	}
	return ok, metric, nil
}

func (storage *MemStorage) GetByIDs(ctx context.Context, IDs []string) (map[string]*metrics.Metric, error) {

	storage.mu.RLock()
	defer storage.mu.RUnlock()

	result := make(map[string]*metrics.Metric)

	for _, ID := range IDs {
		metric, ok := storage.Metrics[ID]
		if ok {
			result[metric.ID] = metric
		}
	}

	return result, nil
}

func (storage *MemStorage) GetAll(ctx context.Context) ([]*metrics.Metric, error) {

	storage.mu.RLock()
	defer storage.mu.RUnlock()

	list := make([]*metrics.Metric, 0, len(storage.Metrics))
	for _, metrics := range storage.Metrics {
		list = append(list, metrics)
	}
	return list, nil
}
