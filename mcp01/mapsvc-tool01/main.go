package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	port     int
	myTracer trace.Tracer
)

func main() {
	flag.IntVar(&port, "p", 0, "Use SSE mode with assigned port")
	flag.Parse()
	// init tracer
	traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		panic(err)
	}
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("MyService"),
			semconv.ServiceVersion("v0.0.1"),
		),
	)
	if err != nil {
		panic(err)
	}
	traceProvider := traceSdk.NewTracerProvider(
		traceSdk.WithBatcher(traceExporter, traceSdk.WithBatchTimeout(2*time.Second)),
		traceSdk.WithResource(res),
	)
	defer func() { _ = traceProvider.Shutdown(context.Background()) }()
	otel.SetTracerProvider(traceProvider)
	// Create & start the tracer
	myTracer = traceProvider.Tracer("MyService")
	// run svc
	if err := run(port); err != nil {
		panic(err)
	}
}

func run(port int) error {
	// Create MCP server with explicit options
	s := server.NewMCPServer(
		"Demo 🚀",
		"1.0.0",
	)

	// Add tool with more explicit configuration
	tool := mcp.NewTool("save_name",
		mcp.WithDescription("Save user name in storage with uuid"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of user"),
		),
	)

	// Add tool handler
	s.AddTool(tool, helloHandler)

	// Debug information
	log.Printf("Registered tool: save_name")

	switch {
	case port == 0:
		srv := server.NewStdioServer(s)
		return srv.Listen(context.Background(), os.Stdin, os.Stdout)
	case port > 0:
		addr := fmt.Sprintf("localhost:%d", port)
		log.Printf("SSE server listening on %s\n", addr)
		srv := server.NewSSEServer(s)

		if err := srv.Start(addr); err != nil {
			return fmt.Errorf("server error: %v", err)
		}
	default:
		return fmt.Errorf("invalid port settings: %d", port)
	}
	return nil
}

func helloHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var span trace.Span
	_, span = myTracer.Start(ctx, "save_name")
	span.SetAttributes(attribute.String("environment", "testing"))
	defer span.End()
	name, ok := request.Params.Arguments["name"].(string)
	if !ok {
		return mcp.NewToolResultError("name must be a string"), nil
	}
	uuid4, _ := uuid.NewRandom()
	span.AddEvent("generated uuid4")
	return mcp.NewToolResultText(fmt.Sprintf("name[%s] has saved as id: %s", name, uuid4)), nil
}
