package client_interceptor

import (
	"context"

	"github.com/gol4ng/logger"
	"google.golang.org/grpc"

	logger_grpc "github.com/gol4ng/logger-grpc"
)

// UnaryInjectLoggerInterceptor returns a new unary client interceptor with logger injected.
func UnaryInjectLoggerInterceptor(log logger.LoggerInterface, opts ...logger_grpc.Option) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		if log != nil && logger.FromContext(ctx, nil) == nil {
			ctx = logger.InjectInContext(ctx, log)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
