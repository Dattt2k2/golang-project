package kafka

import (
	"context"
	"encoding/json"
	"log"

	controller "github.com/Dattt2k2/golang-project/cart-service/controller"
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

func ConsumeOrderSuccess(brokers []string, cartCtrl *controller.CartController) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    OrderSuccessTopic,
		GroupID:  "cart-service",
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

			if event.UserID == "" {
				log.Printf("Received event with empty user ID")
				continue
			}

			// Sử dụng InternalClearCart thay vì ClearCart trực tiếp
			if err := cartCtrl.InternalClearCart(event.UserID); err != nil {
				log.Printf("Error clearing cart: %v", err)
				continue
			}
			log.Printf("Cart cleared for user ID: %s", event.UserID)
		}
	}()

	log.Printf("Order success consumer initialized with brokers: %v", brokers)
}
