package main

import (
	"context"
	"encoding/json"
	"flag"

	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"

	pb "github.com/emuta/message/proto"
)

var (
	brokerURL    string
	exchangeName string
	endpoint     string
)

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

func failOnError(err error, msg string) {
	if err != nil {
		log.WithError(err).Fatal(msg)
	}
}

type RabbitMQClient struct {
	Connection   *amqp.Connection
	Channel      *amqp.Channel
	Queue        *amqp.Queue
	URI          string
	ExchangeName string
	PbAddr string
	PbClient     pb.MessageServiceClient
}

func NewRabbitMQClient(uri, exchangeName, pbAddr string) *RabbitMQClient {
	c := RabbitMQClient{
		URI:          uri,
		ExchangeName: exchangeName,
		PbAddr: pbAddr,
	}

	if err := c.connect(); err != nil {
		failOnError(err, "Failed to connect broker")
	}

	if err := c.newChannel(); err != nil {
		failOnError(err, "Failed to get new channel")
	}

	if err := c.declareExchange(); err != nil {
		failOnError(err, "Failed to declare exchange")
	}

	if err := c.declareQueue(); err != nil {
		failOnError(err, "Failed to declare queue")
	}

	if err := c.newPbClient(); err != nil {
		failOnError(err, "Failed to create pb client")
	}

	return &c
}

func (s *RabbitMQClient) connect() error {
	log.Info("Reay to connect to broker")
	conn, err := amqp.Dial(s.URI)
	if err != nil {
		return err
	}
	s.Connection = conn
	go s.watchConnection(conn)
	log.Info("Connected to broker success")

	return nil
}

func (s *RabbitMQClient) watchConnection(conn *amqp.Connection) {
	ch := conn.NotifyClose(make(chan *amqp.Error))
	go func() {
		defer close(ch)
		err := <-ch
		log.WithError(err).Error("Connection of broker closed, should reconnect channel soon")

		s.connect()
	}()
}
func (s *RabbitMQClient) newChannel() error {
	ch, err := s.Connection.Channel()
	if err != nil {
		return err
	}
	s.Channel = ch
	go s.watchChannel(ch)
	return nil
}

func (s *RabbitMQClient) watchChannel(channel *amqp.Channel) {
	ch := channel.NotifyClose(make(chan *amqp.Error))
	go func() {
		defer close(ch)
		err := <-ch
		log.WithError(err).Error("channel closed, should reconnect channel soon")
		s.newChannel()
	}()
}

func (s *RabbitMQClient) declareExchange() error {
	return s.Channel.ExchangeDeclare(
		s.ExchangeName,      // exchange name
		amqp.ExchangeFanout, // exchange type
		true,                // durable
		false,               // auto deleted
		false,               // internal
		false,               // no wait
		nil,                 // arguments
	)
}

func (s *RabbitMQClient) declareQueue() error {
	q, err := s.Channel.QueueDeclare(
		"message.archive", // queue name
		true,              // durable
		true,              //delete when unused
		false,             // exclusive
		false,             // no wait
		nil,               // arguments
	)

	if err != nil {
		return nil
	}

	err = s.Channel.QueueBind(
		q.Name,         // queue name
		"",             // routing key
		s.ExchangeName, // exchange name
		false,
		nil,
	)

	if err != nil {
		return nil
	}

	s.Queue = &q

	return nil
}

func (s *RabbitMQClient) watchMessage() (<-chan amqp.Delivery, error) {
	return s.Channel.Consume(
		s.Queue.Name,              // queue
		"message.archive.monitor", // consumer
		false,                     // audo ack
		false,                     //exclusive
		false,                     // no local
		false,                     // no wait
		nil,                       // arguments
	)
}

func (s *RabbitMQClient) newPbClient() error {
	conn, err := grpc.Dial(s.PbAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	s.PbClient = pb.NewMessageServiceClient(conn)

	log.Info("Connect to GRPC server success")

	return nil
}

func (s *RabbitMQClient) saveMessage(msg amqp.Delivery) error {
	resp, err := s.PbClient.CreateMessage(context.Background(), &pb.CreateRequest{
		UuidV4: msg.MessageId,
		AppId:  msg.AppId,
		Topic:  msg.Type,
		Data:   msg.Body,
	})
	if err != nil {
		log.WithError(err).Errorf("Failed to save message to db. Message<%s>", msg)
		return err
	}

	if err := msg.Ack(true); err != nil {
		log.WithError(err).Warn(err, "Failed to reply ACK. Message<%s>", msg)
		return err
	}

	log.WithField("id", resp.Id).Info("Save message to db success")

	data, err := json.Marshal(msg)
	if err != nil {
		log.Error(err)
	}
	log.Infof("%s", data)

	return nil
}

func main() {
	forever := make(chan bool)
	c := NewRabbitMQClient(brokerURL, exchangeName, endpoint)

	msgs, err := c.watchMessage()
	failOnError(err, "Failed to consume message")
	for msg := range msgs {
		go c.saveMessage(msg)
	}
	<-forever
}
