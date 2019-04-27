package main

import (
	"context"
	"encoding/json"
	"flag"

	"google.golang.org/grpc"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"

	pb "github.com/emuta/message/proto"
)

var (
	brokerURL    string
	exchangeName string
	endpoint     string
)

func failOnError(err error, msg string) {
	if err != nil {
		log.WithError(err).Fatal(msg)
	}
}

func init() {
	flag.StringVar(&brokerURL, "broker", "amqp://guest:guest@rabbitmq:5672/", "The broker url for RabbitMQ connect")
	flag.StringVar(&exchangeName, "exchange", "pubsub", "The exchange name of RabbitMQ used")
	flag.StringVar(&endpoint, "endpoint", "message-grpc-server:3721", "The message GRP server endpoint")
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

}

func watchConnection(conn *amqp.Connection) {
	ch := conn.NotifyClose(make(chan *amqp.Error))
	go func() {
		defer close(ch)
		err := <-ch
		log.WithError(err).Info("Connection closed, should reconnect channel soon")
	}()
}

func watchChannel(channel *amqp.Channel) {
	ch := channel.NotifyClose(make(chan *amqp.Error))
	go func() {
		defer close(ch)
		err := <-ch
		log.WithError(err).Info("channel closed, should reconnect channel soon")
	}()
}

func main() {
	conn, err := amqp.Dial(brokerURL)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	go watchConnection(conn)

	log.Info("Connected to RabbitMQ server")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	go watchChannel(ch)

	log.Info("[RabbitMQ] Forked a new channel")

	err = ch.ExchangeDeclare(
		exchangeName,        // exchange name
		amqp.ExchangeFanout, // exchange type
		true,                // durable
		false,               // auto deleted
		false,               // internal
		false,               // no wait
		nil,                 // arguments
	)
	failOnError(err, "Failed to declare a exchange")

	log.Infof("[RabbitMQ] Declared a [%s] exchange -> %s", amqp.ExchangeFanout, exchangeName)

	q, err := ch.QueueDeclare(
		"",    // queue name
		true,  // durable
		false, //delete when unused
		false, // exclusive
		false, // no wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name,       // queue name
		"",           // routing key
		exchangeName, // exchange name
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // audo ack
		false,  //exclusive
		false,  // no local
		false,  // no wait
		nil,    // arguments
	)

	// connect to grpc server
	grpcConn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	failOnError(err, "Failed to connect grpc server")
	grpcClient := pb.NewMessageServiceClient(grpcConn)

	log.Info("Connect to GRPC server success")

	forever := make(chan bool)
	go func() {
		defer close(forever)

		for msg := range msgs {

			go func() {
				resp, err := grpcClient.CreateMessage(context.Background(), &pb.CreateMessageReq{
					UuIdV4: msg.MessageId,
					AppId:  msg.AppId,
					Topic:  msg.Type,
					Data:   msg.Body,
				})
				if err != nil {
					log.WithError(err).Errorf("Failed to save message to db. Message<%s>", msg)
				}

				if err := msg.Ack(true); err != nil {
					log.WithError(err).Warn(err, "Failed to reply ACK. Message<%s>", msg)
				}

				log.WithField("id", resp.Id).Info("Save message to db success")
			}()

			data, err := json.Marshal(msg)
			if err != nil {
				log.Error(err)
			}
			log.Infof("%s", data)
		}
	}()
	<-forever
}
