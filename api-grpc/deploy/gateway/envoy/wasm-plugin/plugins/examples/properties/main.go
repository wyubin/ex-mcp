package main

import (
	"encoding/json"
	"slices"

	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm"
	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm/types"
)

func main() {}
func init() {
	proxywasm.SetPluginContext(func(contextID uint32) types.PluginContext {
		return &pluginContext{}
	})
}

type pluginConfig struct {
	PropertyChain []string `json:"propertyChain"`
}

// pluginContext implements types.PluginContext.
type pluginContext struct {
	// Embed the default plugin context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultPluginContext
	propertyChain []string
}

func (ctx *pluginContext) OnPluginStart(pluginConfigurationSize int) types.OnPluginStartStatus {
	// try to load config and set tickPeriod
	cfg := pluginConfig{}
	data, _ := proxywasm.GetPluginConfiguration()
	if len(data) > 0 {
		json.Unmarshal(data, &cfg)
		ctx.propertyChain = cfg.PropertyChain
		proxywasm.LogInfof("set propertyChain: %+v", ctx.propertyChain)
	}
	return types.OnPluginStartStatusOK
}

func (ctx *pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	return &properties{
		contextID:     contextID,
		propertyChain: ctx.propertyChain,
	}
}

// properties implements types.HttpContext.
type properties struct {
	// Embed the default http context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultHttpContext
	contextID     uint32
	propertyChain []string
}

// OnHttpRequestHeaders implements types.HttpContext.
func (ctx *properties) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {
	if len(ctx.propertyChain) == 0 {
		proxywasm.LogInfo("no config to get auth")
		return types.ActionContinue
	}
	auth, err := proxywasm.GetProperty(ctx.propertyChain)
	if err != nil {
		if err == types.ErrorStatusNotFound {
			proxywasm.LogInfo("no auth header for route")
			return types.ActionContinue
		}
		proxywasm.LogCriticalf("failed to read properties: %v", err)
	}
	proxywasm.LogInfof("auth header is \"%s\"", auth)

	hs, err := proxywasm.GetHttpRequestHeaders()
	if err != nil {
		proxywasm.LogCriticalf("failed to get request headers: %v", err)
	}

	// Verify authentication header exists
	authStr := string(auth)
	authHeader := slices.ContainsFunc(hs, func(item [2]string) bool { return item[0] == authStr })

	// Reject requests without authentication header
	if !authHeader {
		_ = proxywasm.SendHttpResponse(401, nil, nil, 16)
		return types.ActionPause
	}

	return types.ActionContinue
}

// OnHttpStreamDone implements types.HttpContext.
func (ctx *properties) OnHttpStreamDone() {
	proxywasm.LogInfof("%d finished", ctx.contextID)
}
