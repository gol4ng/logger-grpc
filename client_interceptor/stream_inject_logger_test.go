package client_interceptor_test

import (
	"context"
	"testing"

	testing_logger "github.com/gol4ng/logger/testing"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/gol4ng/logger-grpc/client_interceptor"
)

func TestStreamInjectLoggerInterceptor_NilLogger(t *testing.T) {
	interceptor := client_interceptor.StreamInjectLoggerInterceptor(nil)
	ctx := context.Background()
	streamMock := func(innerCtx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (stream grpc.ClientStream, e error) {
		assert.Equal(t, ctx, innerCtx)
		return nil, nil
	}
	stream, err := interceptor(ctx, &grpc.StreamDesc{}, &grpc.ClientConn{}, "my_fake_method", streamMock)
	assert.Nil(t, stream)
	assert.NoError(t, err)
}

func TestStreamInjectLoggerInterceptor(t *testing.T) {
	myLogger := &testing_logger.Logger{}

	interceptor := client_interceptor.StreamInjectLoggerInterceptor(myLogger)
	ctx := context.Background()
	streamMock := func(innerCtx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (stream grpc.ClientStream, e error) {
		assert.NotEqual(t, ctx, innerCtx)
		return nil, nil
	}
	stream, err := interceptor(ctx, &grpc.StreamDesc{}, &grpc.ClientConn{}, "my_fake_method", streamMock)
	assert.Nil(t, stream)
	assert.NoError(t, err)
	assert.Empty(t, myLogger.GetEntries())
}
