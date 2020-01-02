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

func TestStreamInterceptor(t *testing.T) {
	myLogger := &testing_logger.Logger{}

	its := &grpc_testing.InterceptorTestSuite{}
	its.Suite.SetT(t)
	its.SetupSuite()

	c := its.NewClient(
		grpc.WithStreamInterceptor(client_interceptor.StreamInterceptor(myLogger)),
	)

	resp, err := c.PingStream(its.SimpleCtx())
	assert.NoError(t, err)

	pingRequest := &pb_testproto.PingRequest{Value: "my_fake_ping_payload"}
	assert.NoError(t, resp.Send(pingRequest))
	pingResponse, err := resp.Recv()
	assert.NoError(t, err)

	assert.Equal(t, "my_fake_ping_payload", pingResponse.Value)
	assert.Equal(t, int32(0), pingResponse.Counter)

	entries := myLogger.GetEntries()
	assert.Len(t, entries, 4)

	for _, e := range entries {
		eCtx := *e.Context
		assert.Contains(t, eCtx, "grpc_start_time")
		assert.Contains(t, eCtx, "grpc_duration")
		assert.Contains(t, eCtx, "grpc_request_deadline")
		assert.Equal(t, "client", eCtx["grpc_kind"].Value)
		assert.Equal(t, "OK", eCtx["grpc_code"].Value)
		assert.Equal(t, "PingStream", eCtx["grpc_method"].Value)
		assert.Equal(t, "mwitkow.testproto.TestService", eCtx["grpc_service"].Value)
	}

	entry1 := entries[0]
	entry1Ctx := *entry1.Context
	assert.NotContains(t, entry1Ctx, "grpc_send_data")
	assert.NotContains(t, entry1Ctx, "grpc_recv_data")
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, "grpc client begin stream call /mwitkow.testproto.TestService/PingStream", entry1.Message)

	entry2 := entries[1]
	entry2Ctx := *entry2.Context
	assert.NotContains(t, entry2Ctx, "grpc_send_data")
	assert.NotContains(t, entry2Ctx, "grpc_recv_data")
	assert.Equal(t, logger.InfoLevel, entry2.Level)
	assert.Regexp(t, `grpc client stream call /mwitkow\.testproto\.TestService/PingStream \[code:OK, duration:.*]`, entry2.Message)

	entry3 := entries[2]
	entry3Ctx := *entry3.Context
	assert.Equal(t, pingRequest, entry3Ctx["grpc_send_data"].Value)
	assert.NotContains(t, entry3Ctx, "grpc_recv_data")
	assert.Equal(t, logger.DebugLevel, entry3.Level)
	assert.Regexp(t, `grpc client stream send message`, entry3.Message)

	entry4 := entries[3]
	entry4Ctx := *entry4.Context
	assert.NotContains(t, entry4Ctx, "grpc_send_data")
	assert.Equal(t, pingResponse, entry4Ctx["grpc_recv_data"].Value)
	assert.Equal(t, logger.DebugLevel, entry4.Level)
	assert.Regexp(t, `grpc client stream receive message`, entry4.Message)
}

