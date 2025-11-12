# intro
要把 restful api 的 api call 轉成能夠傳到後方 grpc server 的 request, 並回傳 json response 到 response body 中

# setup and config
- 設定包含以下

compile and run
```shell
# 指定資料夾進行compile
pluginDir=plugins/manual
pluginPath=$(pwd)/${pluginDir}
docker run --rm -v ${pluginPath}:/workspace go-wasm-builder-exam

# 指定資料夾來進行 envoy 服務
FOLDER_PLUGIN=${pluginPath} ENVOY_CONCURRENCY=1 docker-compose up
```

## example
```shell
# 先測試 user grpc 的回應
grpcurl \
  -plaintext \
  -d '{"name": "Alice", "email": "alice@example.com"}' \
  localhost:50051 \
  user.UserService/CreateUser

# 再打 envoy port 來串到 grpc
curl -X POST -v http://localhost:18000/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'
```

# issue
目前實作狀況為
- 能夠正確地傳遞到後端 grpc 但無法回傳到 response 的部分

## 實作 20251111
- 在 OnHttpRequestBody 用 dispatch 的方式打grpc, 拿到 response 後做 callback
  - dispatch 失敗就 types.ActionContinue, 不然在 OnHttpRequestBody 最後是先 types.ActionPause
  - callback 將 grpc response body 讀出轉成