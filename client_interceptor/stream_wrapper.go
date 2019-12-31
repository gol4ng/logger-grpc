package client_interceptor

import (
	"io"
	"time"

	"github.com/gol4ng/logger"
	"google.golang.org/grpc"

	logger_grpc "github.com/gol4ng/logger-grpc"
)

type StreamWrapper struct {
	grpc.ClientStream
	options       *logger_grpc.Options
	logger        logger.LoggerInterface
	loggerContext logger.Context
}

func (s *StreamWrapper) getLoggerContext() *logger.Context {
	return (&logger.Context{}).Merge(s.loggerContext)
}

func (c *StreamWrapper) SendMsg(m interface{}) error {
	startTime := time.Now()
	err := c.ClientStream.SendMsg(m)
	ctx := c.getLoggerContext().Add("grpc_send_data", m).Add("grpc_duration", time.Since(startTime).Seconds())
	if err != nil {
		code := c.options.CodeFunc(err)
		_ = c.logger.Log("grpc client stream send error", c.options.LevelFunc(code), ctx.Add("grpc_error", err).Add("grpc_code", code.String()))
		return err
	}
	_ = c.logger.Debug("grpc client stream send message", ctx)
	return err
}

func (c *StreamWrapper) RecvMsg(m interface{}) error {
	startTime := time.Now()
	err := c.ClientStream.RecvMsg(m)
	ctx := c.getLoggerContext().Add("grpc_duration", time.Since(startTime).Seconds())
	if err == io.EOF {
		_ = c.logger.Debug("grpc client stream receive EOF", ctx)
		return err
	}
	ctx.Add("grpc_recv_data", m)
	if err != nil {
		code := c.options.CodeFunc(err)
		_ = c.logger.Log("grpc client stream receive error", c.options.LevelFunc(code), ctx.Add("grpc_error", err).Add("grpc_code", code.String()))
		return err
	}
	_ = c.logger.Debug("grpc client stream receive message", ctx)
	return err
}

func NewClientStreamWrapper(stream grpc.ClientStream, options *logger_grpc.Options, l logger.LoggerInterface, loggerContext logger.Context) *StreamWrapper {
	if l == nil {
		l = logger.NewNopLogger()
	}
	return &StreamWrapper{
		ClientStream:  stream,
		options:       options,
		logger:        l,
		loggerContext: loggerContext,
	}
}
