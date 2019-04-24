FROM emuta/images:golang-audi AS build

WORKDIR /go

RUN apk update && apk add git upx \
    && go get -u github.com/emuta/go-rabbitmq/pubsub/publisher \
    && go get -u github.com/emuta/message \
    && CGO_ENABLED=0 go build -ldflags '-w -s' -o server $GOPATH/src/github.com/emuta/message/main/grpc-server/main.go \
    && upx server \
    && apk del git


FROM alpine

ENV TZ Asia/Shanghai
WORKDIR /usr/local/bin

COPY --from=build /go/server .

RUN apk update \
    && add tzdata --no-cache \
    && cp /usr/share/zoneinfo/$TZ /etc/localtime \
    && echo $TZ > /etc/timezone \
    && rm -rf /var/cache/apk

CMD ["./server"]