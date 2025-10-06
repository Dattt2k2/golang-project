package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type EmailMessage struct {
	To       string                 `json:"to"`
	Subject  string                 `json:"subject"`
	Template string                 `json:"template"`
	Data     map[string]interface{} `json:"data"`
}

func NewKafkaWriter(broker, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}

func SendEmailMessage(writer *kafka.Writer, msg EmailMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return writer.WriteMessages(context.Background(), kafka.Message{Value: data})
}

// SendJSONMessage marshals v to JSON and writes to kafka writer.
func SendJSONMessage(writer *kafka.Writer, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return writer.WriteMessages(context.Background(), kafka.Message{Value: data})
}
