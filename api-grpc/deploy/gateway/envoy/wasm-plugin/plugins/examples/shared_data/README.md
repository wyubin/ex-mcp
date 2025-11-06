# intro
- 

# setup and config
- 設定包含以下

compile and run
```shell
# 指定資料夾進行compile
pluginDir=plugins/examples/shared_data
pluginPath=$(pwd)/${pluginDir}
docker run --rm -v ${pluginPath}:/workspace go-wasm-builder-exam

# 指定資料夾來進行 envoy 服務
FOLDER_PLUGIN=${pluginPath} ENVOY_CONCURRENCY=1 docker-compose up
```

## example
- 在同一個 vm context 中可以存取同一個值，每呼叫一次會加1
```shell
curl localhost:18000
```

# main.go structure
VMContext ──▶ PluginContext ──▶ HttpContext
- 
VMContext:
- OnVMStart: 將init value(0) 用 SetSharedData 設定到 sharedDataKey

HttpContext:
- OnHttpRequestHeaders: GetSharedData/SetSharedData 不是 thread-safe, 所以設計一個 cas 來做設定，在 init 之後就必續要用 cas 來做存取
# enovy yaml structure
```shell
┌──────────────────────────────────────────────────────────────┐
│                        Envoy Proxy                           │
│                                                              │
│  Listener: main (0.0.0.0:18000)                              │
│  └── FilterChain                                             │
│      └── envoy.http_connection_manager                       │
│          ├─ stat_prefix: ingress_http                        │
│          ├─ codec_type: auto                                 │
│          ├─ route_config                                     │
│          │   └── virtual_host: local_service                 │
│          │       └── route:                                  │
│          │           match: prefix="/"                       │
│          │           direct_response:                        │
│          │               status: 200                         │
│          │               body: "example body\n"              │
│          │                                                   │
│          └─ http_filters (執行順序自上而下)                  │
│              │                                               │
│              ├── envoy.filters.http.wasm                     │
│              │     ├─ runtime: envoy.wasm.runtime.v8         │
│              │     ├─ code: /etc/envoy/plugin/main.wasm      │
│              │     └─ vm_config.configuration:               │
│              │         { "sharedDataKey": "shared_key_yaml"} │
│              │                                               │
│              └── envoy.filters.http.router                   │
│                    (負責送出 direct_response 結果)           │
│                                                              │
└──────────────────────────────────────────────────────────────┘

```
