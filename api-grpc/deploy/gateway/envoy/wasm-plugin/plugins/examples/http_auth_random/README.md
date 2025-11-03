# intro
另一個 DispatchHttpCall 的應用範例，可以在 OnHttpRequestHeaders 時直接 DispatchHttpCall 到設定好的位置，再以 call back 直接做 response
- 另外，也導入一個 ActionPause 來等 callback 中做 proxywasm.ResumeHttpRequest()

## setup
```shell
# 指定資料夾進行compile
pluginDir=plugins/examples/http_auth_random
pluginPath=$(pwd)/${pluginDir}
docker run --rm -v ${pluginPath}:/workspace go-wasm-builder-exam

# 指定資料夾來進行 envoy 服務
FOLDER_PLUGIN=${pluginPath} ENVOY_CONCURRENCY=1 docker-compose up
```

## example
- 試著打 `localhost:18000/uuid` 到目標位置, 
```shell
curl localhost:18000/uuid -v
```

# main.go structure
VMContext ──▶ PluginContext ──▶ HttpContext
- 如果 VMContext 沒什麼特別的操作，直接用 proxywasm.SetPluginContext 直接取代 func (*vmContext) NewPluginContext 也是可以
- HttpContext 的 OnHttpRequestHeaders 中，先拿 header(如果沒有 header 可能就是不合法的 http request), 然後打 DispatchHttpCall 子 request 去做 callback
  - 如果以上成功就會 return types.ActionPause, 在 callback 中再決定要不要 ResumeHttpRequest
  - 如果以上有一步驟失敗，就會 log error 後 types.ActionContinue
  - httpCallResponseCallback 中 發出 async 到 `httpbin` 這個 cluster, 這個 call back 目前看起來要有 後續，可能要 logger 或是直接用 proxywasm.SendHttpResponse 去改response
    - 先 GetHttpCallResponseHeaders
    - 用 GetHttpCallResponseBody 拿body 然後寫入 New32a hash
    - 用 Sum32 確認是否整除 2
      - 是 => log auth pass => ResumeHttpRequest
      - 否 => 403 => 維持 types.ActionPause

```shell
┌────────────────────────────┐
│ Envoy 接收到 HTTP Request │
└──────────────┬─────────────┘
               │
               ▼
   [WASM VM 啟動 plugin]
               │
               ▼
        ┌───────────────────────┐
        │ OnHttpRequestHeaders() │
        └───────────────────────┘
               │
               ▼
     取得 Request Headers
               │
               ▼
     DispatchHttpCall(cluster=httpbin)
               │
               ▼
      暫停原請求(ActionPause)
               │
               ▼
 ┌──────────────────────────────────┐
 │ 等待 httpCallResponseCallback() │
 └──────────────────────────────────┘
               │
               ▼
     取得 Response Headers & Body
               │
               ▼
      對 body 計算 FNV Hash
               │
               ▼
        Hash % 2 == 0 ?
        ├───────────────┬────────────────┐
        │Yes            │No              │
        ▼                ▼
  ResumeHttpRequest()    SendHttpResponse(403 Forbidden)
  (放行原請求)           (攔截回應，結束流程)

```

# enovy yaml structure
```shell
┌──────────────────────────────────────────┐
│               Client                    │
│          (curl / browser)               │
└────────────────────┬────────────────────┘
                     │
                     ▼
        ┌────────────────────────┐
        │ Listener: main (0.0.0.0:18000) │
        └───────────────┬────────────────┘
                        ▼
        ┌────────────────────────────────┐
        │ HttpConnectionManager (ingress_http) │
        │  • route_config: httpbin              │
        │  • virtual_host: *                    │
        └────────────────┬──────────────────────┘
                         ▼
         ┌────────────────────────────────┐
         │ HTTP Filter Chain               │
         │---------------------------------│
         │ 1️⃣ envoy.filters.http.wasm     │
         │     → 載入 /etc/envoy/plugin/main.wasm │
         │     → 執行 WASM Plugin (auth/hash 檢查) │
         │---------------------------------│
         │ 2️⃣ envoy.filters.http.router   │
         │     → 根據 route 導向 cluster    │
         └────────────────┬────────────────┘
                          ▼
             ┌───────────────────────────────┐
             │ Cluster: httpbin               │
             │  • type: strict_dns            │
             │  • endpoint: httpbin.org:80    │
             │  • lb_policy: round_robin      │
             └───────────────────────────────┘
                          │
                          ▼
             ┌───────────────────────────────┐
             │       Upstream Service         │
             │         (httpbin.org)          │
             └───────────────────────────────┘
```
- 可以把 main listener 的 route 導到 staticreply, 變成實際上 main 在經過 wasm 處理時，會 dispatch 到 httpbin 做 auth 確認後，在 call back 中 valid auth 再決定是 403 或是resume http
  - auth invalid => set response 403 => keep pause http
  - auth valid => resume http => route to staticreply and get response from staticreply