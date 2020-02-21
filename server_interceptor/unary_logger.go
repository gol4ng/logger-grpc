package server_interceptor

import (
	"context"
	"fmt"
	"time"

	"github.com/gol4ng/logger"
	logger_grpc "github.com/gol4ng/logger-grpc"
	"google.golang.org/grpc"
)

// UnaryInterceptor returns a new unary server interceptors that log.
func UnaryInterceptor(log logger.LoggerInterface, opts ...logger_grpc.Option) grpc.UnaryServerInterceptor {
	o := logger_grpc.EvaluateServerOpt(opts)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		startTime := time.Now()

		currentLogger := logger.FromContext(ctx, log)
		currentLoggerContext := logger_grpc.FeedContext(o.LoggerContextProvider(info.FullMethod), ctx, info.FullMethod, startTime).Add("grpc_kind", "server")

		defer func() {
			duration := time.Since(startTime)
			currentLoggerContext.Add("grpc_duration", duration.Seconds())

			if err := recover(); err != nil {
				currentLoggerContext.Add("grpc_panic", err)
				currentLogger.Critical(fmt.Sprintf("grpc server unary panic %s [duration:%s]", info.FullMethod, duration), *currentLoggerContext.Slice()...)
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

			currentLogger.Log(fmt.Sprintf("grpc server unary call %s [code:%s, duration:%s]", info.FullMethod, codeStr, duration), o.LevelFunc(code), *currentLoggerContext.Slice()...)
		}()

		return handler(ctx, req)
	}
}
