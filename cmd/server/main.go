package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/galogen13/yandex-go-metrics/internal/audit"
	"github.com/galogen13/yandex-go-metrics/internal/buildinfo"
	"github.com/galogen13/yandex-go-metrics/internal/config"
	"github.com/galogen13/yandex-go-metrics/internal/logger"
	memstorage "github.com/galogen13/yandex-go-metrics/internal/repository/memstorage"
	pgstorage "github.com/galogen13/yandex-go-metrics/internal/repository/pgstorage"
	"github.com/galogen13/yandex-go-metrics/internal/service/server"
	grpcserver "github.com/galogen13/yandex-go-metrics/internal/service/server/grpc"
	httpserver "github.com/galogen13/yandex-go-metrics/internal/service/server/http"
	"go.uber.org/zap"
)

var (
	buildVersion string = buildinfo.BuildInfoNotAvaluable
	buildDate    string = buildinfo.BuildInfoNotAvaluable
	buildCommit  string = buildinfo.BuildInfoNotAvaluable
)

func main() {

	buildinfo.PrintBuildInfo(buildVersion, buildDate, buildCommit)

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

	logger.Log.Info("Running server service",
		zap.Bool("restore storage", *config.RestoreStorage),
		zap.Bool("use database as storage", config.UseDatabaseAsStorage),
		zap.Bool("store periodically", config.StorePeriodically),
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	var mStorage server.Storage

	if config.UseDatabaseAsStorage {
		mStorage, err = pgstorage.NewPGStorage(ctx, config.DatabaseDSN)
		if err != nil {
			return err
		}
	} else {
		mStorage = memstorage.NewMemStorage()
	}
	defer mStorage.Close()

	auditService := auditService(config)

	serverService, err := server.NewServerService(config, mStorage, auditService)
	if err != nil {
		return fmt.Errorf("failed to create new server service: %w", err)
	}
	if err = serverService.Start(ctx); err != nil {
		return fmt.Errorf("failed to start new server service: %w", err)
	}

	type mServerI interface {
		Start(ctx context.Context) error
	}

	var mServer mServerI

	if config.UseGRPC {
		mServer, err = grpcserver.NewMetricsServer(config, serverService)
		if err != nil {
			return fmt.Errorf("failed to create new metrics service: %w", err)
		}
	} else {
		mServer, err = httpserver.NewMetricsServer(config, serverService)
		if err != nil {
			return fmt.Errorf("failed to create new metrics service: %w", err)
		}
	}

	if err := mServer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start metrics service: %w", err)
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
