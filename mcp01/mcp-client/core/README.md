# intro
設定從 host -> client 的 spec

# specs

## host
- 從 config bytes 讀入 servers 資料來設定 client
- 也要能夠 dump config bytes
- 

## client
- 要能夠 disable server

# note
- server unmarshal 的方法可以寫在外面， host 直接讀 CfgServer 這個 object 來init client 就好