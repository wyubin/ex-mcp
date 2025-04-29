package core

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	urlCfg     = "http://localhost:8081/sse"
	cfgDefault = CfgServer{
		Url:           urlCfg,
		TransportType: TransportSSE,
	}
	clientDefault, _ = NewClient(cfgDefault)
)

func TestClientNew(t *testing.T) {
	cfg := CfgServer{
		Url:           urlCfg,
		TransportType: TransportSTDIO,
	}
	_, err := NewClient(cfg)
	fmt.Printf("error: %s\n", err)
	assert.Error(t, err, "TestClientNew")
	cfg.TransportType = TransportSSE
	client, err := NewClient(cfg)
	assert.NoError(t, err, "TestClientNew")
	defer client.Close()
}

func TestClientInit(t *testing.T) {
	ctx := context.Background()
	info, err := clientDefault.Init(ctx)
	assert.NoError(t, err, "TestFClientInit")
	fmt.Printf("info: %+v\n", *info)
}

func TestClientListTools(t *testing.T) {
	ctx := context.Background()
	clientDefault.Init(ctx)
	tools, err := clientDefault.ListTools(ctx)
	assert.NoError(t, err, "TestFClientListTools")
	fmt.Printf("tools: %+v\n", tools)
}

func TestClientCallTool(t *testing.T) {
	ctx := context.Background()
	clientDefault.Init(ctx)
	args := map[string]interface{}{
		"name": "binbinbin",
	}
	contents, err := clientDefault.CallTool(ctx, "save_name", args)
	assert.NoError(t, err, "TestClientCallTool")
	fmt.Printf("contents: %+v\n", contents)
}

func TestClientEnable(t *testing.T) {
	// client 預設開啟
	assert.False(t, clientDefault.cfg.Disabled, "TestClientEnable - before")
	clientDefault.Enable(false)
	assert.True(t, clientDefault.cfg.Disabled, "TestClientEnable - disable")
	clientDefault.Enable(true)
	assert.False(t, clientDefault.cfg.Disabled, "TestClientEnable - enable")
}
