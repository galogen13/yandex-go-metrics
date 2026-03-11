package connector

import (
	"context"
	"fmt"

	"github.com/galogen13/yandex-go-metrics/internal/config"
	"github.com/galogen13/yandex-go-metrics/internal/logger"
	"github.com/galogen13/yandex-go-metrics/internal/proto"
	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type GRPCConnector struct {
	host   string
	client *grpc.ClientConn
}

func NewGRPCConnector(config config.AgentConfig) (*GRPCConnector, error) {

	client, err := grpc.NewClient(config.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("gRPC client init failure: %w", err)
	}

	return &GRPCConnector{
			host:   config.Host,
			client: client,
		},
		nil
}

func (c GRPCConnector) SendMetrics(ctx context.Context, curMetrics []metrics.Metric, localIP string) error {

	client := proto.NewMetricsClient(c.client)

	pbMetrics := make([]*proto.Metric, 0, len(curMetrics))
	for _, metric := range curMetrics {

		pbMetric := proto.Metric{
			Id:   metric.ID,
			Type: *pbMTypeByMType(metric.MType),
		}

		switch metric.MType {
		case metrics.Counter:
			pbMetric.Delta = *metric.Delta
		case metrics.Gauge:
			pbMetric.Value = *metric.Value
		}

		pbMetrics = append(pbMetrics, &pbMetric)
	}

	ctxData := metadata.AppendToOutgoingContext(ctx, "x-real-ip", localIP)

	req := &proto.UpdateMetricsRequest{Metrics: pbMetrics}

	_, err := client.UpdateMetrics(ctxData, req)
	if err != nil {
		return fmt.Errorf("failed to send metrics via gRPC: %w", err)
	}

	logger.Log.Info("data sent via gRPC")

	return nil
}

func (c *GRPCConnector) Close() {
	c.client.Close()
	logger.Log.Info("GRPC connector closed")
}

func pbMTypeByMType(mType metrics.MetricType) *proto.Metric_MType {
	switch mType {
	case metrics.Counter:
		return proto.Metric_COUNTER.Enum()
	case metrics.Gauge:
		return proto.Metric_GAUGE.Enum()
	}
	return nil
}
