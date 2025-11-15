package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"order-service/repositories"

	"github.com/segmentio/kafka-go"
)

// PaymentEvent is the shape produced by payment-service
type PaymentEvent struct {
	OrderID         string  `json:"order_id"`
	PaymentIntentID string  `json:"payment_intent_id"`
	Amount          float64 `json:"amount"`
	Status          string  `json:"status"`
}

// PaymentEventHandler defines interface for handling payment events
type PaymentEventHandler interface {
	HandlePaymentSuccess(ctx context.Context, orderID string, paymentIntentID string) error
	HandlePaymentFailure(ctx context.Context, orderID string, reason string) error
}

func StartPaymentConsumer(brokers []string, repo *repositories.OrderRepository, handler PaymentEventHandler) *kafka.Reader {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          "checkout_completed",
		GroupID:        "order-service-group",
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.FirstOffset,
	})

	go func() {
		defer r.Close()

		log.Printf("✅ Kafka consumer started, listening to topic: checkout_completed")

		for {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			m, err := r.FetchMessage(ctx)
			cancel()

			if err != nil {
				// Nếu chỉ là hết thời gian chờ (không có message mới) → bỏ qua
				if err == context.DeadlineExceeded {
					continue
				}

				// Các lỗi khác (ví dụ mất kết nối Kafka, group rebalance,...) → log và retry
				log.Printf("❌ Kafka fetch error: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			log.Printf("📨 Received Kafka message: %s", string(m.Value))

			var ev PaymentEvent
			if err := json.Unmarshal(m.Value, &ev); err != nil {
				log.Printf("⚠️ Invalid payment event: %v", err)
				_ = r.CommitMessages(context.Background(), m) // Commit để skip message lỗi
				continue
			}

			log.Printf("🔄 Processing payment event: OrderID=%s, Status=%s", ev.OrderID, ev.Status)

			// Handle payment event based on status
			processCtx := context.Background()
			var processErr error

			switch ev.Status {
			case "checkout_completed":
				log.Printf("✅ Payment successful, calling HandlePaymentSuccess for order: %s", ev.OrderID)
				processErr = handler.HandlePaymentSuccess(processCtx, ev.OrderID, ev.PaymentIntentID)
				if processErr != nil {
					log.Printf("❌ Failed to handle payment success: %v", processErr)
				} else {
					log.Printf("✅ Successfully handled payment success for order: %s", ev.OrderID)
				}

			case "payment_failed", "checkout_failed":
				log.Printf("❌ Payment failed for order: %s", ev.OrderID)
				processErr = handler.HandlePaymentFailure(processCtx, ev.OrderID, "Payment failed")
				if processErr != nil {
					log.Printf("❌ Failed to handle payment failure: %v", processErr)
				}

			default:
				log.Printf("⚠️ Unknown payment status: %s for order: %s", ev.Status, ev.OrderID)
			}

			// ❌ REMOVED: Don't override payment_status here
			// HandlePaymentSuccess already sets the correct status ("HELD")
			// if err := repo.UpdatePaymentStatus(context.Background(), ev.OrderID, ev.Status); err != nil {
			// 	log.Printf("⚠️ Failed to update payment status for order %s: %v", ev.OrderID, err)
			// 	continue
			// }

			// Commit message sau khi xử lý thành công
			if err := r.CommitMessages(context.Background(), m); err != nil {
				log.Printf("⚠️ Failed to commit message for order %s: %v", ev.OrderID, err)
			} else {
				log.Printf("✅ Completed processing payment event for order %s", ev.OrderID)
			}
		}
	}()

	return r
}
