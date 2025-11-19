package main

import (
	"encoding/json"
	"slices"

	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm"
	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm/types"

	"main/internal/grpcio"

	userpb "main/proto/user"
)

const (
	timeOutDispatch = 5000
)

var (
	allRoutes = []grpcio.Route{
		userpb.NewUserGrpcMapper("user"),
	}
)

func main() {}
func init() {
	proxywasm.SetPluginContext(func(contextID uint32) types.PluginContext {
		return &pluginContext{}
	})
}

type pluginContext struct {
	types.DefaultPluginContext
	routes *grpcio.Routes
}

type pluginConfig map[string]string

func (ctx *pluginContext) OnPluginStart(pluginConfigurationSize int) types.OnPluginStartStatus {
	proxywasm.LogDebug("loading plugin config")
	data, err := proxywasm.GetPluginConfiguration()

	if err != nil {
		proxywasm.LogCriticalf("error reading plugin configuration: %v", err)
		return types.OnPluginStartStatusFailed
	}

	if data == nil { // 沒有設定就直接 pass
		proxywasm.LogCriticalf("no plugin configuration")
		return types.OnPluginStartStatusFailed
	}

	var cfg pluginConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		proxywasm.LogCriticalf("failed to parse plugin config: %v", err)
		return types.OnPluginStartStatusFailed
	}

	routes := grpcio.NewRoutes()
	for nameRoute, clusterRoute := range cfg {
		proxywasm.LogInfof("[plugin] get config[%s: %s]", nameRoute, clusterRoute)
		idx := slices.IndexFunc(allRoutes, func(route grpcio.Route) bool { return route.Name() == nameRoute })
		if idx == -1 {
			proxywasm.LogWarnf("no route found for name: %s", nameRoute)
			continue
		}
		route := allRoutes[idx]
		route.SetClusterName(clusterRoute)
		routes.Add(route)
	}
	ctx.routes = routes
	return types.OnPluginStartStatusOK
}

func (ctx *pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	return &httpContext{routes: ctx.routes}
}

// --- HTTP Context ---
type httpContext struct {
	types.DefaultHttpContext
	routes     *grpcio.Routes
	routeReq   grpcio.Route
	info       grpcio.InfoRequest
	remoteBody []byte
}

func (ctx *httpContext) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {
	method, _ := proxywasm.GetHttpRequestHeader(":method")
	path, _ := proxywasm.GetHttpRequestHeader(":path")
	proxywasm.LogInfof("get method[%s], path[%s]", method, path)
	route, info := ctx.routes.Match(method, path)
	if route == nil {
		proxywasm.SendHttpResponse(500, nil, []byte("no match route"), 0)
		return types.ActionContinue
	}
	ctx.routeReq = route
	ctx.info = info

	proxywasm.LogInfof("[plugin] setup callback info → ctx.info[%s]", ctx.info)
	return types.ActionContinue
}

func (ctx *httpContext) OnHttpRequestBody(bodySize int, endOfStream bool) types.Action {
	if ctx.routeReq == nil {
		return types.ActionContinue
	}
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
	grpcBody, err := ctx.routeReq.RequestCov(ctx.info, body)
	if err != nil {
		proxywasm.LogCriticalf("failed to gen grpc body: %v", err)
		proxywasm.SendHttpResponse(500, nil, []byte(err.Error()), 0)
		return types.ActionContinue
	}

	proxywasm.LogInfof("[plugin] OnHttpRequestBody → grpcBody[%+v]", grpcBody)
	headers := [][2]string{
		{":method", "POST"}, {":path", ctx.info.PathGrpc}, {":authority", "localhost"},
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

	jsonBody, err := ctx.routeReq.ResponseCov(ctx.info, ctx.remoteBody)
	if err != nil {
		proxywasm.LogCriticalf("failed to encode response: %v", err)
		proxywasm.SendHttpResponse(500, nil, []byte(err.Error()), 0)
		return types.ActionContinue
	}
	proxywasm.ReplaceHttpResponseBody(jsonBody)

	return types.ActionContinue
}
