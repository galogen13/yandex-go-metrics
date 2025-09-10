package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/galogen13/yandex-go-metrics/internal/logger"
	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
	"go.uber.org/zap"
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

func (storage *MemStorage) Insert(ctx context.Context, metrics *metrics.Metric) error {
	return storage.Update(ctx, metrics)
}

func (storage *MemStorage) Update(ctx context.Context, metrics *metrics.Metric) error {

	storage.mu.Lock()
	defer storage.mu.Unlock()
	storage.Metrics[metrics.ID] = metrics
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

func (storage *MemStorage) GetByID(ctx context.Context, ID string) (bool, *metrics.Metric, error) {

	storage.mu.RLock()
	defer storage.mu.RUnlock()

	metric, ok := storage.Metrics[ID]
	return ok, metric, nil
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

func (storage *MemStorage) RestoreFromFile(ctx context.Context, fileStoragePath string) error {

	if fileStoragePath == "" {
		return fmt.Errorf("fileStoragePath is not filled")
	}

	if _, err := os.Stat(fileStoragePath); os.IsNotExist(err) {
		logger.Log.Info("storage not exists", zap.String("fileStoragePath", fileStoragePath))
		return nil
	}

	file, err := os.Open(fileStoragePath)
	if err != nil {
		return fmt.Errorf("error while opening file to restore: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	metrics := []*metrics.Metric{}
	err = decoder.Decode(&metrics)
	if err != nil {
		return fmt.Errorf("error while marshalling file store: %w", err)
	}
	for _, metric := range metrics {
		err := storage.Update(ctx, metric)
		if err != nil {
			return fmt.Errorf("error while updating metrics when restoring from file: %w", err)
		}
	}
	return nil
}

func (storage *MemStorage) SaveToFile(ctx context.Context, fileStoragePath string) error {

	if fileStoragePath == "" {
		return fmt.Errorf("fileStoragePath is not filled")
	}

	metrics, err := storage.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("error while getting all metrics from storage: %w", err)
	}
	if len(metrics) == 0 {
		logger.Log.Info("no metrics to save in file storage")
		return nil
	}

	file, err := os.Create(fileStoragePath)
	if err != nil {
		return fmt.Errorf("error while create store file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	if err = encoder.Encode(metrics); err != nil {
		return fmt.Errorf("error while encode metrics to file: %w", err)
	}
	logger.Log.Info("metrics saved to file", zap.String("fileStoragePath", fileStoragePath))
	return nil
}
