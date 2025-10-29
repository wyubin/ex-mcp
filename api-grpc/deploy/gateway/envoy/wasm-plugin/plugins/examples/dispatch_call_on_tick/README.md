# intro
基於設定檔來決定是否要隨機換到另一個 route 到哪一個 cluster

## setup
```shell
# 指定資料夾進行compile
pluginDir=plugins/examples/dispatch_call_on_tick
pluginPath=$(pwd)/${pluginDir}
docker run --rm -v ${pluginPath}:/workspace go-wasm-builder-exam

# 指定資料夾來進行 envoy 服務
FOLDER_PLUGIN=${pluginPath} docker-compose up
```

## example
- 只有 build 到 PluginContext 的部分，因此 wasm 不處理 http context, 只要觀察 envoy log 就好

# main.go structure
VMContext ──▶ PluginContext
- 在 new PluginContext 先把 contextID 設定進去，所以起一個 http.wasm 時，就會 new 一個 PluginContext
- 在啟動時才會做以下動作
  - SetTickPeriodMilliSeconds: 開始設定 ticker
  - 把call back 記在 struct 中，等到 DispatchHttpCall 才會 call
- OnTick: 設定時間到時，會做的事
  - 設定 headers, 然後 trigger DispatchHttpCall
  - DispatchHttpCall 是設定 打 web_service, path 是 /ok or /fail

# enovy yaml structure
```shell
[Client]
   |
   v
+-------------------------+
| Listener: main (0.0.0.0:18000) |
|  └─ HTTP Connection Manager   |
|      ├─ Route Config           |
|      |    └─ Virtual Host "*" |
|      |         └─ Route "/" → cluster: web_service
|      └─ HTTP Filters           |
|           ├─ wasm plugin1      |
|           ├─ wasm plugin2      |
|           └─ router filter     |
+-------------------------+
   |
   v
+-------------------------+
| Cluster: web_service     |
|  └─ Endpoint: 127.0.0.1:8099|
+-------------------------+
   |
   v
+-------------------------+
| Listener: staticreply (127.0.0.1:8099) |
|  └─ HTTP Connection Manager          |
|      ├─ Route Config                 |
|      |    └─ Virtual Host "*"       |
|      |         ├─ Route "/ok"  → 200 "example body" |
|      |         └─ Route "/fail" → 503             |
|      └─ HTTP Filters                   |
|           └─ router filter             |
+-------------------------+

```
- tick 功能看起來是在 plugin context 設定，流程上在 envoy 啟動後就會把 設定的 plugin context gen 起來，並做 OnPluginStart, 所以一次會做個 plugin1 跟 plugin2 兩個 plugin context
- 因為預設會有8 workers, 所以就同時有 8*2 個 ticks 被啟動