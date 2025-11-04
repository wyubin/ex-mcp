package main

import (
	"strconv"

	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm"
	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm/types"
)

const clusterName = "httpbin"

func main() {}
func init() {
	proxywasm.SetPluginContext(func(contextID uint32) types.PluginContext {
		return &pluginContext{}
	})
}

// pluginContext implements types.PluginContext.
type pluginContext struct {
	// Embed the default plugin context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultPluginContext
}

// NewHttpContext implements types.PluginContext.
func (*pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	return &httpContext{contextID: contextID}
}

// httpContext implements types.HttpContext.
type httpContext struct {
	// Embed the default http context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultHttpContext
	// contextID is the unique identifier assigned to each httpContext.
	contextID uint32
	// pendingDispatchedRequest is the number of pending dispatched requests.
	pendingDispatchedRequest int
}

const totalDispatchNum = 10

// OnHttpResponseHeaders implements types.HttpContext.
func (ctx *httpContext) OnHttpResponseHeaders(numHeaders int, endOfStream bool) types.Action {
	// On each request response, we dispatch the http calls `totalDispatchNum` times.
	// Note: DispatchHttpCall is asynchronously processed, so each loop is non-blocking.
	for i := 0; i < totalDispatchNum; i++ {
		if _, err := proxywasm.DispatchHttpCall(clusterName, [][2]string{
			{":path", "/"},
			{":method", "GET"},
			{":authority", ""}},
			nil, nil, 50000, ctx.dispatchCallback); err != nil {
			panic(err)
		}
		// Now we have made a dispatched request, so we record it.
		ctx.pendingDispatchedRequest++
	}
	return types.ActionPause
}

// dispatchCallback is the callback function called in response to the response arrival from the dispatched request.
func (ctx *httpContext) dispatchCallback(numHeaders, bodySize, numTrailers int) {
	// Decrement the pending request counter.
	ctx.pendingDispatchedRequest--
	if ctx.pendingDispatchedRequest == 0 {
		// This case, all the dispatched request was processed.
		// Adds a response header to the original response.
		_ = proxywasm.AddHttpResponseHeader("total-dispatched", strconv.Itoa(totalDispatchNum))
		// And then contniue the original reponse.
		_ = proxywasm.ResumeHttpResponse()
		proxywasm.LogInfof("response resumed after processed %d dispatched request", totalDispatchNum)
	} else {
		proxywasm.LogInfof("pending dispatched requests: %d", ctx.pendingDispatchedRequest)
	}
}
