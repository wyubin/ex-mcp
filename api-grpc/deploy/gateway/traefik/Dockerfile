FROM golang:1.24-alpine AS builder

# get src and build
ENV SRC="/var/src"
COPY . $SRC
RUN cd $SRC \
    && go mod tidy \
    && CGO_ENABLED=0 GOOS=linux go build -buildmode=plugin -o /opt/usergateway.so plugin/gateway/traefik.go

# --------- Runtime Stage ---------
FROM alpine:latest
WORKDIR /plugins
COPY --from=builder /opt/usergateway.so .
# Plugin 不需要其他啟動命令，Traefik container 會 mount 這個 /plugins
