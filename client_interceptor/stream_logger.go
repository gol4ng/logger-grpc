package client_interceptor

import (
	"context"
	"fmt"
	"time"

	"github.com/gol4ng/logger"
	"google.golang.org/grpc"

	logger_grpc "github.com/gol4ng/logger-grpc"
)

// StreamInterceptor returns a new streaming client interceptor that optionally logs the execution of external gRPC calls.
func StreamInterceptor(log logger.LoggerInterface, opts ...logger_grpc.Option) grpc.StreamClientInterceptor {
	o := logger_grpc.EvaluateClientOpt(opts)
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (stream grpc.ClientStream, err error) {
		startTime := time.Now()

		currentLogger := logger.FromContext(ctx, log)
		currentLoggerContext := logger_grpc.FeedContext(o.LoggerContextProvider(method), ctx, method, startTime).Add("grpc_kind", "client")

		defer func() {
			duration := time.Since(startTime)
			currentLoggerContext.Add("grpc_duration", duration.Seconds())

			if err := recover(); err != nil {
				currentLoggerContext.Add("grpc_panic", err)
				currentLogger.Critical(fmt.Sprintf("grpc client stream panic %s [duration:%s]", method, duration), *currentLoggerContext.Slice()...)
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

			currentLogger.Log(fmt.Sprintf("grpc client stream call %s [code:%s, duration:%s]", method, codeStr, duration), o.LevelFunc(code), *currentLoggerContext.Slice()...)
		}()

		currentLogger.Debug("grpc client begin stream call "+method, *currentLoggerContext.Slice()...)
		stream, err = streamer(ctx, desc, cc, method, opts...)
		if err == nil {
			stream = NewClientStreamWrapper(stream, o, currentLogger, *currentLoggerContext)
		}
		return stream, err
	}
}
