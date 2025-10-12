package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	log.Printf("🔧 Initializing Kafka producer: brokers=%v, topic=%s", brokers, topic)

	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
		KeepAlive: 30 * time.Second,
		Resolver: &net.Resolver{
			PreferGo:     true,
			StrictErrors: false,
		},
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		RequiredAcks: kafka.RequireAll,
		Transport:    &kafka.Transport{Dial: dialer.DialFunc},
	}

	return &Producer{writer: writer}
}

func (p *Producer) PublishRatingUpdate(ctx context.Context, message interface{}) error {
	var err error
	maxRetries := 3
	backoff := 2 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = p.publishOnce(ctx, message)
		if err == nil {
			return nil
		}

		if attempt < maxRetries {
			log.Printf("⚠️  Kafka publish attempt %d/%d failed: %v. Retrying in %v...",
				attempt, maxRetries, err, backoff)
			time.Sleep(backoff)
			backoff *= 2
		}
	}

	return fmt.Errorf("failed to publish after %d retries: %w", maxRetries, err)
}

func (p *Producer) publishOnce(ctx context.Context, message interface{}) error {
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

	log.Printf("✅ Published rating update to Kafka: %s", string(data))
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
