package storage

import (
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
	Metrics map[string]metrics.Metric
}

func NewMemStorage() *MemStorage {
	newStorage := MemStorage{}
	newStorage.Metrics = map[string]metrics.Metric{}
	return &newStorage
}

func (storage *MemStorage) Update(metrics metrics.Metric) {

	storage.mu.Lock()
	defer storage.mu.Unlock()
	storage.Metrics[metrics.ID] = metrics
}

func (storage *MemStorage) Get(incomingMetric metrics.Metric) (bool, metrics.Metric) {
	storage.mu.RLock()
	defer storage.mu.RUnlock()
	metric, ok := storage.Metrics[incomingMetric.ID]
	if !ok {
		return false, metrics.Metric{}
	}
	if metric.MType != incomingMetric.MType {
		return false, metrics.Metric{}
	}
	return ok, metric
}

func (storage *MemStorage) GetByID(ID string) (bool, metrics.Metric) {

	storage.mu.RLock()
	defer storage.mu.RUnlock()

	metrics, ok := storage.Metrics[ID]
	return ok, metrics
}

func (storage *MemStorage) GetAll() []metrics.Metric {

	storage.mu.RLock()
	defer storage.mu.RUnlock()

	list := make([]metrics.Metric, 0, len(storage.Metrics))
	for _, metrics := range storage.Metrics {
		list = append(list, metrics)
	}
	return list
}

func (storage *MemStorage) RestoreFromFile(fileStoragePath string) error {

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
	metrics := []metrics.Metric{}
	err = decoder.Decode(&metrics)
	if err != nil {
		return fmt.Errorf("error while marshalling file store: %w", err)
	}
	for _, metric := range metrics {
		storage.Update(metric)
	}
	return nil
}

func (storage *MemStorage) SaveToFile(fileStoragePath string) error {

	if fileStoragePath == "" {
		return fmt.Errorf("fileStoragePath is not filled")
	}

	metrics := storage.GetAll()
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
