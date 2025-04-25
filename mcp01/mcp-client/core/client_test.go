package core

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	urlCfg, _ = url.Parse("http://localhost:8081/sse")
	cfg       = CfgServer{
		URL:           urlCfg,
		TransportType: TransportSSE,
	}
	client = NewClient(cfg)
)

func TestClientInit(t *testing.T) {
	ctx := context.Background()
	info, err := client.Init(ctx)
	assert.NoError(t, err, "TestFClientInit")
	fmt.Printf("info: %+v\n", *info)
}

func TestClientListTools(t *testing.T) {
	ctx := context.Background()
	client.Init(ctx)
	tools, err := client.ListTools(ctx)
	assert.NoError(t, err, "TestFClientListTools")
	fmt.Printf("tools: %+v\n", tools)
}

func TestClientCallTool(t *testing.T) {
	ctx := context.Background()
	client.Init(ctx)
	args := map[string]interface{}{
		"name": "binbinbin",
	}
	contents, err := client.CallTool(ctx, "save_name", args)
	assert.NoError(t, err, "TestClientCallTool")
	fmt.Printf("contents: %+v\n", contents)
}
