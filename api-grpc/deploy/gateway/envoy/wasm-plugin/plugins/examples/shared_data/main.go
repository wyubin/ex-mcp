package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"

	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm"
	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm/types"
)

const (
	defaultSharedDataKey = "shared_data_key"
)

func main() {}
func init() {
	proxywasm.SetVMContext(&vmContext{})
}

type vmConfig struct {
	SharedDataKey string `json:"sharedDataKey"`
}

var configVM = &vmConfig{SharedDataKey: defaultSharedDataKey} // 全域變數 (屬於 VMContext)

type vmContext struct {
	types.DefaultVMContext
}

// OnVMStart implements types.VMContext.
func (ctx *vmContext) OnVMStart(vmConfigurationSize int) types.OnVMStartStatus {
	data, err := proxywasm.GetVMConfiguration()
	if err == nil {
		json.Unmarshal(data, configVM)
	}
	proxywasm.LogInfof("set vmContext.sharedKey: %+v", configVM.SharedDataKey)

	initialValueBuf := make([]byte, 0)
	// Empty data to indicate that the data is not initialized.
	if err := proxywasm.SetSharedData(configVM.SharedDataKey, initialValueBuf, 0); err != nil {
		proxywasm.LogWarnf("error setting shared data on OnVMStart: %v", err)
	}
	return types.OnVMStartStatusOK
}

// NewPluginContext implements types.VMContext.
func (ctx *vmContext) NewPluginContext(contextID uint32) types.PluginContext {
	proxywasm.LogInfof("set pluginContext.sharedKey: %+v", configVM.SharedDataKey)
	return &pluginContext{}
}

type pluginContext struct {
	types.DefaultPluginContext
}

// NewHttpContext implements types.PluginContext.
func (ctx *pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	proxywasm.LogInfof("set HttpContext.sharedKey: %+v", configVM.SharedDataKey)
	return &httpContext{}
}

type httpContext struct {
	types.DefaultHttpContext
}

// OnHttpRequestHeaders implements types.HttpContext.
func (ctx *httpContext) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {
	for {
		value, err := ctx.incrementData()
		if err == nil {
			proxywasm.LogInfof("shared value: %d", value)
		} else if errors.Is(err, types.ErrorStatusCasMismatch) {
			continue
		}
		break
	}
	return types.ActionContinue
}

// incrementData increments the shared data value by 1.
func (ctx *httpContext) incrementData() (uint64, error) {
	data, cas, err := proxywasm.GetSharedData(configVM.SharedDataKey)
	if err != nil {
		proxywasm.LogWarnf("error getting shared data on OnHttpRequestHeaders: %v", err)
		return 0, err
	}

	var nextValue uint64
	if len(data) > 0 {
		nextValue = binary.LittleEndian.Uint64(data) + 1
	} else {
		nextValue = 1
	}

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, nextValue)
	if err := proxywasm.SetSharedData(configVM.SharedDataKey, buf, cas); err != nil {
		proxywasm.LogWarnf("error setting shared data on OnHttpRequestHeaders: %v", err)
		return 0, err
	}
	return nextValue, err
}
