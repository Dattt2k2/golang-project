package service

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokerAddress string, topic string) *KafkaProducer {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{brokerAddress},
		Topic:   topic,
	})
	return &KafkaProducer{writer: writer}
}

func (kp *KafkaProducer) SendMessage(ctx context.Context, message interface{}) error {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return err
	}

	err = kp.writer.WriteMessages(ctx, kafka.Message{
		Value: messageBytes,
	})
	if err != nil {
		log.Printf("Error sending message to Kafka: %v", err)
		return err
	}

	log.Println("Message sent to Kafka successfully")
	return nil
}

func (kp *KafkaProducer) Close() error {
	return kp.writer.Close()
}