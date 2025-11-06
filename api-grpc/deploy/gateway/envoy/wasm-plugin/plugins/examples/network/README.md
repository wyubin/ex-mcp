# intro
學習如何用 在envoy 中做 network 操作

# setup and config
- 設定包含以下


compile and run
```shell
# 指定資料夾進行compile
pluginDir=plugins/examples/network
pluginPath=$(pwd)/${pluginDir}
docker run --rm -v ${pluginPath}:/workspace go-wasm-builder-exam

# 指定資料夾來進行 envoy 服務
FOLDER_PLUGIN=${pluginPath} ENVOY_CONCURRENCY=1 docker-compose up
```

## example
- 
```shell
# 在 curl 回覆中附加 header
curl --head localhost:18000 --data '[initial body]'

curl -s 'localhost:8001/stats/prometheus'| grep connection_counter
```

# main.go structure
這裡用一個更低階的 context 類型: TcpContext
PluginContext ──▶ TcpContext
pluginContext: 設定好counter 給 TcpContext 用
- NewTcpContext: 目前看起來 TcpContext 定義 counter 是有效果的

TcpContext: 直接實作了 interface 中幾個 func
- onNewConnection: downstream 建立連線時
- OnDownstreamData: client 傳入資料, 會把 header 跟body 都放在一起 -> tcp_proxy:web_service
```plaintext
envoy-1  | POST / HTTP/1.1
envoy-1  | Host: localhost:18000
envoy-1  | User-Agent: curl/8.7.1
envoy-1  | Accept: */*
envoy-1  | Content-Length: 14
envoy-1  | Content-Type: application/x-www-form-urlencoded
envoy-1  | 
envoy-1  | [initial body]
```
- onUpstreamData: 在 tcp_proxy 那邊拿到回應
  - 目前並沒有 metadata, 所以log `failed to get upstream location metadata`
- 再來才會把 socket 關掉， trigger OnDownstreamClose -> OnStreamDone
- 跟 httpcontext 不同， data 就是 header + body 一大包
- 在 OnDownstreamData/OnUpstreamData 一樣，有讀資料時 endOfStream 就是 false, 會讀完後再 trigger 一次 OnUpstreamData with endOfStream[true], 但那時 dataSize 會是 0

# enovy yaml structure
```shell
             ┌────────────────────────────────────────────────────┐
             │                    Downstream                      │
             │                Client (curl / TCP)                  │
             └────────────────────────────────────────────────────┘
                                   │
                                   ▼
           ┌────────────────────────────────────────────────────┐
           │                Envoy Listener :18000                │
           │----------------------------------------------------│
           │ [1] envoy.filters.network.wasm                     │
           │     • 初始化 WASM VM (/etc/envoy/plugin/main.wasm) │
           │     • 可攔截/修改 downstream data                  │
           │                                                    │
           │ [2] envoy.tcp_proxy                                │
           │     • 將連線導向 cluster:web_service               │
           │     • 建立 upstream 連線                          │
           └────────────────────────────────────────────────────┘
                                   │
                                   ▼
             ┌────────────────────────────────────────────────┐
             │              Cluster : web_service              │
             │------------------------------------------------│
             │  • Type: STATIC                                 │
             │  • Endpoint: 127.0.0.1:8099                    │
             │  • Metadata: aws / ap-northeast-1a              │
             └────────────────────────────────────────────────┘
                                   │
                                   ▼
             ┌────────────────────────────────────────────────┐
             │           Envoy Listener :8099 (HTTP)           │
             │------------------------------------------------│
             │  envoy.http_connection_manager                 │
             │   ├─ VirtualHost: local_service                │
             │   ├─ Route: prefix "/"                         │
             │   └─ direct_response:                          │
             │        status: 200                             │
             │        body: "example body\n"                  │
             │  envoy.filters.http.router                     │
             └────────────────────────────────────────────────┘
                                   │
                                   ▼
             ┌────────────────────────────────────────────────┐
             │                 Downstream Response             │
             │                 "example body\n"                │
             └────────────────────────────────────────────────┘

```
- 在 tcp context 的 filter_chains 中，是可以直接設定 envoy.filters.network.wasm 然後接 envoy.tcp_proxy

