# intro
紀錄 traefix 如何簡單部署及使用config 及 plugin

# deply
## docker compose
```shell
docker-compose up -d
# down
docker-compose down
```

# settings
基本設定就靜態(traefik.yml)跟動態設定(providers.*)，如果沒有 plugin, 靜態部分如下
```yaml
entryPoints:
  web:
    address: ":80"

providers:
  file:
    directory: /etc/traefik/dynamic
    watch: true

api:
  insecure: true
```
動態則是可以隨需要加 yaml, 會自動 merge 不同的 routers 跟 services 再進行轉發
```yaml
http:
  routers:
    hello:
      rule: "PathPrefix(`/hello`)"
      service: hello-service

  services:
    hello-service:
      loadBalancer:
        servers:
          - url: "http://example.com"
```

# plugin
目前主要會思考如何自己開發 plugins

# minitor
服務起來後可以直接 access http://localhost:8080/dashboard/#/ 可以顯示服務狀態