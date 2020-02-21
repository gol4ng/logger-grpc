package server_interceptor_test

import (
	"context"
	"testing"

	testing_logger "github.com/gol4ng/logger/testing"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/gol4ng/logger-grpc/server_interceptor"
)

func TestUnaryInjectLoggerInterceptor_NilLogger(t *testing.T) {
	interceptor := server_interceptor.UnaryInjectLoggerInterceptor(nil)
	ctx := context.Background()
	handlerMock := func(innerCtx context.Context, req interface{}) (interface{}, error) {
		assert.Equal(t, ctx, innerCtx)
		return nil, nil
	}
	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handlerMock)
	assert.NoError(t, err)
	assert.Nil(t, resp)
}

func TestUnaryInjectLoggerInterceptor(t *testing.T) {
	myLogger, store := testing_logger.NewLogger()
	interceptor := server_interceptor.UnaryInjectLoggerInterceptor(myLogger)
	ctx := context.TODO()
	handlerMock := func(innerCtx context.Context, req interface{}) (interface{}, error) {
		assert.NotEqual(t, ctx, innerCtx)
		return nil, nil
	}
	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handlerMock)
	assert.NoError(t, err)
	assert.Nil(t, resp)
	assert.Empty(t, store.GetEntries())
}
