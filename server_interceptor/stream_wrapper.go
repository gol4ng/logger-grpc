package server_interceptor

import (
	"context"
	"io"
	"time"

	"github.com/gol4ng/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	logger_grpc "github.com/gol4ng/logger-grpc"
)

type StreamWrapper struct {
	grpc.ServerStream
	options       *logger_grpc.Options
	context       context.Context
	logger        logger.LoggerInterface
	loggerContext logger.Context
}

func (s *StreamWrapper) getLoggerContext() *logger.Context {
	return (&logger.Context{}).Merge(s.loggerContext)
}

func (s *StreamWrapper) SendHeader(md metadata.MD) error {
	err := s.ServerStream.SendHeader(md)
	lContext := s.getLoggerContext().Add("grpc_metadata", md)
	if err != nil {
		code := s.options.CodeFunc(err)
		s.logger.Log("grpc server stream send header error", s.options.LevelFunc(code), *lContext.Add("grpc_error", err).Add("grpc_code", code.String()).Slice()...)
		return err
	}
	s.logger.Debug("grpc server stream send header", *lContext.Slice()...)
	return err
}

func (s *StreamWrapper) Context() context.Context {
	return s.context
}

func (s *StreamWrapper) SendMsg(m interface{}) error {
	startTime := time.Now()
	err := s.ServerStream.SendMsg(m)
	lContext := s.getLoggerContext().Add("grpc_send_data", m).Add("grpc_duration", time.Since(startTime).Seconds())
	if err != nil {
		code := s.options.CodeFunc(err)
		s.logger.Log("grpc server stream send error", s.options.LevelFunc(code), *lContext.Add("grpc_error", err).Add("grpc_code", code.String()).Slice()...)
		return err
	}
	s.logger.Debug("grpc server stream send message", *lContext.Slice()...)
	return err
}

func (s *StreamWrapper) RecvMsg(m interface{}) error {
	startTime := time.Now()
	err := s.ServerStream.RecvMsg(m)
	lContext := s.getLoggerContext().Add("grpc_duration", time.Since(startTime).Seconds())
	if err == io.EOF {
		s.logger.Debug("grpc server stream receive EOF", *lContext.Slice()...)
		return err
	}
	lContext.Add("grpc_recv_data", m)
	if err != nil {
		code := s.options.CodeFunc(err)
		s.logger.Log("grpc server stream receive error", s.options.LevelFunc(code), *lContext.Add("grpc_error", err).Add("grpc_code", code.String()).Slice()...)
		return err
	}
	s.logger.Debug("grpc server stream receive message", *lContext.Slice()...)
	return err
}

func NewServerStreamWrapper(stream grpc.ServerStream, context context.Context, options *logger_grpc.Options, l logger.LoggerInterface, loggerContext logger.Context) *StreamWrapper {
	if l == nil {
		l = logger.NewNopLogger()
	}
	return &StreamWrapper{
		ServerStream:  stream,
		context:       context,
		options:       options,
		logger:        l,
		loggerContext: loggerContext,
	}
}
