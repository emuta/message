FROM golang:alpine AS build

ENV GOBIN "/go/bin"

RUN apk update && apk add git upx \
    && go get -u github.com/emuta/message \
    && CGO_ENABLED=0 go build -ldflags '-w -s' -o server $GOPATH/src/github.com/emuta/message/main/grpc-server/main.go \
    && upx server \
    && apk del git && rm -rf /var/cache/apk

FROM scratch

COPY --from=build $GOBIN/server /usr/local/bin/server

CMD ["server"]