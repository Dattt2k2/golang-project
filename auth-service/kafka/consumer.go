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

func NewKafkaReader(broker, topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
		GroupID: groupID,
	})
}

func ConsumeUserDeleted(reader *kafka.Reader, handleFunc func(payload UserDeletedPayload)) {
    logger.Info("Started consuming user.deleted topic")
    for {
        msg, err := reader.ReadMessage(context.Background())
        if err != nil {
            logger.Err("Error reading message", err)
            time.Sleep(time.Second)
            continue
        }
        logger.Info("Received raw message: " + string(msg.Value))

        var payload UserDeletedPayload
        if err := json.Unmarshal(msg.Value, &payload); err != nil {
            logger.Err("Unmarshal failed", err)
            continue
        }
        logger.Info("Processed user.deleted payload id=" + payload.UserID + " email=" + payload.Email)

        // xử lý
        handleFunc(payload)

        // commit offset (kafka-go)
        if err := reader.CommitMessages(context.Background(), msg); err != nil {
            logger.Err("Commit message failed", err)
        }
    }
}
