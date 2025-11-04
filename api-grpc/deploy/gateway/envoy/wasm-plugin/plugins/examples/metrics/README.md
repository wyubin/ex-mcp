# intro
學習如何用 envoy 中的 metric 模組(內建 prometheus)來進行計算

## setup
```shell
# 指定資料夾進行compile
pluginDir=plugins/examples/metrics
pluginPath=$(pwd)/${pluginDir}
docker run --rm -v ${pluginPath}:/workspace go-wasm-builder-exam

# 指定資料夾來進行 envoy 服務
FOLDER_PLUGIN=${pluginPath} ENVOY_CONCURRENCY=1 docker-compose up
```

## example
- 模組功能為以定義的tag 來進行 request 次數的計算
```shell
# 累計 `custom-tag` 的次數
curl localhost:18000 -v -H "custom-tag: foo"
# 詢問目前 metric 狀態, prometheus 預設會記一些其他的 metric, 因此會grep custom_header_value_counts
curl -s 'localhost:8001/stats/prometheus'| grep custom_header_value_counts
# TYPE custom_header_value_counts counter
custom_header_value_counts{value="foo",reporter="wasmgosdk"} 4

# 因為 dockercompose 沒mount volume, 重開就會重新計算
```

# main.go structure
PluginContext ──▶ HttpContext
pluginContext
- OnPluginStart: GetPluginConfiguration -> parse config -> ctx.config, 然後在 NewHttpContext 放進 HttpContext 

HttpContext
- OnHttpRequestHeaders: 因為資料來源是 request header, 只要實作在這個interface 就可以
  - GetHttpRequestHeader: 拿到需要作為tag 的 value, 到 counters 找是否已經 init
  - proxywasm.DefineCounterMetric: 如果沒有init 要先 DefineCounterMetric 再放進去 counters
  - counter.Increment(1)

# enovy yaml structure
```shell
[Client]
   │  HTTP Request
   ▼
[Listener:18000]
   ▼
[HttpConnectionManager]
   ▼
 ┌──────────────────────────────┐
 │ HTTP Filters Chain            │
 │   1️⃣ wasm filter (main.wasm)  │
 │   2️⃣ router filter            │
 └──────────────────────────────┘
   ▼
[Route Config]
   ▼
[Direct Response 200: "example body\n"]
   ▼
[Client receives response]

```
- 直接在 route_config 裡面也可以定義 direct_response, 不一定要指定一個 cluster(除非wasm 內部要用)

