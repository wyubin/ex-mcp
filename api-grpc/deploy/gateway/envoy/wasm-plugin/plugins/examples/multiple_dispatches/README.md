# intro
學習如何用 在envoy 中 做多個 dispatches,, 並先 puase 在正確都收到後才進行 response

# setup and config
- 設定包含以下
  - dispatchCluster[string]: 設定做 dispatch 的 clusterName
  - dispatchCount[int]: 會發出設定數目的dispatch

compile and run
```shell
# 指定資料夾進行compile
pluginDir=plugins/examples/multiple_dispatches
pluginPath=$(pwd)/${pluginDir}
docker run --rm -v ${pluginPath}:/workspace go-wasm-builder-exam

# 指定資料夾來進行 envoy 服務
FOLDER_PLUGIN=${pluginPath} ENVOY_CONCURRENCY=1 docker-compose up
```

## example
- 模組功能為以定義的tag 來進行 request 次數的計算
```shell
# 在 curl 回覆中附加 header
curl --head localhost:18000
```

# main.go structure
PluginContext ──▶ HttpContext
pluginContext: 通常都是做好載入設定的工作, 是有可能有多個不同設定的 單一wasm plugin 被啟動
- OnPluginStart: 解析 pluginContext 並存到 config

HttpContext: 真正地處理每次 http request 的 scope
- OnHttpResponseHeaders: 是在 response 進來後才做 DispatchHttp, 直接以 for loop 呼叫數次 DispatchHttpCall, 同時累加 pendingDispatchedRequest
  - 設計上用 httpContext 接上 dispatchCallback, 就可以有共同的 pendingDispatchedRequest 狀態來進行改變, 就可以確認每個都 到call back 再做 AddHttpResponseHeader -> ResumeHttpResponse

# enovy yaml structure
```shell
                 ┌────────────────────────────┐
                 │        Client (HTTP)       │
                 │ e.g., curl localhost:18000 │
                 └──────────────┬─────────────┘
                                │
                                ▼
              ┌─────────────────────────────────────┐
              │           Envoy Listener             │
              │         0.0.0.0:18000 (main)        │
              └─────────────────────────────────────┘
                                │
                                ▼
     ┌──────────────────────────────────────────────────────────┐
     │  HTTP Connection Manager (envoy.http_connection_manager) │
     │  stat_prefix: ingress_http                               │
     └──────────────────────────────────────────────────────────┘
                                │
              ┌─────────────────┴────────────────┐
              ▼                                  ▼
   ┌──────────────────────┐         ┌──────────────────────────────┐
   │ HTTP Filter #1: WASM │         │ HTTP Filter #2: Router       │
   │ (envoy.filters.http. │         │ (envoy.filters.http.router)  │
   │ wasm)                │         │ routes to cluster httpbin    │
   └──────────────────────┘         └──────────────────────────────┘
              │
              │ WASM plugin (/etc/envoy/plugin/main.wasm)
              │  → 讀取設定：
              │     {
              │       "dispatchCluster": "httpbin",
              │       "dispatchCount": 3
              │     }
              ▼
      (可能對 downstream request 做攔截/修改/額外 dispatch)

                                │
                                ▼
              ┌────────────────────────────────────┐
              │        Cluster: httpbin            │
              │  type: strict_dns                  │
              │  lb_policy: round_robin            │
              │  endpoint: httpbin.org:80          │
              └────────────────────────────────────┘
                                │
                                ▼
                 ┌────────────────────────────┐
                 │       httpbin.org:80       │
                 │ (external upstream target) │
                 └────────────────────────────┘

```
- 也可以把 httpbin 改成內部的靜態回覆 `staticreply`, 就不會檢查 header 直接回 200 了

