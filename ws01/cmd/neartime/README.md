# intro
模擬 音訊如何傳入並傳到 asura 進行即時輸出逐字稿

# plan
- 先建立一個純 client 打 asura token -> ws/v3 可以完成整個流程的 case
  - 需要包含來源 pcm 每秒切開並 fmt 輸出到 log

# script
```shell
# 將 mp3 轉成 pcm
ffmpeg -i ~/Downloads/recording2.mp3 -f s16le -acodec pcm_s16le -ac 1 -ar 16000 ./ws01/tmp/recording2.pcm

# run server
go run server/main.go
# run client
go run client/main.go -session=abc123 -file=/workspaces/ex-mcp/ws01/tmp/test-5.pcm -host=localhost:8080

# run asura client
go run asura-client/main.go -token=W8eDsvxL0hR7I94yK7BFs -file=/workspaces/ex-mcp/ws01/tmp/recording2.pcm -host=fedgpt-dev.corp.ailabs.tw
```