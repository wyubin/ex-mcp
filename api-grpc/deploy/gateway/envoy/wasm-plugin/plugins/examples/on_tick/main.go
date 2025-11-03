package main

import (
	"math/rand"
	"time"

	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm"
	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm/types"
)

const tickMilliseconds uint32 = 20000

func main() {}
func init() {
	// proxywasm.SetVMContext(&vmContext{})
	proxywasm.SetPluginContext(func(contextID uint32) types.PluginContext {
		return &helloWorld{}
	})
}

// type vmContext struct {
// 	types.DefaultVMContext
// }
// func (*vmContext) NewPluginContext(contextID uint32) types.PluginContext {
// 	return &helloWorld{}
// }

// helloWorld implements types.PluginContext.
type helloWorld struct {
	// Embed the default plugin context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultPluginContext
}

// OnPluginStart implements types.PluginContext.
func (ctx *helloWorld) OnPluginStart(pluginConfigurationSize int) types.OnPluginStartStatus {
	proxywasm.LogInfo("OnPluginStart from Go!")
	if err := proxywasm.SetTickPeriodMilliSeconds(tickMilliseconds); err != nil {
		proxywasm.LogCriticalf("failed to set tick period: %v", err)
	}

	return types.OnPluginStartStatusOK
}

// OnTick implements types.PluginContext.
func (ctx *helloWorld) OnTick() {
	t := time.Now().UnixNano()
	proxywasm.LogInfof("It's %d: random value: %d", t, rand.Uint64())
	proxywasm.LogInfof("OnTick called")
}

// NewHttpContext implements types.PluginContext.
func (*helloWorld) NewHttpContext(uint32) types.HttpContext { return &types.DefaultHttpContext{} }
