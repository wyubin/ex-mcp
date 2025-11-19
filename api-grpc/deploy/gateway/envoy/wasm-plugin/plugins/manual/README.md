# intro
要把 restful api 的 api call 轉成能夠傳到後方 grpc server 的 request, 並回傳 json response 到 response body 中

# setup and config
- 設定包含以下
  - `pluginConfig`: JSON 格式，對應 route name 到 envoy cluster name。
    ```json
    {
      "user": "user_service"
    }
    ```
    - Key ("user"): 對應代碼中 `NewUserGrpcMapper("user")` 的名稱。
    - Value ("user_service"): Envoy 中設定的 Cluster Name。

# Analysis & Functionality
本插件是一個 HTTP 轉 gRPC 的 Transcoder，主要功能是在 Envoy 層將 RESTful API 請求轉換為 gRPC 請求發送到後端服務，並將 gRPC 回應轉換回 JSON 格式回傳給客戶端。

## Architecture
- **Lifecycle (`main.go`)**:
  - 負責插件的初始化 (`OnPluginStart`)，讀取設定並建立路由表。
  - 處理 HTTP 請求生命週期 (`OnHttpRequestHeaders`, `OnHttpRequestBody`, `OnHttpResponseBody`)。
  - 使用 `DispatchHttpCall` 發送 gRPC 請求。
- **Routing (`internal/grpcio`)**:
  - `RequestMapper`: 核心路由邏輯。
  - 支援 Regex 路徑比對 (e.g., `/v1/user/{id}`)。
  - 自動提取路徑參數 (`PathParams`) 和 Query String。
- **Protocol Conversion (`proto/`)**:
  - 針對個別服務 (e.g., `user`) 實作具體的轉換邏輯。
  - `Json2Grpc`: JSON Body -> gRPC Protobuf Binary。
  - `Grpc2Json`: gRPC Protobuf Binary -> JSON Body。

## Workflow
1. **Request Matching (`OnHttpRequestHeaders`)**:
   - 攔截 HTTP Header，比對 Method 和 Path。
   - 找到對應的 `Route` 和提取參數 (`InfoRequest`)。
2. **Request Transcoding (`OnHttpRequestBody`)**:
   - 讀取完整 HTTP Body。
   - 呼叫 `RequestCov` 將 JSON 轉為 gRPC Binary。
   - 構造 gRPC Header (`:method`, `:path`, `content-type`)。
   - `DispatchHttpCall` 發送請求到後端 gRPC Cluster。
3. **Response Handling (`callback`)**:
   - 接收 gRPC 回應，暫存 Body 到 `ctx.remoteBody`。
   - 恢復原始請求處理 (`ResumeHttpRequest`)。
4. **Response Transcoding (`OnHttpResponseHeaders` / `OnHttpResponseBody`)**:
   - 修改 Response Header (`content-type: application/json`, `status: 200`)。
   - 呼叫 `ResponseCov` 將 gRPC Binary 轉回 JSON。
   - 替換 Response Body (`ReplaceHttpResponseBody`)。

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

### struct
最外層直接讓 main 使用的 grpcs
- AddRoute: 將 route 加到需要掃描的 routes 中
- Match: 輸出 InfoRequest 及對應 route, 如果route 為nil 就是沒有對到可以用的route, 直接 response 不存在
- route 跟 info 可以直接存在http context 來進行後續動作
內層針對每個 grpc 的實作 interface
- 以 http.ServeMux 實作一個 requestMapper
  - New(clusterName): 設定 cluster name
  - MatchInfo(method, path(with querystring)) (InfoRequest, error)
  - RequestBody2grpc(InfoRequest, jsonBody) ([]btye, error)
    - endpoint 可能會需要 dto 跟實作轉成 grpc request body
    - 這裡才真正做 field validation
  - ResponseBody2json(InfoRequest, grpcBody) ([]btye, error)
  - RegisterRoutes(): 將 http path 與 對應的 grpc path 跟 body converter 聯繫起來
