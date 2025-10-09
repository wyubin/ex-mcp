# intro
基於 http header 來進行 request or response 的 rewrite
```shell
# 預設會修改 request body，log 會有 "original request body"
curl -XPUT localhost:18000 --data '[initial body]' -H "buffer-operation: prepend"
# 不改 request, 而是在 response 時修改 response body
curl -XPUT localhost:18000 --data '[initial body]' -H "buffer-operation: prepend" -H "buffer-replace-at: response"
# 如果不指定 header, 則預設為 replace body
curl -XPUT localhost:18000 --data '[initial body]'
``` 

# main.go structure
VMContext ──▶ PluginContext ──▶ HttpContext
- VMContext：對應到整個 WASM VM 的生命周期（通常只有一個）
- PluginContext：對應到 Envoy 中的 plugin 實例（可能有多個）
- HttpContext：對應到每個 HTTP stream（每次 request-response 流程）

```scss
Envoy (Request)
   │
   ▼
[WASM VM]
   │
   ├─> vmContext
   │    └─> 建立 pluginContext
   │
   ├─> pluginContext
   │    ├─> 解析 plugin config ("echo" or default)
   │    └─> 決定使用 echoBodyContext 或 setBodyContext
   │
   └─> HttpContext (每個 request)
        ├─> OnHttpRequestHeaders()
        ├─> OnHttpRequestBody()
        ├─> OnHttpResponseHeaders()
        └─> OnHttpResponseBody()

```

# enovy yaml structure
```yaml
┌─────────────────────┐
│        Client        │
│ curl localhost:18000 │
└─────────┬───────────┘
          │
          ▼
   ┌──────────────────────────────┐
   │  Envoy Listener: main:18000  │
   │  Filter Chain:               │
   │    - http_connection_manager │
   │    - wasm (config=body-set)  │  ← main.wasm (修改 body)
   │    - router                  │
   └─────────────┬────────────────┘
                 │
                 ▼
          [Cluster: echo]
                 │
                 ▼
   ┌──────────────────────────────┐
   │  Envoy Listener: echo:38140  │
   │  Filter Chain:               │
   │    - http_connection_manager │
   │    - wasm (config=echo)      │  ← main.wasm (echo 模式)
   │    - router                  │
   └─────────────┬────────────────┘
                 │
                 ▼
           [Cluster: admin] (port 8001)
```