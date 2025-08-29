# ========== build go .so ==========
FROM golang:1.24-alpine AS builder
ENV SRC="/var"
COPY ./plugin "${SRC}/src"
RUN apk add --no-cache build-base gcc musl-dev libc-dev
RUN cd "${SRC}/src" \
    && mkdir -p "${SRC}/bin" \
    && go mod tidy \
    && CGO_ENABLED=1 go build -buildmode=c-shared -o ../bin/libgolang.so .
# c-shared 需要 CGO=1 與編譯器

# ========== envoy ==========
FROM envoyproxy/envoy:contrib-v1.35-latest
COPY --from=builder /var/bin/libgolang.so /usr/local/lib/libgolang.so
COPY envoy.yaml /etc/envoy/envoy.yaml
