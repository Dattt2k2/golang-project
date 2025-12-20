package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	logger "product-service/log"
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
	VariantID string  `json:"variant_id"`
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
		for {
			message, err := reader.ReadMessage(context.Background())
			if err != nil {
				time.Sleep(5 * time.Second) // Wait before retrying
				continue
			}

			var event OrderSuccessEvent
			if err := json.Unmarshal(message.Value, &event); err != nil {
				continue
			}

			stockItems := make([]models.StockUpdateItem, len(event.Items))
			for i, item := range event.Items {
				log.Printf("  Item[%d]: ProductID=%s, VariantID=%s, Quantity=%d", i, item.ProductID, item.VariantID, item.Quantity)

				variantID := item.VariantID
				// Fallback: nếu VariantID rỗng (cart items cũ), lấy variant đầu tiên
				if variantID == "" {
					if productGetter, ok := updater.(interface {
						GetProductForgRPC(ctx context.Context, id string) (*models.Product, error)
					}); ok {
						if product, err := productGetter.GetProductForgRPC(context.Background(), item.ProductID); err == nil && len(product.Variants) > 0 {
							variantID = product.Variants[0].ID
							log.Printf("⚠️ VariantID was empty, using first variant: %s for product %s", variantID, item.ProductID)
						}
					}
				}

				stockItems[i] = models.StockUpdateItem{
					ProductID: item.ProductID,
					VariantID: variantID,
					Quantity:  item.Quantity,
				}
			}

			// Increase sold count (cộng số lượng đã bán)
			for _, item := range stockItems {
				if err := updater.IncrementSoldCount(context.Background(), item.ProductID, item.Quantity); err != nil {
					logger.Err("Failed to increment sold count", err)
				}

				// Decrease stock (trừ stock)
				if err := updater.UpdateProductStock(context.Background(), item.ProductID, item.VariantID, item.Quantity); err != nil {
					logger.Err("Failed to update product stock", err)
				}
			}
		}
	}()
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
				variantID := item.VariantID
				// Fallback: nếu VariantID rỗng, lấy variant đầu tiên
				if variantID == "" {
					if productGetter, ok := updater.(interface {
						GetProductForgRPC(ctx context.Context, id string) (*models.Product, error)
					}); ok {
						if product, err := productGetter.GetProductForgRPC(context.Background(), item.ProductID); err == nil && len(product.Variants) > 0 {
							variantID = product.Variants[0].ID
							log.Printf("⚠️ [Returned] VariantID was empty, using first variant: %s for product %s", variantID, item.ProductID)
						}
					}
				}

				stockItems[i] = models.StockUpdateItem{
					ProductID: item.ProductID,
					VariantID: variantID,
					Quantity:  item.Quantity,
				}
			}
			for _, item := range stockItems {
				log.Printf("⬆Increasing stock for product %s, variant %s by %d (order returned)", item.ProductID, item.VariantID, item.Quantity)
				// For returns, we need to INCREASE stock, so pass negative quantity to UpdateProductStock
				if err := updater.UpdateProductStock(context.Background(), item.ProductID, item.VariantID, -item.Quantity); err != nil {
					log.Printf("Error increasing product stock: %v", err)
				} else {
					log.Printf("Stock increased for product %s, variant %s", item.ProductID, item.VariantID)
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
