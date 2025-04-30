# intro
建立一個能夠用 command line 基於一個 config.json 來 access mcp server 的 tool 並試圖 run 他

# spec
以一個 cobra 套件來建立 cmd interface
- 一定會需要 config, 所以唯一 args 是 path of config json
- subcmd: list - 列出目前可用工具
  - 應該不用參數
- subcmd: call - 使用工具及參數
  - 必須輸入 -t, --tool 跟 -a, --args(這個要是 jsonstring)

# compile
```shell
go build -o bin/mcptools mcp-client/cmd/tools-access/main.go
```

# usage
```shell
# check tools
bin/mcptools list tmp/mcp-servers.json | yq -P
# run tools
bin/mcptools call tmp/mcp-servers.json \
  --name-tool 'mcp01.save_name' \
  --params '{"name": "yubin"}'
```