package server

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/emuta/go-rabbitmq/pubsub/publisher"
	
	"message/pbtype"
	pb "message/proto"
	"message/repository"
)

type messageServiceServer struct {
	repo   *repository.Repository
	broker *publisher.Publisher
}

func NewMessageServiceServer(repo *repository.Repository, broker *publisher.Publisher) *messageServiceServer {
	log.Info("Services loaded")
	return &messageServiceServer{
		repo:   repo,
		broker: broker,
	}
}

func (s *messageServiceServer) CreateMessage(ctx context.Context, req *pb.CreateRequest) (*pb.Message, error) {
	resp, err := s.repo.CreateMessage(ctx, req.UuidV4, req.AppId, req.Topic, req.Data, req.Emit)
	if err != nil {
		return nil, err
	}
	return pbtype.MessageProto(resp), nil
}

func (s *messageServiceServer) ListMessage(ctx context.Context, req *pb.FindRequest) (*pb.ListResponse, error) {
	args := map[string]interface{}{}
	results, err := s.repo.ListMessage(ctx, args)
	if err != nil {
		return nil, err
	}

	var resp pb.ListResponse
	for _, result := range *results {
		resp.Messages = append(resp.Messages, pbtype.MessageProto(&result))
	}
	return &resp, nil
}

func (s *messageServiceServer) GetMessage(ctx context.Context, req *pb.GetRequest) (*pb.Message, error) {
	resp, err := s.repo.GetMessage(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return pbtype.MessageProto(resp), nil
}

func (s *messageServiceServer) DeleteMessage(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	if err := s.repo.DeleteMessage(ctx, req.Id); err != nil {
		return nil, err
	}
	return &pb.DeleteResponse{Success: true}, nil
}

func (s *messageServiceServer) PublishMessage(ctx context.Context, req *pb.PublishMessageRequest) (*pb.PublishResponse, error) {
	resp, err := s.repo.PublishMessage(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	go func() {
		msg, err := s.repo.GetMessage(context.Background(), req.Id)
		if err != nil {
			log.WithError(err).Errorf("Failed to get message of id: %d", req.Id)
		}
		body, err := msg.Data.MarshalJSON()
		if err != nil {
			log.WithError(err).Errorf("Failed to publish message of id: %d", req.Id)
		}
		s.broker.Publish(msg.UuidV4, msg.AppId, msg.Topic, body)
	}()
	return pbtype.PublishProto(resp), nil
}
