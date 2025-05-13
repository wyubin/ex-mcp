package core

import (
	"context"
	"fmt"

	mcpClient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// 實作 mcp client 內部的操作，主要會在跟 mcp server 互動或是輸出相關工具列表等等
type Client struct {
	cfg    CfgServer
	client *mcpClient.Client
}

// 以 CfgServer init 一個 mcp client(也要確認是否可運作)
func (s *Client) Init(ctx context.Context) (*InfoServer, error) {
	// init
	result, err := s.client.Initialize(ctx, mcp.InitializeRequest{})
	if err != nil {
		return nil, fmt.Errorf("%w -> %s", ErrMcpClientNew, err)
	}
	info := InfoServer{
		Name:    result.ServerInfo.Name,
		Version: result.ServerInfo.Version,
	}
	return &info, nil
}

// Config, return CfgServer 提供外部使用(儲存)
func (s *Client) Config() CfgServer {
	return s.cfg
}

// 設定是否使用 mcp server, 不輸入參數預設要開啟
func (s *Client) Enable(willEnables ...bool) {
	if len(willEnables) == 0 {
		s.cfg.Disabled = false
	} else {
		s.cfg.Disabled = !willEnables[0]
	}
}

// 實作其他與 mcp server 的互動, 需要從 init 那邊拿到ctx, 再從裡面抽 session id 做request
func (s *Client) ListTools(ctx context.Context) ([]mcp.Tool, error) {
	toolListResult, err := s.client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("%w -> %s", ErrMcpClientNoTools, err)
	}
	return toolListResult.Tools, nil
}

// call 完拿 content
func (s *Client) CallTool(ctx context.Context, name string, args map[string]interface{}) ([]mcp.Content, error) {
	request := mcp.CallToolRequest{}
	request.Params.Name = name
	request.Params.Arguments = args

	result, err := s.client.CallTool(ctx, request)
	if err != nil {
		return nil, err
	}

	return result.Content, nil
}

// close transport
func (s *Client) Close() error {
	return s.client.Close()
}

// setup client as sse or stdio
func NewClient(cfg CfgServer) (*Client, error) {
	var client *mcpClient.Client
	var err error
	switch cfg.TransportType {
	case TransportSSE:
		client, err = mcpClient.NewSSEMCPClient(cfg.Url)
	case TransportSTDIO:
		envs := []string{}
		for k, v := range cfg.Env {
			envs = append(envs, fmt.Sprintf("%s=%s", k, v))
		}
		client, err = mcpClient.NewStdioMCPClient(cfg.Command, envs, cfg.Args...)
	}
	if err != nil {
		return nil, fmt.Errorf("%w -> %s", ErrMcpClientNew, err)
	}
	// start
	ctx := context.Background()
	if err := client.Start(ctx); err != nil {
		return nil, fmt.Errorf("%w -> %s", ErrMcpClientStart, err)
	}
	inst := Client{
		cfg:    cfg,
		client: client,
	}
	return &inst, nil
}
