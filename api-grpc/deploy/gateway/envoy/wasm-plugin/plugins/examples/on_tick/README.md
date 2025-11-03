# intro
最簡單的 tick 範例

## setup
```shell
# 指定資料夾進行compile
pluginDir=plugins/examples/on_tick
pluginPath=$(pwd)/${pluginDir}
docker run --rm -v ${pluginPath}:/workspace go-wasm-builder-exam

# 指定資料夾來進行 envoy 服務
FOLDER_PLUGIN=${pluginPath} ENVOY_CONCURRENCY=1 docker-compose up
```

## example
- 只有 build 到 PluginContext 的部分，因此 wasm 不處理 http context, 只要觀察 envoy log 就好

# main.go structure
PluginContext
- 如果 VMContext 沒什麼特別的操作，直接用 proxywasm.SetPluginContext 直接取代 func (*vmContext) NewPluginContext 也是可以


# enovy yaml structure
```shell
        ┌────────────────────────────┐
        │        Client (curl)       │
        │   curl http://localhost:18000/  │
        └──────────────┬─────────────┘
                       │
                       ▼
        ┌────────────────────────────┐
        │ Listener: main             │
        │ Port: 18000                │
        └──────────────┬─────────────┘
                       │
                       ▼
        ┌────────────────────────────┐
        │ FilterChain:               │
        │ envoy.http_connection_manager │
        └──────────────┬─────────────┘
                       │
                       ▼
        ┌────────────────────────────┐
        │ HTTP Filter #1             │
        │ envoy.filters.http.wasm    │
        │  → 執行 /etc/envoy/plugin/main.wasm │
        │  → 可攔截: OnRequestHeaders / OnResponseBody 等 │
        └──────────────┬─────────────┘
                       │
                       ▼
        ┌────────────────────────────┐
        │ HTTP Filter #2             │
        │ envoy.filters.http.router  │
        └──────────────┬─────────────┘
                       │
                       ▼
        ┌────────────────────────────┐
        │ Route Config: local_service│
        │ match prefix "/"            │
        │ direct_response:            │
        │   200 + "example body\n"   │
        └──────────────┬─────────────┘
                       │
                       ▼
        ┌────────────────────────────┐
        │ Response → Wasm (onResponse)│
        │ → 回傳到 Client             │
        └────────────────────────────┘

```
