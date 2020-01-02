package server_interceptor

import (
	"context"

	"github.com/gol4ng/logger"
	"google.golang.org/grpc"
)

// UnaryInjectLoggerInterceptor returns a new unary server interceptors with logger injected.
func UnaryInjectLoggerInterceptor(log logger.LoggerInterface) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if log != nil && logger.FromContext(ctx, nil) == nil {
			ctx = logger.InjectInContext(ctx, log)
		}
		return handler(ctx, req)
	}
}
