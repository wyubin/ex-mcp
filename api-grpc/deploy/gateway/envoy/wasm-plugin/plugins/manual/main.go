package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm"
	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm/types"

	"google.golang.org/protobuf/proto"

	userpb "main/proto/user"
)

const (
	timeOutDispatch = 5000
)

func main() {}
func init() {
	proxywasm.SetPluginContext(func(contextID uint32) types.PluginContext {
		return &pluginContext{}
	})
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
	remoteBody []byte
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

	proxywasm.LogInfof("[plugin] setup callback path → ctx.grpcPath[%s]", ctx.grpcPath)
	return types.ActionContinue
}

func (ctx *httpContext) OnHttpRequestBody(bodySize int, endOfStream bool) types.Action {
	proxywasm.LogInfof("[plugin] OnHttpRequestBody → received chunk, size=%d, endOfStream=%t", bodySize, endOfStream)
	if !endOfStream {
		return types.ActionPause
	}

	// 取得完整 body
	body, err := proxywasm.GetHttpRequestBody(0, bodySize)
	if err != nil {
		proxywasm.LogCriticalf("failed to get body: %v", err)
		proxywasm.SendHttpResponse(500, nil, []byte("failed to read body"), 0)
		return types.ActionContinue
	}
	grpcBody, err := genGrpcBody(ctx.grpcPath, ctx.pathParams, body)
	if err != nil {
		proxywasm.LogCriticalf("failed to gen grpc body: %v", err)
		proxywasm.SendHttpResponse(500, nil, []byte("failed to gen grpc body"), 0)
		return types.ActionContinue
	}

	proxywasm.LogInfof("[plugin] OnHttpRequestBody → grpcBody[%+v]", grpcBody)
	headers := [][2]string{
		{":method", "POST"}, {":path", ctx.grpcPath}, {":authority", "localhost"},
		{"content-type", "application/grpc"},
	}
	_, err = proxywasm.DispatchHttpCall("user_service", headers, grpcBody, nil, timeOutDispatch, ctx.callback)
	if err != nil {
		proxywasm.LogCriticalf("dipatch httpcall failed: %v", err)
		proxywasm.SendHttpResponse(500, nil, []byte("dipatch httpcall failed"), 0)
		return types.ActionContinue
	}
	return types.ActionPause
}

func (ctx *httpContext) callback(numHeaders, bodySize, numTrailers int) {
	bodyCB, err := proxywasm.GetHttpCallResponseBody(0, bodySize)
	if err != nil {
		proxywasm.LogCriticalf("failed to get response body: %v", err)
		proxywasm.SendHttpResponse(500, nil, []byte("failed to get "), 0)
		_ = proxywasm.ResumeHttpRequest()
		return
	}
	ctx.remoteBody = append([]byte(nil), bodyCB...)
	_ = proxywasm.ResumeHttpRequest()
}

func (ctx *httpContext) OnHttpResponseHeaders(numHeaders int, endOfStream bool) types.Action {
	if len(ctx.remoteBody) == 0 {
		return types.ActionContinue
	}
	proxywasm.LogInfof("[plugin] replace header")
	proxywasm.ReplaceHttpResponseHeader("content-type", "application/json")
	proxywasm.RemoveHttpResponseHeader("content-length")
	proxywasm.ReplaceHttpResponseHeader(":status", "200")
	return types.ActionContinue
}

func (ctx *httpContext) OnHttpResponseBody(bodySize int, endOfStream bool) types.Action {
	if len(ctx.remoteBody) == 0 {
		return types.ActionContinue
	}
	proxywasm.LogInfof("[plugin] OnHttpResponseBody → received chunk, size=%d, endOfStream=%t", bodySize, endOfStream)
	if !endOfStream {
		return types.ActionPause
	}
	proxywasm.LogInfof("[plugin] OnHttpResponseBody → grpcData from ctx [%s]", ctx.remoteBody)

	jsonBody, err := genJsonResponseBody(ctx.grpcPath, ctx.remoteBody)
	if err != nil {
		proxywasm.LogCriticalf("failed to encode response: %v", err)
		proxywasm.SendHttpResponse(500, nil, []byte("failed to encode response"), 0)
		return types.ActionPause
	}
	proxywasm.ReplaceHttpResponseBody(jsonBody)

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

// genGrpcBody 基於 bodyJson 轉成 grpc body 提供remote call
func genGrpcBody(pathGrpc string, pathParams map[string]string, bodyJson []byte) ([]byte, error) {
	var grpcData []byte
	var err error
	switch pathGrpc {
	// POST /v1/users → CreateUser
	case "/user.UserService/CreateUser":
		var req userpb.CreateUserRequest
		if err := json.Unmarshal(bodyJson, &req); err != nil {
			return nil, fmt.Errorf("invalid JSON for CreateUser\n->%w", err)
		}
		grpcData, err = proto.Marshal(&req)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal gRPC request\n->%w", err)
		}

	// GET /v1/users/{id} → GetUser
	case "/user.UserService/GetUser":
		req := &userpb.GetUserRequest{Id: pathParams["id"]}
		grpcData, err = proto.Marshal(req)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal gRPC request\n->%w", err)
		}

	// GET /v1/users → ListUsers
	case "/user.UserService/ListUsers":
		req := &userpb.ListUsersRequest{}
		grpcData, err = proto.Marshal(req)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal gRPC request\n->%w", err)
		}

	// DELETE /v1/users/{id} → DeleteUser
	case "/user.UserService/DeleteUser":
		req := &userpb.DeleteUserRequest{Id: pathParams["id"]}
		grpcData, err = proto.Marshal(req)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal gRPC request\n->%w", err)
		}

	default:
		return nil, fmt.Errorf("no match grpc service path\n-> path: %s", pathGrpc)
	}
	return grpcFrame(grpcData), nil
}

// --- gRPC framing helper ---
func grpcFrame(data []byte) []byte {
	frame := make([]byte, 5+len(data))
	frame[0] = 0 // compression flag
	binary.BigEndian.PutUint32(frame[1:5], uint32(len(data)))
	copy(frame[5:], data)
	return frame
}

// genJsonResponseBody 將 grpc response body 轉成 json body
func genJsonResponseBody(pathGrpc string, bodyGrpc []byte) ([]byte, error) {
	if len(bodyGrpc) < 5 {
		return nil, fmt.Errorf("body less than 5 length\n-> byte: %s", bodyGrpc)
	}
	grpcPayload := bodyGrpc[5:]
	var (
		jsonData  []byte
		errEncode error
	)
	switch pathGrpc {
	case "/user.UserService/CreateUser":
		var resp userpb.CreateUserResponse
		errEncode = proto.Unmarshal(grpcPayload, &resp)
		jsonData, _ = json.Marshal(&resp)
	case "/user.UserService/GetUser":
		var resp userpb.GetUserResponse
		errEncode = proto.Unmarshal(grpcPayload, &resp)
		jsonData, _ = json.Marshal(&resp)
	case "/user.UserService/ListUsers":
		var resp userpb.ListUsersResponse
		errEncode = proto.Unmarshal(grpcPayload, &resp)
		jsonData, _ = json.Marshal(&resp)
	case "/user.UserService/DeleteUser":
		var resp userpb.DeleteUserResponse
		errEncode = proto.Unmarshal(grpcPayload, &resp)
		jsonData, _ = json.Marshal(&resp)
	default:
		return nil, fmt.Errorf("no match grpc service path\n-> path: %s", pathGrpc)
	}
	if errEncode != nil {
		return nil, fmt.Errorf("grpcBody encode error\n->%w", errEncode)
	}
	return jsonData, nil
}
