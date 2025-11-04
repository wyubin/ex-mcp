# intro
學習如何用 envoy wasm 結合 envoy yaml 的設定來操作 http header

## setup
```shell
# 指定資料夾進行compile
pluginDir=plugins/examples/http_headers
pluginPath=$(pwd)/${pluginDir}
docker run --rm -v ${pluginPath}:/workspace go-wasm-builder-exam

# 指定資料夾來進行 envoy 服務
FOLDER_PLUGIN=${pluginPath} ENVOY_CONCURRENCY=1 docker-compose up
```

## example
- 先準備一堆大檔案到 /tmp/file.txt, 然後送給 localhost:18000 進行處理
```shell
curl 'localhost:18000' -v
# 會有預設 header, x-proxy-wasm-go-sdk-example/x-wasm-header 則是設定加的
# < x-envoy-upstream-service-time: 1
```

# main.go structure
PluginContext ──▶ HttpContext
pluginContext
- NewHttpContext: 會把 contextId 跟設定的 header name/value 寫到 HttpContext
- OnPluginStart: 進行以下處理，如果error 會 return OnPluginStartStatusFailed
  - config == nil 直接 OnPluginStartStatusOK, headerName/headerValue 就維持空值
  - 將設定 decode 為 PluginConfig
  - 設定 headerName/headerValue, 如果有空值就 fail
HttpContext
- OnHttpRequestHeaders: (其實沒有具體操作) 簡單加上 `test` header 並 log  request headers
- OnHttpResponseHeaders: 在 target response 後, 加上 x-proxy-wasm-go-sdk-example 跟設定的 header

# enovy yaml structure
```shell
        ┌───────────────────────────────┐
        │        Client (HTTP)          │
        └──────────────┬────────────────┘
                       │  (port 18000)
                       ▼
        ┌────────────────────────────────────────────┐
        │          Listener: main (0.0.0.0:18000)    │
        │--------------------------------------------│
        │ FilterChain:                               │
        │  ├─ HttpConnectionManager                  │
        │  │   ├─ Route Config: local_route          │
        │  │   │   └─ Route: "/" → Cluster web_service
        │  │   │
        │  │   └─ HTTP Filters (順序執行):           │
        │  │       1️⃣ envoy.filters.http.wasm        │
        │  │           • 載入 /etc/envoy/plugin/main.wasm
        │  │           • 設定 header = x-wasm-header
        │  │           • 設定 value  = demo-wasm
        │  │       2️⃣ envoy.filters.http.router      │
        │  │           • 將請求轉送至目標 cluster    │
        │  │
        │  └────────────────────────────────────────│
        │
        ▼
┌──────────────────────────────────────────────┐
│           Cluster: web_service               │
│----------------------------------------------│
│ type: STATIC                                 │
│ lb_policy: ROUND_ROBIN                       │
│ endpoints:                                   │
│   127.0.0.1:8099                             │
└──────────────────────────────────────────────┘
        │
        ▼
┌──────────────────────────────────────────────┐
│     Listener: staticreply (127.0.0.1:8099)   │
│----------------------------------------------│
│ HttpConnectionManager                        │
│   ├─ Route: "/" → DirectResponse (200 OK)    │
│   │    body: "example body\n"                │
│   └─ HTTP Filter: envoy.filters.http.router  │
└──────────────────────────────────────────────┘
        │
        ▼
        💬 回傳 Response: "example body\n"

```
- `web_service` 對應到 Listener staticreply, 是靜態回應的 server

