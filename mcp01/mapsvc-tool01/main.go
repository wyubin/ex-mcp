package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/server"
)

var (
	port int
)

func main() {
	flag.IntVar(&port, "p", 0, "Use SSE mode with assigned port")
	flag.Parse()
	s := NewServer()
	if err := s.Init(); err != nil {
		panic(fmt.Errorf("server init error: %v", err))
	}
	// defer shutdown
	defer s.CleanUp()
	// run svc
	if err := run(s.MCPServer()); err != nil {
		panic(err)
	}
}

func run(serv *server.MCPServer) error {
	switch {
	case port == 0:
		srv := server.NewStdioServer(serv)
		return srv.Listen(context.Background(), os.Stdin, os.Stdout)
	case port > 0:
		addr := fmt.Sprintf("localhost:%d", port)
		log.Printf("SSE server listening on %s\n", addr)
		srv := server.NewSSEServer(serv)

		if err := srv.Start(addr); err != nil {
			return fmt.Errorf("server error: %v", err)
		}
	default:
		return fmt.Errorf("invalid port settings: %d", port)
	}
	return nil
}
