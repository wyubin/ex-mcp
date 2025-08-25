# intro
紀錄多個服務的部署方式

# service
```shell
docker build -t api-grpc/svc-user \
    -f deploy/service/user/Dockerfile
```

# plugin
```shell
docker build -t usergateway-plugin \
    -f deploy/plugin/gateway/Dockerfile
```