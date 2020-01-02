package server_interceptor_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gol4ng/logger"
	testing_logger "github.com/gol4ng/logger/testing"
	grpc_testing "github.com/grpc-ecosystem/go-grpc-middleware/testing"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	logger_grpc "github.com/gol4ng/logger-grpc"
	"github.com/gol4ng/logger-grpc/server_interceptor"
)

func TestUnaryInterceptor(t *testing.T) {
	myLogger := &testing_logger.Logger{}

	its := &grpc_testing.InterceptorTestSuite{
		ServerOpts: []grpc.ServerOption{
			grpc.UnaryInterceptor(server_interceptor.UnaryInterceptor(myLogger)),
		},
	}
	its.Suite.SetT(t)
	its.SetupSuite()

	c := its.NewClient()
	resp, err := c.Ping(its.SimpleCtx(), &pb_testproto.PingRequest{Value: "my_fake_ping_payload"})

	assert.NoError(t, err)
	assert.Equal(t, "my_fake_ping_payload", resp.Value)
	assert.Equal(t, int32(42), resp.Counter)

	entries := myLogger.GetEntries()
	assert.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, logger.InfoLevel, entry.Level)
	assert.Regexp(t, `grpc server unary call /mwitkow\.testproto\.TestService/Ping \[code:OK, duration:.*\]`, entry.Message)

	assert.Equal(t, "server", (*entry.Context)["grpc_kind"].Value)
	assert.Equal(t, "OK", (*entry.Context)["grpc_code"].Value)
	assert.Equal(t, "mwitkow.testproto.TestService", (*entry.Context)["grpc_service"].Value)
	assert.Equal(t, "Ping", (*entry.Context)["grpc_method"].Value)
	assert.Contains(t, *entry.Context, "grpc_start_time")
	assert.Contains(t, *entry.Context, "grpc_request_deadline")
	assert.Contains(t, *entry.Context, "grpc_duration")
}

func TestUnaryInterceptor_WithContext(t *testing.T) {
	myLogger := &testing_logger.Logger{}

	its := &grpc_testing.InterceptorTestSuite{
		ServerOpts: []grpc.ServerOption{
			grpc.UnaryInterceptor(server_interceptor.UnaryInterceptor(myLogger, logger_grpc.WithLoggerContext(func(fullMethodName string) *logger.Context {
				return logger.NewContext().Add("base_context_key", "base_context_value")
			}))),
		},
	}
	its.Suite.SetT(t)
	its.SetupSuite()

	c := its.NewClient()
	resp, err := c.Ping(its.SimpleCtx(), &pb_testproto.PingRequest{Value: "my_fake_ping_payload"})

	assert.NoError(t, err)
	assert.Equal(t, "my_fake_ping_payload", resp.Value)
	assert.Equal(t, int32(42), resp.Counter)

	entries := myLogger.GetEntries()
	assert.Len(t, entries, 1)

	entry := entries[0]
	assert.Contains(t, (*entry.Context)["base_context_key"].Value, "base_context_value")
	assert.Equal(t, logger.InfoLevel, entry.Level)
	assert.Regexp(t, `grpc server unary call /mwitkow\.testproto\.TestService/Ping \[code:OK, duration:.*\]`, entry.Message)

	assert.Equal(t, "server", (*entry.Context)["grpc_kind"].Value)
	assert.Equal(t, "OK", (*entry.Context)["grpc_code"].Value)
	assert.Equal(t, "mwitkow.testproto.TestService", (*entry.Context)["grpc_service"].Value)
	assert.Equal(t, "Ping", (*entry.Context)["grpc_method"].Value)
	assert.Contains(t, *entry.Context, "grpc_start_time")
	assert.Contains(t, *entry.Context, "grpc_request_deadline")
	assert.Contains(t, *entry.Context, "grpc_duration")
}

