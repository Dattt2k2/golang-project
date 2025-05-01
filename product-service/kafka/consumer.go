package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Dattt2k2/golang-project/product-service/models"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	OrderSuccessTopic  = "order_success"
	OrderReturnedTopic = "order_returned"
)

type OrderSuccessEvent struct {
	OrderID    string          `json:"order_id"`
	UserID     string          `json:"user_id"`
	Items      []OrderItemInfo `json:"items"`
	TotalPrice float64         `json:"total_price"`
}

type OrderItemInfo struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type OrderReturnedEvent struct {
	OrderID    string          `json:"order_id"`
	UserID     string          `json:"user_id"`
	Items      []OrderItemInfo `json:"items"`
	TotalPrice float64         `json:"total_price"`
}

func ConsumeOrderSuccess(brokers []string, updater models.ProductStockUpdater) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    OrderSuccessTopic,
		GroupID:  "product-service",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	go func() {
		for {
			message, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Printf("Error reading message: %v", err)
				continue
			}

			var event OrderSuccessEvent
			if err := json.Unmarshal(message.Value, &event); err != nil {
				log.Printf("Error unmarshalling message: %v", err)
				continue
			}

			userID, err := primitive.ObjectIDFromHex(event.UserID)
			if err != nil {
				log.Printf("Error converting user ID: %v", err)
				continue
			}

			stockItems := make([]models.StockUpdateItem, len(event.Items))
			for i, item := range event.Items {
				stockItems[i] = models.StockUpdateItem{
					ProductID: item.ProductID,
					Quantity:  item.Quantity,
				}
			}
			for _, item := range stockItems {
				id, err := primitive.ObjectIDFromHex(item.ProductID)
				if err != nil {
					log.Printf("Invalid product ID: %v", item.ProductID)
					continue
				}
				if err := updater.UpdateProductStock(context.Background(), id, -item.Quantity); err != nil {
					log.Printf("Error updating product stock: %v", err)
				}
			}

			for _, item := range stockItems {
				if err := updater.IncrementSoldCount(context.Background(), item.ProductID, item.Quantity); err != nil {
					log.Printf("Error incrementing sold count: %v", err)
				}
			}

			log.Printf("Product stock updated successfully for user: %v", userID)
		}
	}()

	log.Printf("Kafka consumer started for topic: %s", OrderSuccessTopic)
}

func ConsumerOrderReturned(brokers []string, updater models.ProductStockUpdater) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    OrderReturnedTopic,
		GroupID:  "product-service",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	go func() {
		for {
			message, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Printf("Error reading message: %v", err)
				continue
			}
			var event OrderReturnedEvent
			if err := json.Unmarshal(message.Value, &event); err != nil {
				log.Printf("Error unmarshalling message: %v", err)
				continue
			}

			userID, err := primitive.ObjectIDFromHex(event.UserID)
			if err != nil {
				log.Printf("Error converting user ID: %v", err)
				continue
			}

			stockItems := make([]models.StockUpdateItem, len(event.Items))
			for i, item := range event.Items {
				stockItems[i] = models.StockUpdateItem{
					ProductID: item.ProductID,
					Quantity:  item.Quantity,
				}
			}
			for _, item := range stockItems {
				id, err := primitive.ObjectIDFromHex(item.ProductID)
				if err != nil {
					log.Printf("Invalid product ID: %v", item.ProductID)
					continue
				}
				if err := updater.UpdateProductStock(context.Background(), id, item.Quantity); err != nil {
					log.Printf("Error updating product stock: %v", err)
				}
			}

			for _, item := range stockItems {
				if err := updater.DecrementSoldCount(context.Background(), item.ProductID, item.Quantity); err != nil {
					log.Printf("Error decrementing sold count: %v", err)
				}
			}

			log.Printf("Product stock updated successfully for user: %v", userID)
		}
	}()
	log.Printf("Kafka consumer started for topic: %s", OrderReturnedTopic)
}
