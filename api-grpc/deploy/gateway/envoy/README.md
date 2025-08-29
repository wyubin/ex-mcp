# intro
研究 envoy 的 plugin 使用方式

# plugin
在 ./plugin 中直接以 golang 實作， compile 為 .so 來進行使用

# compile
```shell
cd plugin
CGO_ENABLED=1 go build -buildmode=c-shared -o /var/tmp/libgolang.so .
```

# deploy
- 直接從 docker compose 將 plugin 進行 compile 並塞到 envoy image 中並進行服務
```shell
docker build -t envoy-with-plugin .
docker run --rm -it -p 8080:8080 envoy-with-plugin
# check version
# check extension support
docker run --rm -it --entrypoint envoy  --help envoy-with-plugin | grep golang

```

# ref
base on [MoE 系列 - 如何使用 Golang 扩展 Envoy [一]](https://mosn.io/blog/posts/moe-extend-envoy-using-golang-1/)