func TestStreamInterceptoor_WithContext(t *testing.T) {
	myLogger := &testing_logger.Logger{}

	its := &grpc_testing.InterceptorTestSuite{}
	its.Suite.SetT(t)
	its.SetupSuite()

	c := its.NewClient(
		grpc.WithStreamInterceptor(client_interceptor.StreamInterceptor(myLogger, logger_grpc.WithLoggerContext(func(fullMethodName string) *logger.Context {
			return logger.NewContext().Add("base_context_key", "base_context_value")
		}))),
	)

	resp, err := c.PingStream(its.SimpleCtx())
	assert.NoError(t, err)

	pingRequest := &pb_testproto.PingRequest{Value: "my_fake_ping_payload"}
	assert.NoError(t, resp.Send(pingRequest))
	pingResponse, err := resp.Recv()
	assert.NoError(t, err)

	assert.Equal(t, "my_fake_ping_payload", pingResponse.Value)
	assert.Equal(t, int32(0), pingResponse.Counter)

	entries := myLogger.GetEntries()
	assert.Len(t, entries, 4)

	for _, e := range entries {
		eCtx := *e.Context
		assert.Equal(t, "base_context_value", eCtx["base_context_key"].Value)

		assert.Contains(t, eCtx, "grpc_start_time")
		assert.Contains(t, eCtx, "grpc_duration")
		assert.Contains(t, eCtx, "grpc_request_deadline")
		assert.Equal(t, "client", eCtx["grpc_kind"].Value)
		assert.Equal(t, "OK", eCtx["grpc_code"].Value)
		assert.Equal(t, "PingStream", eCtx["grpc_method"].Value)
		assert.Equal(t, "mwitkow.testproto.TestService", eCtx["grpc_service"].Value)
	}

	entry1 := entries[0]
	entry1Ctx := *entry1.Context
	assert.NotContains(t, entry1Ctx, "grpc_send_data")
	assert.NotContains(t, entry1Ctx, "grpc_recv_data")
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, "grpc client begin stream call /mwitkow.testproto.TestService/PingStream", entry1.Message)

	entry2 := entries[1]
	entry2Ctx := *entry2.Context
	assert.NotContains(t, entry2Ctx, "grpc_send_data")
	assert.NotContains(t, entry2Ctx, "grpc_recv_data")
	assert.Equal(t, logger.InfoLevel, entry2.Level)
	assert.Regexp(t, `grpc client stream call /mwitkow\.testproto\.TestService/PingStream \[code:OK, duration:.*]`, entry2.Message)

	entry3 := entries[2]
	entry3Ctx := *entry3.Context
	assert.Equal(t, pingRequest, entry3Ctx["grpc_send_data"].Value)
	assert.NotContains(t, entry3Ctx, "grpc_recv_data")
	assert.Equal(t, logger.DebugLevel, entry3.Level)
	assert.Regexp(t, `grpc client stream send message`, entry3.Message)

	entry4 := entries[3]
	entry4Ctx := *entry4.Context
	assert.NotContains(t, entry4Ctx, "grpc_send_data")
	assert.Equal(t, pingResponse, entry4Ctx["grpc_recv_data"].Value)
	assert.Equal(t, logger.DebugLevel, entry4.Level)
	assert.Regexp(t, `grpc client stream receive message`, entry4.Message)
}

func TestStreamInterceptoor_WithLevels(t *testing.T) {
	myLogger := &testing_logger.Logger{}

	its := &grpc_testing.InterceptorTestSuite{}
	its.Suite.SetT(t)
	its.SetupSuite()

	c := its.NewClient(
		grpc.WithStreamInterceptor(client_interceptor.StreamInterceptor(myLogger, logger_grpc.WithLevels(func(code codes.Code) logger.Level {
			return logger.EmergencyLevel
		}))),
	)

	resp, err := c.PingStream(its.SimpleCtx())
	assert.NoError(t, err)

	pingRequest := &pb_testproto.PingRequest{Value: "my_fake_ping_payload",}
	assert.NoError(t, resp.Send(pingRequest))
	pingResponse, err := resp.Recv()
	assert.NoError(t, err)

	assert.Equal(t, "my_fake_ping_payload", pingResponse.Value)
	assert.Equal(t, int32(0), pingResponse.Counter)

	entries := myLogger.GetEntries()
	assert.Len(t, entries, 4)

	for _, e := range entries {
		eCtx := *e.Context
		assert.Contains(t, eCtx, "grpc_start_time")
		assert.Contains(t, eCtx, "grpc_duration")
		assert.Contains(t, eCtx, "grpc_request_deadline")
		assert.Equal(t, "client", eCtx["grpc_kind"].Value)
		assert.Equal(t, "OK", eCtx["grpc_code"].Value)
		assert.Equal(t, "PingStream", eCtx["grpc_method"].Value)
		assert.Equal(t, "mwitkow.testproto.TestService", eCtx["grpc_service"].Value)
	}

	entry1 := entries[0]
	entry1Ctx := *entry1.Context
	assert.NotContains(t, entry1Ctx, "grpc_send_data")
	assert.NotContains(t, entry1Ctx, "grpc_recv_data")
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, "grpc client begin stream call /mwitkow.testproto.TestService/PingStream", entry1.Message)

	entry2 := entries[1]
	entry2Ctx := *entry2.Context
	assert.NotContains(t, entry2Ctx, "grpc_send_data")
	assert.NotContains(t, entry2Ctx, "grpc_recv_data")
	assert.Equal(t, logger.EmergencyLevel, entry2.Level)
	assert.Regexp(t, `grpc client stream call /mwitkow\.testproto\.TestService/PingStream \[code:OK, duration:.*]`, entry2.Message)

	entry3 := entries[2]
	entry3Ctx := *entry3.Context
	assert.Equal(t, pingRequest, entry3Ctx["grpc_send_data"].Value)
	assert.NotContains(t, entry3Ctx, "grpc_recv_data")
	assert.Equal(t, logger.DebugLevel, entry3.Level)
	assert.Regexp(t, `grpc client stream send message`, entry3.Message)

	entry4 := entries[3]
	entry4Ctx := *entry4.Context
	assert.NotContains(t, entry4Ctx, "grpc_send_data")
	assert.Equal(t, pingResponse, entry4Ctx["grpc_recv_data"].Value)
	assert.Equal(t, logger.DebugLevel, entry4.Level)
	assert.Regexp(t, `grpc client stream receive message`, entry4.Message)
}

