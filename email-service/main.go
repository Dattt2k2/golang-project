package main

import (
	"email-service/logger"
	"email-service/service"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load("./email-service/.env")
	logger.InitLogger()
	defer logger.Sync()
	emailService := service.NewEmailService()

	err := service.StartKafkaConsumer(emailService)
	if err != nil {
		log.Fatalf("Failed to start Kafka consumer: %v", err)
	}
}