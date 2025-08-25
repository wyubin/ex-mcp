# intro
基於 goalng grpc buf, connect-go 架構建立一個 user item 的 CRUD service

# install
## requirements
```shell
# 安裝相關 protoc plugin
go install github.com/bufbuild/protovalidate-go/cmd/protoc-gen-validate@latest
# go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go get github.com/grpc-ecosystem/grpc-gateway/v2@v2.27.1

# 下載 swagger cli 進行 generate api yaml
go install github.com/swaggo/swag/cmd/swag@latest

# 安裝 buf
BIN="/usr/local/bin" && \
VERSION="1.56.0" && \
curl -sSL \
"https://github.com/bufbuild/buf/releases/download/v${VERSION}/buf-$(uname -s)-$(uname -m)" \
-o "${BIN}/buf" && \
chmod +x "${BIN}/buf"

# 下載 buf 相關模組
buf dep update
buf generate
```

# buf
- buf.gen.yaml 中 plugins 的設定可以直接去看 remote 的 src code, opt 下面通常會是 plugins 的 main 的 flags

# swagger
需要考慮如何從 proto 到產生swagger json 的流程，會需要做以下步驟
- 在 buf lib 會自動產生 yaml
- 在 makefile 裡面，可以針對不同 service 把相關的 yaml 做 merge, 然後放到 swagger dir 就可以

# build
```shell
DESTDIR=/var/local NAME=user make svc
```

# test
```shell
/var/local/apisvc
# swagger: http://localhost:8080/static/docs.html
# for example create a user
curl -X POST http://localhost:8080/api/user.UserService/CreateUser \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "email": "alice@example.com"}'
```