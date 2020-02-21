package client_interceptor_test

import (
	"errors"
	"io"
	"testing"

	"github.com/gol4ng/logger"
	logger_grpc "github.com/gol4ng/logger-grpc"
	"github.com/gol4ng/logger-grpc/client_interceptor"
	"github.com/gol4ng/logger-grpc/mocks"
	testing_logger "github.com/gol4ng/logger/testing"
	"github.com/stretchr/testify/assert"
)

func TestStreamWrapper_SendMsg(t *testing.T) {
	mMock := "my_fake_message"
	myLogger, store := testing_logger.NewLogger()
	clientStreamMock := &mocks.ClientStream{}
	clientStreamMock.On("SendMsg", mMock).Return(nil)

	clientStreamWrapper := client_interceptor.NewClientStreamWrapper(clientStreamMock, logger_grpc.EvaluateClientOpt([]logger_grpc.Option{}), myLogger, nil)
	err := clientStreamWrapper.SendMsg(mMock)
	assert.NoError(t, err)

	entries := store.GetEntries()
	assert.Len(t, entries, 1)

	entry1 := entries[0]
	entry1Ctx := *entry1.Context
	assert.Contains(t, entry1Ctx, "grpc_duration")
	assert.Equal(t, mMock, entry1Ctx["grpc_send_data"].Value)
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, `grpc client stream send message`, entry1.Message)
}

func TestStreamWrapper_SendMsg_WithError(t *testing.T) {
	mMock := "my_fake_message"
	errorMock := errors.New("my_fake_error")
	myLogger, store := testing_logger.NewLogger()
	clientStreamMock := &mocks.ClientStream{}
	clientStreamMock.On("SendMsg", mMock).Return(errorMock)

	clientStreamWrapper := client_interceptor.NewClientStreamWrapper(clientStreamMock, logger_grpc.EvaluateClientOpt([]logger_grpc.Option{}), myLogger, nil)
	err := clientStreamWrapper.SendMsg(mMock)
	assert.EqualError(t, err, "my_fake_error")

	entries := store.GetEntries()
	assert.Len(t, entries, 1)

	entry1 := entries[0]
	entry1Ctx := *entry1.Context
	assert.Contains(t, entry1Ctx, "grpc_duration")
	assert.Equal(t, mMock, entry1Ctx["grpc_send_data"].Value)
	assert.Equal(t, "Unknown", entry1Ctx["grpc_code"].Value)
	assert.Equal(t, errorMock, entry1Ctx["grpc_error"].Value)
	assert.Equal(t, logger.ErrorLevel, entry1.Level)
	assert.Equal(t, `grpc client stream send error`, entry1.Message)
}

func TestStreamWrapper_RecvMsg(t *testing.T) {
	mMock := "my_fake_message"
	myLogger, store := testing_logger.NewLogger()
	clientStreamMock := &mocks.ClientStream{}
	clientStreamMock.On("RecvMsg", mMock).Return(nil)

	clientStreamWrapper := client_interceptor.NewClientStreamWrapper(clientStreamMock, logger_grpc.EvaluateClientOpt([]logger_grpc.Option{}), myLogger, nil)
	err := clientStreamWrapper.RecvMsg(mMock)
	assert.NoError(t, err)

	entries := store.GetEntries()
	assert.Len(t, entries, 1)

	entry1 := entries[0]
	entry1Ctx := *entry1.Context
	assert.Contains(t, entry1Ctx, "grpc_duration")
	assert.Equal(t, mMock, entry1Ctx["grpc_recv_data"].Value)
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, `grpc client stream receive message`, entry1.Message)
}

func TestStreamWrapper_RecvMsg_WithEOF(t *testing.T) {
	mMock := "my_fake_message"
	myLogger, store := testing_logger.NewLogger()
	clientStreamMock := &mocks.ClientStream{}
	clientStreamMock.On("RecvMsg", mMock).Return(io.EOF)

	clientStreamWrapper := client_interceptor.NewClientStreamWrapper(clientStreamMock, logger_grpc.EvaluateClientOpt([]logger_grpc.Option{}), myLogger, nil)
	err := clientStreamWrapper.RecvMsg(mMock)
	assert.Equal(t, io.EOF, err)

	entries := store.GetEntries()
	assert.Len(t, entries, 1)

	entry1 := entries[0]
	entry1Ctx := *entry1.Context
	assert.Contains(t, entry1Ctx, "grpc_duration")
	assert.Equal(t, logger.DebugLevel, entry1.Level)
	assert.Equal(t, `grpc client stream receive EOF`, entry1.Message)
}

func TestStreamWrapper_RecvMsg_WithError(t *testing.T) {
	mMock := "my_fake_message"
	errorMock := errors.New("my_fake_error")
	myLogger, store := testing_logger.NewLogger()
	clientStreamMock := &mocks.ClientStream{}
	clientStreamMock.On("RecvMsg", mMock).Return(errorMock)

	clientStreamWrapper := client_interceptor.NewClientStreamWrapper(clientStreamMock, logger_grpc.EvaluateClientOpt([]logger_grpc.Option{}), myLogger, nil)
	err := clientStreamWrapper.RecvMsg(mMock)
	assert.EqualError(t, err, "my_fake_error")

	entries := store.GetEntries()
	assert.Len(t, entries, 1)

	entry1 := entries[0]
	entry1Ctx := *entry1.Context
	assert.Contains(t, entry1Ctx, "grpc_duration")
	assert.Equal(t, mMock, entry1Ctx["grpc_recv_data"].Value)
	assert.Equal(t, "Unknown", entry1Ctx["grpc_code"].Value)
	assert.Equal(t, errorMock, entry1Ctx["grpc_error"].Value)
	assert.Equal(t, logger.ErrorLevel, entry1.Level)
	assert.Equal(t, `grpc client stream receive error`, entry1.Message)
}
