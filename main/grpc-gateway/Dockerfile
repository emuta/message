FROM emuta/images:golang-audi AS build


RUN apk update && apk add git upx \
    && go get -u github.com/emuta/message \
    && CGO_ENABLED=0 go build -ldflags '-w -s' -o /go/bin/server $GOPATH/src/github.com/emuta/message/main/grpc-gateway/main.go \
    && upx server \
    && apk del git && rm -rf /var/cache/apk

FROM scratch

COPY --from=build /go/bin/server /usr/local/bin/server

CMD ["server"]