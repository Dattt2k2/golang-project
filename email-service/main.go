package main

import (
	"log"
	"email-service/service"
)

func main() {
	emailService := service.NewEmailService()

	err := service.StartKafkaConsumer(emailService)
	if err != nil {
		log.Fatalf("Failed to start Kafka consumer: %v", err)
	}
}