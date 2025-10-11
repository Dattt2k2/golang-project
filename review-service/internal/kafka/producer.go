package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		RequiredAcks: kafka.RequireAll,
	}

	return &Producer{writer: writer}
}

func (p *Producer) PublishRatingUpdate(ctx context.Context, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(fmt.Sprintf("rating-update-%d", time.Now().Unix())),
		Value: data,
		Time:  time.Now(),
	})

	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	log.Printf("Published rating update to Kafka: %s", string(data))
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
