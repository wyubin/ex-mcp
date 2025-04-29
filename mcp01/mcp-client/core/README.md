# intro
設定從 host -> client 的 spec

# specs

## host
- new 一個空的 host
- setClient 每次新加一個 client
- list tools, 列出目前所有 client 的 tools(only enable), 但名稱
- CallTool, 從指定的 <name-server>.<name-tools> 及 arg 來執行tools
- 當 server disable 時， list tools 跟call tool 都會 raise error

## client
- 要能夠 disable server
- 輸出 cfg
- 列出tools
- 執行並拿到對應的結果

### use
- 如果是 sse, 都必須要 init 之後才會拿到有 client_id 的ctx 才能繼續做 list 跟 execute
- 目前看 mark3labs 實作, sse 只有做 tools 方面？ 

# note
- server unmarshal 的方法可以寫在外面， host 直接讀 CfgServer 這個 object 來init client 就好