# intro
基於設定檔來決定是否要隨機換到另一個 route 到哪一個 cluster

## setup
```shell
# 指定資料夾進行compile
pluginDir=plugins/examples/http_routing
pluginPath=$(pwd)/${pluginDir}
docker run --rm -v ${pluginPath}:/workspace go-wasm-builder-exam

# 指定資料夾來進行 envoy 服務
FOLDER_PLUGIN=${pluginPath} docker-compose up
```

## example
```shell
# 如果在 http_filters 的 configuration 沒有設定就會隨機 到 "-canary"
# 在 configuration 上設定的數值，整除於 2 就會固定route 到 "-canary", 不然就維持原來 route
# 在 HTTP/2 協議裡，:authority 是一個 pseudo-header，相當於傳統 HTTP/1.1 的 Host header。
# 它用來告訴 上游伺服器（upstream server）這個請求是針對哪個主機。
curl localhost:18000
```

# main.go structure
VMContext ──▶ PluginContext ──▶ HttpContext
- 在 PluginContext 時可以依據設定來決定如何 new 一個 HttpContext(httpRouting)
- httpRouting 中則是只實作 `OnHttpRequestHeaders`
  - 如果 diceOverride % 2 == 0 就會 ReplaceHttpRequestHeader(":authority", value) 到 value += "-canary"

# enovy yaml structure
```shell
                  +----------------------------+
                  |        Envoy Proxy         |
                  | (port: 18000, listener main)|
                  +-------------+--------------+
                                |
                                v
                    [HttpConnectionManager]
                                |
                      +----------------------+
                      | virtual_hosts 判斷域名 |
                      +----------------------+
                        |                 |
                domain: *-canary       domain: *
                        |                 |
                        v                 v
                   cluster: canary   cluster: primary
                        |                 |
            +-----------------+   +-----------------+
            | 127.0.0.1:31000 |   | 127.0.0.1:8099  |
            |  staticreply_canary |  staticreply     |
            | returns "hello..."  | returns "hello..."|
            +-----------------+   +-----------------+
```
- 用 listeners 設定好兩個靜態回覆的 `socket_address` 分別是 `staticreply` 跟 `staticreply_canary`
- 然後用 clusters 來掛這兩個address, 命名 `primary` 及 `canary`
- 然後 `main` listener 在 wasm 處理過後就會依據設定走 :authority 到 "*" or  "*-canary" 然後在 route_config 分流到 `cluster: primary` or `cluster: canary`