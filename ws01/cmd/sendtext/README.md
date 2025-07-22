# intro
建立並測試 websocket server 的輸入及輸出

# test
```shell
# run server
go run server/main.go

# run client with multiple msgs
go run client/main.go -session abc123 -msgcount 10 -host localhost:8080
```