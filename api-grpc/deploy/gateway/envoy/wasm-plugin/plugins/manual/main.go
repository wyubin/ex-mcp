package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm"
	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm/types"

	"google.golang.org/protobuf/proto"

	userpb "main.go/proto"
)

func main() {
	proxywasm.SetVMContext(&vmContext{})
}

// --- VM Context ---
type vmContext struct {
	types.DefaultVMContext
}

func (*vmContext) NewHttpContext(contextID uint32) types.HttpContext {
	return &httpContext{contextID: contextID}
}

// --- HTTP Context ---
type httpContext struct {
	types.DefaultHttpContext
	contextID uint32
	method    string
	path      string
}

func (ctx *httpContext) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {
	method, err := proxywasm.GetHttpRequestHeader(":method")
	if err != nil {
		method = ""
	}
	path, err := proxywasm.GetHttpRequestHeader(":path")
	if err != nil {
		path = ""
	}
	ctx.method = method
	ctx.path = path

	if endOfStream {
		return types.ActionContinue
	}
	return types.ActionPause
}

func (ctx *httpContext) OnHttpRequestBody(bodySize int, endOfStream bool) types.Action {
	if !endOfStream {
		return types.ActionPause
	}

	// 取得完整 body
	body, err := proxywasm.GetHttpRequestBody(0, 0)
	if err != nil {
		proxywasm.LogCriticalf("failed to get body: %v", err)
		proxywasm.SendHttpResponse(500, nil, []byte("failed to read body"), 0)
		return types.ActionPause
	}

	var grpcData []byte
	var grpcPath string

	// --- 動態 REST → gRPC 映射 ---
	switch {
	// POST /v1/users → CreateUser
	case ctx.method == "POST" && ctx.path == "/v1/users":
		var req userpb.CreateUserRequest
		if err := json.Unmarshal(body, &req); err != nil {
			proxywasm.SendHttpResponse(400, nil, []byte("invalid JSON for CreateUser"), 0)
			return types.ActionPause
		}
		grpcData, err = proto.Marshal(&req)
		if err != nil {
			proxywasm.SendHttpResponse(500, nil, []byte("failed to marshal gRPC request"), 0)
			return types.ActionPause
		}
		grpcPath = "/user.UserService/CreateUser"

	// GET /v1/users/{id} → GetUser
	case ctx.method == "GET" && strings.HasPrefix(ctx.path, "/v1/users/"):
		id := strings.TrimPrefix(ctx.path, "/v1/users/")
		req := &userpb.GetUserRequest{Id: id}
		grpcData, err = proto.Marshal(req)
		if err != nil {
			proxywasm.SendHttpResponse(500, nil, []byte("failed to marshal gRPC request"), 0)
			return types.ActionPause
		}
		grpcPath = "/user.UserService/GetUser"

	// GET /v1/users → ListUsers
	case ctx.method == "GET" && ctx.path == "/v1/users":
		req := &userpb.ListUsersRequest{}
		grpcData, err = proto.Marshal(req)
		if err != nil {
			proxywasm.SendHttpResponse(500, nil, []byte("failed to marshal gRPC request"), 0)
			return types.ActionPause
		}
		grpcPath = "/user.UserService/ListUsers"

	// DELETE /v1/users/{id} → DeleteUser
	case ctx.method == "DELETE" && strings.HasPrefix(ctx.path, "/v1/users/"):
		id := strings.TrimPrefix(ctx.path, "/v1/users/")
		req := &userpb.DeleteUserRequest{Id: id}
		grpcData, err = proto.Marshal(req)
		if err != nil {
			proxywasm.SendHttpResponse(500, nil, []byte("failed to marshal gRPC request"), 0)
			return types.ActionPause
		}
		grpcPath = "/user.UserService/DeleteUser"

	default:
		// 動態解析任意 /{service}/{method} pattern
		parts := strings.Split(strings.Trim(ctx.path, "/"), "/")
		if len(parts) >= 2 {
			service := parts[len(parts)-2]
			method := parts[len(parts)-1]
			grpcPath = fmt.Sprintf("/%s/%s", service, method)
			grpcData = body // 原始 body 當作 gRPC payload
		} else {
			proxywasm.SendHttpResponse(404, nil, []byte("unsupported path"), 0)
			return types.ActionPause
		}
	}

	// 替換 header → gRPC
	proxywasm.ReplaceHttpRequestHeader(":path", grpcPath)
	proxywasm.ReplaceHttpRequestHeader(":method", "POST")
	proxywasm.ReplaceHttpRequestHeader("content-type", "application/grpc")

	// 替換 body
	proxywasm.ReplaceHttpRequestBody(grpcData)

	return types.ActionContinue
}
