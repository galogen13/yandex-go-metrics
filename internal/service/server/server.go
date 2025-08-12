package server

import (
	"github.com/galogen13/yandex-go-metrics/internal/config"
	models "github.com/galogen13/yandex-go-metrics/internal/model"
	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
)

type ServerService struct {
	Storage models.Storage
	Config  config.ServerConfig
}

func NewServerService(config config.ServerConfig, storage models.Storage) *ServerService {
	return &ServerService{Config: config, Storage: storage}
}

func (server *ServerService) UpdateMetric(ID string, MType string, Value any) error {

	ok, metric := server.Storage.Get(ID)
	if ok {
		err := metric.CheckType(MType)
		if err != nil {
			return err
		}
	} else {
		metric = metrics.NewMetrics(ID, MType)
	}

	err := metric.UpdateValue(Value)
	if err != nil {
		return err
	}

	server.Storage.Update(metric)

	return nil

}

func (server ServerService) GetMetricValue(ID string, MType string) (any, error) {

	ok, metric := server.Storage.Get(ID)
	if ok {
		err := metric.CheckType(MType)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, metrics.ErrorMetricsNotExists
	}

	value := metric.GetValue()
	return value, nil

}

func (server ServerService) GetAllMetricsValues() map[string]any {

	allMetrics := server.Storage.GetAll()
	metricsValues := metrics.GetMetricsValues(allMetrics)
	return metricsValues

}

func (server ServerService) Host() string {

	return server.Config.Host

}
