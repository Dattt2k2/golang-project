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
	Source     string          `json:"source"` // "cart" or "direct"
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
				log.Printf("[CartService] Error reading message from order_success: %v", err)
				continue
			}

			var event OrderSuccessEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("[CartService] Failed to unmarshal order success event: %v", err)
				continue
			}

			log.Printf("[CartService] Received order success event for UserID=%s, OrderID=%s, Source=%s", event.UserID, event.OrderID, event.Source)

			// Chỉ xóa cart items nếu order từ cart
			if event.Source != "cart" {
				log.Printf("[CartService] Skipping cart deletion because order source is '%s', not 'cart'", event.Source)
				continue
			}

			// Extract product IDs from order items
			productIDs := make([]string, 0, len(event.Items))
			for _, item := range event.Items {
				productIDs = append(productIDs, item.ProductID)
			}

			if err := cartRepo.DeleteCartItems(context.Background(), event.UserID, productIDs); err != nil {
				log.Printf("[CartService] Failed to delete cart items for user %s: %v", event.UserID, err)
			} else {
				log.Printf("[CartService] Successfully deleted %d cart items for user %s", len(productIDs), event.UserID)
			}
		}
	}()

	log.Printf("Order success consumer initialized with brokers: %v", brokers)
}
