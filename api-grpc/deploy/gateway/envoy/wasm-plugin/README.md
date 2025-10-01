# intro
建立 enovy plugin 開發環境，主要以golang compile 成 wasm 的 plugin 提供 envoy 作為 plugin 使用，會需要兩步驟開發
- 先在 local compile wasm
```shell
# 先在 container 中 compile wasm
docker build -t go-wasm-builder -f Dockerfile.build-wasm .
docker run --rm -v $(pwd):/app go-wasm-builder cp /var/main.wasm /app/main.wasm
# 啟動enovy
docker-compose up
# 如果有重新 compile wasm, restart
docker-compose restart
```
