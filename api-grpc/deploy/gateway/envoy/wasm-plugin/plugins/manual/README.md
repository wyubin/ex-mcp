# intro
要把 restful api 的 api call 轉成能夠傳到後方 grpc server 的 request, 並回傳 json response 到 response body 中

# issue
目前實作狀況為
- 能夠正確地傳遞到後端 grpc 但無法回傳到 response 的部分