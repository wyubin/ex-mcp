package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm"
	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm/types"

	"google.golang.org/protobuf/proto"

	userpb "main/proto"
)

func main() {}
func init() {
	proxywasm.SetVMContext(&vmContext{})
}

// --- VM Context ---
type vmContext struct {
	types.DefaultVMContext
}

func (*vmContext) NewPluginContext(contextID uint32) types.PluginContext {
	return &pluginContext{}
}

type pluginContext struct {
	types.DefaultPluginContext
}

func (ctx *pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	return &httpContext{contextID: contextID}
}

// --- HTTP Context ---
type httpContext struct {
	types.DefaultHttpContext
	contextID  uint32
	grpcPath   string
	pathParams map[string]string

	respBuffer []byte
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
	ctx.grpcPath, ctx.pathParams = http2grpcMapping(method, path)

	// 替換 header → gRPC
	proxywasm.ReplaceHttpRequestHeader(":path", ctx.grpcPath)
	proxywasm.ReplaceHttpRequestHeader(":method", "POST")
	proxywasm.ReplaceHttpRequestHeader("content-type", "application/grpc")

	proxywasm.LogInfof("[plugin] ReplaceHttpRequestHeader → ctx.grpcPath[%s]", ctx.grpcPath)
	return types.ActionContinue
}

func (ctx *httpContext) OnHttpRequestBody(bodySize int, endOfStream bool) types.Action {
	if !endOfStream {
		return types.ActionPause
	}

	// 取得完整 body
	body, err := proxywasm.GetHttpRequestBody(0, bodySize)
	if err != nil {
		proxywasm.LogCriticalf("failed to get body: %v", err)
		proxywasm.SendHttpResponse(500, nil, []byte("failed to read body"), 0)
		return types.ActionPause
	}

	var grpcData []byte

	// --- 動態 REST → gRPC 映射 ---
	switch ctx.grpcPath {
	// POST /v1/users → CreateUser
	case "/user.UserService/CreateUser":
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

	// GET /v1/users/{id} → GetUser
	case "/user.UserService/GetUser":
		req := &userpb.GetUserRequest{Id: ctx.pathParams["id"]}
		grpcData, err = proto.Marshal(req)
		if err != nil {
			proxywasm.SendHttpResponse(500, nil, []byte("failed to marshal gRPC request"), 0)
			return types.ActionPause
		}

	// GET /v1/users → ListUsers
	case "/user.UserService/ListUsers":
		req := &userpb.ListUsersRequest{}
		grpcData, err = proto.Marshal(req)
		if err != nil {
			proxywasm.SendHttpResponse(500, nil, []byte("failed to marshal gRPC request"), 0)
			return types.ActionPause
		}

	// DELETE /v1/users/{id} → DeleteUser
	case "/user.UserService/DeleteUser":
		req := &userpb.DeleteUserRequest{Id: ctx.pathParams["id"]}
		grpcData, err = proto.Marshal(req)
		if err != nil {
			proxywasm.SendHttpResponse(500, nil, []byte("failed to marshal gRPC request"), 0)
			return types.ActionPause
		}

	default:
		grpcData = body
	}

	proxywasm.LogInfof("[plugin] OnHttpRequestBody → grpcData[%s]", grpcData)

	// 替換 body
	proxywasm.ReplaceHttpRequestBody(grpcFrame(grpcData))

	return types.ActionContinue
}

func (ctx *httpContext) OnHttpResponseHeaders(numHeaders int, endOfStream bool) types.Action {
	proxywasm.ReplaceHttpResponseHeader("content-type", "application/json")
	proxywasm.LogInfof("[plugin] gRPC headers → JSON headers")
	return types.ActionContinue
}

