# intro
基於設定檔來決定是否要隨機換到另一個 route 到哪一個 cluster
```shell
# 如果在 http_filters 的 configuration 沒有設定就會隨機 到 "-canary"
# 在 HTTP/2 協議裡，:authority 是一個 pseudo-header，相當於傳統 HTTP/1.1 的 Host header。
# 它用來告訴 上游伺服器（upstream server）這個請求是針對哪個主機。
curl localhost:18000
``` 