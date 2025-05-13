package core

import (
	"context"
	"fmt"
	"testing"

	"github.com/wyubin/ex-mcp/mcp01/utils/testtool"

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
	caseA := testtool.AssertCase{ErrorExpect: ErrMcpClientNew}
	_, caseA.ErrorActual = NewClient(cfg)
	caseA.Assert(t, nil)
	// caseB
	caseB := testtool.AssertCase{}
	cfg.TransportType = TransportSSE
	var client *Client
	client, caseB.ErrorActual = NewClient(cfg)
	caseB.Assert(t, nil)
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
	testtool.AssertCase{A: false, B: clientDefault.cfg.Disabled,
		Description: "TestClientEnable - default",
	}.Assert(t, assert.Equal)
	// disable
	clientDefault.Enable(false)
	testtool.AssertCase{A: true, B: clientDefault.cfg.Disabled,
		Description: "TestClientEnable - disable",
	}.Assert(t, assert.Equal)
	// enable
	clientDefault.Enable(true)
	testtool.AssertCase{A: false, B: clientDefault.cfg.Disabled,
		Description: "TestClientEnable - disable",
	}.Assert(t, assert.Equal)
}
