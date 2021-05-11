package otelgrpc

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto" // nolint:staticcheck
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc"
	grpc_codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Metric names reported by the MetricsReporter.
const (
	ServerMsgReceivedCounter  = "grpc.server.msg.received.total"
	ServerMsgReceivedBytes    = "grpc.server.msg.received.bytes"
	ServerMsgSentCounter      = "grpc.server.msg.sent.total"
	ServerMsgSentBytes        = "grpc.server.msg.sent.bytes"
	ServerLatencyMilliseconds = "grpc.server.latency.milliseconds"
)

// MetricsReporter exposes methods to record metrics for gRPC servers.
type MetricsReporter struct {
	serverMsgReceivedCounter  metric.Int64UpDownCounter
	serverMsgReceivedBytes    metric.Int64ValueRecorder
	serverMsgSentCounter      metric.Int64UpDownCounter
	serverMsgSentBytes        metric.Int64ValueRecorder
	serverLatencyMilliseconds metric.Int64ValueRecorder
}

// NewMetricsReporter uses the MeterProvider to register metrics and
// collect them in a MetricsReporter instance, that can be used with grpc.NewServer
// for instrumenting communication.
func NewMetricsReporter(meterProvider metric.MeterProvider) (*MetricsReporter, error) {
	meter := meterProvider.Meter(instrumentationName)

	withMessage := func(err error, metric string) error {
		return fmt.Errorf("otelgrpc: failed to register metric '%s': %w", metric, err)
	}

	serverMsgReceivedCounter, err := meter.NewInt64UpDownCounter(
		ServerMsgReceivedCounter,
		metric.WithDescription("Total number of messages received on the server"),
	)

	if err != nil {
		return nil, withMessage(err, ServerMsgReceivedCounter)
	}

	serverMsgReceivedBytes, err := meter.NewInt64ValueRecorder(
		ServerMsgReceivedBytes,
		metric.WithDescription("Number of bytes received by the server"),
	)

	if err != nil {
		return nil, withMessage(err, ServerMsgReceivedBytes)
	}

	serverMsgSentCounter, err := meter.NewInt64UpDownCounter(
		ServerMsgSentCounter,
		metric.WithDescription("Total number of messages sent by the server"),
	)

	if err != nil {
		return nil, withMessage(err, ServerMsgSentCounter)
	}

	serverMsgSentBytes, err := meter.NewInt64ValueRecorder(
		ServerMsgSentBytes,
		metric.WithDescription("Number of bytes sent by the server"),
	)

	if err != nil {
		return nil, withMessage(err, ServerMsgSentBytes)
	}

	serverLatencyMilliseconds, err := meter.NewInt64ValueRecorder(
		ServerLatencyMilliseconds,
		metric.WithDescription("Latency recorded by the server to handle a gRPC request"),
	)

	if err != nil {
		return nil, withMessage(err, ServerLatencyMilliseconds)
	}

	return &MetricsReporter{
		serverMsgReceivedCounter:  serverMsgReceivedCounter,
		serverMsgReceivedBytes:    serverMsgReceivedBytes,
		serverMsgSentCounter:      serverMsgSentCounter,
		serverMsgSentBytes:        serverMsgSentBytes,
		serverLatencyMilliseconds: serverLatencyMilliseconds,
	}, nil
}

// UnaryServerInterceptor returns a grpc.UnaryServerInterceptor suitable
// for use in a grpc.NewServer call.
func (mr *MetricsReporter) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// These attributes contain 'rpc.method' and 'rpc.service.
		_, attributes := parseFullMethod(info.FullMethod)

		mr.serverMsgReceivedCounter.Add(ctx, 1, attributes...)

		if p, ok := req.(proto.Message); ok {
			mr.serverMsgReceivedBytes.Record(ctx, int64(proto.Size(p)), attributes...)
		}

		code := grpc_codes.OK
		resp, err := handler(ctx, req)
		latency := time.Since(start)

		if err != nil {
			status, _ := status.FromError(err)
			code = status.Code()
		}

		attributes = append(attributes, statusCodeAttr(code))

		mr.serverMsgSentCounter.Add(ctx, 1, attributes...)
		mr.serverLatencyMilliseconds.Record(ctx, latency.Milliseconds(), attributes...)

		if p, ok := resp.(proto.Message); ok {
			mr.serverMsgSentBytes.Record(ctx, int64(proto.Size(p)), attributes...)
		}

		return resp, err
	}
}
