package server_interceptor_test

import (
	"context"
	"testing"

	testing_logger "github.com/gol4ng/logger/testing"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/gol4ng/logger-grpc/server_interceptor"
)

type StreamMock struct {
	grpc.ServerStream
	context context.Context
}

func (s *StreamMock) Context() context.Context {
	return s.context
}

func TestStreamInjectLoggerInterceptor_NilLogger(t *testing.T) {
	interceptor := server_interceptor.StreamInjectLoggerInterceptor(nil)
	ctx := context.TODO()
	handlerMock := func(srv interface{}, stream grpc.ServerStream) error {
		assert.Equal(t, ctx, stream.Context())
		return nil
	}
	err := interceptor(nil, &StreamMock{context: ctx}, &grpc.StreamServerInfo{}, handlerMock)
	assert.NoError(t, err)
}

func TestStreamInjectLoggerInterceptor(t *testing.T) {
	myLogger, store := testing_logger.NewLogger()
	interceptor := server_interceptor.StreamInjectLoggerInterceptor(myLogger)
	ctx := context.TODO()
	handlerMock := func(srv interface{}, stream grpc.ServerStream) error {
		assert.NotEqual(t, ctx, stream.Context())
		return nil
	}
	err := interceptor(nil, &StreamMock{context: ctx}, &grpc.StreamServerInfo{}, handlerMock)
	assert.NoError(t, err)
	assert.Empty(t, store.GetEntries())
}
