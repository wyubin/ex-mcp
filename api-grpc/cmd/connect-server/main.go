package main

import (
	"log"
	"net/http"

	"github.com/wyubin/ex-mcp/api-grpc/src/gen/user/v1/userconnect"
)

func main() {
	// create svc
	// userSvc := svc.NewUserService()

	// 建立 gRPC-Gateway mux
	mux := http.NewServeMux()

	// 會需要另外撰寫userSvc for connect-go 的 handler, 目前不推薦
	path, handler := userconnect.NewUserServiceHandler(nil)

	mux.Handle(path, handler)

	// 啟動 connect-go server（port: 8080）
	log.Println("🌐 connect-go listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("failed to serve HTTP: %v", err)
	}
}
