package main

import (
	"encoding/json"

	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm"
	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm/types"
)

const defaultTickMilliseconds uint32 = 1000

func main() {}
func init() {
	proxywasm.SetPluginContext(func(contextID uint32) types.PluginContext {
		return &pluginContext{
			contextID: contextID,
			postponed: make([]uint32, 0, 1024),
		}
	})
}

type pluginConfig struct {
	TickPeriod uint32 `json:"tickPeriod"`
}

// pluginContext implements types.PluginContext.
type pluginContext struct {
	// Embed the default plugin context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultPluginContext
	contextID uint32
	postponed []uint32
}

// OnPluginStart implements types.PluginContext.
func (ctx *pluginContext) OnPluginStart(pluginConfigurationSize int) types.OnPluginStartStatus {
	// try to load config and set tickPeriod
	cfg := pluginConfig{}
	tickPeriodSet := defaultTickMilliseconds
	data, _ := proxywasm.GetPluginConfiguration()
	if len(data) > 0 {
		json.Unmarshal(data, &cfg)
		if cfg.TickPeriod > 0 {
			tickPeriodSet = cfg.TickPeriod
		}
	}
	if err := proxywasm.SetTickPeriodMilliSeconds(tickPeriodSet); err != nil {
		proxywasm.LogCriticalf("failed to set tick period: %v", err)
	}

	return types.OnPluginStartStatusOK
}

// OnTick implements types.PluginContext.
func (ctx *pluginContext) OnTick() {
	for len(ctx.postponed) > 0 {
		httpCtxId, tail := ctx.postponed[0], ctx.postponed[1:]
		proxywasm.LogInfof("resume request with contextID=%v", httpCtxId)
		_ = proxywasm.SetEffectiveContext(httpCtxId)
		_ = proxywasm.ResumeHttpRequest()
		ctx.postponed = tail
	}
}

// NewHttpContext implements types.PluginContext.
func (ctx *pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	return &httpContext{
		contextID: contextID,
		pluginCtx: ctx,
	}
}

// httpContext implements types.HttpContext.
type httpContext struct {
	// Embed the default http context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultHttpContext
	contextID uint32
	pluginCtx *pluginContext
}

// OnHttpRequestHeaders implements types.HttpContext.
func (ctx *httpContext) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {
	proxywasm.LogInfof("postpone request with contextID=%d", ctx.contextID)
	ctx.pluginCtx.postponed = append(ctx.pluginCtx.postponed, ctx.contextID)
	return types.ActionPause
}

func (ctx *httpContext) OnHttpRequestBody(numHeaders int, endOfStream bool) types.Action {
	if !endOfStream {
		// Wait until we see the entire body to replace.
		return types.ActionPause
	}
	proxywasm.LogInfof("tick continue request with contextID=%d; with %d ctx in queue: %+v", ctx.contextID, len(ctx.pluginCtx.postponed), ctx.pluginCtx.postponed)
	return types.ActionContinue
}

func (ctx *httpContext) OnHttpResponseHeaders(numHeaders int, endOfStream bool) types.Action {
	proxywasm.LogInfof("get response with contextID=%d; with %d ctx in queue: %+v", ctx.contextID, len(ctx.pluginCtx.postponed), ctx.pluginCtx.postponed)
	return types.ActionContinue
}
