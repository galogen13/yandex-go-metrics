package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/galogen13/yandex-go-metrics/internal/audit"
	"github.com/galogen13/yandex-go-metrics/internal/config"
	"github.com/galogen13/yandex-go-metrics/internal/logger"
	addinfo "github.com/galogen13/yandex-go-metrics/internal/service/additional-info"
	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=mocks/storage_mock.go . Storage
type Storage interface {
	Update(ctx context.Context, metrics []*metrics.Metric) error
	Insert(ctx context.Context, metrics []*metrics.Metric) error
	Get(ctx context.Context, metric *metrics.Metric) (*metrics.Metric, error)
	GetByIDs(ctx context.Context, IDs []string) (map[string]*metrics.Metric, error)
	GetAll(ctx context.Context) ([]*metrics.Metric, error)
	Ping(ctx context.Context) error
	Close() error
}

type ServerService struct {
	Storage           Storage
	AuditService      *audit.AuditService
	storeOnUpdate     bool
	fileStoragePath   string
	storeInterval     int
	restoreStorage    bool
	storePeriodically bool
}

func NewServerService(config *config.ServerConfig, storage Storage, auditService *audit.AuditService) (*ServerService, error) {

	return &ServerService{
			Storage:           storage,
			AuditService:      auditService,
			storeOnUpdate:     config.StoreOnUpdate,
			fileStoragePath:   config.FileStoragePath,
			storeInterval:     *config.StoreInterval,
			restoreStorage:    *config.RestoreStorage,
			storePeriodically: config.StorePeriodically,
		},
		nil
}

func (serverService *ServerService) Start(ctx context.Context) error {
	logger.Log.Info("Running server service",
		zap.Bool("restore storage", serverService.restoreStorage),
		zap.String("file storage path", serverService.fileStoragePath),
		zap.Bool("store on update", serverService.storeOnUpdate),
		zap.Bool("store periodically", serverService.storePeriodically),
		zap.Int("storeInterval", serverService.storeInterval),
	)

	if serverService.restoreStorage {
		serverService.restoreFromFile(ctx)
	}

	if serverService.storePeriodically {
		go serverService.startPeriodicSave(ctx)
	}

	return nil
}

func (serverService *ServerService) UpdateMetric(ctx context.Context, incomingMetric *metrics.Metric, addInfo addinfo.AddInfo) error {

	return serverService.UpdateMetrics(ctx, []*metrics.Metric{incomingMetric}, addInfo)

}

func (serverService *ServerService) UpdateMetrics(ctx context.Context, incomingMetrics []*metrics.Metric, addInfo addinfo.AddInfo) error {

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

	metricsUpdate := make([]*metrics.Metric, 0, len(incomingMetrics)/2+1)
	metricsInsert := make([]*metrics.Metric, 0, len(incomingMetrics)/2+1)

	for _, incomingMetric := range incomingMetrics {

		metric, ok := metricsFound[incomingMetric.ID]
		if ok {
			err := metric.CompareTypes(incomingMetric.MType)
			if err != nil {
				return errUpdatingMetrics(err)
			}
			metric.UpdateValue(incomingMetric.GetValue())
			metricsUpdate = append(metricsUpdate, metric)

		} else {
			metricsInsert = append(metricsInsert, incomingMetric)
		}

	}

	if len(metricsInsert) > 0 {
		if err := serverService.Storage.Insert(ctx, metricsInsert); err != nil {
			return errUpdatingMetrics(err)
		}
	}

	if len(metricsUpdate) > 0 {
		if err := serverService.Storage.Update(ctx, metricsUpdate); err != nil {
			return errUpdatingMetrics(err)
		}
	}

	if serverService.storeOnUpdate {
		err := serverService.saveStorageToFile(ctx)
		if err != nil {
			logger.Log.Info("cant save metrics to file on update", zap.Error(err))
		}
	}

	auditLog := audit.NewAuditLog(metrics.GetMetricIDs(incomingMetrics), addInfo.RemoteAddr)
	serverService.AuditService.Notify(auditLog)

	return nil

}

func (serverService *ServerService) GetMetric(ctx context.Context, incomingMetric *metrics.Metric) (*metrics.Metric, error) {

	if err := incomingMetric.Check(false); err != nil {
		return nil, errGettingMetrics(err)
	}

	metric, err := serverService.Storage.Get(ctx, incomingMetric)
	if err != nil {
		return nil, errGettingMetrics(err)
	}
	if metric == nil {
		return nil, fmt.Errorf("%w: ID: %s, mType: %s", metrics.ErrMetricNotFound, incomingMetric.ID, incomingMetric.MType)
	}

	return metric, nil
}

func (serverService *ServerService) GetAllMetrics(ctx context.Context) ([]*metrics.Metric, error) {

	allMetrics, err := serverService.Storage.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting all metrics: %w", err)
	}
	return allMetrics, nil

}

func (serverService *ServerService) PingStorage(ctx context.Context) error {

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	return serverService.Storage.Ping(ctx)

}

func (serverService *ServerService) restoreFromFile(ctx context.Context) {

	ctxTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	err := serverService.restoreStorageFromFile(ctxTimeout)
	if err != nil {
		logger.Log.Info("error while restoring from file", zap.Error(err))
		return
	}
	switch ctxTimeout.Err() {
	case context.Canceled:
		logger.Log.Info("restoring from file cancelled", zap.Error(ctxTimeout.Err()))
	case context.DeadlineExceeded:
		logger.Log.Info("error while restoring from file", zap.Error(ctxTimeout.Err()))
	}

}

func (serverService *ServerService) restoreStorageFromFile(ctx context.Context) error {

	if serverService.fileStoragePath == "" {
		return fmt.Errorf("fileStoragePath is not filled")
	}

	file, err := os.Open(serverService.fileStoragePath)
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

func (serverService *ServerService) startPeriodicSave(ctx context.Context) {

	ctxTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	ticker := time.NewTicker(time.Second * time.Duration(serverService.storeInterval))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := serverService.saveStorageToFile(ctxTimeout); err != nil {
				logger.Log.Info("cant save metrics to file periodically", zap.Error(err))
			}
		case <-ctxTimeout.Done():
			logger.Log.Info("periodic save stopped")
			logger.Log.Info("last save to file")
			if err := serverService.saveStorageToFile(ctxTimeout); err != nil {
				logger.Log.Info("error occurred while last saving to file", zap.Error(err))
			}

			return
		}
	}
}

func (serverService *ServerService) saveStorageToFile(ctx context.Context) error {

	if serverService.fileStoragePath == "" {
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

	file, err := os.Create(serverService.fileStoragePath)
	if err != nil {
		return fmt.Errorf("error while create store file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	if err = encoder.Encode(metrics); err != nil {
		return fmt.Errorf("error while encode metrics to file: %w", err)
	}
	logger.Log.Info("metrics saved to file", zap.String("fileStoragePath", serverService.fileStoragePath))
	return nil
}

func errUpdatingMetrics(err error) error {
	return fmt.Errorf("error updating metrics: %w", err)
}

func errGettingMetrics(err error) error {
	return fmt.Errorf("error getting metrics: %w", err)
}
