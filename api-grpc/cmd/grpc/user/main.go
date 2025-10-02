package main

import (
	"log"
	"net"

	userv1 "github.com/wyubin/ex-mcp/api-grpc/src/gen/user/v1"
	"github.com/wyubin/ex-mcp/api-grpc/src/svc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// create svc
	userSvc := svc.NewUserService()

	// 建立 gRPC server
	grpcServer := grpc.NewServer()
	userv1.RegisterUserServiceServer(grpcServer, userSvc)
	reflection.Register(grpcServer)

	// 啟動 gRPC server（port: 50051）
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("🚀 gRPC server listening on :50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}
