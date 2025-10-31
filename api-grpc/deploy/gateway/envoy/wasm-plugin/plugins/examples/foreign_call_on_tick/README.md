# intro
學習如何從 pluginContext 啟動後, 使用 SetTickPeriodMilliSeconds 來啟動固定時距的資料處理

## setup
```shell
# 指定資料夾進行compile
pluginDir=plugins/examples/foreign_call_on_tick
pluginPath=$(pwd)/${pluginDir}
docker run --rm -v ${pluginPath}:/workspace go-wasm-builder-exam

# 指定資料夾來進行 envoy 服務
FOLDER_PLUGIN=${pluginPath} docker-compose up
```

## example
- 只有 build 到 PluginContext 的部分，因此 wasm 不處理 http context, 只要觀察 envoy log 就好

# vm thread
雖然設定的 concurrency 跟 worker 數目相同，但 pluginContext 的 ticker 基本上數目是 worker num +1 因為以下
```shell
               ┌──────────────────────────────┐
               │       Main Thread            │
               │                              │
               │  Base VM (global)            │
               │  │                           │
               │  └── Thread-Local VM #1      │
               └────────┬──────────┬──────────┘
                        │          │
     ┌──────────────────┘          └──────────────────┐
     ▼                                               ▼
Worker Thread 0                               Worker Thread 1
Thread-Local VM #2                            Thread-Local VM #3
```
Envoy 先在主線程 (main thread) 建立了「Base Wasm VM」。
- 這個 VM 用來載入 .wasm bytecode，初始化 runtime（例如 v8、wavm），
- 並在後續被 worker threads 拿來作為模板 clone。
- 它的目的不是直接處理請求，而是作為 初始化參考 (prototype VM)。

main thread 自身也會有一個 thread-local VM，用於：
- 控制 tick（例如 set tick period milliseconds: 10000）
- 處理不屬於 worker I/O 的事件，例如 timer callback 或 async grpc callback。
- 也就是說，main thread 本身也啟一個 thread-local VM。

每個 worker 是獨立 event loop，因此：
- 各 worker 都會從 Base VM clone 一份自己的 Wasm VM，
- 各自維護 Context（例如 per-thread global data, tick timer, stats）。

三個 Wasm log 是完全正常的現象。

# main.go structure
VMContext ──▶ PluginContext
- 在 new PluginContext 先把 contextID 設定進去，所以起一個 http.wasm 時，就會 new 一個 PluginContext
- 在啟動時才會做以下動作
  - SetTickPeriodMilliSeconds: 開始設定 ticker
- OnTick: 設定時間到時，會做的事
  - 設定 headers, 然後 trigger CallForeignFunction
  - CallForeignFunction 第一個參數是 外部 function name, 由外部直接跟 envoy 註冊，像是 compress 是內建的外部功能，第二個參數是輸入參數，然後會拿到 return byte

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
