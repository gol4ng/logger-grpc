package client_interceptor_test

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
	"github.com/gol4ng/logger-grpc/client_interceptor"
)

func TestUnaryInterceptor(t *testing.T) {
	myLogger := &testing_logger.Logger{}

	its := &grpc_testing.InterceptorTestSuite{}
	its.Suite.SetT(t)
	its.SetupSuite()

	c := its.NewClient(
		grpc.WithUnaryInterceptor(client_interceptor.UnaryInterceptor(myLogger)),
	)
	resp, err := c.Ping(its.SimpleCtx(), &pb_testproto.PingRequest{Value: "my_fake_ping_payload"})

	assert.NoError(t, err)
	assert.Equal(t, "my_fake_ping_payload", resp.Value)
	assert.Equal(t, int32(42), resp.Counter)

	entries := myLogger.GetEntries()
	assert.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, logger.InfoLevel, entry.Level)
	assert.Regexp(t, `grpc client unary call /mwitkow\.testproto\.TestService/Ping \[code:OK, duration:.*\]`, entry.Message)

	assert.Equal(t, "client", (*entry.Context)["grpc_kind"].Value)
	assert.Equal(t, "OK", (*entry.Context)["grpc_code"].Value)
	assert.Equal(t, "mwitkow.testproto.TestService", (*entry.Context)["grpc_service"].Value)
	assert.Equal(t, "Ping", (*entry.Context)["grpc_method"].Value)
	assert.Contains(t, *entry.Context, "grpc_start_time")
	assert.Contains(t, *entry.Context, "grpc_request_deadline")
	assert.Contains(t, *entry.Context, "grpc_duration")
}

func TestUnaryInterceptor_WithContext(t *testing.T) {
	myLogger := &testing_logger.Logger{}

	its := &grpc_testing.InterceptorTestSuite{}
	its.Suite.SetT(t)
	its.SetupSuite()

	c := its.NewClient(
		grpc.WithUnaryInterceptor(client_interceptor.UnaryInterceptor(myLogger, logger_grpc.WithLoggerContext(func(fullMethodName string) *logger.Context {
			return logger.NewContext().Add("base_context_key", "base_context_value")
		}))),
	)

	resp, err := c.Ping(its.SimpleCtx(), &pb_testproto.PingRequest{Value: "my_fake_ping_payload"})

	assert.NoError(t, err)
	assert.Equal(t, "my_fake_ping_payload", resp.Value)
	assert.Equal(t, int32(42), resp.Counter)

	entries := myLogger.GetEntries()
	assert.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, (*entry.Context)["base_context_key"].Value, "base_context_value")
	assert.Equal(t, logger.InfoLevel, entry.Level)
	assert.Regexp(t, `grpc client unary call /mwitkow\.testproto\.TestService/Ping \[code:OK, duration:.*\]`, entry.Message)

	assert.Equal(t, (*entry.Context)["grpc_kind"].Value, "client")
	assert.Equal(t, (*entry.Context)["grpc_code"].Value, "OK")
	assert.Equal(t, (*entry.Context)["grpc_service"].Value, "mwitkow.testproto.TestService")
	assert.Equal(t, (*entry.Context)["grpc_method"].Value, "Ping")
	assert.Contains(t, *entry.Context, "grpc_start_time")
	assert.Contains(t, *entry.Context, "grpc_request_deadline")
	assert.Contains(t, *entry.Context, "grpc_duration")
}

func TestUnaryInterceptor_WithLevels(t *testing.T) {
	myLogger := &testing_logger.Logger{}

	its := &grpc_testing.InterceptorTestSuite{}
	its.Suite.SetT(t)
	its.SetupSuite()

	c := its.NewClient(
		grpc.WithUnaryInterceptor(client_interceptor.UnaryInterceptor(myLogger, logger_grpc.WithLevels(func(code codes.Code) logger.Level {
			return logger.EmergencyLevel
		}))),
	)
	resp, err := c.Ping(its.SimpleCtx(), &pb_testproto.PingRequest{Value: "my_fake_ping_payload"})

	assert.NoError(t, err)
	assert.Equal(t, "my_fake_ping_payload", resp.Value)
	assert.Equal(t, int32(42), resp.Counter)

	entries := myLogger.GetEntries()
	assert.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, logger.EmergencyLevel, entry.Level)
	assert.Regexp(t, `grpc client unary call /mwitkow\.testproto\.TestService/Ping \[code:OK, duration:.*\]`, entry.Message)

	assert.Equal(t, (*entry.Context)["grpc_kind"].Value, "client")
	assert.Equal(t, (*entry.Context)["grpc_code"].Value, "OK")
	assert.Equal(t, (*entry.Context)["grpc_service"].Value, "mwitkow.testproto.TestService")
	assert.Equal(t, (*entry.Context)["grpc_method"].Value, "Ping")
	assert.Contains(t, *entry.Context, "grpc_start_time")
	assert.Contains(t, *entry.Context, "grpc_request_deadline")
	assert.Contains(t, *entry.Context, "grpc_duration")
}

