# intro
試著建立一個 mcp

# compile
```shell
go build -o bin/mcp01-server main.go

# 也可以試著用 go install 來安裝
GOBIN="$HOME/go/bin" go install github.com/wyubin/ex-mcp/mcp01
```

# usage
執行檔預設以 stdio 的模式執行，因此應用程式啟動後就直接從stdin 輸入

```shell
./bin/mymcp-server
# list tools
{"jsonrpc": "2.0", "method": "tools/list", "params": {}, "id": 1}
# call tool
{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "hello_world", "arguments": {"name": "yubin"}}, "id": 1}
```

也可以用 SSE 做

```shell
# 啟動服務
./bin/mcp01-server -t sse
# 設定mcp server 時要用 `http://localhost:8081/sse`
```