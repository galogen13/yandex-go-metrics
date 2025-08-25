package server

import (
	"fmt"

	"github.com/galogen13/yandex-go-metrics/internal/config"
	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
)

type Storage interface {
	Update(metric metrics.Metric)
	Get(metrics.Metric) (bool, metrics.Metric)
	GetByID(ID string) (bool, metrics.Metric)
	GetAll() []metrics.Metric
}

type ServerService struct {
	Storage Storage
	Config  config.ServerConfig
}

func NewServerService(config config.ServerConfig, storage Storage) *ServerService {
	return &ServerService{Config: config, Storage: storage}
}

func (server *ServerService) UpdateMetric(incomingMetric metrics.Metric) error {

	if err := incomingMetric.Check(true); err != nil {
		return errUpdatingMetrics(err)
	}

	ok, metric := server.Storage.GetByID(incomingMetric.ID)
	if ok {
		err := metric.CompareTypes(incomingMetric.MType)
		if err != nil {
			return errUpdatingMetrics(err)
		}
		metric.UpdateValue(incomingMetric.GetValue())
	} else {
		metric = incomingMetric
	}

	server.Storage.Update(metric)

	return nil

}

func (server ServerService) GetMetric(incomingMetric metrics.Metric) (metrics.Metric, error) {

	if err := incomingMetric.Check(false); err != nil {
		return metrics.Metric{}, errGettingMetrics(err)
	}

	ok, metric := server.Storage.Get(incomingMetric)
	if !ok {
		return metrics.Metric{}, fmt.Errorf("%w: ID: %s, mType: %s", metrics.ErrMetricNotFound, incomingMetric.ID, incomingMetric.MType)
	}

	return metric, nil
}

func (server ServerService) GetAllMetricsValues() map[string]any {

	allMetrics := server.Storage.GetAll()
	metricsValues := metrics.GetMetricsValues(allMetrics)
	return metricsValues

}

func (server ServerService) Host() string {
	return server.Config.Host
}

func errUpdatingMetrics(err error) error {
	return fmt.Errorf("error updating metrics: %w", err)
}

func errGettingMetrics(err error) error {
	return fmt.Errorf("error getting metrics: %w", err)
}