func TestUnaryInterceptor_WithLevel(t *testing.T) {
	myLogger := &testing_logger.Logger{}

	its := &grpc_testing.InterceptorTestSuite{
		ServerOpts: []grpc.ServerOption{
			grpc.UnaryInterceptor(server_interceptor.UnaryInterceptor(myLogger, logger_grpc.WithLevels(func(code codes.Code) logger.Level {
				return logger.EmergencyLevel
			}))),
		},
	}
	its.Suite.SetT(t)
	its.SetupSuite()

	c := its.NewClient()
	resp, err := c.Ping(its.SimpleCtx(), &pb_testproto.PingRequest{Value: "my_fake_ping_payload"})

	assert.NoError(t, err)
	assert.Equal(t, "my_fake_ping_payload", resp.Value)
	assert.Equal(t, int32(42), resp.Counter)

	entries := myLogger.GetEntries()
	assert.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, logger.EmergencyLevel, entry.Level)
	assert.Regexp(t, `grpc server unary call /mwitkow\.testproto\.TestService/Ping \[code:OK, duration:.*\]`, entry.Message)

	assert.Equal(t, "server", (*entry.Context)["grpc_kind"].Value)
	assert.Equal(t, "OK", (*entry.Context)["grpc_code"].Value)
	assert.Equal(t, "mwitkow.testproto.TestService", (*entry.Context)["grpc_service"].Value)
	assert.Equal(t, "Ping", (*entry.Context)["grpc_method"].Value)
	assert.Contains(t, *entry.Context, "grpc_start_time")
	assert.Contains(t, *entry.Context, "grpc_request_deadline")
	assert.Contains(t, *entry.Context, "grpc_duration")
}

func TestUnaryInterceptor_WillPanic(t *testing.T) {
	myLogger := &testing_logger.Logger{}
	interceptor := server_interceptor.UnaryInterceptor(myLogger)
	ctx, _ := context.WithTimeout(context.TODO(), 2*time.Second)

	handlerMock := func(innerCtx context.Context, req interface{}) (interface{}, error) {
		assert.Equal(t, ctx, innerCtx)
		panic("my_fake_panic_message")
	}
	assert.PanicsWithValue(t,
		"my_fake_panic_message",
		func() { _, _ = interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/mwitkow.testproto.TestService/Ping"}, handlerMock) },
	)

	entries := myLogger.GetEntries()
	assert.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, logger.CriticalLevel, entry.Level)
	assert.Regexp(t, `grpc server unary panic /mwitkow\.testproto\.TestService/Ping \[duration:.*\]`, entry.Message)

	assert.Equal(t, "server", (*entry.Context)["grpc_kind"].Value)
	assert.Equal(t, "my_fake_panic_message", (*entry.Context)["grpc_panic"].Value)
	assert.NotContains(t, *entry.Context, "grpc_code")
	assert.Equal(t, "mwitkow.testproto.TestService", (*entry.Context)["grpc_service"].Value)
	assert.Equal(t, "Ping", (*entry.Context)["grpc_method"].Value)
	assert.Contains(t, *entry.Context, "grpc_start_time")
	assert.Contains(t, *entry.Context, "grpc_request_deadline")
	assert.Contains(t, *entry.Context, "grpc_duration")
}

func TestUnaryInterceptor_WithError(t *testing.T) {
	myLogger := &testing_logger.Logger{}
	interceptor := server_interceptor.UnaryInterceptor(myLogger)
	ctx, _ := context.WithTimeout(context.TODO(), 2*time.Second)

	handlerMock := func(innerCtx context.Context, req interface{}) (interface{}, error) {
		assert.Equal(t, ctx, innerCtx)
		return nil, errors.New("my_fake_error_message")
	}
	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/mwitkow.testproto.TestService/Ping"}, handlerMock)
	assert.Nil(t, resp)
	assert.EqualError(t, err, "my_fake_error_message")

	entries := myLogger.GetEntries()
	assert.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, logger.ErrorLevel, entry.Level)
	assert.Regexp(t, `grpc server unary call /mwitkow\.testproto\.TestService/Ping \[code:Unknown, duration:.*\]`, entry.Message)

	assert.Equal(t, "server", (*entry.Context)["grpc_kind"].Value)
	assert.EqualError(t, (*entry.Context)["grpc_error"].Value.(error), "my_fake_error_message")
	assert.Equal(t, "Unknown", (*entry.Context)["grpc_code"].Value)
	assert.Equal(t, "mwitkow.testproto.TestService", (*entry.Context)["grpc_service"].Value)
	assert.Equal(t, "Ping", (*entry.Context)["grpc_method"].Value)
	assert.Contains(t, *entry.Context, "grpc_start_time")
	assert.Contains(t, *entry.Context, "grpc_request_deadline")
	assert.Contains(t, *entry.Context, "grpc_duration")
}
