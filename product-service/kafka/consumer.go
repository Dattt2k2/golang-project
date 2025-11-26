package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"product-service/models"

	"github.com/segmentio/kafka-go"
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
	log.Printf("🔄 Starting Kafka consumer for order_success with brokers: %v", brokers)
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    OrderSuccessTopic,
		GroupID:  "product-service",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	go func() {
		log.Printf("Kafka reader created for topic: %s", OrderSuccessTopic)
		for {
			log.Printf("Waiting for message on topic: %s", OrderSuccessTopic)
			message, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Printf("Error reading message from Kafka (topic: %s): %v", OrderSuccessTopic, err)
				time.Sleep(5 * time.Second) // Wait before retrying
				continue
			}

			log.Printf("Received message from Kafka: key=%s, topic=%s, partition=%d, offset=%d",
				string(message.Key), message.Topic, message.Partition, message.Offset)

			var event OrderSuccessEvent
			if err := json.Unmarshal(message.Value, &event); err != nil {
				log.Printf(" Error unmarshalling message: %v", err)
				continue
			}

			log.Printf("Received order_success event: OrderID=%s, Items=%d", event.OrderID, len(event.Items))

			stockItems := make([]models.StockUpdateItem, len(event.Items))
			for i, item := range event.Items {
				stockItems[i] = models.StockUpdateItem{
					ProductID: item.ProductID,
					Quantity:  item.Quantity,
				}
			}

			// Decrease stock (trừ số lượng tồn kho)
			for _, item := range stockItems {
				log.Printf("Decreasing stock for product %s by %d", item.ProductID, item.Quantity)
				if err := updater.UpdateProductStock(context.Background(), item.ProductID, item.Quantity); err != nil {
					log.Printf("Error updating product stock: %v", err)
				} else {
					log.Printf("Stock decreased for product %s", item.ProductID)
				}
			}

			// Increase sold count (cộng số lượng đã bán)
			for _, item := range stockItems {
				log.Printf("⬆Increasing sold count for product %s by %d", item.ProductID, item.Quantity)
				if err := updater.IncrementSoldCount(context.Background(), item.ProductID, item.Quantity); err != nil {
					log.Printf("Error incrementing sold count: %v", err)
				} else {
					log.Printf("Sold count increased for product %s", item.ProductID)
				}
			}

			log.Printf("Finished processing order_success: OrderID=%s", event.OrderID)

			// Commit the message
			if err := reader.CommitMessages(context.Background(), message); err != nil {
				log.Printf("Error committing message: %v", err)
			} else {
				log.Printf("Message committed for OrderID=%s", event.OrderID)
			}
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
		log.Printf(" Starting Kafka consumer for topic: %s with brokers: %v", OrderReturnedTopic, brokers)
		for {
			message, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Printf("Error reading message from Kafka: %v", err)
				continue
			}

			log.Printf("📨 Received message from topic %s: key=%s, offset=%d", OrderReturnedTopic, string(message.Key), message.Offset)

			var event OrderReturnedEvent
			if err := json.Unmarshal(message.Value, &event); err != nil {
				log.Printf("Error unmarshalling message: %v, raw message: %s", err, string(message.Value))
				continue
			}

			log.Printf("📨 Received order_returned event: OrderID=%s, UserID=%s, Items=%d, TotalPrice=%.2f",
				event.OrderID, event.UserID, len(event.Items), event.TotalPrice)

			stockItems := make([]models.StockUpdateItem, len(event.Items))
			for i, item := range event.Items {
				stockItems[i] = models.StockUpdateItem{
					ProductID: item.ProductID,
					Quantity:  item.Quantity,
				}
			}
			for _, item := range stockItems {
				log.Printf("⬆Increasing stock for product %s by %d (order returned)", item.ProductID, item.Quantity)
				// For returns, we need to INCREASE stock, so pass negative quantity to UpdateProductStock
				if err := updater.UpdateProductStock(context.Background(), item.ProductID, -item.Quantity); err != nil {
					log.Printf("Error increasing product stock: %v", err)
				} else {
					log.Printf("Stock increased for product %s", item.ProductID)
				}
			}

			for _, item := range stockItems {
				log.Printf("⬇Decreasing sold count for product %s by %d (order returned)", item.ProductID, item.Quantity)
				if err := updater.DecrementSoldCount(context.Background(), item.ProductID, item.Quantity); err != nil {
					log.Printf("Error decrementing sold count: %v", err)
				} else {
					log.Printf("Sold count decreased for product %s", item.ProductID)
				}
			}

			log.Printf("Finished processing order_returned: OrderID=%s", event.OrderID)

			if err := reader.CommitMessages(context.Background(), message); err != nil {
				log.Printf("Error committing message: %v", err)
			} else {
				log.Printf("Message committed for OrderID=%s", event.OrderID)
			}
		}
	}()
	log.Printf("Kafka consumer started for topic: %s", OrderReturnedTopic)
}
