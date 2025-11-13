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
  - callback 將 grpc response body 讀出先存到 httpContext 的 property
  - OnHttpResponseHeaders: 如果 callback 有拿到 body(可能要在 callback 做 grpc body -> json body) 就要改 header
  - OnHttpResponseBody: 如果 callback 有拿到 body 就要改 response

## 重構
### plan
- 會需要一個 service 把對應的設定讀進去後(每個 grpc 的cluster name), 放到 http context 作為工具使用，用來轉換 http <--> grpc 的 request/response，至少需要以下介面
  - GetGrpcPath and query[OnHttpRequestHeaders]: 基於http 的 method 跟 path 來對應到 grpc 要打的 clusterName, Path, 跟 querystring
  - GenGrpcRequestBody[OnHttpRequestBody]: 依據對應的 grpcPath 跟 QueryString 跟 httpRequestBody, 產生 GrpcRequestBody
  - GenHttpResponseBody[DispatchHttpCall]: 在打 grpc 之後，將其 response body 轉為預計想要的 http response body 來回應, 存到 http context
    - callback 裡面需要處理是否轉 httpBody 成功的錯誤處理, 也要處理 call grpc 的 status 來進行確認，來決定是否處理 body 及 怎麼 SendHttpResponse(有 grpc --> http status code, 並且要抓 grpc-message)
- 在 callback 如果可以把各種狀況都 `SendHttpResponse` 的話，應該可以直接不用實作 OnHttpResponseHeaders/OnHttpResponseBody, 反正沒有要處理 admin 回傳的東西
