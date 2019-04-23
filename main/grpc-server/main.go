package main

import (
    "flag"
    "fmt"
    "net"
    "os"

    "github.com/jinzhu/gorm"
    _ "github.com/lib/pq"
    log "github.com/sirupsen/logrus"
    "google.golang.org/grpc"

    "github.com/emuta/go-rabbitmq/pubsub/publisher"
    
    pb "github.com/emuta/message/proto"
    "github.com/emuta/message/repository"
    "github.com/emuta/message/server"
)

var (
    port   string
    pubsub string

    broker *publisher.Publisher
    db     *gorm.DB
)

func init() {
    flag.StringVar(&port, "port", "3721", "The port of service listen")
    flag.StringVar(&pubsub, "pubsub", "pubsub", "The exchange name  of rabbitmq")
    flag.Parse()

    initLogger()
    initPostgres()
    initPublisherBroker()
}

func initLogger() {
    log.SetFormatter(&log.TextFormatter{
        FullTimestamp:   true,
        TimestampFormat: "2006-01-02 15:04:05",
    })
}

func initPublisherBroker() {
    url, ok := os.LookupEnv("RABBITMQ_URL")
    if !ok {
        log.Fatal("Not found postgresql URL from environment")
    }

    broker = publisher.NewPublisher(url, pubsub)
}

func initPostgres() {
    var err error
    url, ok := os.LookupEnv("PG_URL")
    if !ok {
        log.Fatal("Not found postgresql URL from environment")
    }
    db, err = gorm.Open("postgres", url)
    if err != nil {
        log.WithError(err).Fatal("Failed to connect PostgreSQL server")
    }
    log.Info("[PostgreSQL] Connected successfully")
    db.LogMode(true)
}

func main() {
    repo := repository.NewRepository(db)

    s := grpc.NewServer()
    pb.RegisterMessageServiceServer(s, server.NewMessageServiceServer(repo, broker))

    addr := fmt.Sprintf("0.0.0.0:%s", port)
    l, err := net.Listen("tcp", addr)
    if err != nil {
        log.WithError(err).Fatal("Failed to listen %s", addr)
    }

    log.Info("Initialized all components")
    log.Infof("Server starting with addr -> %s", addr)
    log.Fatal(s.Serve(l))
}
