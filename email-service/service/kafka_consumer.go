package service

import (
	"context"
	logger "email-service/log"
	"encoding/json"
	"os"

	"github.com/segmentio/kafka-go"
)

type EmailKafkaMessage struct {
	To string `json:"to"`
	Subject string `json:"subject"`
	TemplatePath string `json:"template_path"`
	Data interface{} `json:"data"`
}

func StartKafkaConsumer(emailService *EmailService) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{os.Getenv("KAFKA_BROKER")},
		Topic: "email_topic",
		GroupID: "email_service_group",
	})

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			logger.Error("Failed to read message from Kafka", logger.ErrField(err))
			continue
		}

		var msg EmailKafkaMessage

		if err := json.Unmarshal(m.Value, &msg); err != nil {
			logger.Error("Failed to unmarshal Kafka message", logger.ErrField(err))
			continue
		}

		if err := emailService.SendEmail(msg.To, msg.Subject, msg.TemplatePath, msg.Data); err != nil {
			logger.Error("Failed to send email", logger.ErrField(err))
		}
	}
}