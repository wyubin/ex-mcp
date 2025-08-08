# intro
說明 proto 建立的基本方法

# steps
- 可以先以 gpt 產生一個簡單的 proto, 基於需求架上以下描述
  - 需要基於 "buf/validate/validate.proto" 建立基本的欄位 validate
  - 詳細說明每個 service function 需要怎樣的 request 跟response
  - rpc 的 option (google.api.http) 定義想要用的 http method 跟 path, 這樣比較方便，就不需要重新定義，直接用 handler
- 後面再加 service 去 implement *_grpc.pb.go 的 UserServiceServer

# gateway
`buf.build/grpc-ecosystem/gateway` 會建立 *.pb.gw.go, 提供 RegisterUserServiceHandlerServer, 可以把 implSvc 跟 gateway route 接在一起，之後這個route 可以再往上接到 main http route
- `RegisterUserServiceHandlerServer` 並沒有另外建立 grpc client 再打 grpc server, 

# connect-go
- Pros:
  - connect-go 可以同時用單一 port 服務 http protocol 跟 grpc, 而且因為 http protocol 會直接走到 connect 內部轉換，所以會稍微低一點(但 gateway 也可以直接內部轉沒有外轉另一個 port, 有可能差不多)
- Cons:
  - svcImpl interface 與 grpc 不同，會需要另外再寫一個類似實作的服務或是 adaptor, 會有點麻煩
  - buf.gen.yaml 本來的 openapi 轉出的 swagger 並不會透過 connect 的 request/response 去轉換，要額外寫
