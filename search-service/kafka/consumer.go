package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"search-service/models"
	"search-service/service"
	"github.com/segmentio/kafka-go"
)

type ProductEvent struct {
	Type    string          `json:"type"`
	Product *models.Product `json:"product"`
	ID      string          `json:"id"`
}

func InitProductEventConsumer(svc service.SearchService, brokers []string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    "product-events",
		GroupID:  "search-service",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	defer r.Close()

	log.Printf("Product event consumer initialized with brokers: %v", brokers)
	log.Printf("Kafka consumer starting for topic: product-events with brokers: %v", brokers)

	isConnected := false // Biến để chỉ log kết nối thành công 1 lần

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message: %v", err) // Log lỗi đọc/kết nối
			isConnected = false                          // Đánh dấu mất kết nối nếu có lỗi
			time.Sleep(5 * time.Second)                  // Chờ chút trước khi thử lại
			continue
		}

		// Log khi kết nối/đọc thành công lần đầu hoặc sau khi mất kết nối
		if !isConnected {
			log.Println("Kafka consumer connected and reading messages...")
			isConnected = true
		}

		log.Printf("Received message offset %d: %s", m.Offset, string(m.Value)) // Log khi đọc được message

		var event ProductEvent
		if err := json.Unmarshal(m.Value, &event); err != nil {
			log.Printf("Error unmarshalling message at offset %d: %v", m.Offset, err)
			continue
		}

		log.Printf("Successfully unmarshalled event at offset %d: %+v", m.Offset, event)

		switch event.Type {
		case "created", "updated":
			if event.Product != nil {
				// Sửa log lỗi index: lấy lỗi từ svc.IndexProduct
				if indexErr := svc.IndexProduct(event.Product); indexErr != nil {
					log.Printf("Error indexing product %s: %v", event.Product.ID, indexErr) // Log lỗi index đúng
				} else {
					log.Printf("Successfully indexed product: %s", event.Product.ID)
				}
			}
		case "deleted":
			if event.ID != "" {
				// Sửa log lỗi delete: lấy lỗi từ svc.DeleteProduct
				if deleteErr := svc.DeleteProduct(event.ID); deleteErr != nil {
					log.Printf("Error deleting product %s: %v", event.ID, deleteErr) // Log lỗi delete đúng
				} else {
					log.Printf("Successfully deleted product: %s", event.ID)
				}
			}
		case "INITIAL_SYNC":
			if event.Product != nil {
				if indexErr := svc.IndexProduct(event.Product); indexErr != nil {
					log.Printf("Error indexing product during INITIAL_SYNC %s: %v", event.Product.ID, indexErr)
				} else {
					log.Printf("Indexed product during INITIAL_SYNC: %s", event.Product.ID)
				}
			}
		}
	}
}
