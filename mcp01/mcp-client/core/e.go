package core

import "errors"

var (
	ErrMcpClientNew     = errors.New("create mcp client error")
	ErrMcpClientStart   = errors.New("start mcp client fail")
	ErrMcpClientInit    = errors.New("init mcp client fail")
	ErrMcpClientNoTools = errors.New("mcp has no tools")

	ErrMcpHostClientNotExist = errors.New("client not exist")
)
