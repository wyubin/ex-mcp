package main

import (
	"context"
	"log"
	"net"
	"net/http"

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

	// 建立 gRPC server
	grpcServer := grpc.NewServer()
	userSvc := svc.NewUserService()
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

	// 使用 RegisterHandlerServer（避免建立 grpc client）
	err := userv1.RegisterUserServiceHandlerServer(ctx, mux, userSvc)
	if err != nil {
		log.Fatalf("failed to register gRPC-Gateway handler: %v", err)
	}

	// 啟動 HTTP Gateway server（port: 8080）
	log.Println("🌐 HTTP Gateway listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("failed to serve HTTP: %v", err)
	}
}
