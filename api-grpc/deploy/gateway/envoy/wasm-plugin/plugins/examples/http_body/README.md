# intro
基於 http header 來進行 request or response 的 rewrite
```shell
# 預設會修改 request body，log 會有 "original request body"
curl -XPUT localhost:18000 --data '[initial body]' -H "buffer-operation: prepend"
# 不改 request, 而是在 response 時修改 response body
curl -XPUT localhost:18000 --data '[initial body]' -H "buffer-operation: prepend" -H "buffer-replace-at: response"
# 如果不指定 header, 則預設為 replace body
curl -XPUT localhost:18000 --data '[initial body]'
``` 

# main.go structure
VMContext ──▶ PluginContext ──▶ HttpContext
- VMContext：對應到整個 WASM VM 的生命周期（通常只有一個）
- PluginContext：對應到 Envoy 中的 plugin 實例（可能有多個），只要在envoy.yaml 讀進來一次就會基於config 進行一次 init，在此次範例就有兩個 instance, 分別是 setBody Mode 跟 echo mode
- HttpContext：對應到每個 HTTP stream（每次 request-response 流程），所以單個 request 經過一個 PluginContext NewHttpContext 後，就會產生一個 HttpContext

```scss
Envoy (Request)
   │
   ▼
[WASM VM]
   │
   ├─> vmContext
   │    └─> 建立 pluginContext
   │
   ├─> pluginContext
   │    ├─> 解析 plugin config ("echo" or default)
   │    └─> 決定使用 echoBodyContext 或 setBodyContext
   │
   └─> HttpContext (每個 request)
        ├─> OnHttpRequestHeaders()
        ├─> OnHttpRequestBody()
        ├─> OnHttpResponseHeaders()
        └─> OnHttpResponseBody()

```

## learn
- 在 pluginContext 中 OnPluginStart -> NewHttpContext
- 在 OnHttpRequestHeaders
  - 通常 不會檢查 endOfStream
  - `GetHttpRequestHeader("buffer-replace-at")` 確定是否在 OnHttpResponseBody 才改回應
  - 中會先 RemoveHttpRequestHeader("content-length") 來避免在 OnHttpRequestBody 中修改 RequestBody(ex. proxywasm.ReplaceHttpRequestBody) 造成 Error
  - 用 GetHttpRequestHeader("buffer-operation") 先設定 bufferOperation
- OnHttpRequestBody
  - 用 ctx.bufferOperation 來決定 如何更新 RequestBody
  - 會在 endOfStream 之後一次 `proxywasm.GetHttpRequestBody(0, bodySize)` 拿到全部 body, 然後再進行 revise body
- OnHttpResponseHeaders
  - 一樣沒有 endOfStream 確認，預設會修改 body, 所以先 `RemoveHttpResponseHeader("content-length")`
- OnHttpResponseBody
  - 跟 OnHttpRequestBody 類似操作
- echoBodyContext
  - 在 `OnHttpRequestBody` 中 直接 `GetHttpRequestBody(0, bodySize)`, 然後直送 `SendHttpResponse(200, nil, body, -1)`, 但之後會用 `types.ActionPause` 結束資料傳遞

結論
- 如果要改 RequestBody/ResponseBody, 要在 header 先 remove "content-length"
- 看起來好像不用在 !endOfStream 時先 get body, 等 endOfStream 再一次讀

# enovy yaml structure
```yaml
┌─────────────────────┐
│        Client        │
│ curl localhost:18000 │
└─────────┬───────────┘
          │
          ▼
   ┌──────────────────────────────┐
   │  Envoy Listener: main:18000  │
   │  Filter Chain:               │
   │    - http_connection_manager │
   │    - wasm (config=body-set)  │  ← main.wasm (修改 body)
   │    - router                  │
   └─────────────┬────────────────┘
                 │
                 ▼
          [Cluster: echo]
                 │
                 ▼
   ┌──────────────────────────────┐
   │  Envoy Listener: echo:38140  │
   │  Filter Chain:               │
   │    - http_connection_manager │
   │    - wasm (config=echo)      │  ← main.wasm (echo 模式) (return types.ActionPause)
   │    - router                  │
   └─────────────┬────────────────┘
                 │
                 ▼
           [Cluster: admin] (port 8001)
```

# question
- 為什麼不直接從 set response 回去，還要多一個 echo 來接？
  - 要該要把 http_filters 想成是一個 middleware, 會在request 時分別 header 跟 body 處理完傳到 router 後，再等 router 收到response(在 prefix: "/" 會進到 cluster: echo) 
  - 但在 echo httpContext 實作中，則是直接在 OnHttpRequestBody 時就把 body 完整用 SendHttpResponse 送回，並接下去回 types.ActionPause 停止往下傳, 雖然還是指定一個 `cluster: admin`, 但實際上並不會傳過去
  - 也就是說，照理說，在 http_filters 中，本來就應該在 wasm plugin 之後接下去看要 route 到哪一個 cluster, 基本上 response 時也應該維持原來的 url 會比較正常(不然就要記 origin url 然後 response 時在再轉回原來的 url 做 response)
  - 另外，因為也有在 response body 做修改的功能，所以要有一個 echo 來做直接回應