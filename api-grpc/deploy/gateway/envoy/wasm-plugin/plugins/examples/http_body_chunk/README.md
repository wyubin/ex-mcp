# intro
學習如何用 envoy wasm 操作 request body chunk by chunk 地進行回應

## setup
```shell
# 指定資料夾進行compile
pluginDir=plugins/examples/http_body_chunk
pluginPath=$(pwd)/${pluginDir}
docker run --rm -v ${pluginPath}:/workspace go-wasm-builder-exam

# 指定資料夾來進行 envoy 服務
FOLDER_PLUGIN=${pluginPath} ENVOY_CONCURRENCY=1 docker-compose up
```

## example
- 先準備一堆大檔案到 /tmp/file.txt, 然後送給 localhost:18000 進行處理
```shell
head -c 700000 /dev/urandom | base64 > /tmp/file.txt && echo "pattern" >> /tmp/file.txt
curl 'localhost:18000/anything' -d @/tmp/file.txt
# 每次chunk 的長度不同，所以會找到 pattern 字串的 chunk num 會改變
# 如果不帶 -d body, 連 OnHttpRequestBody 都不會 trigger, 就會直接到 admin 也沒有 wasm log
head -c 700000 /dev/urandom | base64 > /tmp/file-no-pattern.txt
curl 'localhost:18000/anything' -d @/tmp/file-no-pattern.txt
```

# main.go structure
VMContext ──▶ PluginContext ──▶ HttpContext
- 如果 VMContext 沒什麼特別的操作，直接用 proxywasm.SetPluginContext 直接取代 func (*vmContext) NewPluginContext 也是可以
- PluginContext 的 OnPluginStart 中，如果沒有要讀設定，也可以不用實作直接用 default
- HttpContext
  - totalRequestBodyReadSize: 記錄已處理的 chunk 長度
  - receivedChunks: 紀錄已處理的 chunk 次數
  - OnHttpRequestBody:
    - bodySize 是目前 已 request 的 body 長度，所以此次 chunkSize := bodySize - ctx.totalRequestBodyReadSize
    - chunkSize > 0 就會先記錄 ctx.receivedChunks++, 用 GetHttpRequestBody 拿 chunkSize 到 chunk
    - 再 更新 ctx.totalRequestBodyReadSize, 並確認 chunk 裡面有沒有 "pattern"
      - 有 "pattern" 的話，會跑到 types.ActionPause, 並直接送 SendHttpResponse 403, 並寫body 為 <patternFound>
    - 沒有 pattern 則會確認 endOfStream 做 types.ActionPause
    - 如果 endOfStream 已經 true 也沒有 pattern, 就會 log "pattern not found"
    - 如果 body 有 pattern 就會在最後一個 stream 回403, 所以不會走到 endOfStream
    - 在 OnHttpRequestBody 中，endOfStream 為true 的那次 OnHttpRequestBody, 當次 chunk 為空，因為已經在前一次讀完

# enovy yaml structure
```shell
          ┌────────────────────────────────────────────┐
          │                Client (Browser/curl)       │
          │            HTTP Request to :18000          │
          └────────────────────────────────────────────┘
                                │
                                ▼
 ┌──────────────────────────────────────────────────────────────┐
 │                   Envoy Listener (:18000)                    │
 │   (listener name: main, protocol: HTTP via HttpConnectionMgr)│
 └──────────────────────────────────────────────────────────────┘
                                │
                                ▼
 ┌──────────────────────────────────────────────────────────────┐
 │          Http Connection Manager (stat_prefix=ingress_http)  │
 │                                                              │
 │  http_filters chain:                                          │
 │   1️⃣ envoy.filters.http.wasm                                 │
 │         ↳ load /etc/envoy/plugin/main.wasm                    │
 │         ↳ execute plugin with config "body-set"               │
 │   2️⃣ envoy.filters.http.router                               │
 │         ↳ route request to cluster "admin"                    │
 └──────────────────────────────────────────────────────────────┘
                                │
                                ▼
 ┌──────────────────────────────────────────────────────────────┐
 │                   Cluster: admin                             │
 │   type: strict_dns                                            │
 │   connect_timeout: 5000s                                      │
 │   load_assignment:                                            │
 │     0.0.0.0:8001                                              │
 └──────────────────────────────────────────────────────────────┘
                                │
                                ▼
 ┌──────────────────────────────────────────────────────────────┐
 │             Upstream Admin Service (:8001)                   │
 │   (Envoy admin interface / custom service)                   │
 └──────────────────────────────────────────────────────────────┘
                                │
                                ▼
          ┌────────────────────────────────────────────┐
          │                HTTP Response                │
          │     returned through Router → Wasm → Client │
          └────────────────────────────────────────────┘

```

