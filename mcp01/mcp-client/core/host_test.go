package core

import (
	"slices"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

var (
	hostDefault = NewHost()
)

func TestHostSetClient(t *testing.T) {
	err := hostDefault.SetClient("sseTest", cfgSSE)
	assert.NoError(t, err, "TestHostSetClient")
}

func TestHostGetClient(t *testing.T) {
	hostDefault.SetClient("sseTest", cfgSSE)
	client := hostDefault.GetClient("stdio")
	assert.True(t, client == nil)
	client = hostDefault.GetClient("sseTest")
	assert.Equal(t, cfgSSE, client.Config())
}

func TestHostListTools(t *testing.T) {
	hostDefault.SetClient("sseTest", cfgSSE)
	tools, err := hostDefault.ListTools()
	assert.NoError(t, err)
	idx := slices.IndexFunc(tools, func(x mcp.Tool) bool { return x.Name == "sseTest.save_name" })
	assert.NotEqual(t, -1, idx, "TestHostListTools")
}

func TestHostCallTool(t *testing.T) {
	hostDefault.SetClient("sseTest", cfgSSE)
	args := map[string]interface{}{
		"name": "binbinbin",
	}
	_, err := hostDefault.CallTool("sse.save_name", nil)
	assert.ErrorIs(t, err, ErrMcpHostClientNotExist)
	_, err = hostDefault.CallTool("sseTest.save_world", nil)
	assert.Error(t, err)
	rawContents, err := hostDefault.CallTool("sseTest.save_name", args)
	assert.NoError(t, err)
	content := rawContents[0].(mcp.TextContent)
	assert.Contains(t, content.Text, "binbinbin")
	hostDefault.GetClient("sseTest").Enable(false)
	_, err = hostDefault.CallTool("sseTest.save_name", args)
	assert.ErrorIs(t, err, ErrMcpHostClientDisabled)
}
