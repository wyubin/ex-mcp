# intro
建立 enovy plugin 開發環境，主要以golang compile 成 wasm 的 plugin 提供 envoy 作為 plugin 使用，會需要兩步驟開發

大部分的 wasm plugin 都是把資料中的 main.go build 成 wasm 後，在 envoy 中藉由 envoy.yaml 讀入
```shell
# 建立通用 docker image
prjDir=deploy/gateway/envoy/wasm-plugin
cd $prjDir
docker build -t go-wasm-builder-exam -f Dockerfile.build-wasm .

# 指定資料夾進行compile
pluginDir=plugins/manual
pluginPath=$(pwd)/${pluginDir}
docker run --rm -v ${pluginPath}:/workspace go-wasm-builder-exam

# 指定資料夾來進行 envoy 服務
FOLDER_PLUGIN=${pluginPath} docker-compose up
```

# learn
## use grpc_json_transcoder
使用 grpc_json_transcoder 直接把 http json 制式化轉成 grpc request
```shell
# 轉出 user.pb
buf build --path src/proto/user \
 -o deploy/gateway/envoy/wasm-plugin/plugins/grpc/user.pb

# 並加入對應設定檔 並啟動 docker compose

# test (需要在 proto 加上 option(api) 設定)
curl -X POST -v http://localhost:18000/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'
```

## by wasm route
如果沒設定 proto annotation 基本上還是沒辦法做 route, 所以看起來還是只能自己寫來做 http -> grpc, 要試試看才知道限制在哪裡
```shell
# 將 proto 轉出的 golang pb script 複製到 plugin, 看起來只要把 *.pb.go 複製過去就好
cd deploy/gateway/envoy/wasm-plugin/plugins/manual
# compile wasm
docker build -t go-wasm-builder -f Dockerfile.build-wasm .
docker run --rm -v $(pwd):/app go-wasm-builder cp /var/main.wasm /app/main.wasm
# run docker compose
docker-compose up
```