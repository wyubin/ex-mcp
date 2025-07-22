package mcp

import (
	"context"
	"fmt"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/wyubin/ex-mcp/mcp01/utils/testtool"

	"github.com/stretchr/testify/assert"
)

var (
	urlCfg = "http://localhost:8081/sse"
	cfgSSE = CfgServer{
		Url:           urlCfg,
		TransportType: TransportSSE,
	}
	cfgStdio = CfgServer{
		Command:       "/workspaces/ex-mcp/mcp01/bin/mcp01-server",
		Env:           map[string]string{},
		Args:          []string{},
		TransportType: TransportSTDIO,
	}
)

func TestClientNew(t *testing.T) {
	cfg := CfgServer{
		Url:           urlCfg,
		TransportType: TransportSTDIO,
	}
	// fail case
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
	caseSSE := testtool.AssertCase{Description: "TestFClientInit - SSE"}
	clientSSE, _ := NewClient(cfgSSE)
	var info *InfoServer
	info, caseSSE.ErrorActual = clientSSE.Init(context.Background())
	caseSSE.Assert(t, nil)
	defer clientSSE.Close()
	fmt.Printf("info: %+v\n", *info)
	// stdio
	caseStdio := testtool.AssertCase{Description: "TestFClientInit - stdio"}
	clientStdio, _ := NewClient(cfgStdio)
	info, caseStdio.ErrorActual = clientStdio.Init(context.Background())
	caseStdio.Assert(t, nil)
	defer clientStdio.Close()
	fmt.Printf("info: %+v\n", *info)
}

func TestClientListTools(t *testing.T) {
	caseSSE := testtool.AssertCase{A: "save_name", Description: "ClientListTools - SSE"}
	clientSSE, _ := NewClient(cfgSSE)
	ctx := context.Background()
	clientSSE.Init(ctx)
	var tools []mcp.Tool
	tools, caseSSE.ErrorActual = clientSSE.ListTools(ctx)
	caseSSE.B = tools[0].Name
	caseSSE.Assert(t, assert.Equal)
	defer clientSSE.Close()

	// stdio
	caseStdio := testtool.AssertCase{A: "save_name", Description: "ClientListTools - stdio"}
	clientStdio, _ := NewClient(cfgStdio)
	ctx = context.Background()
	clientStdio.Init(ctx)
	tools, caseStdio.ErrorActual = clientStdio.ListTools(ctx)
	caseStdio.B = tools[0].Name
	caseStdio.Assert(t, assert.Equal)
	defer clientStdio.Close()
}

func TestClientListPrompts(t *testing.T) {
	caseSSE := testtool.AssertCase{A: "sql_query_builder", Description: "ClientListPrompts - SSE"}
	clientSSE, _ := NewClient(cfgSSE)
	ctx := context.Background()
	clientSSE.Init(ctx)
	var prompts []mcp.Prompt
	prompts, caseSSE.ErrorActual = clientSSE.ListPrompts(ctx)
	fmt.Printf("prompts: %+v\n", prompts)

	// get prompt
	msgs, err := clientSSE.GetPrompt(ctx, "sql_query_builder", map[string]string{"table": "users"})
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("msgs: %+v\n", msgs)
	caseSSE.B = prompts[0].Name
	caseSSE.Assert(t, assert.Equal)
	defer clientSSE.Close()
}

func TestClientCallTool(t *testing.T) {
	// sse
	caseSSE := testtool.AssertCase{Description: "ClientCallTool - SSE"}
	ctx := context.Background()
	clientSSE, _ := NewClient(cfgSSE)
	clientSSE.Init(ctx)
	args := map[string]interface{}{
		"name": "binbinbin",
	}
	caseSSE.B = "binbinbin"
	var contents []mcp.Content
	contents, caseSSE.ErrorActual = clientSSE.CallTool(ctx, "save_name", args)
	caseSSE.A = contents[0].(mcp.TextContent).Text
	caseSSE.Assert(t, assert.Contains)
	fmt.Printf("contents: %+v\n", contents)
	// stdio
	caseStdio := testtool.AssertCase{Description: "ClientCallTool - stdio"}
	ctx = context.Background()
	clientStdio, _ := NewClient(cfgStdio)
	clientStdio.Init(ctx)
	caseStdio.B = "binbinbin"
	contents, caseStdio.ErrorActual = clientStdio.CallTool(ctx, "save_name", args)
	caseStdio.A = contents[0].(mcp.TextContent).Text
	caseStdio.Assert(t, assert.Contains)
	fmt.Printf("contents: %+v\n", contents)
}

func TestClientEnable(t *testing.T) {
	clientSSE, _ := NewClient(cfgSSE)
	// client 預設開啟
	testtool.AssertCase{A: false, B: clientSSE.cfg.Disabled,
		Description: "TestClientEnable - default",
	}.Assert(t, assert.Equal)
	// disable
	clientSSE.Enable(false)
	testtool.AssertCase{A: true, B: clientSSE.cfg.Disabled,
		Description: "TestClientEnable - disable",
	}.Assert(t, assert.Equal)
	// enable
	clientSSE.Enable(true)
	testtool.AssertCase{A: false, B: clientSSE.cfg.Disabled,
		Description: "TestClientEnable - disable",
	}.Assert(t, assert.Equal)
}
