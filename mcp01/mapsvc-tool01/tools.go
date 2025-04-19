package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/wyubin/ex-mcp/mcp01/pkg/oteltracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func helloTool(name string) mcp.Tool {
	return mcp.NewTool(name,
		mcp.WithDescription("Save user name in storage with uuid"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of user"),
		),
	)
}

func helloHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tracer := oteltracer.FromContext(ctx)
	var span trace.Span
	if tracer != nil {
		_, span = tracer.Start(ctx, "save_name")
		span.SetAttributes(attribute.String("environment", "testing"))
		defer span.End()
	}

	name, ok := request.Params.Arguments["name"].(string)
	if !ok {
		return mcp.NewToolResultError("name must be a string"), nil
	}
	uuid4, _ := uuid.NewRandom()
	span.AddEvent("generated uuid4") // do nothing if span is nil

	return mcp.NewToolResultText(fmt.Sprintf("name[%s] has saved as id: %s", name, uuid4)), nil
}
