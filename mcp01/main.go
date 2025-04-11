package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	var transport string
	flag.StringVar(&transport, "t", "stdio", "Transport type (stdio or sse)")
	flag.StringVar(
		&transport,
		"transport",
		"stdio",
		"Transport type (stdio or sse)",
	)
	addr := flag.String("sse-address", "localhost:8081", "The host and port to start the sse server on")
	flag.Parse()
	fmt.Println(*addr)

	if err := run(transport, *addr); err != nil {
		panic(err)
	}
}

func run(transport, addr string) error {
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
	// s.AddTools(server.ServerTool{Tool: tool, Handler: helloHandler})

	// Debug information
	log.Printf("Registered tool: save_name")

	switch transport {
	case "stdio":
		srv := server.NewStdioServer(s)
		return srv.Listen(context.Background(), os.Stdin, os.Stdout)
	case "sse":
		// Create the SSE server with explicit debugging
		srv := server.NewSSEServer(s)

		log.Printf("SSE server listening on %s", addr)
		if err := srv.Start(addr); err != nil {
			return fmt.Errorf("server error: %v", err)
		}
		// This code is unreachable as Start() blocks until error
	default:
		return fmt.Errorf(
			"invalid transport type: %s. Must be 'stdio' or 'sse'",
			transport,
		)
	}
	return nil
}

func helloHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, ok := request.Params.Arguments["name"].(string)
	if !ok {
		return mcp.NewToolResultError("name must be a string"), nil
	}
	uuid4, _ := uuid.NewRandom()
	return mcp.NewToolResultText(fmt.Sprintf("name[%s] has saved as id: %s", name, uuid4)), nil
}
