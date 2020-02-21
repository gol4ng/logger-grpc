package client_interceptor

import (
	"context"
	"fmt"
	"time"

	"github.com/gol4ng/logger"
	"google.golang.org/grpc"

	logger_grpc "github.com/gol4ng/logger-grpc"
)

// UnaryInterceptor returns a new unary client interceptor that optionally logs the execution of external gRPC calls.
func UnaryInterceptor(log logger.LoggerInterface, opts ...logger_grpc.Option) grpc.UnaryClientInterceptor {
	o := logger_grpc.EvaluateClientOpt(opts)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		startTime := time.Now()

		currentLogger := logger.FromContext(ctx, log)
		currentLoggerContext := logger_grpc.FeedContext(o.LoggerContextProvider(method), ctx, method, startTime).Add("grpc_kind", "client")

		defer func() {
			duration := time.Since(startTime)
			currentLoggerContext.Add("grpc_duration", duration.Seconds())

			if err := recover(); err != nil {
				currentLoggerContext.Add("grpc_panic", err)
				currentLogger.Critical(fmt.Sprintf("grpc client unary panic %s [duration:%s]", method, duration), *currentLoggerContext.Slice()...)
				panic(err)
			}

			code := o.CodeFunc(err)
			codeStr := code.String()
			currentLoggerContext.Add("grpc_code", codeStr)
			if err != nil {
				currentLoggerContext.
					Add("grpc_error", err).
					Add("grpc_error_message", err.Error())
			}

			currentLogger.Log(fmt.Sprintf("grpc client unary call %s [code:%s, duration:%s]", method, codeStr, duration), o.LevelFunc(code), *currentLoggerContext.Slice()...)
		}()

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
