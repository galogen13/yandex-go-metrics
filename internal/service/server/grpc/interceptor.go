package grpcserver

import (
	"context"
	"net"
	"time"

	"github.com/galogen13/yandex-go-metrics/internal/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func SubnetInterceptor(trustedSubnet *net.IPNet) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if trustedSubnet == nil {
			return handler(ctx, req)
		}

		ip, err := getClientIP(ctx)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "failed to extract client IP")
		}

		if !trustedSubnet.Contains(ip) {
			return nil, status.Errorf(codes.PermissionDenied, "IP %s is not in trusted subnet", ip)
		}

		return handler(ctx, req)
	}
}

func LoggerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

		startTime := time.Now()

		res, err := handler(ctx, req)

		duration := time.Since(startTime)

		logger.Log.Info("incoming request",
			zap.String("method", info.FullMethod),
			zap.String("duration", duration.String()),
			zap.Int("size", getResponseSize(res)),
			zap.Error(err),
		)
		return res, err
	}
}

func getResponseSize(res any) int {
	if res == nil {
		return 0
	}
	if msg, ok := res.(proto.Message); ok {
		return proto.Size(msg)
	}
	return 0
}
