package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
)

type TransportType string

const (
	TransportSSE   TransportType = "sse"
	TransportSTDIO TransportType = "stdio"
)

// InfoServer, 列出 server 資訊
type InfoServer mcp.Implementation

// 針對 sse/stdio 兩種模式的共通 struct
type CfgServer struct {
	Disabled      bool              `json:"disabled"`
	Timeout       int               `json:"timeout"`
	Command       string            `json:"command,omitempty"` // stdio 專用
	Args          []string          `json:"args,omitempty"`    // stdio 專用
	Env           map[string]string `json:"env,omitempty"`     // stdio 專用
	Url           string            `json:"url,omitempty"`     // sse 專用
	TransportType TransportType     `json:"transportType"`     // stdio 或 sse
}

type CfgServers map[string]CfgServer
