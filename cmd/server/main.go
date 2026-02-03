package main

import (
	"context"
	"log"

	"github.com/galogen13/yandex-go-metrics/internal/audit"
	"github.com/galogen13/yandex-go-metrics/internal/config"
	"github.com/galogen13/yandex-go-metrics/internal/logger"
	memstorage "github.com/galogen13/yandex-go-metrics/internal/repository/memstorage"
	pgstorage "github.com/galogen13/yandex-go-metrics/internal/repository/pgstorage"
	"github.com/galogen13/yandex-go-metrics/internal/service/server"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {

	config, err := config.GetServerConfig()
	if err != nil {
		return err
	}

	if err := logger.Initialize(config.LogLevel); err != nil {
		return err
	}
	defer logger.Log.Sync()

	var mStorage server.Storage

	if config.UseDatabaseAsStorage {
		mStorage, err = pgstorage.NewPGStorage(context.Background(), config.DatabaseDSN)
		if err != nil {
			return err
		}
	} else {
		mStorage = memstorage.NewMemStorage()
	}
	defer mStorage.Close()

	auditService := auditService(config)

	serverService := server.NewServerService(config, mStorage, auditService)

	if err := serverService.Start(); err != nil {
		return err
	}

	return nil
}

func auditService(config *config.ServerConfig) *audit.AuditService {

	auditService := audit.NewAuditService()

	fileAuditor, err := audit.NewFileAuditor(config.AuditFile)
	if err != nil {
		logger.Log.Info("auditor not working", zap.Error(err))
	} else {
		auditService.Register(fileAuditor)
	}

	urlAuditor, err := audit.NewURLAuditor(config.AuditURL)
	if err != nil {
		logger.Log.Info("auditor not working", zap.Error(err))
	} else {
		auditService.Register(urlAuditor)
	}

	return auditService
}
