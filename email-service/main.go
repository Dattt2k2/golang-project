package main

import (
	"email-service/logger"
	"email-service/service"
	"log"
)

func main() {
	logger.InitLogger()
	defer logger.Sync()
	emailService := service.NewEmailService()

	err := service.StartKafkaConsumer(emailService)
	if err != nil {
		log.Fatalf("Failed to start Kafka consumer: %v", err)
	}
}