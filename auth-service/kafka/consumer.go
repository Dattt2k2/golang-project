package kafka

import (
	"auth-service/logger"
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

type UserDeletedPayload struct {
	UserID string `json:"id"`
	Email  string `json:"email"`
}

type UserDisabledPayload struct {
	UserID     string `json:"id"`
	Email      string `json:"email"`
	IsDisabled bool   `json:"is_disabled"`
}

func NewKafkaReader(broker, topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
		GroupID: groupID,
	})
}

func ConsumeUserDeleted(reader *kafka.Reader, handleFunc func(payload UserDeletedPayload)) {
	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			logger.Err("Error reading message", err)
			time.Sleep(time.Second)
			continue
		}

		var payload UserDeletedPayload
		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			logger.Err("Unmarshal failed", err)
			continue
		}

		// xử lý
		handleFunc(payload)

		// commit offset (kafka-go)
		if err := reader.CommitMessages(context.Background(), msg); err != nil {
			logger.Err("Commit message failed", err)
		}
	}
}

func ConsumeUserDisabled(reader *kafka.Reader, handleFunc func(payload UserDisabledPayload)) {
	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			logger.Err("Error reading message", err)
			time.Sleep(time.Second)
			continue
		}

		var payload UserDisabledPayload
		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			logger.Err("Unmarshal failed", err)
			continue
		}

		// xử lý
		handleFunc(payload)

		// commit offset (kafka-go)
		if err := reader.CommitMessages(context.Background(), msg); err != nil {
			logger.Err("Commit message failed", err)
		}
	}
}
