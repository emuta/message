FROM golang:alpine AS build

WORKDIR /go/src

ADD ./ ./cmd

RUN sed -i "s/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g" /etc/apk/repositories \
    && apk add --update --no-cache git upx \
    && go get -u google.golang.org/grpc
    && go get -u github.com/streadway/amqp \
    && go get -u github.com/sirupsen/logrus \
    && go get -u github.com/emuta/message \
    && CGO_ENABLED=0 go build -ldflags '-w -s' -o server $GOPATH/src/cmd/main.go \
    && upx server \
    && apk del git


FROM alpine

ENV TZ Asia/Shanghai
WORKDIR /usr/local/bin

COPY --from=build /go/src/server .

RUN sed -i "s/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g" /etc/apk/repositories \
    &&apk add --update --no-cache tzdata\
    && cp /usr/share/zoneinfo/$TZ /etc/localtime \
    && echo $TZ > /etc/timezone \
    && rm -rf /var/cache/apk

CMD ["./server"]