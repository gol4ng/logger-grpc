package server_interceptor

import (
	"github.com/gol4ng/logger"
	"google.golang.org/grpc"
)

// StreamInjectLoggerInterceptor returns a new streaming server interceptor with logger injected.
func StreamInjectLoggerInterceptor(log logger.LoggerInterface) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		if log != nil {
			ctx := stream.Context()
			if logger.FromContext(ctx, nil) == nil {
				stream = newServerStreamWithContext(stream, logger.InjectInContext(ctx, log))
			}
		}
		return handler(srv, stream)
	}
}
