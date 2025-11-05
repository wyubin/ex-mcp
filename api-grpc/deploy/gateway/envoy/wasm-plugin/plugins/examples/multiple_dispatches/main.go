package main

import (
	"encoding/json"
	"strconv"

	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm"
	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm/types"
)

func main() {}
func init() {
	proxywasm.SetPluginContext(func(contextID uint32) types.PluginContext {
		return &pluginContext{}
	})
}

// pluginContext implements types.PluginContext.
type pluginContext struct {
	types.DefaultPluginContext
	config pluginConfig
}

// NewHttpContext implements types.PluginContext.
func (s *pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	return &httpContext{config: s.config}
}

func (s *pluginContext) OnPluginStart(pluginConfigurationSize int) types.OnPluginStartStatus {
	data, err := proxywasm.GetPluginConfiguration()
	if err != nil && err != types.ErrorStatusNotFound {
		proxywasm.LogCriticalf("error reading plugin configuration: %v", err)
		return types.OnPluginStartStatusFailed
	}
	var cfg pluginConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		proxywasm.LogCriticalf("failed to parse plugin config: %v", err)
		return types.OnPluginStartStatusFailed
	}
	s.config = cfg
	proxywasm.LogInfof("parsed config: %+v\n", cfg)
	return types.OnPluginStartStatusOK
}

type pluginConfig struct {
	DispatchCluster string `json:"dispatchCluster"`
	DispatchCount   int    `json:"dispatchCount"`
}

// httpContext implements types.HttpContext.
type httpContext struct {
	types.DefaultHttpContext
	config pluginConfig
	// pendingDispatchedRequest is the number of pending dispatched requests.
	pendingDispatchedRequest int
}

// OnHttpResponseHeaders implements types.HttpContext.
func (s *httpContext) OnHttpResponseHeaders(numHeaders int, endOfStream bool) types.Action {
	// On each request response, we dispatch the http calls `totalDispatchNum` times.
	// Note: DispatchHttpCall is asynchronously processed, so each loop is non-blocking.
	for i := 0; i < s.config.DispatchCount; i++ {
		if _, err := proxywasm.DispatchHttpCall(s.config.DispatchCluster, [][2]string{
			{":path", "/"},
			{":method", "GET"},
			{":authority", ""}},
			nil, nil, 50000, s.dispatchCallback); err != nil {
			panic(err)
		}
		// Now we have made a dispatched request, so we record it.
		s.pendingDispatchedRequest++
	}
	return types.ActionPause
}

// dispatchCallback is the callback function called in response to the response arrival from the dispatched request.
func (s *httpContext) dispatchCallback(numHeaders, bodySize, numTrailers int) {
	// Decrement the pending request counter.
	s.pendingDispatchedRequest--
	if s.pendingDispatchedRequest == 0 {
		// This case, all the dispatched request was processed.
		// Adds a response header to the original response.
		_ = proxywasm.AddHttpResponseHeader("total-dispatched", strconv.Itoa(s.config.DispatchCount))
		// And then contniue the original reponse.
		_ = proxywasm.ResumeHttpResponse()
		proxywasm.LogInfof("response resumed after processed %d dispatched request", s.config.DispatchCount)
	} else {
		proxywasm.LogInfof("pending dispatched requests: %d", s.pendingDispatchedRequest)
	}
}