func TestStreamInterceptor_WillPanic(t *testing.T) {
	myLogger := &testing_logger.Logger{}

	interceptor := client_interceptor.StreamInterceptor(myLogger)
	ctx, _ := context.WithTimeout(context.TODO(), 2*time.Second)
	streamerMock := func(innerCtx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		assert.Equal(t, ctx, innerCtx)
		panic("my_fake_panic_message")
	}
	assert.PanicsWithValue(t,
		"my_fake_panic_message",
		func() { _, _ = interceptor(ctx, &grpc.StreamDesc{}, &grpc.ClientConn{}, "/mwitkow.testproto.TestService/PingStream", streamerMock) },
	)

	entries := myLogger.GetEntries()
	assert.Len(t, entries, 2)

	for _, e := range entries {
		eCtx := *e.Context
		assert.NotContains(t, eCtx, "grpc_code")

		assert.Contains(t, eCtx, "grpc_start_time")
		assert.Contains(t, eCtx, "grpc_duration")
		assert.Contains(t, eCtx, "grpc_request_deadline")
		assert.Equal(t, "client", eCtx["grpc_kind"].Value)
		assert.Equal(t, "PingStream", eCtx["grpc_method"].Value)
		assert.Equal(t, "mwitkow.testproto.TestService", eCtx["grpc_service"].Value)
	}

	entry1 := entries[0]
	entry1Ctx := *entry1.Context
	assert.NotContains(t, entry1Ctx, "grpc_send_data")
	assert.NotContains(t, entry1Ctx, "grpc_recv_data")
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, "grpc client begin stream call /mwitkow.testproto.TestService/PingStream", entry1.Message)

	entry2 := entries[1]
	entry2Ctx := *entry2.Context
	assert.NotContains(t, entry2Ctx, "grpc_send_data")
	assert.NotContains(t, entry2Ctx, "grpc_recv_data")
	assert.Equal(t, "my_fake_panic_message", entry2Ctx["grpc_panic"].Value)
	assert.Equal(t, logger.CriticalLevel, entry2.Level)
	assert.Regexp(t, `grpc client stream panic /mwitkow\.testproto\.TestService/PingStream \[duration:.*]`, entry2.Message)
}

func TestStreamInterceptor_WithError(t *testing.T) {
	myLogger := &testing_logger.Logger{}

	interceptor := client_interceptor.StreamInterceptor(myLogger)
	ctx, _ := context.WithTimeout(context.TODO(), 2*time.Second)
	streamerMock := func(innerCtx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		assert.Equal(t, ctx, innerCtx)
		return nil, errors.New("my_fake_error_message")
	}

	client, err := interceptor(ctx, &grpc.StreamDesc{}, &grpc.ClientConn{}, "/mwitkow.testproto.TestService/PingStream", streamerMock)
	assert.Nil(t, client)
	assert.EqualError(t, err, "my_fake_error_message")

	entries := myLogger.GetEntries()
	assert.Len(t, entries, 2)

	for _, e := range entries {
		eCtx := *e.Context

		assert.Contains(t, eCtx, "grpc_start_time")
		assert.Contains(t, eCtx, "grpc_duration")
		assert.Contains(t, eCtx, "grpc_request_deadline")
		assert.Equal(t, "client", eCtx["grpc_kind"].Value)
		assert.Equal(t, "Unknown", eCtx["grpc_code"].Value)
		assert.Equal(t, "PingStream", eCtx["grpc_method"].Value)
		assert.Equal(t, "mwitkow.testproto.TestService", eCtx["grpc_service"].Value)
	}

	entry1 := entries[0]
	entry1Ctx := *entry1.Context
	assert.NotContains(t, entry1Ctx, "grpc_send_data")
	assert.NotContains(t, entry1Ctx, "grpc_recv_data")
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, "grpc client begin stream call /mwitkow.testproto.TestService/PingStream", entry1.Message)

	entry2 := entries[1]
	entry2Ctx := *entry2.Context
	assert.NotContains(t, entry2Ctx, "grpc_send_data")
	assert.NotContains(t, entry2Ctx, "grpc_recv_data")
	assert.EqualError(t,  entry2Ctx["grpc_error"].Value.(error), "my_fake_error_message")
	assert.Equal(t, logger.ErrorLevel, entry2.Level)
	assert.Regexp(t, `grpc client stream call /mwitkow\.testproto\.TestService/PingStream \[code:Unknown, duration:.*]`, entry2.Message)
}
