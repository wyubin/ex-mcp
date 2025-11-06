# intro
- 

# setup and config
- 設定包含以下

compile and run
```shell
# 指定資料夾進行compile
pluginDir=plugins/examples/properties
pluginPath=$(pwd)/${pluginDir}
docker run --rm -v ${pluginPath}:/workspace go-wasm-builder-exam

# 指定資料夾來進行 envoy 服務
FOLDER_PLUGIN=${pluginPath} ENVOY_CONCURRENCY=1 docker-compose up
```

## example
- 藉由envoy yaml 的 設定來做 auth, 因此，可以基於各種 metadata 來做設定切換
```shell
# 沒有任何header
curl localhost:18000/one -v
# 401
curl localhost:18000/one -v -H 'cookie: value'
# 200 OK
curl localhost:18000/two -v -H 'cookie: value'
# 401
curl localhost:18000/two -v -H 'authorization: token'
# 200 OK

curl localhost:18000/three -v
# 200 OK
# wasm log: no auth header for route
```

# main.go structure
HttpContext
也可以直接用 SetHttpContext 來直接建立 HttpContext, 但如果有設定還是有一個plugin config 來讀會比較好，不然就要寫死在wasm 了

HttpContext: 真正地處理每次 http request 的 scope
- OnHttpRequestHeaders: 基於 propertyPrefix 設定去試著讀相關的 property, 可以用 GetProperty 或是 GetPropertyMap, 可以參考 https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/advanced/attributes, 通常在 HttpContext 裡面用

# enovy yaml structure
```shell
                         ┌───────────────────────────────┐
                         │         Envoy Proxy           │
                         └───────────────────────────────┘
                                       │
                  ┌────────────────────┴────────────────────┐
                  │                                         │
          ┌──────────────┐                         ┌────────────────┐
          │ Listener:     │                         │ Listener:       │
          │ main (18000)  │                         │ staticreply(8099)│
          └──────────────┘                         └────────────────┘
                  │                                         │
         ┌────────▼─────────┐                     ┌────────▼──────────┐
         │ HTTP ConnManager │                     │ HTTP ConnManager  │
         │ (ingress_http)   │                     │ (ingress_http)    │
         └────────┬─────────┘                     └────────┬──────────┘
                  │                                         │
     ┌────────────▼───────────────┐              ┌──────────▼────────────┐
     │   Route Config (local_route)│              │ Route Config          │
     │   VirtualHost: local_service│              │ VirtualHost: local_service │
     └────────────┬───────────────┘              └──────────┬────────────┘
                  │                                         │
   ┌──────────────▼────────────────────┐          ┌─────────▼────────────────┐
   │ Routes:                           │          │ Route: prefix="/"        │
   │                                   │          │ └─ direct_response:      │
   │ 1️⃣ /one   → cluster:web_service  │          │    └─ 200 OK             │
   │     └─ metadata.auth: cookie      │          │       body:"example body"│
   │                                   │          └──────────────────────────┘
   │ 2️⃣ /two   → cluster:web_service  │
   │     └─ metadata.auth: authorization │
   │                                   │
   │ 3️⃣ /three → cluster:web_service │
   │     └─ no metadata                │
   └──────────────────────────────────┘
                  │
        ┌─────────▼────────────┐
        │ HTTP Filters Chain   │
        ├──────────────────────┤
        │ 1️⃣ Wasm Filter      │
        │    - wasm: /etc/envoy/plugin/main.wasm
        │    - propertyChain:
        │      ["xds","route_metadata",
        │       "filter_metadata","envoy.filters.http.wasm","auth"]
        │    → 讀取 route_metadata.envoy.filters.http.wasm.auth
        │      (例: "cookie" / "authorization")
        │    → 可用於身份驗證 / 驗證邏輯
        │
        │ 2️⃣ Router Filter
        │    → 將請求路由至 cluster:web_service
        └──────────────────────┘
                  │
        ┌─────────▼────────────┐
        │ Cluster: web_service │
        │ type: STATIC         │
        │ lb_policy: ROUND_ROBIN│
        │ connect_timeout:0.25s│
        │                      │
        │ → endpoint: 127.0.0.1:8099 │
        └─────────┬────────────┘
                  │
       ┌──────────▼────────────┐
       │ Listener: staticreply │
       │ → direct_response(200,"example body") │
       └────────────────────────┘

```
- 在 route_config 的 routes 可以每個 match 都自己帶 metadata
