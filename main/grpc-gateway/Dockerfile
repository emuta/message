FROM emuta/images:golang-audi AS build

WORKDIR /go

RUN apk update && apk add git upx \
    && go get -u github.com/emuta/go-rabbitmq/pubsub/publisher \
    && go get -u github.com/emuta/message \
    && CGO_ENABLED=0 go build -ldflags '-w -s' -o server $GOPATH/src/github.com/emuta/message/main/grpc-gateway/main.go \
    && upx server \
    && apk del git && rm -rf /var/cache/apk

FROM scratch

WORKDIR /usr/local/bin

COPY --from=build /go/server .

CMD ["./server"]