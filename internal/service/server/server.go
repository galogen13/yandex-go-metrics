package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/galogen13/yandex-go-metrics/internal/config"
	"github.com/galogen13/yandex-go-metrics/internal/logger"
	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
	"go.uber.org/zap"
)

type Storage interface {
	Update(ctx context.Context, metrics []*metrics.Metric) error
	Insert(ctx context.Context, metrics []*metrics.Metric) error
	Get(ctx context.Context, metric *metrics.Metric) (bool, *metrics.Metric, error)
	GetByIDs(ctx context.Context, IDs []string) (map[string]*metrics.Metric, error)
	GetAll(ctx context.Context) ([]*metrics.Metric, error)
	Ping(ctx context.Context) error
	Close() error
}

type ServerService struct {
	Storage Storage
	Config  *config.ServerConfig
}

func NewServerService(config *config.ServerConfig, storage Storage) *ServerService {
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
		zap.Bool("restore storage", *serverService.Config.RestoreStorage),
		zap.Bool("use database as storage", serverService.Config.UseDatabaseAsStorage),
		zap.Bool("store on update", serverService.Config.StoreOnUpdate),
		zap.Bool("store periodically", serverService.Config.StorePeriodically),
	)
	return http.ListenAndServe(serverService.Config.Host, r)
}

func (serverService *ServerService) UpdateMetric(ctx context.Context, incomingMetric *metrics.Metric) error {

	return serverService.UpdateMetrics(ctx, []*metrics.Metric{incomingMetric})

}

func (serverService *ServerService) UpdateMetrics(ctx context.Context, incomingMetrics []*metrics.Metric) error {

	IDs := make([]string, 0, len(incomingMetrics))
	for _, incomingMetric := range incomingMetrics {
		if err := incomingMetric.Check(true); err != nil {
			return errUpdatingMetrics(err)
		}
		IDs = append(IDs, incomingMetric.ID)
	}

	metricsFound, err := serverService.Storage.GetByIDs(ctx, IDs)
	if err != nil {
		return errUpdatingMetrics(err)
	}

	metricsUpdate := make(map[string]*metrics.Metric, len(incomingMetrics)/2+1)
	metricsInsert := make(map[string]*metrics.Metric, len(incomingMetrics)/2+1)

	for _, incomingMetric := range incomingMetrics {

		metric, ok := metricsInsert[incomingMetric.ID]
		if ok {
			metric.UpdateValue(incomingMetric.GetValue())
			metricsInsert[metric.ID] = metric
			continue
		}

		metric, ok = metricsUpdate[incomingMetric.ID]
		if ok {
			metric.UpdateValue(incomingMetric.GetValue())
			metricsUpdate[metric.ID] = metric
			continue
		}

		metric, ok = metricsFound[incomingMetric.ID]
		if ok {
			err := metric.CompareTypes(incomingMetric.MType)
			if err != nil {
				return errUpdatingMetrics(err)
			}
			metric.UpdateValue(incomingMetric.GetValue())
			metricsUpdate[metric.ID] = metric

		} else {
			metricsInsert[incomingMetric.ID] = incomingMetric
		}

	}

	if len(metricsInsert) > 0 {
		if err := serverService.Storage.Insert(ctx, metrics.GetMetricsFromMap(metricsInsert)); err != nil {
			return errUpdatingMetrics(err)
		}
	}

	if len(metricsUpdate) > 0 {
		if err := serverService.Storage.Update(ctx, metrics.GetMetricsFromMap(metricsUpdate)); err != nil {
			return errUpdatingMetrics(err)
		}
	}

	if serverService.Config.StoreOnUpdate {
		err := serverService.saveStorageToFile(ctx, serverService.Config.FileStoragePath)
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

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := serverService.restoreStorageFromFile(ctx, serverService.Config.FileStoragePath)
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

func (serverService *ServerService) restoreStorageFromFile(ctx context.Context, fileStoragePath string) error {

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
	err = serverService.Storage.Update(ctx, metrics)
	if err != nil {
		return fmt.Errorf("error while updating metrics when restoring from file: %w", err)
	}

	return nil
}

func (serverService *ServerService) startPeriodicSave(stopChan <-chan struct{}) {

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	ticker := time.NewTicker(time.Second * time.Duration(*serverService.Config.StoreInterval))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := serverService.saveStorageToFile(ctx, serverService.Config.FileStoragePath); err != nil {
				logger.Log.Info("cant save metrics to file periodically", zap.Error(err))
			}
		case <-stopChan:
			logger.Log.Info("periodic save stopped")
			return
		}
	}
}

func (serverService *ServerService) saveStorageToFile(ctx context.Context, fileStoragePath string) error {

	if fileStoragePath == "" {
		return fmt.Errorf("fileStoragePath is not filled")
	}

	metrics, err := serverService.Storage.GetAll(ctx)
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

func errUpdatingMetrics(err error) error {
	return fmt.Errorf("error updating metrics: %w", err)
}

func errGettingMetrics(err error) error {
	return fmt.Errorf("error getting metrics: %w", err)
}
