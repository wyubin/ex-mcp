# intro
學習如何用 在envoy 中 做多個 dispatches,, 並先 puase 在正確都收到後才進行 response

## setup
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
curl --head localhost:18000
```

# main.go structure
PluginContext ──▶ HttpContext
pluginContext

HttpContext

# enovy yaml structure
```shell

```
- 直接在 route_config 裡面也可以定義 direct_response, 不一定要指定一個 cluster(除非wasm 內部要用)

