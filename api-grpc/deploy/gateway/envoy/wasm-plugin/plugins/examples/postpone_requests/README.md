# intro
學習如何用 在envoy 中 做多個 dispatches,, 並先 puase 在正確都收到後才進行 response

# setup and config
- 設定包含以下
  - dispatchCluster[string]: 設定做 dispatch 的 clusterName
  - dispatchCount[int]: 會發出設定數目的dispatch

compile and run
```shell
# 指定資料夾進行compile
pluginDir=plugins/examples/postpone_requests
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
- OnPluginStart: 解析 pluginContext 並存到 config, 如果出錯就用預設值
- OnTick: 如果有 postponed, 就會把第一個拿出來，設定 proxywasm 到那個ctx, 做 ResumeHttpRequest, 如果還有 postponed, 就會直接拿下一個，看起來會在時間到時把在 queue 裡面的ctx 都一次一一地放出來
  - ResumeHttpRequest: 會直接 route 到下一個, 所以如果在 OnHttpRequestHeaders 做pause, resume 會直接跳過 OnHttpRequestBody

HttpContext: 真正地處理每次 http request 的 scope
- OnHttpRequestHeaders: 將 contextID 放到 pluginContext 的 postponed 就先停著, 等 OnTick 叫到所屬ctx 之後才會繼續

# enovy yaml structure
```shell
```
- 也可以把 httpbin 改成內部的靜態回覆 `staticreply`, 就不會檢查 header 直接回 200 了

