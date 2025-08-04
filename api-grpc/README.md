# intro
基於 goalng grpc buf, connect-go 架構建立一個 user item 的 CRUD service

# install
## requirements
```shell
# 安裝相關 protoc plugin
go install github.com/bufbuild/protovalidate-go/cmd/protoc-gen-validate@latest
go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# 安裝 buf
BIN="/usr/local/bin" && \
VERSION="1.56.0" && \
curl -sSL \
"https://github.com/bufbuild/buf/releases/download/v${VERSION}/buf-$(uname -s)-$(uname -m)" \
-o "${BIN}/buf" && \
chmod +x "${BIN}/buf"

# 下載 buf 相關模組
buf mod update
buf generate
```

