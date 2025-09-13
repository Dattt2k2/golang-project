package service

import (
	"context"
	"encoding/json"
	logger "payment-service/src/utils"
	"github.com/segmentio/kafka-go"
	"log"
)

type PaymentMessage struct {
	OrderID string `json:"order_id"`
	Amount  float64 `json:"amount"`
	Status  string `json:"status"`
}

func NewKafkaConsumer(brokerAddress string, topic string) *kafka.Reader {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokerAddress},
		Topic:   topic,
		GroupID: "payment-service-group",
	})

	return r
}
func RunKafkaConsumer(r *kafka.Reader) {
	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message from Kafka: %v", err)
			continue
		}

		var paymentMsg PaymentMessage
		if err := json.Unmarshal(m.Value, &paymentMsg); err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			continue
		}

		processPayment(paymentMsg)
	}
}

func processPayment(msg PaymentMessage) {
	// Implement payment processing logic here
	logger.Info("Processing payment for Order ID: " + msg.OrderID)
	// Add further processing logic as needed
}