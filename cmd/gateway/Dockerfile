FROM golang:alpine AS build

WORKDIR /go/src

ADD ./ ./cmd

RUN apk add --update --no-cache git upx \
    && go get -u google.golang.org/grpc \
    && go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway \
    && go get -u github.com/streadway/amqp \
    && go get -u github.com/sirupsen/logrus \
    && go get -u github.com/jinzhu/gorm \
    && go get -u github.com/lib/pq \
    && go get -u github.com/emuta/message \
    && go get -u github.com/emuta/go-rabbitmq/pubsub/publisher \
    && CGO_ENABLED=0 go build -ldflags '-w -s' -o server $GOPATH/src/cmd/main.go \
    && upx server \
    && apk del git


FROM alpine

ENV TZ Asia/Shanghai
WORKDIR /usr/local/bin

COPY --from=build /go/src/server .

RUN apk add --update --no-cache tzdata\
    && cp /usr/share/zoneinfo/$TZ /etc/localtime \
    && echo $TZ > /etc/timezone \
    && rm -rf /var/cache/apk

CMD ["./server"]