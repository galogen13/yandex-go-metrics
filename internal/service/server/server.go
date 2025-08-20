package server

import (
	"fmt"

	"github.com/galogen13/yandex-go-metrics/internal/config"
	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
)

type Storage interface {
	Update(metric metrics.Metric)
	Get(ID string) (bool, metrics.Metric)
	GetAll() []metrics.Metric
}

type ServerService struct {
	Storage Storage
	Config  config.ServerConfig
}

func NewServerService(config config.ServerConfig, storage Storage) *ServerService {
	return &ServerService{Config: config, Storage: storage}
}

func (server *ServerService) UpdateMetric(ID string, MType string, Value any) error {

	ok, metric := server.Storage.Get(ID)
	if ok {
		err := metric.CheckType(MType)
		if err != nil {
			return errMetricTypeDoesNotMatch(err)
		}
	} else {
		metric = metrics.NewMetrics(ID, MType)
	}

	err := metric.UpdateValue(Value)
	if err != nil {
		return fmt.Errorf("error updating metrics: %w", err)
	}

	server.Storage.Update(metric)

	return nil

}

func (server ServerService) GetMetricValue(ID string, MType string) (any, error) {

	ok, metric := server.Storage.Get(ID)
	if ok {
		err := metric.CheckType(MType)
		if err != nil {
			return nil, errMetricTypeDoesNotMatch(err)
		}
	} else {
		return nil, fmt.Errorf("metric does not exist in storage. ID: %s, mType: %s", ID, MType)
	}

	value := metric.GetValue()
	return value, nil

}

func errMetricTypeDoesNotMatch(err error) error {
	newVar := fmt.Errorf("metric type does not match an existing metric in the repository: %w", err)
	return newVar
}

func (server ServerService) GetAllMetricsValues() map[string]any {

	allMetrics := server.Storage.GetAll()
	metricsValues := metrics.GetMetricsValues(allMetrics)
	return metricsValues

}

func (server ServerService) Host() string {

	return server.Config.Host

}
