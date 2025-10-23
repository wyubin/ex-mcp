# intro
用 proxy-wasm 的範例來學習如何 build envoy 的 plugin

# learn
大部分的 wasm plugin 都是把資料中的 main.go build 成 wasm 後，在 envoy 中藉由 envoy.yaml 讀入
```shell
# create docker image for build wasm
prjDir=deploy/gateway/envoy/wasm-plugin
cd $prjDir
docker build -t go-wasm-builder-exam -f Dockerfile.build-wasm .

# 指定資料夾進行compile
pluginDir=plugins/examples/http_body
pluginPath=$(pwd)/${pluginDir}
docker run --rm -v ${pluginPath}:/workspace go-wasm-builder-exam

# 指定資料夾來進行 envoy 服務
FOLDER_PLUGIN=${pluginPath} docker-compose up
```

# ref
[examples](https://github.com/proxy-wasm/proxy-wasm-go-sdk/tree/main/examples)