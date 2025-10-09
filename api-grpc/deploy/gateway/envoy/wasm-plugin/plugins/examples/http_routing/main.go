package main

import (
	"crypto/rand"
	"encoding/binary"

	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm"
	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm/types"
)

func main() {}
func init() {
	proxywasm.SetVMContext(&vmContext{})
}

// vmContext implements types.VMContext.
type vmContext struct {
	// Embed the default VM context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultVMContext
}

// NewPluginContext implements types.VMContext.
func (*vmContext) NewPluginContext(contextID uint32) types.PluginContext {
	return &pluginContext{}
}

// pluginContext implements types.PluginContext.
type pluginContext struct {
	// Embed the default plugin context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultPluginContext

	diceOverride uint32 // For unit test
}

// OnPluginStart implements types.PluginContext.
func (ctx *pluginContext) OnPluginStart(pluginConfigurationSize int) types.OnPluginStartStatus {
	data, err := proxywasm.GetPluginConfiguration()
	if err != nil && err != types.ErrorStatusNotFound {
		proxywasm.LogCriticalf("error reading plugin configuration: %v", err)
		return types.OnPluginStartStatusFailed
	}

	// If the configuration data is not empty, we use its value to override the routing
	// decision for unit tests.
	if len(data) > 0 {
		ctx.diceOverride = uint32(data[0])
	}
	proxywasm.LogInfof("GetPluginConfiguration: %+v\n", data)
	proxywasm.LogInfof("ctx.diceOverride: %d\n", ctx.diceOverride)
	return types.OnPluginStartStatusOK
}

// NewHttpContext implements types.PluginContext.
func (ctx *pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	return &httpRouting{diceOverride: ctx.diceOverride}
}

// httpRouting implements types.HttpContext.
type httpRouting struct {
	// Embed the default http context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultHttpContext

	diceOverride uint32 // For unit test
}

// dice returns a random value to be used to determine the route.
func dice() uint32 {
	buf := make([]byte, 4)
	_, _ = rand.Read(buf)
	return binary.LittleEndian.Uint32(buf)
}

// OnHttpRequestHeaders implements types.HttpContext.
func (ctx *httpRouting) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {
	// Randomly routing to the canary cluster.
	var value uint32
	if ctx.diceOverride != 0 {
		value = ctx.diceOverride
	} else {
		value = dice()
	}
	proxywasm.LogInfof("value: %d\n", value)
	if value%2 == 0 {
		const authorityKey = ":authority"
		value, err := proxywasm.GetHttpRequestHeader(authorityKey)
		if err != nil {
			proxywasm.LogCritical("failed to get request header: ':authority'")
			return types.ActionPause
		}
		// Append "-canary" suffix to route this request to the canary cluster.
		value += "-canary"
		if err := proxywasm.ReplaceHttpRequestHeader(":authority", value); err != nil {
			proxywasm.LogCritical("failed to set request header: test")
			return types.ActionPause
		}
	}
	return types.ActionContinue
}
