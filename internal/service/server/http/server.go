package httpserver

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/galogen13/yandex-go-metrics/internal/config"
	"github.com/galogen13/yandex-go-metrics/internal/crypto"
	"github.com/galogen13/yandex-go-metrics/internal/handler"
	"github.com/galogen13/yandex-go-metrics/internal/logger"
	"github.com/galogen13/yandex-go-metrics/internal/trusted"
	"go.uber.org/zap"
)

type MetricsServer struct {
	serverService handler.ServerService
	decryptor     *crypto.Decryptor
	host          string
	key           string
	trustedSubnet *net.IPNet
}

func NewMetricsServer(config *config.ServerConfig, ss handler.ServerService) (*MetricsServer, error) {

	trustedSubnet, err := trusted.GetTrustedSubnet(config.TrustedSubnet)
	if err != nil {
		return nil, fmt.Errorf("failed to parse trusted subnet: %w", err)
	}

	decryptor, err := crypto.NewDecryptor(config.CryptoKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create decryptor: %w", err)
	}
	return &MetricsServer{
			serverService: ss,
			decryptor:     decryptor,
			host:          config.Host,
			key:           config.Key,
			trustedSubnet: trustedSubnet,
		},
		nil
}

func (mServer *MetricsServer) Start(ctx context.Context) error {

	r := metricsRouter(mServer)

	httpServer := &http.Server{
		Addr:    mServer.host,
		Handler: r,
	}

	httpServerErrChan := make(chan error)

	go func() {
		defer close(httpServerErrChan)

		logger.Log.Info("Running HTTP server",
			zap.String("address", mServer.host),
			zap.String("trusted subnet", mServer.trustedSubnet.String()),
		)
		if err := httpServer.ListenAndServe(); err != nil {
			httpServerErrChan <- err
		}
	}()

	select {
	case err := <-httpServerErrChan:
		return err
	case <-ctx.Done():
		logger.Log.Info("shutdown signal received, stopping server gracefully...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("HTTP server shutdown error: %w", err)
		}

		logger.Log.Info("Server stopped gracefully, all data saved")
	}

	return nil

}
