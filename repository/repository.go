package repository

import (
	"context"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"

	"github.com/emuta/message/model"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (s *Repository) CreateMessage(ctx context.Context, uuidV4, appId, topic string, data []byte, emit bool) (*model.Message, error) {
	result := model.Message{
		UuidV4:    uuidV4,
		AppId:     appId,
		Topic:     topic,
		Data:      postgres.Jsonb{data},
		CreatedAt: time.Now(),
		AutoEmit:  emit,
	}

	ch := make(chan error)

	go func() {
		defer close(ch)

		if err := s.db.Take(&model.Message{}, "uuid_v4 = ?", uuidV4).Error; err != nil {
			ch <- err

			log.WithError(err).WithFields(log.Fields{
				"uuidV4": uuidV4,
				"appId":  appId,
				"topic":  topic,
				"data":   data,
			}).Error("Failed to create message")
			return
		}

		ch <- s.db.Create(&result).Error

	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-ch:
		if err != nil {
			return nil, err
		}
	}

	return &result, nil
}

func (s *Repository) GetMessage(ctx context.Context, id int64) (*model.Message, error) {
	result := model.Message{Id: id}
	ch := make(chan error)
	go func() {
		defer close(ch)
		ch <- s.db.Take(&result).Error
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-ch:
		if err != nil {
			log.WithError(err).Errorf("Failed to get message %d", id)
			return nil, err
		}
	}

	return &result, nil
}

func (s *Repository) DeleteMessage(ctx context.Context, id int64) error {
	result := model.Message{Id: id}
	ch := make(chan error)
	go func() {
		defer close(ch)
		ch <- s.db.Delete(&result).Error
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-ch:
		if err != nil {
			log.WithError(err).Errorf("Failed to delete message %d", id)
			return err
		}
	}

	return nil
}

func getListMessageDB(db *gorm.DB, args map[string]interface{}) *gorm.DB {
	tx := db.Model(&model.Message{})

	if id, ok := args["id"]; ok {
		tx = tx.Where("id = ?", id)
	}

	if uuid_v4, ok := args["uuid_v4"]; ok {
		tx = tx.Where("uuid_v4 = ?", uuid_v4)
	}

	if app_id, ok := args["app_id"]; ok {
		tx = tx.Where("app_id = ?", app_id)
	}

	if topic, ok := args["topic"]; ok {
		tx = tx.Where("topic = ?", topic)
	}

	if created_from, ok := args["created_from"]; ok {
		tx = tx.Where("created_at >= ?", created_from)
	}

	if created_to, ok := args["created_to"]; ok {
		tx = tx.Where("created_at <= ?", created_to)
	}

	return tx
}

func (s *Repository) ListMessage(ctx context.Context, args map[string]interface{}) (*[]model.Message, error) {
	var result []model.Message
	ch := make(chan error)
	go func() {
		defer close(ch)

		tx := getListMessageDB(s.db, args)

		if limit, ok := args["limit"]; ok {
			tx = tx.Limit(limit)
		} else {
			tx = tx.Limit(40)
		}

		if offset, ok := args["offset"]; ok {
			tx = tx.Offset(offset)
		}

		if orderBy, ok := args["order_by"]; ok {
			tx = tx.Order(orderBy)
		} else {
			tx = tx.Order("id DESC")
		}

		ch <- tx.Find(&result).Error

	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-ch:
		if err != nil {
			log.WithError(err).WithField("args", args).Error("Failed to list message")
			return nil, err
		}
	}

	return &result, nil
}

func (s *Repository) PublishMessage(ctx context.Context, id int64) (*model.Publish, error) {
	result := model.Publish{MsgId: id, PublishAt: time.Now()}
	ch := make(chan error)
	go func() {
		defer close(ch)
		ch <- s.db.Create(&result).Error
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-ch:
		if err != nil {
			return nil, err
		}
	}

	return &result, nil
}
