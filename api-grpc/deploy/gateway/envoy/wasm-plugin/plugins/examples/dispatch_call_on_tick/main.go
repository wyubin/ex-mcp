package main

import (
	"crypto/rand"

	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm"
	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm/types"
)

const tickMilliseconds uint32 = 30000

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
	return &pluginContext{contextID: contextID}
}

// pluginContext implements types.PluginContext.
type pluginContext struct {
	// Embed the default plugin context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultPluginContext
	contextID uint32
	callBack  func(numHeaders, bodySize, numTrailers int)
	cnt       int
}

// OnPluginStart implements types.PluginContext.
func (ctx *pluginContext) OnPluginStart(pluginConfigurationSize int) types.OnPluginStartStatus {
	if err := proxywasm.SetTickPeriodMilliSeconds(tickMilliseconds); err != nil {
		proxywasm.LogCriticalf("failed to set tick period: %v", err)
		return types.OnPluginStartStatusFailed
	}
	proxywasm.LogInfof("set tick period milliseconds: %d", tickMilliseconds)
	data, _ := proxywasm.GetPluginConfiguration()
	namePlugin := string(data)

	ctx.callBack = func(numHeaders, bodySize, numTrailers int) {
		ctx.cnt++
		proxywasm.LogInfof("into call back")
		proxywasm.LogInfof("called %d for contextID=%d; namePlugin=%s", ctx.cnt, ctx.contextID, namePlugin)
		// headers, err := proxywasm.GetHttpCallResponseHeaders()
		// if err != nil && err != types.ErrorStatusNotFound {
		// 	panic(err)
		// }
		// for _, h := range headers {
		// 	proxywasm.LogInfof("response header for the dispatched call: %s: %s", h[0], h[1])
		// }
		// headers, err = proxywasm.GetHttpCallResponseTrailers()
		// if err != nil && err != types.ErrorStatusNotFound {
		// 	panic(err)
		// }
		// for _, h := range headers {
		// 	proxywasm.LogInfof("response trailer for the dispatched call: %s: %s", h[0], h[1])
		// }
	}
	return types.OnPluginStartStatusOK
}

// OnTick implements types.PluginContext.
func (ctx *pluginContext) OnTick() {
	headers := [][2]string{
		{":method", "GET"}, {":authority", "some_authority"}, {"accept", "*/*"},
	}
	// Pick random value to select the request path.
	buf := make([]byte, 1)
	_, _ = rand.Read(buf)
	if buf[0]%2 == 0 {
		headers = append(headers, [2]string{":path", "/ok"})
	} else {
		headers = append(headers, [2]string{":path", "/fail"})
	}
	if _, err := proxywasm.DispatchHttpCall("web_service", headers, nil, nil, 5000, ctx.callBack); err != nil {
		proxywasm.LogCriticalf("dispatch httpcall failed: %v", err)
	}
}
