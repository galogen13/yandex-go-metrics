package grpcserver

import (
	"context"
	"fmt"
	"net"

	"github.com/galogen13/yandex-go-metrics/internal/config"
	"github.com/galogen13/yandex-go-metrics/internal/logger"
	"github.com/galogen13/yandex-go-metrics/internal/proto"
	addinfo "github.com/galogen13/yandex-go-metrics/internal/service/additional-info"
	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
	"github.com/galogen13/yandex-go-metrics/internal/trusted"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ServerService interface {
	// UpdateMetrics обновляет несколько метрик за один запрос.
	// Принимает контекст, слайс метрик и дополнительную информацию.
	// Возвращает ошибку в случае неудачи.
	UpdateMetrics(ctx context.Context, metrics []*metrics.Metric, addInfo addinfo.AddInfo) error
}

type MetricsServer struct {
	proto.UnimplementedMetricsServer

	host          string
	serverService ServerService
	trustedSubnet *net.IPNet
}

func NewMetricsServer(config *config.ServerConfig, ss ServerService) (*MetricsServer, error) {

	trustedSubnet, err := trusted.GetTrustedSubnet(config.TrustedSubnet)
	if err != nil {
		return nil, fmt.Errorf("failed to parse trusted subnet: %w", err)
	}

	return &MetricsServer{
			serverService: ss,
			trustedSubnet: trustedSubnet,
			host:          config.Host,
		},
		nil
}

func (mServer *MetricsServer) Start(ctx context.Context) error {

	listen, err := net.Listen("tcp", mServer.host)
	if err != nil {
		return fmt.Errorf("failed to init listener: %ww", err)
	}

	interceptors := []grpc.UnaryServerInterceptor{
		loggerInterceptor(),
	}
	if mServer.trustedSubnet != nil {
		interceptors = append(interceptors, subnetInterceptor(mServer.trustedSubnet))
	}

	s := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))

	proto.RegisterMetricsServer(s, mServer)

	grpcServerErrChan := make(chan error, 1)

	go func() {
		defer close(grpcServerErrChan)

		logger.Log.Info("Running gRPC server",
			zap.String("address", mServer.host),
			zap.String("trusted subnet", mServer.trustedSubnet.String()),
		)

		if err := s.Serve(listen); err != nil {
			grpcServerErrChan <- err
		}

	}()

	select {
	case err := <-grpcServerErrChan:
		return err
	case <-ctx.Done():
		logger.Log.Info("shutdown signal received, stopping server gracefully...")

		s.GracefulStop()

		logger.Log.Info("Server stopped gracefully, all data saved")
	}

	return nil
}

func (mServer *MetricsServer) UpdateMetrics(ctx context.Context, in *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	var response proto.UpdateMetricsResponse

	clientIP, err := getClientIPStringFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to get client IP")
	}

	newMetrics := make([]*metrics.Metric, 0, len(in.GetMetrics()))
	for _, rMetric := range in.GetMetrics() {
		var mType metrics.MetricType
		var mValue any
		switch rMetric.GetType() {
		case proto.Metric_GAUGE:
			mType = metrics.Gauge
			mValue = rMetric.GetValue()
		case proto.Metric_COUNTER:
			mType = metrics.Counter
			mValue = rMetric.GetDelta()
		}
		newMetric := metrics.NewMetrics(rMetric.GetId(), mType)
		err := newMetric.UpdateValue(mValue)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument,
				"invalid metric value update: %v (metric ID: %s, type: %v)",
				err, newMetric.ID, newMetric.MType)
		}
		newMetrics = append(newMetrics, newMetric)
	}

	err = mServer.serverService.UpdateMetrics(ctx, newMetrics, addinfo.AddInfo{RemoteAddr: clientIP})
	if err != nil {
		logger.Log.Error("failed to update metrics", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update metrics")
	}
	return &response, nil
}

func getClientIPStringFromContext(ctx context.Context) (string, error) {

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if values := md.Get("x-real-ip"); len(values) > 0 {
			return values[0], nil
		}
	}

	return "", status.Error(codes.InvalidArgument, "cannot determine client IP")
}
