package main

import (
	"encoding/json"
	"fmt"

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
	// Embed the default plugin context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultPluginContext
	configuration pluginConfiguration
}

// pluginConfiguration is a type to represent an example configuration for this wasm plugin.
type pluginConfiguration struct {
	RequiredKeys []string `json:"requiredKeys"`
}

// OnPluginStart implements types.PluginContext.
func (ctx *pluginContext) OnPluginStart(pluginConfigurationSize int) types.OnPluginStartStatus {
	data, err := proxywasm.GetPluginConfiguration()
	if err != nil && err != types.ErrorStatusNotFound {
		proxywasm.LogCriticalf("error reading plugin configuration: %v", err)
		return types.OnPluginStartStatusFailed
	}
	config, err := parsePluginConfiguration(data)
	if err != nil {
		proxywasm.LogCriticalf("error parsing plugin configuration: %v", err)
		return types.OnPluginStartStatusFailed
	}
	ctx.configuration = config
	proxywasm.LogInfof("parsed config: %+v\n", config)
	return types.OnPluginStartStatusOK
}

// parsePluginConfiguration parses the json plugin configuration data and returns pluginConfiguration.
// Note that this parses the json data by gjson, since TinyGo doesn't support encoding/json.
// You can also try https://github.com/mailru/easyjson, which supports decoding to a struct.
func parsePluginConfiguration(data []byte) (pluginConfiguration, error) {
	if len(data) == 0 {
		return pluginConfiguration{}, nil
	}

	var config pluginConfiguration
	if err := json.Unmarshal(data, &config); err != nil {
		proxywasm.LogCriticalf("failed to parse plugin config: %v", err)
		return pluginConfiguration{}, fmt.Errorf("the plugin configuration is not a valid json: %q", string(data))
	}

	return config, nil
}

// NewHttpContext implements types.PluginContext.
func (ctx *pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	return &payloadValidationContext{requiredKeys: ctx.configuration.RequiredKeys}
}

// payloadValidationContext implements types.HttpContext.
type payloadValidationContext struct {
	// Embed the default root http context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultHttpContext
	requiredKeys []string
}

// OnHttpRequestHeaders implements types.HttpContext.
func (*payloadValidationContext) OnHttpRequestHeaders(numHeaders int, _ bool) types.Action {
	contentType, err := proxywasm.GetHttpRequestHeader("content-type")
	if err != nil || contentType != "application/json" {
		// If the header doesn't have the expected content value, send the 403 response,
		if err := proxywasm.SendHttpResponse(403, nil, []byte("content-type must be provided"), -1); err != nil {
			panic(err)
		}
		// and terminates the further processing of this traffic by ActionPause.
		return types.ActionPause
	}

	// ActionContinue lets the host continue the processing the body.
	return types.ActionContinue
}

// OnHttpRequestBody implements types.HttpContext.
func (ctx *payloadValidationContext) OnHttpRequestBody(bodySize int, endOfStream bool) types.Action {
	proxywasm.LogInfof("OnHttpRequestBody called. BodySize: %d, endOfStream: %v", bodySize, endOfStream)
	if !endOfStream {
		// OnHttpRequestBody may be called each time a part of the body is received.
		// Wait until we see the entire body to replace.
		return types.ActionPause
	}

	body, err := proxywasm.GetHttpRequestBody(0, bodySize)
	if err != nil {
		proxywasm.LogErrorf("failed to get request body: %v", err)
		return types.ActionContinue
	}
	if !ctx.validatePayload(body) {
		// If the validation fails, send the 403 response,
		if err := proxywasm.SendHttpResponse(403, nil, []byte("invalid payload"), -1); err != nil {
			proxywasm.LogErrorf("failed to send the 403 response: %v", err)
		}
		// and terminates this traffic.
		return types.ActionPause
	}

	return types.ActionContinue
}

// validatePayload validates the given json payload.
// Note that this function parses the json data by gjson, since TinyGo doesn't support encoding/json.
func (ctx *payloadValidationContext) validatePayload(body []byte) bool {
	var jsonBody map[string]any
	if err := json.Unmarshal(body, &jsonBody); err != nil {
		proxywasm.LogCriticalf("failed to parse plugin config: %v", err)
		return false
	}

	// Do any validation on the json. Check if required keys exist here as an example.
	// The required keys are configurable via the plugin configuration.
	for _, requiredKey := range ctx.requiredKeys {
		if _, found := jsonBody[requiredKey]; !found {
			proxywasm.LogErrorf("required key (%v) is missing: %v", requiredKey, jsonBody)
			return false
		}
	}

	return true
}
