package model

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type Message struct {
	Id    int64  `gorm:"primary_key:true"`
	UuidV4 string
	AppId string
	Topic string
	Data  postgres.Jsonb
	CreatedAt time.Time
	AutoEmit  bool
}

func (Message) TableName() string {
	return "message.message"
}

type Publish struct {
	Id        int64     `gorm:"primary_key:true"`
	MsgId     int64
	PublishAt time.Time
}

func (Publish) TableName() string {
	return "message.publish"
}