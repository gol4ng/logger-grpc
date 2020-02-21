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
	its.Regexp(`grpc client unary call /mwitkow\.testproto\.TestService/Ping \[code:OK, duration:.*\]`, entry.Message)

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
	its.Equal((*entry.Context)["base_context_key"].Value, "base_context_value")
	its.Equal(logger.InfoLevel, entry.Level)
	its.Regexp(`grpc client unary call /mwitkow\.testproto\.TestService/Ping \[code:OK, duration:.*\]`, entry.Message)

	its.Equal("OK", (*entry.Context)["grpc_code"].Value)
}

func TestUnaryInterceptor_WithLevels(t *testing.T) {
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
	its.Equal((*entry.Context)["grpc_code"].Value, "OK")
	its.Contains(*entry.Context, "grpc_duration")
	its.Regexp(`grpc client unary call /mwitkow\.testproto\.TestService/Ping \[code:OK, duration:.*\]`, entry.Message)
}

func TestUnaryInterceptor_WillPanic(t *testing.T) {
	myLogger, store := testing_logger.NewLogger()
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

	entries := store.GetEntries()
	assert.Len(t, entries, 1)

	AssertUnaryDefaultEntries(t, entries)

	for _, e := range entries {
		eCtx := *e.Context
		assert.NotContains(t, eCtx, "grpc_code")
	}

	entry1 := entries[0]
	entry1Ctx := *entry1.Context
	assert.NotContains(t, entry1Ctx, "grpc_send_data")
	assert.NotContains(t, entry1Ctx, "grpc_recv_data")
	assert.Equal(t, logger.CriticalLevel, entry1.Level)
	assert.Regexp(t, `grpc client unary panic /mwitkow\.testproto\.TestService/Ping \[duration:.*\]`, entry1.Message)
}

func TestUnaryInterceptor_WithError(t *testing.T) {
	myLogger, store := testing_logger.NewLogger()
	interceptor := client_interceptor.UnaryInterceptor(myLogger)
	ctx, _ := context.WithTimeout(context.TODO(), 2*time.Second)
	unaryInvokerMock := func(innerCtx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) (err error) {
		assert.Equal(t, ctx, innerCtx)
		return errors.New("my_fake_panic_message")
	}
	err := interceptor(ctx, "/mwitkow.testproto.TestService/Ping", nil, nil, &grpc.ClientConn{}, unaryInvokerMock)
	assert.EqualError(t, err, "my_fake_panic_message")

	entries := store.GetEntries()
	assert.Len(t, entries, 1)

	AssertUnaryDefaultEntries(t, entries)

	entry1 := entries[0]
	entry1Ctx := *entry1.Context
	assert.Equal(t, "Unknown", entry1Ctx["grpc_code"].Value)
	assert.NotContains(t, entry1Ctx, "grpc_send_data")
	assert.NotContains(t, entry1Ctx, "grpc_recv_data")
	assert.Equal(t, logger.ErrorLevel, entry1.Level)
	assert.Regexp(t, `grpc client unary call /mwitkow\.testproto\.TestService/Ping \[code:Unknown, duration:.*\]`, entry1.Message)
}

func getUnaryInterceptorTestSuite(t *testing.T, logger logger.LoggerInterface, opts ...logger_grpc.Option) *grpc_testing.InterceptorTestSuite {
	its := &grpc_testing.InterceptorTestSuite{
		ClientOpts: []grpc.DialOption{
			grpc.WithUnaryInterceptor(client_interceptor.UnaryInterceptor(logger, opts...)),
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
		assert.Equal(t, "client", eCtx["grpc_kind"].Value)
		assert.Equal(t, "Ping", eCtx["grpc_method"].Value)
		assert.Equal(t, "mwitkow.testproto.TestService", eCtx["grpc_service"].Value)
	}
}
