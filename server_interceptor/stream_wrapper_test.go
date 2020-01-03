package server_interceptor_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/gol4ng/logger"
	testing_logger "github.com/gol4ng/logger/testing"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"

	logger_grpc "github.com/gol4ng/logger-grpc"
	"github.com/gol4ng/logger-grpc/mocks"
	"github.com/gol4ng/logger-grpc/server_interceptor"
)

func TestStreamWrapper_SendHeader(t *testing.T) {
	mMock := metadata.MD{
		"first": []string{"test"},
	}
	myLogger := &testing_logger.Logger{}
	serverStreamMock := &mocks.ServerStream{}
	serverStreamMock.On("SendHeader", mMock).Return(nil)
	ctx, _ := context.WithTimeout(context.TODO(), 2*time.Second)

	serverStreamWrapper := server_interceptor.NewServerStreamWrapper(serverStreamMock, ctx, logger_grpc.EvaluateServerOpt([]logger_grpc.Option{}), myLogger, nil)
	err := serverStreamWrapper.SendHeader(mMock)
	assert.NoError(t, err)

	entries := myLogger.GetEntries()
	assert.Len(t, entries, 1)

	entry1 := entries[0]
	entry1Ctx := *entry1.Context
	assert.NotContains(t, entry1Ctx, "grpc_duration")
	assert.Equal(t, mMock, entry1Ctx["grpc_metadata"].Value)
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, `grpc server stream send header`, entry1.Message)
}

func TestStreamWrapper_SendHeader_WithError(t *testing.T) {
	mMock := metadata.MD{
		"first": []string{"test"},
	}
	errorMock := errors.New("my_fake_error")
	myLogger := &testing_logger.Logger{}
	serverStreamMock := &mocks.ServerStream{}
	serverStreamMock.On("SendHeader", mMock).Return(errorMock)
	ctx, _ := context.WithTimeout(context.TODO(), 2*time.Second)

	serverStreamWrapper := server_interceptor.NewServerStreamWrapper(serverStreamMock, ctx, logger_grpc.EvaluateServerOpt([]logger_grpc.Option{}), myLogger, nil)
	err := serverStreamWrapper.SendHeader(mMock)
	assert.EqualError(t, err, "my_fake_error")

	entries := myLogger.GetEntries()
	assert.Len(t, entries, 1)

	entry1 := entries[0]
	entry1Ctx := *entry1.Context
	assert.NotContains(t, entry1Ctx, "grpc_duration")
	assert.Equal(t, mMock, entry1Ctx["grpc_metadata"].Value)
	assert.Equal(t, logger.ErrorLevel, entry1.Level)
	assert.Equal(t, `grpc server stream send header error`, entry1.Message)
}

func TestStreamWrapper_SendMsg(t *testing.T) {
	mMock := "my_fake_message"
	myLogger := &testing_logger.Logger{}
	serverStreamMock := &mocks.ServerStream{}
	serverStreamMock.On("SendMsg", mMock).Return(nil)
	ctx, _ := context.WithTimeout(context.TODO(), 2*time.Second)

	serverStreamWrapper := server_interceptor.NewServerStreamWrapper(serverStreamMock, ctx, logger_grpc.EvaluateServerOpt([]logger_grpc.Option{}), myLogger, nil)
	err := serverStreamWrapper.SendMsg(mMock)
	assert.NoError(t, err)

	entries := myLogger.GetEntries()
	assert.Len(t, entries, 1)

	entry1 := entries[0]
	entry1Ctx := *entry1.Context
	assert.Contains(t, entry1Ctx, "grpc_duration")
	assert.Equal(t, mMock, entry1Ctx["grpc_send_data"].Value)
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, `grpc server stream send message`, entry1.Message)
}

