package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	userv1 "github.com/wyubin/ex-mcp/api-grpc/src/gen/user/v1"
	"github.com/wyubin/ex-mcp/api-grpc/src/svc"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

func main() {
	// 初始化 context
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// create svc
	userSvc := svc.NewUserService()

	// 建立 gRPC server
	grpcServer := grpc.NewServer()
	userv1.RegisterUserServiceServer(grpcServer, userSvc)

	// 啟動 gRPC server（port: 50051）
	go func() {
		listener, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		log.Println("🚀 gRPC server listening on :50051")
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// 建立 gRPC-Gateway mux
	mux := runtime.NewServeMux()

	// 使用 RegisterHandlerServer（避免建立 grpc client, 內部會建立 grpc server）
	err := userv1.RegisterUserServiceHandlerServer(ctx, mux, userSvc)
	if err != nil {
		log.Fatalf("failed to register gRPC-Gateway handler: %v", err)
	}

	// 如果是要直接call 另一個 runtime 的 grpc, 要用 RegisterUserServiceHandlerFromEndpoint

	// 包一層 /api router
	httpMux := http.NewServeMux()
	// httpMux.Handle("/api/", http.StripPrefix("/api", mux))
	httpMux.Handle("/", http.StripPrefix("/api", mux))

	// add static
	wd, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	fmt.Printf("current pwd: %s\n", wd)
	fs := http.FileServer(http.Dir(filepath.Join(wd, "static")))
	httpMux.Handle("/static/", http.StripPrefix("/static/", fs))

	// 啟動 HTTP Gateway server（port: 8080）
	log.Println("🌐 HTTP Gateway listening on :8080")
	if err := http.ListenAndServe(":8080", httpMux); err != nil {
		log.Fatalf("failed to serve HTTP: %v", err)
	}
}