func TestUnaryInterceptor_WillPanic(t *testing.T) {
	myLogger := &testing_logger.Logger{}
	interceptor := client_interceptor.UnaryInterceptor(myLogger)
	ctx, _ := context.WithTimeout(context.TODO(), 2*time.Second)
	unaryInvokerMock := func(innerCtx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) (err error) {
		assert.Equal(t, ctx, innerCtx)
		panic("my_fake_panic_message")
	}
	assert.PanicsWithValue(t,
		"my_fake_panic_message",
		func() { _ = interceptor(ctx, "/mwitkow.testproto.TestService/Ping", nil, nil, &grpc.ClientConn{}, unaryInvokerMock) },
	)

	entries := myLogger.GetEntries()
	assert.Len(t, entries, 1)

	for _, e := range entries {
		eCtx := *e.Context
		assert.NotContains(t, eCtx, "grpc_code")

		assert.Contains(t, eCtx, "grpc_start_time")
		assert.Contains(t, eCtx, "grpc_duration")
		assert.Contains(t, eCtx, "grpc_request_deadline")
		assert.Equal(t, "client", eCtx["grpc_kind"].Value)
		assert.Equal(t, "Ping", eCtx["grpc_method"].Value)
		assert.Equal(t, "mwitkow.testproto.TestService", eCtx["grpc_service"].Value)
	}

	entry1 := entries[0]
	entry1Ctx := *entry1.Context
	assert.NotContains(t, entry1Ctx, "grpc_send_data")
	assert.NotContains(t, entry1Ctx, "grpc_recv_data")
	assert.Equal(t, logger.CriticalLevel, entry1.Level)
	assert.Regexp(t, `grpc client unary panic /mwitkow\.testproto\.TestService/Ping \[duration:.*\]`, entry1.Message)
}

func TestUnaryInterceptor_WithError(t *testing.T) {
	myLogger := &testing_logger.Logger{}
	interceptor := client_interceptor.UnaryInterceptor(myLogger)
	ctx, _ := context.WithTimeout(context.TODO(), 2*time.Second)
	unaryInvokerMock := func(innerCtx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) (err error) {
		assert.Equal(t, ctx, innerCtx)
		return errors.New("my_fake_panic_message")
	}
	err := interceptor(ctx, "/mwitkow.testproto.TestService/Ping", nil, nil, &grpc.ClientConn{}, unaryInvokerMock)
	assert.EqualError(t, err, "my_fake_panic_message")

	entries := myLogger.GetEntries()
	assert.Len(t, entries, 1)

	for _, e := range entries {
		eCtx := *e.Context
		assert.Contains(t, eCtx, "grpc_start_time")
		assert.Contains(t, eCtx, "grpc_duration")
		assert.Contains(t, eCtx, "grpc_request_deadline")
		assert.Equal(t, "Unknown", eCtx["grpc_code"].Value)
		assert.Equal(t, "client", eCtx["grpc_kind"].Value)
		assert.Equal(t, "Ping", eCtx["grpc_method"].Value)
		assert.Equal(t, "mwitkow.testproto.TestService", eCtx["grpc_service"].Value)
	}

	entry1 := entries[0]
	entry1Ctx := *entry1.Context
	assert.NotContains(t, entry1Ctx, "grpc_send_data")
	assert.NotContains(t, entry1Ctx, "grpc_recv_data")
	assert.Equal(t, logger.ErrorLevel, entry1.Level)
	assert.Regexp(t, `grpc client unary call /mwitkow\.testproto\.TestService/Ping \[code:Unknown, duration:.*\]`, entry1.Message)
}
