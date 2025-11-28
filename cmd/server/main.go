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

	auditService := audit.NewAuditServise()
	auditService.Register(audit.NewFileAuditor(config.AuditFile))
	auditService.Register(audit.NewURLAuditor(config.AuditURL))

	serverService := server.NewServerService(config, mStorage, auditService)

	if err := serverService.Start(); err != nil {
		return err
	}

	return nil
}
