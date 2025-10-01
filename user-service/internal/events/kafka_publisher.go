package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaPublisher publishes events to Kafka using segmentio/kafka-go.
type KafkaPublisher struct {
	brokers      []string
	defaultTopic string
}

func NewKafkaPublisher(brokers []string, defaultTopic string) *KafkaPublisher {
	return &KafkaPublisher{brokers: brokers, defaultTopic: defaultTopic}
}

func (p *KafkaPublisher) Publish(topic string, payload interface{}) error {
	if topic == "" {
		topic = p.defaultTopic
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	w := &kafka.Writer{
		Addr:         kafka.TCP(p.brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
	}
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msg := kafka.Message{
		Key:   nil,
		Value: b,
		Time:  time.Now(),
	}

	return w.WriteMessages(ctx, msg)
}
