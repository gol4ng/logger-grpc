package client_interceptor

import (
	"context"

	"github.com/gol4ng/logger"
	"google.golang.org/grpc"
)

// StreamInjectLoggerInterceptor returns a new streaming client interceptor with logger injected.
func StreamInjectLoggerInterceptor(log logger.LoggerInterface) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (stream grpc.ClientStream, err error) {
		if log != nil && logger.FromContext(ctx, nil) == nil {
			ctx = logger.InjectInContext(ctx, log)
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}
