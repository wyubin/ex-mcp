# intro
學習如何用 envoy wasm 結合 envoy yaml 的設定來進行 body json 的 validation

## setup
```shell
# 指定資料夾進行compile
pluginDir=plugins/examples/json_validation
pluginPath=$(pwd)/${pluginDir}
docker run --rm -v ${pluginPath}:/workspace go-wasm-builder-exam

# 指定資料夾來進行 envoy 服務
FOLDER_PLUGIN=${pluginPath} ENVOY_CONCURRENCY=1 docker-compose up
```

## example
- 先準備一堆大檔案到 /tmp/file.txt, 然後送給 localhost:18000 進行處理
```shell
curl -X POST localhost:18000 -H 'Content-Type: application/json' --data '{"id": "xxx", "token": "xxx"}' -v
# 收到 hello from the server
curl -X POST localhost:18000 -H 'Content-Type: application/json' --data '{"id": "xxx"}' -v
# invalid payload
curl -X POST localhost:18000 --data '{"id": "xxx", "token": "xxx"}' -v
# content-type must be provided
```

# main.go structure
PluginContext ──▶ HttpContext
pluginContext
- GetPluginConfiguration -> parse config

HttpContext
- OnHttpRequestHeaders: 有檢查 content-type 要是 `application/json`
- OnHttpRequestBody: bodySize 應該就等於已讀進來的 body length, 不需要 totalRequestBodySize 做累加, 直到 endOfStream 再 GetHttpRequestBody 就可以
  - validatePayload 把 body 轉成 map[string]any 再檢查 keys 就結束

# enovy yaml structure
```shell
[ Client ]
     │  (HTTP request)
     ▼
┌─────────────────────────────┐
│ Listener :18000 (main)      │
│  └─ HttpConnectionManager   │
│     ├─ Wasm Filter (驗證 header id/token)
│     └─ Router Filter → web_service
└─────────────────────────────┘
     │
     ▼
┌─────────────────────────────┐
│ Cluster: web_service        │
│ → 127.0.0.1:8099            │
└─────────────────────────────┘
     │
     ▼
┌─────────────────────────────┐
│ Listener :8099 (staticreply)│
│  └─ Direct Response: "hello from the server" │
└─────────────────────────────┘
     │
     ▼
[ Client receives response ]
```
- `web_service` 對應到 Listener staticreply, 是靜態回應的 server

