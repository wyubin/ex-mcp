# intro
用 proxy-wasm 的範例來學習如何 build envoy 的 plugin

# learn
大部分的 wasm plugin 都是把資料中的 main.go build 成 wasm 後，在 envoy 中藉由 envoy.yaml 讀入
```shell
# 建立通用 docker image
docker build -t go-wasm-builder-exam -f Dockerfile.build-wasm .

# 指定資料夾進行compile
docker run --rm -v $(pwd)/http_body:/workspace go-wasm-builder-exam

# 指定資料夾來進行 envoy 服務
FOLDER_PLUGIN=./http_body docker-compose up
```

## http_body
