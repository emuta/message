package pbtype

import (
    "github.com/golang/protobuf/ptypes"

    pb "github.com/emuta/message/proto"
    "github.com/emuta/message/model"
)

func MessageProto(m *model.Message) *pb.Message {
    p := pb.Message{
        Id:     m.Id,
        UuidV4: m.UuidV4,
        AppId:  m.AppId,
        Topic:  m.Topic,
    }

    if t, err := ptypes.TimestampProto(m.CreatedAt); err != nil {
        p.CreatedAt = t
    }

    return &p
}

func PublishProto(m *model.Publish) *pb.PublishResponse {
    p := pb.PublishResponse{
        Id: m.Id,
        MsgId: m.MsgId,
    }

    if t, err := ptypes.TimestampProto(m.PublishAt); err != nil {
        p.Timestamp = t
    }
    
    return &p
}