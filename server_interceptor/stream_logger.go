package server_interceptor

import (
	"fmt"
	"time"

	"github.com/gol4ng/logger"
	"google.golang.org/grpc"

	logger_grpc "github.com/gol4ng/logger-grpc"
)

// StreamInterceptor returns a new streaming server interceptor that log.
func StreamInterceptor(log logger.LoggerInterface, opts ...logger_grpc.Option) grpc.StreamServerInterceptor {
	o := logger_grpc.EvaluateServerOpt(opts)
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		ctx := stream.Context()
		startTime := time.Now()

		currentLogger := logger.FromContext(ctx, log)
		currentLoggerContext := logger_grpc.FeedContext(o.LoggerContextProvider(info.FullMethod), ctx, info.FullMethod, startTime).Add("grpc_kind", "server")

		defer func() {
			duration := time.Since(startTime)
			currentLoggerContext.Add("grpc_duration", duration.Seconds())

			if err := recover(); err != nil {
				currentLoggerContext.Add("grpc_panic", err)
				currentLogger.Critical(fmt.Sprintf("grpc server stream panic %s [duration:%s]", info.FullMethod, duration), *currentLoggerContext.Slice()...)
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

			currentLogger.Log(fmt.Sprintf("grpc server stream call %s [code:%s, duration:%s]", info.FullMethod, codeStr, duration), o.LevelFunc(code), *currentLoggerContext.Slice()...)
		}()
		currentLogger.Debug("grpc server begin stream call "+info.FullMethod, *currentLoggerContext.Slice()...)
		return handler(srv, NewServerStreamWrapper(stream, ctx, o, currentLogger, *currentLoggerContext))
	}
}
