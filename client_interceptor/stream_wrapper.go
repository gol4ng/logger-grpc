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

func (s *StreamWrapper) SendMsg(m interface{}) error {
	startTime := time.Now()
	err := s.ClientStream.SendMsg(m)
	lContext := s.getLoggerContext().Add("grpc_send_data", m).Add("grpc_duration", time.Since(startTime).Seconds())
	if err != nil {
		code := s.options.CodeFunc(err)
		s.logger.Log("grpc client stream send error", s.options.LevelFunc(code), *lContext.Add("grpc_error", err).Add("grpc_code", code.String()).Slice()...)
		return err
	}
	s.logger.Debug("grpc client stream send message", *lContext.Slice()...)
	return err
}

func (s *StreamWrapper) RecvMsg(m interface{}) error {
	startTime := time.Now()
	err := s.ClientStream.RecvMsg(m)
	lContext := s.getLoggerContext().Add("grpc_duration", time.Since(startTime).Seconds())
	if err == io.EOF {
		s.logger.Debug("grpc client stream receive EOF", *lContext.Slice()...)
		return err
	}
	lContext.Add("grpc_recv_data", m)
	if err != nil {
		code := s.options.CodeFunc(err)
		s.logger.Log("grpc client stream receive error", s.options.LevelFunc(code), *lContext.Add("grpc_error", err).Add("grpc_code", code.String()).Slice()...)
		return err
	}
	s.logger.Debug("grpc client stream receive message", *lContext.Slice()...)
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
