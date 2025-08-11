package main

import (
	"log"
	"payment-service/src/config"
	"payment-service/src/service"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize Kafka producer
	producer, err := service.NewKafkaProducer(cfg.Kafka.Broker)
	if err != nil {
		log.Fatalf("Error initializing Kafka producer: %v", err)
	}
	defer producer.Close()

	// Initialize Kafka consumer
	consumer, err := service.NewKafkaConsumer(cfg.Kafka.Broker, cfg.Kafka.Topic)
	if err != nil {
		log.Fatalf("Error initializing Kafka consumer: %v", err)
	}
	defer consumer.Close()

	// Start the Kafka consumer in a separate goroutine
	go func() {
		if err := consumer.Start(); err != nil {
			log.Fatalf("Error starting Kafka consumer: %v", err)
		}
	}()

	// Start the payment service (e.g., HTTP server)
	if err := service.StartPaymentService(cfg); err != nil {
		log.Fatalf("Error starting payment service: %v", err)
	}
}