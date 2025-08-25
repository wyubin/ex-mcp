FROM golang:1.24-alpine AS builder

# get src and build
ENV SRC="/var/src"
COPY . $SRC
RUN cd $SRC \
    && go mod tidy \
    && DESTDIR=/opt NAME=user make svc

FROM alpine:latest
COPY --from=builder /opt /opt
EXPOSE 50051
CMD ["/opt/apisvc"]
