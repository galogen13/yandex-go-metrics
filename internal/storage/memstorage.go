package storage

import (
	models "github.com/galogen13/yandex-go-metrics/internal/model"
	"github.com/galogen13/yandex-go-metrics/internal/service"
)

type MemStorage struct {
	Metrics map[string]models.Metrics
}

func NewMemStorage() MemStorage {
	newStorage := MemStorage{}
	newStorage.Metrics = map[string]models.Metrics{}
	return newStorage
}

func (storage MemStorage) Update(ID string, MType string, Value any) (err error) {

	metrics, ok := storage.Metrics[ID]
	if ok {
		err := service.CheckMetricsType(metrics, MType)
		if err != nil {
			return err
		}
		metrics, err = service.UpdateMetricsValue(metrics, Value)
		if err != nil {
			return err
		}

	} else {
		metrics = service.NewMetrics(ID, MType)
		metrics, err = service.UpdateMetricsValue(metrics, Value)
		if err != nil {
			return err
		}
	}
	storage.Metrics[ID] = metrics

	return nil
}
