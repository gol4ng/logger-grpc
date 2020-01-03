# logger-grpc

[![Build Status](https://travis-ci.org/gol4ng/logger-grpc.svg?branch=master)](https://travis-ci.org/gol4ng/logger-grpc)
[![Maintainability](https://api.codeclimate.com/v1/badges/5dfa47faa129518b9b9d/maintainability)](https://codeclimate.com/github/gol4ng/logger-grpc/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/5dfa47faa129518b9b9d/test_coverage)](https://codeclimate.com/github/gol4ng/logger-grpc/test_coverage)
[![Go Report Card](https://goreportcard.com/badge/github.com/gol4ng/logger-grpc?)](https://goreportcard.com/report/github.com/gol4ng/logger-grpc?)
[![GoDoc](https://godoc.org/github.com/gol4ng/logger-grpc?status.svg)](https://godoc.org/github.com/gol4ng/logger-grpc)

Gol4ng logger sub package for logging grpc
Related package [gol4ng/logger](https://github.com/gol4ng/logger)  

## Installation

`go get -u github.com/gol4ng/logger-grpc`

### Client interceptor 

Log you're `grpc.Client` request

```
<debug> grpc client begin stream call /mwitkow.testproto.TestService/PingStream {"grpc_method":"PingStream","grpc_start_time":"2019-12-27T11:26:56+01:00","grpc_request_deadline":"2019-12-27T11:26:58+01:00","grpc_kind":"client","grpc_service":"mwitkow.testproto.TestService"}
<info> grpc client stream call /mwitkow.testproto.TestService/PingStream [code:OK, duration:99.615µs] {"grpc_duration":0.000099615,"grpc_code":"OK","grpc_service":"mwitkow.testproto.TestService","grpc_method":"PingStream","grpc_start_time":"2019-12-27T11:26:56+01:00","grpc_request_deadline":"2019-12-27T11:26:58+01:00","grpc_kind":"client"}
<debug> grpc client stream send message {"grpc_kind":"client","grpc_duration":0.0000594,"grpc_code":"OK","grpc_service":"mwitkow.testproto.TestService","grpc_method":"PingStream","grpc_start_time":"2019-12-27T11:26:56+01:00","grpc_send_data":{"value":"my_fake_ping_payload"},"grpc_request_deadline":"2019-12-27T11:26:58+01:00"}
<debug> grpc client stream receive message {"grpc_recv_data":{"Value":"my_fake_ping_payload"},"grpc_start_time":"2019-12-27T11:26:56+01:00","grpc_request_deadline":"2019-12-27T11:26:58+01:00","grpc_kind":"client","grpc_duration":0.000409316,"grpc_code":"OK","grpc_service":"mwitkow.testproto.TestService","grpc_method":"PingStream"}
```

```go
package main

import (
	"context"
	"net"
	"os"
	"time"

	"github.com/gol4ng/logger"
	logger_grpc "github.com/gol4ng/logger-grpc"
	"github.com/gol4ng/logger-grpc/client_interceptor"
	"github.com/gol4ng/logger/formatter"
	"github.com/gol4ng/logger/handler"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"google.golang.org/grpc"
)

func main() {
	myLogger := logger.NewLogger(
		handler.Stream(os.Stdout, formatter.NewDefaultFormatter()),
	)

	serverListener, _ := net.Listen("tcp", "127.0.0.1:0")
	serverAddr := serverListener.Addr().String()
	server := grpc.NewServer()
	pb_testproto.RegisterTestServiceServer(server, &pb_testproto.UnimplementedTestServiceServer{})

	go func() {
		server.Serve(serverListener)
	}()

	clientConn, _ := grpc.Dial(
		serverAddr,
		grpc.WithInsecure(),
		grpc.WithStreamInterceptor(client_interceptor.StreamInterceptor(myLogger, logger_grpc.WithLoggerContext(func(fullMethodName string) *logger.Context {
			return logger.NewContext().Add("base_context_key", "base_context_value")
		}))),
		grpc.WithUnaryInterceptor(client_interceptor.UnaryInterceptor(myLogger, logger_grpc.WithLoggerContext(func(fullMethodName string) *logger.Context {
			return logger.NewContext().Add("base_context_key", "base_context_value")
		}))),
	)
	client := pb_testproto.NewTestServiceClient(clientConn)

	ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
	defer cancel()
	pingRequest := &pb_testproto.PingRequest{Value: "my_fake_ping_payload"}
	client.Ping(ctx, pingRequest)

    //<error> grpc client unary call /mwitkow.testproto.TestService/Ping [code:Unimplemented, duration:73.078µs]
}
```

### Server Interceptor

Log you're incoming grpc server request

```
<debug> grpc server begin stream call /mwitkow.testproto.TestService/PingStream {"grpc_service":"mwitkow.testproto.TestService","grpc_method":"PingStream","grpc_start_time":"2019-12-27T11:28:30+01:00","grpc_request_deadline":"2019-12-27T11:28:32+01:00","grpc_kind":"server"}
<debug> grpc server stream receive message {"grpc_duration":0.000062933,"grpc_recv_data":{"value":"my_fake_ping_payload"},"grpc_request_deadline":"2019-12-27T11:28:32+01:00","grpc_kind":"server","grpc_service":"mwitkow.testproto.TestService","grpc_method":"PingStream","grpc_start_time":"2019-12-27T11:28:30+01:00"}
<debug> grpc server stream send message {"grpc_service":"mwitkow.testproto.TestService","grpc_method":"PingStream","grpc_start_time":"2019-12-27T11:28:30+01:00","grpc_request_deadline":"2019-12-27T11:28:32+01:00","grpc_kind":"server","grpc_send_data":{"Value":"my_fake_ping_payload"},"grpc_duration":0.000019383}
<debug> grpc server stream receive EOF {"grpc_start_time":"2019-12-27T11:28:30+01:00","grpc_request_deadline":"2019-12-27T11:28:32+01:00","grpc_kind":"server","grpc_duration":0.00016474,"grpc_service":"mwitkow.testproto.TestService","grpc_method":"PingStream"}
<info> grpc server stream call /mwitkow.testproto.TestService/PingStream [code:OK, duration:425.384µs] {"grpc_kind":"server","grpc_duration":0.000425384,"grpc_code":"OK","grpc_service":"mwitkow.testproto.TestService","grpc_method":"PingStream","grpc_start_time":"2019-12-27T11:28:30+01:00","grpc_request_deadline":"2019-12-27T11:28:32+01:00"}
```

```go
package main

import (
	"context"
	"net"
	"os"
	"time"

	"github.com/gol4ng/logger"
	logger_grpc "github.com/gol4ng/logger-grpc"
	"github.com/gol4ng/logger-grpc/server_interceptor"
	"github.com/gol4ng/logger/formatter"
	"github.com/gol4ng/logger/handler"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"google.golang.org/grpc"
)

func main() {
	myLogger := logger.NewLogger(
		handler.Stream(os.Stdout, formatter.NewDefaultFormatter()),
	)

	serverListener, _ := net.Listen("tcp", "127.0.0.1:0")
	serverAddr := serverListener.Addr().String()
	server := grpc.NewServer(
		grpc.StreamInterceptor(server_interceptor.StreamInterceptor(myLogger, logger_grpc.WithLoggerContext(func(fullMethodName string) *logger.Context {
			return logger.NewContext().Add("base_context_key", "base_context_value")
		}))),
		grpc.UnaryInterceptor(server_interceptor.UnaryInterceptor(myLogger, logger_grpc.WithLoggerContext(func(fullMethodName string) *logger.Context {
			return logger.NewContext().Add("base_context_key", "base_context_value")
		}))),

	)
	pb_testproto.RegisterTestServiceServer(server, &pb_testproto.UnimplementedTestServiceServer{})

	go func() {
		server.Serve(serverListener)
	}()

	clientConn, _ := grpc.Dial(serverAddr, grpc.WithInsecure())
	client := pb_testproto.NewTestServiceClient(clientConn)

	ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
	defer cancel()
	pingRequest := &pb_testproto.PingRequest{Value: "my_fake_ping_payload"}
	client.Ping(ctx, pingRequest)

    //<error> grpc server unary call /mwitkow.testproto.TestService/Ping [code:Unimplemented, duration:73.078µs]
}
```
