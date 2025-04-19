package main

import (
	"context"

	"github.com/mark3labs/mcp-go/server"
	"github.com/wyubin/ex-mcp/mcp01/pkg/oteltracer"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
)

type Server struct {
	server         *server.MCPServer
	tracerProvider *trace.TracerProvider
}

func (s *Server) Init() error {
	var err error
	// Create MCP server with explicit options
	s.server = server.NewMCPServer(
		"Demo 🚀",
		"1.0.0",
	)
	// setup tracer
	traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return err
	}
	traceProvider, err := oteltracer.GetProvider("MyService", "v0.0.1", traceExporter)
	if err != nil {
		return err
	}
	s.tracerProvider = traceProvider
	// add tools
	tracer := traceProvider.Tracer("MyService")
	mw := oteltracer.McpMWTracer(tracer)
	s.server.AddTool(helloTool("save_name"), mw(helloHandler))

	// add prompt
	return nil
}

// defer shutdown
func (s *Server) CleanUp() {
	if s.tracerProvider != nil {
		s.tracerProvider.Shutdown(context.Background())
	}
}

func (s *Server) MCPServer() *server.MCPServer {
	return s.server
}

// setup server spec
func NewServer() *Server {
	return &Server{}
}
