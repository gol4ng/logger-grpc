package server_interceptor_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gol4ng/logger"
	logger_grpc "github.com/gol4ng/logger-grpc"
	testing_logger "github.com/gol4ng/logger/testing"
	grpc_testing "github.com/grpc-ecosystem/go-grpc-middleware/testing"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/gol4ng/logger-grpc/server_interceptor"
)

func TestUnaryInterceptor(t *testing.T) {
	myLogger, store := testing_logger.NewLogger()

	its := getUnaryInterceptorTestSuite(t, myLogger)
	defer its.TearDownSuite()

	resp, err := its.Client.Ping(its.SimpleCtx(), &pb_testproto.PingRequest{Value: "my_fake_ping_payload"})

	its.NoError(err)
	its.Equal("my_fake_ping_payload", resp.Value)
	its.Equal(int32(42), resp.Counter)

	entries := store.GetEntries()
	its.Len(entries, 1)

	AssertUnaryDefaultEntries(t, entries)

	entry := entries[0]
	its.Equal(logger.InfoLevel, entry.Level)
	its.Regexp(`grpc server unary call /mwitkow\.testproto\.TestService/Ping \[code:OK, duration:.*\]`, entry.Message)

	its.Equal("OK", (*entry.Context)["grpc_code"].Value)
}

func TestUnaryInterceptor_WithContext(t *testing.T) {
	myLogger, store := testing_logger.NewLogger()

	its := getUnaryInterceptorTestSuite(t, myLogger, logger_grpc.WithLoggerContext(func(fullMethodName string) *logger.Context {
		return logger.NewContext().Add("base_context_key", "base_context_value")
	}))
	defer its.TearDownSuite()

	resp, err := its.Client.Ping(its.SimpleCtx(), &pb_testproto.PingRequest{Value: "my_fake_ping_payload"})

	its.NoError(err)
	its.Equal("my_fake_ping_payload", resp.Value)
	its.Equal(int32(42), resp.Counter)

	entries := store.GetEntries()
	its.Len(entries, 1)

	AssertUnaryDefaultEntries(t, entries)

	entry := entries[0]
	its.Contains((*entry.Context)["base_context_key"].Value, "base_context_value")
	its.Equal(logger.InfoLevel, entry.Level)
	its.Regexp(`grpc server unary call /mwitkow\.testproto\.TestService/Ping \[code:OK, duration:.*\]`, entry.Message)

	its.Equal("OK", (*entry.Context)["grpc_code"].Value)
}

func TestUnaryInterceptor_WithLevel(t *testing.T) {
	myLogger, store := testing_logger.NewLogger()

	its := getUnaryInterceptorTestSuite(t, myLogger, logger_grpc.WithLevels(func(code codes.Code) logger.Level {
		return logger.EmergencyLevel
	}))
	defer its.TearDownSuite()

	resp, err := its.Client.Ping(its.SimpleCtx(), &pb_testproto.PingRequest{Value: "my_fake_ping_payload"})

	its.NoError(err)
	its.Equal("my_fake_ping_payload", resp.Value)
	its.Equal(int32(42), resp.Counter)

	entries := store.GetEntries()
	its.Len(entries, 1)

	AssertUnaryDefaultEntries(t, entries)

	entry := entries[0]
	its.Equal(logger.EmergencyLevel, entry.Level)
	its.Regexp(`grpc server unary call /mwitkow\.testproto\.TestService/Ping \[code:OK, duration:.*\]`, entry.Message)

	its.Equal("OK", (*entry.Context)["grpc_code"].Value)
}

func TestUnaryInterceptor_WillPanic(t *testing.T) {
	myLogger, store := testing_logger.NewLogger()
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

	entries := store.GetEntries()
	assert.Len(t, entries, 1)

	AssertUnaryDefaultEntries(t, entries)

	entry := entries[0]
	assert.Equal(t, logger.CriticalLevel, entry.Level)
	assert.Regexp(t, `grpc server unary panic /mwitkow\.testproto\.TestService/Ping \[duration:.*\]`, entry.Message)

	assert.Equal(t, "my_fake_panic_message", (*entry.Context)["grpc_panic"].Value)
	assert.NotContains(t, *entry.Context, "grpc_code")
}

func TestUnaryInterceptor_WithError(t *testing.T) {
	myLogger, store := testing_logger.NewLogger()
	interceptor := server_interceptor.UnaryInterceptor(myLogger)
	ctx, _ := context.WithTimeout(context.TODO(), 2*time.Second)

	handlerMock := func(innerCtx context.Context, req interface{}) (interface{}, error) {
		assert.Equal(t, ctx, innerCtx)
		return nil, errors.New("my_fake_error_message")
	}
	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/mwitkow.testproto.TestService/Ping"}, handlerMock)
	assert.Nil(t, resp)
	assert.EqualError(t, err, "my_fake_error_message")

	entries := store.GetEntries()
	assert.Len(t, entries, 1)

	AssertUnaryDefaultEntries(t, entries)

	entry := entries[0]
	assert.Equal(t, logger.ErrorLevel, entry.Level)
	assert.Regexp(t, `grpc server unary call /mwitkow\.testproto\.TestService/Ping \[code:Unknown, duration:.*\]`, entry.Message)

	assert.EqualError(t, (*entry.Context)["grpc_error"].Value.(error), "my_fake_error_message")
	assert.Equal(t, "Unknown", (*entry.Context)["grpc_code"].Value)
}

func getUnaryInterceptorTestSuite(t *testing.T, logger logger.LoggerInterface, opts ...logger_grpc.Option) *grpc_testing.InterceptorTestSuite {
	its := &grpc_testing.InterceptorTestSuite{
		ServerOpts: []grpc.ServerOption{
			grpc.UnaryInterceptor(server_interceptor.UnaryInterceptor(logger, opts...)),
		},
	}
	its.Suite.SetT(t)
	its.SetupSuite()
	return its
}

func AssertUnaryDefaultEntries(t *testing.T, entries []logger.Entry) {
	for _, e := range entries {
		eCtx := *e.Context
		assert.Contains(t, eCtx, "grpc_start_time")
		assert.Contains(t, eCtx, "grpc_duration")
		assert.Contains(t, eCtx, "grpc_request_deadline")
		assert.Equal(t, "server", eCtx["grpc_kind"].Value)
		assert.Equal(t, "Ping", eCtx["grpc_method"].Value)
		assert.Equal(t, "mwitkow.testproto.TestService", eCtx["grpc_service"].Value)
	}
}
