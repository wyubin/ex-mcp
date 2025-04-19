package oteltracer

import (
	"context"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.opentelemetry.io/otel/sdk/resource"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type tracerContextKey struct{}

type McpMW func(next server.ToolHandlerFunc) server.ToolHandlerFunc

func ChainMcpMW(mws ...McpMW) McpMW {
	if len(mws) == 0 {
		return func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
			return next
		}
	}
	outer, others := mws[0], mws[1:]
	return func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
		for i := len(others) - 1; i >= 0; i-- { // reverse
			next = others[i](next)
		}
		return outer(next)
	}
}

// GetProvider with name and version
func GetProvider(name, ver string, exporter traceSdk.SpanExporter) (*traceSdk.TracerProvider, error) {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(name),
			semconv.ServiceVersion(ver),
		),
	)
	if err != nil {
		return nil, err
	}
	traceProvider := traceSdk.NewTracerProvider(
		traceSdk.WithBatcher(exporter, traceSdk.WithBatchTimeout(2*time.Second)),
		traceSdk.WithResource(res),
	)
	return traceProvider, err
}

// WithTracer, set tracer in context
func WithTracer(ctx context.Context, tracer trace.Tracer) context.Context {
	return context.WithValue(ctx, tracerContextKey{}, tracer)
}

// FromContext, get tracer from context
func FromContext(ctx context.Context) trace.Tracer {
	if tracer, ok := ctx.Value(tracerContextKey{}).(trace.Tracer); ok {
		return tracer
	}
	return nil
}

// McpMWTracer, middleware to set tracer in mcp handler with context
func McpMWTracer(tracer trace.Tracer) McpMW {
	return func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			ctx = WithTracer(ctx, tracer)
			return next(ctx, request)
		}
	}
}
