package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/galogen13/yandex-go-metrics/internal/config"
	"github.com/galogen13/yandex-go-metrics/internal/logger"
	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
	"go.uber.org/zap"
)

type Storage interface {
	Update(ctx context.Context, metric *metrics.Metric) error
	Insert(ctx context.Context, metric *metrics.Metric) error
	Get(ctx context.Context, metric *metrics.Metric) (bool, *metrics.Metric, error)
	GetByID(ctx context.Context, ID string) (bool, *metrics.Metric, error)
	GetAll(ctx context.Context) ([]*metrics.Metric, error)
	RestoreFromFile(ctx context.Context, fileStoragePath string) error
	SaveToFile(ctx context.Context, fileStoragePath string) error
	Ping(ctx context.Context) error
	Close() error
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
		serverService.restoreFromFile()
	}

	if serverService.Config.StorePeriodically {
		stopChan := make(chan struct{})
		defer close(stopChan)

		go serverService.startPeriodicSave(stopChan)
	}

	r := metricsRouter(serverService)
	logger.Log.Info("Running server",
		zap.String("address", serverService.Config.Host),
		zap.String("logLevel", serverService.Config.LogLevel),
		zap.Int("storeInterval", *serverService.Config.StoreInterval),
		zap.String("file storage path", serverService.Config.FileStoragePath),
		zap.String("database DSN", serverService.Config.DatabaseDSN),
		zap.Bool("restore storage", *serverService.Config.RestoreStorage),
		zap.Bool("use database as storage", serverService.Config.UseDatabaseAsStorage),
		zap.Bool("store on update", serverService.Config.StoreOnUpdate),
		zap.Bool("store periodically", serverService.Config.StorePeriodically),
	)
	return http.ListenAndServe(serverService.Config.Host, r)
}

func (serverService *ServerService) UpdateMetric(ctx context.Context, incomingMetric *metrics.Metric) error {

	if err := incomingMetric.Check(true); err != nil {
		return errUpdatingMetrics(err)
	}

	ok, metric, err := serverService.Storage.GetByID(ctx, incomingMetric.ID)
	if err != nil {
		return errUpdatingMetrics(err)
	}
	if ok {
		err := metric.CompareTypes(incomingMetric.MType)
		if err != nil {
			return errUpdatingMetrics(err)
		}
		metric.UpdateValue(incomingMetric.GetValue())
		if err := serverService.Storage.Update(ctx, metric); err != nil {
			return errUpdatingMetrics(err)
		}
	} else {
		metric = incomingMetric
		if err := serverService.Storage.Insert(ctx, metric); err != nil {
			return errUpdatingMetrics(err)
		}
	}

	if serverService.Config.StoreOnUpdate {
		err := serverService.Storage.SaveToFile(ctx, serverService.Config.FileStoragePath)
		if err != nil {
			logger.Log.Info("cant save metrics to file on update", zap.Error(err))
		}
	}

	return nil

}

func (serverService *ServerService) GetMetric(ctx context.Context, incomingMetric *metrics.Metric) (*metrics.Metric, error) {

	if err := incomingMetric.Check(false); err != nil {
		return nil, errGettingMetrics(err)
	}

	ok, metric, err := serverService.Storage.Get(ctx, incomingMetric)
	if err != nil {
		return nil, errGettingMetrics(err)
	}
	if !ok {
		return nil, fmt.Errorf("%w: ID: %s, mType: %s", metrics.ErrMetricNotFound, incomingMetric.ID, incomingMetric.MType)
	}

	return metric, nil
}

func (serverService *ServerService) GetAllMetricsValues(ctx context.Context) (map[string]any, error) {

	allMetrics, err := serverService.Storage.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting all metrics: %w", err)
	}
	metricsValues := metrics.GetMetricsValues(allMetrics)
	return metricsValues, nil

}

func (serverService *ServerService) PingStorage(ctx context.Context) error {

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	return serverService.Storage.Ping(ctx)

}

func (serverService *ServerService) restoreFromFile() {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := serverService.Storage.RestoreFromFile(ctx, serverService.Config.FileStoragePath)
	if err != nil {
		logger.Log.Info("error while restoring from file", zap.Error(err))
		return
	}
	switch ctx.Err() {
	case context.Canceled:
		logger.Log.Info("restoring from file cancelled", zap.Error(ctx.Err()))
	case context.DeadlineExceeded:
		logger.Log.Info("error while restoring from file", zap.Error(ctx.Err()))
	}

}

func (serverService *ServerService) startPeriodicSave(stopChan <-chan struct{}) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ticker := time.NewTicker(time.Second * time.Duration(*serverService.Config.StoreInterval))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := serverService.Storage.SaveToFile(ctx, serverService.Config.FileStoragePath); err != nil {
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
