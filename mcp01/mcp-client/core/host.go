package core

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// 實作 host 控制每個 client 的操作，也會將不同 client 的工具描述再一次輸出給外部
type Host struct {
	clients map[string]*Client
}

// new
func NewHost() *Host {
	inst := Host{
		clients: map[string]*Client{},
	}
	return &inst
}

// add single server with name
// 如果有重複名字的server 就會 overwrite
func (s *Host) SetClient(name string, cfgServ CfgServer) error {
	client, err := NewClient(cfgServ)
	if err != nil {
		return err
	}
	s.clients[name] = client
	return nil
}

// Get Client, 如果是 nil 代表不存在
func (s *Host) GetClient(name string) *Client {
	return s.clients[name]
}

// ListTools
func (s *Host) ListTools() ([]mcp.Tool, error) {
	toolsAll := []mcp.Tool{}
	errs := []error{}
	for name, client := range s.clients {
		if client.cfg.Disabled {
			continue
		}
		ctx := context.Background()
		_, err := client.Init(ctx)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		tools, err := client.ListTools(ctx)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		for _, tool := range tools {
			tool.Name = strings.Join([]string{name, tool.Name}, ".")
			toolsAll = append(toolsAll, tool)
		}
	}
	return toolsAll, errors.Join(errs...)
}

// CallTool
func (s *Host) CallTool(name string, args map[string]interface{}) ([]mcp.Content, error) {
	names := strings.SplitN(name, ".", 2)
	// check nameServer exists
	client := s.GetClient(names[0])
	if client == nil {
		return nil, fmt.Errorf("%w -> client[%s]", ErrMcpHostClientNotExist, names[0])
	}
	ctx := context.Background()
	_, err := client.Init(ctx)
	if err != nil {
		return nil, err
	}
	return client.CallTool(ctx, names[1], args)
}
