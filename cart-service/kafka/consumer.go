package kafka

import (
	"context"
	"encoding/json"
	"log"

	// controller "cart-service/controller"
	repositories "cart-service/repository"

	"github.com/segmentio/kafka-go"
)

const (
	OrderSuccessTopic = "order_success"
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

type CartDeleteEvent struct {
	UserID     string   `json:"user_id"`
	ProductIDs []string `json:"product_ids"`
}

func StartCartDeleteConsumer(brokers []string, groupID string, cartRepo repositories.CartRepository) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   "cart_delete_items",
		GroupID: groupID,
	})

	log.Printf("[CartService] Kafka consumer started for topic: cart-delete-items")

	go func() {
		for {
			msg, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Printf("[CartService] Error reading message: %v", err)
				continue
			}

			var event CartDeleteEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("[CartService] Failed to unmarshal cart delete event: %v", err)
				continue
			}

			log.Printf("[CartService] Received cart delete event: %+v", event)

			if err := cartRepo.DeleteCartItems(context.Background(), event.UserID, event.ProductIDs); err != nil {
				log.Printf("[CartService] Failed to delete cart items for user %s: %v", event.UserID, err)
			} else {
				log.Printf("[CartService] Successfully deleted cart items for user %s", event.UserID)
			}

			if err := reader.CommitMessages(context.Background(), msg); err != nil {
				log.Printf("[CartService] Error committing message: %v", err)
			}
		}
	}()
}

func ConsumeOrderSuccess(brokers []string, cartRepo repositories.CartRepository) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    OrderSuccessTopic,
		GroupID:  "cart-service",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	go func() {
		for {
			msg, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Printf("[CartService] Error reading message: %v", err)
				continue
			}

			var event CartDeleteEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("[CartService] Failed to unmarshal cart delete event: %v", err)
				continue
			}

			log.Printf("[CartService] Received cart delete event: %+v", event)

			if err := cartRepo.DeleteCartItems(context.Background(), event.UserID, event.ProductIDs); err != nil {
				log.Printf("[CartService] Failed to delete cart items for user %s: %v", event.UserID, err)
			} else {
				log.Printf("[CartService] Successfully deleted cart items for user %s", event.UserID)
			}
		}
	}()

	log.Printf("Order success consumer initialized with brokers: %v", brokers)
}
