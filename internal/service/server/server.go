package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/galogen13/yandex-go-metrics/internal/config"
	"github.com/galogen13/yandex-go-metrics/internal/logger"
	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
	"go.uber.org/zap"
)

type Storage interface {
	Update(metric metrics.Metric)
	Get(metrics.Metric) (bool, metrics.Metric)
	GetByID(ID string) (bool, metrics.Metric)
	GetAll() []metrics.Metric
	RestoreFromFile(fileStoragePath string) error
	SaveToFile(fileStoragePath string) error
}

type ServerService struct {
	Storage Storage
	Config  config.ServerConfig
}

func NewServerService(config config.ServerConfig, storage Storage) *ServerService {
	return &ServerService{Config: config, Storage: storage}
}

func (serverService *ServerService) Start() error {

	if *serverService.Config.RestoreStorage {
		err := serverService.Storage.RestoreFromFile(serverService.Config.FileStoragePath)
		if err != nil {
			return err
		}
	}

	stopChan := make(chan struct{})
	defer close(stopChan)

	if !serverService.Config.StoreOnUpdate {
		go serverService.startPeriodicSave(stopChan)
	}

	r := metricsRouter(serverService)
	logger.Log.Info("Running server",
		zap.String("address", serverService.Config.Host),
		zap.String("logLevel", serverService.Config.LogLevel),
		zap.Int("storeInterval", *serverService.Config.StoreInterval),
		zap.String("fileStoragePath", serverService.Config.FileStoragePath),
		zap.Bool("RestoreStorage", *serverService.Config.RestoreStorage),
	)
	return http.ListenAndServe(serverService.Config.Host, r)
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

	if server.Config.StoreOnUpdate {
		err := server.Storage.SaveToFile(server.Config.FileStoragePath)
		if err != nil {
			logger.Log.Info("cant save metrics to file on update", zap.Error(err))
		}
	}

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

func (server *ServerService) startPeriodicSave(stopChan <-chan struct{}) {
	if *server.Config.StoreInterval == 0 {
		return
	}

	ticker := time.NewTicker(time.Second * time.Duration(*server.Config.StoreInterval))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := server.Storage.SaveToFile(server.Config.FileStoragePath); err != nil {
				logger.Log.Info("cant save metrics to file periodicaly", zap.Error(err))
			}
		case <-stopChan:
			logger.Log.Info("periodic save stopped")
			return
		}
	}
}

func errUpdatingMetrics(err error) error {
	return fmt.Errorf("error updating metrics: %w", err)
}

func errGettingMetrics(err error) error {
	return fmt.Errorf("error getting metrics: %w", err)
}
