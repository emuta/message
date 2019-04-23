package main

import (
    "context"
    "flag"
    "net/http"

    "github.com/golang/glog"
    "github.com/grpc-ecosystem/grpc-gateway/runtime"
    "google.golang.org/grpc"

    gw "message/proto"
)

var (
    addr string
    endpoint string
)

func init() {
    flag.StringVar(&addr, "addr", ":80", "the address of service listen")
    flag.StringVar(&endpoint, "endpoint", ":3721", "the GRP address of service")
    flag.Parse()
}

func run() error {
    ctx := context.Background()
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    mux := runtime.NewServeMux()
    opts := []grpc.DialOption{grpc.WithInsecure()}
    err := gw.RegisterMessageServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
    if err != nil {
        return err
    }

    return http.ListenAndServe(addr, mux)
}

func main() {
    flag.Parse()
    defer glog.Flush()

    if err := run(); err != nil {
        glog.Fatal(err)
    }
}