func (ctx *httpContext) OnHttpResponseBody(bodySize int, endOfStream bool) types.Action {
	body, err := proxywasm.GetHttpResponseBody(0, bodySize)
	if err != nil {
		proxywasm.LogCriticalf("[plugin] failed to get response body: %v", err)
		return types.ActionContinue
	}

	// 累積 chunk
	if len(body) > 0 {
		ctx.respBuffer = append(ctx.respBuffer, body...)
		proxywasm.LogInfof("[plugin] OnHttpResponseBody → received chunk, size=%d", len(body))
	}

	if !endOfStream {
		return types.ActionPause
	}
	// 到了 stream 結尾
	proxywasm.LogInfof("[plugin] OnHttpResponseBody → Start!! totalSize=%d", len(ctx.respBuffer))

	if len(ctx.respBuffer) < 5 {
		proxywasm.LogCriticalf("[plugin] invalid gRPC response, size=%d", len(ctx.respBuffer))
		return types.ActionContinue
	}

	grpcPayload := ctx.respBuffer[5:]

	var jsonData []byte
	switch ctx.grpcPath {
	case "/user.UserService/CreateUser":
		var resp userpb.CreateUserResponse
		_ = proto.Unmarshal(grpcPayload, &resp)
		jsonData, _ = json.Marshal(resp)
	case "/user.UserService/GetUser":
		var resp userpb.GetUserResponse
		_ = proto.Unmarshal(grpcPayload, &resp)
		jsonData, _ = json.Marshal(resp)
	case "/user.UserService/ListUsers":
		var resp userpb.ListUsersResponse
		_ = proto.Unmarshal(grpcPayload, &resp)
		jsonData, _ = json.Marshal(resp)
	case "/user.UserService/DeleteUser":
		var resp userpb.DeleteUserResponse
		_ = proto.Unmarshal(grpcPayload, &resp)
		jsonData, _ = json.Marshal(resp)
	default:
		jsonData = grpcPayload
	}

	proxywasm.LogInfof("[plugin] OnHttpResponseHeaders → jsonData[%s]", jsonData)

	// --- 清除 gRPC trailers，確保純 JSON ---
	proxywasm.ReplaceHttpResponseBody(jsonData)

	// 清除 gRPC trailers
	if err := proxywasm.ReplaceHttpResponseTrailers(nil); err != nil {
		proxywasm.LogCriticalf("[plugin] failed to clear trailers: %v", err)
	} else {
		proxywasm.LogInfof("[plugin] Cleared gRPC trailers")
	}

	return types.ActionContinue
}

// --- private method -- //
func http2grpcMapping(method, path string) (string, map[string]string) {
	var grpcPath string
	params := map[string]string{}
	switch {
	// POST /v1/users → CreateUser
	case method == "POST" && path == "/v1/users":
		grpcPath = "/user.UserService/CreateUser"

	// GET /v1/users/{id} → GetUser
	case method == "GET" && strings.HasPrefix(path, "/v1/users/"):
		params["id"] = strings.TrimPrefix(path, "/v1/users/")
		grpcPath = "/user.UserService/GetUser"

	// GET /v1/users → ListUsers
	case method == "GET" && path == "/v1/users":
		grpcPath = "/user.UserService/ListUsers"

	// DELETE /v1/users/{id} → DeleteUser
	case method == "DELETE" && strings.HasPrefix(path, "/v1/users/"):
		params["id"] = strings.TrimPrefix(path, "/v1/users/")
		grpcPath = "/user.UserService/DeleteUser"

	default:
		// 動態解析任意 /{service}/{method} pattern
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) >= 2 {
			service := parts[len(parts)-2]
			method := parts[len(parts)-1]
			grpcPath = fmt.Sprintf("/%s/%s", service, method)
		}
	}
	return grpcPath, params
}

// --- gRPC framing helper ---
func grpcFrame(data []byte) []byte {
	frame := make([]byte, 5+len(data))
	frame[0] = 0 // compression flag
	binary.BigEndian.PutUint32(frame[1:5], uint32(len(data)))
	copy(frame[5:], data)
	return frame
}
