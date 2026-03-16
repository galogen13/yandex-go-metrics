package grpcserver

import (
	"context"
	"net"
	"time"

	"github.com/galogen13/yandex-go-metrics/internal/logger"
	"github.com/galogen13/yandex-go-metrics/internal/trusted"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func subnetInterceptor(trustedSubnet *net.IPNet) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if trustedSubnet == nil {
			return handler(ctx, req)
		}

		ipStr, err := getClientIPStringFromContext(ctx)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "failed to extract client IP")
		}

		err = trusted.CheckAccessToTrustedNetwork(ipStr, trustedSubnet)
		if err != nil {
			return nil, status.Errorf(codes.PermissionDenied, "access denied: %v", err.Error())
		}

		return handler(ctx, req)
	}
}

func loggerInterceptor() grpc.UnaryServerInterceptor {
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
