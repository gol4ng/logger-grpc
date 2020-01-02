package client_interceptor_test

import (
	"context"
	"testing"

	testing_logger "github.com/gol4ng/logger/testing"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/gol4ng/logger-grpc/client_interceptor"
)

func TestUnaryInjectLoggerInterceptor_NilLogger(t *testing.T) {
	interceptor := client_interceptor.UnaryInjectLoggerInterceptor(nil)
	ctx := context.Background()
	unaryInvokerMock := func(innerCtx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) (err error) {
		assert.Equal(t, ctx, innerCtx)
		return nil
	}
	err := interceptor(ctx, "my_fake_method", nil, nil, &grpc.ClientConn{}, unaryInvokerMock)
	assert.NoError(t, err)
}

func TestUnaryInjectLoggerInterceptor(t *testing.T) {
	myLogger := &testing_logger.Logger{}
	interceptor := client_interceptor.UnaryInjectLoggerInterceptor(myLogger)
	ctx := context.Background()
	unaryInvokerMock := func(innerCtx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) (err error) {
		assert.NotEqual(t, ctx, innerCtx)
		return nil
	}
	err := interceptor(ctx, "my_fake_method", nil, nil, &grpc.ClientConn{}, unaryInvokerMock)
	assert.NoError(t, err)
	assert.Empty(t, myLogger.GetEntries())
}
