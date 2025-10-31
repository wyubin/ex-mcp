package main

import (
	"encoding/hex"

	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm"
	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm/types"
)

const tickMilliseconds uint32 = 10000

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
	callNum uint32
}

// OnPluginStart implements types.PluginContext.
func (ctx *pluginContext) OnPluginStart(pluginConfigurationSize int) types.OnPluginStartStatus {
	if err := proxywasm.SetTickPeriodMilliSeconds(tickMilliseconds); err != nil {
		proxywasm.LogCriticalf("failed to set tick period: %v", err)
		return types.OnPluginStartStatusFailed
	}
	proxywasm.LogInfof("set tick period milliseconds: %d", tickMilliseconds)
	return types.OnPluginStartStatusOK
}

// OnTick implements types.PluginContext.
func (ctx *pluginContext) OnTick() {
	ctx.callNum++
	ret, err := proxywasm.CallForeignFunction("compress", []byte("hello world!"))
	if err != nil {
		proxywasm.LogCriticalf("foreign function (compress) failed: %v", err)
	}
	proxywasm.LogInfof("foreign function (compress) called: %d, result: %s", ctx.callNum, hex.EncodeToString(ret))
}
