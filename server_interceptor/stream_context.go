package server_interceptor

import (
	"context"

	"google.golang.org/grpc"
)

type StreamWithContext struct {
	grpc.ServerStream
	context       context.Context
}

func (s *StreamWithContext) Context() context.Context {
	return s.context
}

func newServerStreamWithContext(stream grpc.ServerStream, context context.Context) *StreamWithContext {
	return &StreamWithContext{
		ServerStream:  stream,
		context:       context,
	}
}
