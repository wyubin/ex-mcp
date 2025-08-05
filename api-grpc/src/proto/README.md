# intro
說明 proto 建立的基本方法

# steps
- 可以先以 gpt 產生一個簡單的 proto, 基於需求架上以下描述
  - 需要基於 "buf/validate/validate.proto" 建立基本的欄位 validate
  - 詳細說明每個 service function 需要怎樣的 request 跟response
  - rpc 的 option (google.api.http) 定義想要用的 http method 跟 path, 這樣比較方便，就不需要重新定義，直接用 handler
- 後面再加 service 去 implement *_grpc.pb.go 的 