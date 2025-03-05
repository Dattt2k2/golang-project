package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type PaymentOrder struct{
	UserId string `json:"user_id"`
	Amount float64 `json:"amount"`
	Products interface{} `json:"products"`
}

var writer *kafka.Writer

func init(){
	writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic: "payment",
		Balancer: &kafka.LeastBytes{},
	})
}

func ProducePaymentOrder(order PaymentOrder) error{
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	message, err := json.Marshal(order)
	if err != nil{
		log.Printf("Failed to marshal order: %v", err)
		return err
	}

	err = writer.WriteMessages(ctx, kafka.Message{
		Key: []byte(order.UserId),
		Value: message,
	})

	if err != nil{
		log.Printf("Failed to write kafka message: %v", err)
		return err
	}
	return nil
}