func TestStreamWrapper_SendMsg_WithError(t *testing.T) {
	mMock := "my_fake_message"
	errorMock := errors.New("my_fake_error")
	myLogger := &testing_logger.Logger{}
	serverStreamMock := &mocks.ServerStream{}
	serverStreamMock.On("SendMsg", mMock).Return(errorMock)
	ctx, _ := context.WithTimeout(context.TODO(), 2*time.Second)

	serverStreamWrapper := server_interceptor.NewServerStreamWrapper(serverStreamMock, ctx, logger_grpc.EvaluateServerOpt([]logger_grpc.Option{}), myLogger, nil)
	err := serverStreamWrapper.SendMsg(mMock)
	assert.EqualError(t, err, "my_fake_error")

	entries := myLogger.GetEntries()
	assert.Len(t, entries, 1)

	entry1 := entries[0]
	entry1Ctx := *entry1.Context
	assert.Contains(t, entry1Ctx, "grpc_duration")
	assert.Equal(t, mMock, entry1Ctx["grpc_send_data"].Value)
	assert.Equal(t, "Unknown", entry1Ctx["grpc_code"].Value)
	assert.Equal(t, errorMock, entry1Ctx["grpc_error"].Value)
	assert.Equal(t, logger.ErrorLevel, entry1.Level)
	assert.Equal(t, `grpc server stream send error`, entry1.Message)
}

func TestStreamWrapper_RecvMsg(t *testing.T) {
	mMock := "my_fake_message"
	myLogger := &testing_logger.Logger{}
	serverStreamMock := &mocks.ServerStream{}
	serverStreamMock.On("RecvMsg", mMock).Return(nil)
	ctx, _ := context.WithTimeout(context.TODO(), 2*time.Second)

	serverStreamWrapper := server_interceptor.NewServerStreamWrapper(serverStreamMock, ctx, logger_grpc.EvaluateServerOpt([]logger_grpc.Option{}), myLogger, nil)
	err := serverStreamWrapper.RecvMsg(mMock)
	assert.NoError(t, err)

	entries := myLogger.GetEntries()
	assert.Len(t, entries, 1)

	entry1 := entries[0]
	entry1Ctx := *entry1.Context
	assert.Contains(t, entry1Ctx, "grpc_duration")
	assert.Equal(t, mMock, entry1Ctx["grpc_recv_data"].Value)
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, `grpc server stream receive message`, entry1.Message)
}

func TestStreamWrapper_RecvMsg_WithEOF(t *testing.T) {
	mMock := "my_fake_message"
	myLogger := &testing_logger.Logger{}
	serverStreamMock := &mocks.ServerStream{}
	serverStreamMock.On("RecvMsg", mMock).Return(io.EOF)
	ctx, _ := context.WithTimeout(context.TODO(), 2*time.Second)

	serverStreamWrapper := server_interceptor.NewServerStreamWrapper(serverStreamMock, ctx, logger_grpc.EvaluateServerOpt([]logger_grpc.Option{}), myLogger, nil)
	err := serverStreamWrapper.RecvMsg(mMock)
	assert.Equal(t, io.EOF, err)

	entries := myLogger.GetEntries()
	assert.Len(t, entries, 1)

	entry1 := entries[0]
	entry1Ctx := *entry1.Context
	assert.Contains(t, entry1Ctx, "grpc_duration")
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, `grpc server stream receive EOF`, entry1.Message)
}

func TestStreamWrapper_RecvMsg_WithError(t *testing.T) {
	mMock := "my_fake_message"
	errorMock := errors.New("my_fake_error")
	myLogger := &testing_logger.Logger{}
	serverStreamMock := &mocks.ServerStream{}
	serverStreamMock.On("RecvMsg", mMock).Return(errorMock)
	ctx, _ := context.WithTimeout(context.TODO(), 2*time.Second)

	serverStreamWrapper := server_interceptor.NewServerStreamWrapper(serverStreamMock, ctx, logger_grpc.EvaluateServerOpt([]logger_grpc.Option{}), myLogger, nil)
	err := serverStreamWrapper.RecvMsg(mMock)
	assert.EqualError(t, err, "my_fake_error")

	entries := myLogger.GetEntries()
	assert.Len(t, entries, 1)

	entry1 := entries[0]
	entry1Ctx := *entry1.Context
	assert.Contains(t, entry1Ctx, "grpc_duration")
	assert.Equal(t, mMock, entry1Ctx["grpc_recv_data"].Value)
	assert.Equal(t, "Unknown", entry1Ctx["grpc_code"].Value)
	assert.Equal(t, errorMock, entry1Ctx["grpc_error"].Value)
	assert.Equal(t, logger.ErrorLevel, entry1.Level)
	assert.Equal(t, `grpc server stream receive error`, entry1.Message)
}
