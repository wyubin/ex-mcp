package mcp

import "errors"

var (
	ErrMcpClientNew       = errors.New("create mcp client error")
	ErrMcpClientStart     = errors.New("start mcp client fail")
	ErrMcpClientInit      = errors.New("init mcp client fail")
	ErrMcpClientNoTools   = errors.New("mcp has no tools")
	ErrMcpClientNoPrompts = errors.New("mcp has no prompt")

	ErrMcpHostClientNotExist = errors.New("client not exist")
	ErrMcpHostClientDisabled = errors.New("client has been disabled")

	ErrInValidCfgServers = errors.New("not valid config")
)
