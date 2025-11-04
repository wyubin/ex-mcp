package main

import (
	"encoding/json"
	"strings"

	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm"
	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm/types"
)

const (
	METRIC_NAME   = "custom_header_value_counts"
	REPORTER_NAME = "wasmgosdk"
)

func main() {}
func init() {
	proxywasm.SetPluginContext(func(contextID uint32) types.PluginContext {
		return &metricPluginContext{}
	})
}

// metricPluginContext implements types.PluginContext.
type metricPluginContext struct {
	// Embed the default plugin context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultPluginContext
	config pluginConfig
}

// NewHttpContext implements types.PluginContext.
func (ctx *metricPluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	return &metricHttpContext{config: ctx.config}
}

func (ctx *metricPluginContext) OnPluginStart(pluginConfigurationSize int) types.OnPluginStartStatus {
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
	ctx.config = cfg
	proxywasm.LogInfof("parsed config: %+v\n", cfg)
	return types.OnPluginStartStatusOK
}

type pluginConfig struct {
	HeaderName string `json:"headerName"`
	ValueTag   string `json:"valueTag"`
	ReportTag  string `json:"reporterTag"`
}

// metricHttpContext implements types.HttpContext.
type metricHttpContext struct {
	// Embed the default http context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultHttpContext
	config pluginConfig
}

// counters is a map from custom header value to a counter metric.
// Note that Proxy-Wasm plugins are single threaded, so no need to use a lock.
var counters = map[string]proxywasm.MetricCounter{}

// OnHttpRequestHeaders implements types.HttpContext.
func (ctx *metricHttpContext) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {
	customHeaderValue, err := proxywasm.GetHttpRequestHeader(ctx.config.HeaderName)
	if err == nil {
		counter, ok := counters[customHeaderValue]
		if !ok {
			// This metric is processed as: custom_header_value_counts{value="foo",reporter="wasmgosdk"} n.
			// The extraction rule is defined in envoy.yaml as a bootstrap configuration.
			// See https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/metrics/v3/stats.proto#config-metrics-v3-statsconfig.
			counter = proxywasm.DefineCounterMetric(ctx.genMetricFQN(customHeaderValue))
			counters[customHeaderValue] = counter
		}
		counter.Increment(1)
	}
	return types.ActionContinue
}

func (ctx *metricHttpContext) genMetricFQN(tagValue string) string {
	parts := []string{
		METRIC_NAME,
		strings.Join([]string{ctx.config.ValueTag, tagValue}, "="),
		strings.Join([]string{ctx.config.ReportTag, REPORTER_NAME}, "="),
	}
	return strings.Join(parts, "_")
}
