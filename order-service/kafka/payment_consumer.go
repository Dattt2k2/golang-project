package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"order-service/repositories"

	"github.com/segmentio/kafka-go"
)

// PaymentEvent is the shape produced by payment-service
type PaymentEvent struct {
	OrderID string  `json:"order_id"`
	Amount  float64 `json:"amount"`
	Status  string  `json:"status"`
}

func StartPaymentConsumer(brokers []string, repo *repositories.OrderRepository) *kafka.Reader {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   "payment",
		GroupID: "order-service-group",
	})

	go func() {
		for {
			m, err := r.ReadMessage(context.Background())
			if err != nil {
				log.Printf("error reading payment message: %v", err)
				continue
			}

			var ev PaymentEvent
			if err := json.Unmarshal(m.Value, &ev); err != nil {
				log.Printf("invalid payment event: %v", err)
				continue
			}

			// attempt to parse OrderID as uint (order IDs are numeric)
			var oid uint64
			if oid, err = parseUint(ev.OrderID); err != nil {
				log.Printf("invalid order id in payment event: %v", err)
				continue
			}

			// update payment status in DB using repo
			if err := repo.UpdatePaymentStatus(context.Background(), uint(oid), ev.Status); err != nil {
				log.Printf("failed to update payment status for order %s: %v", ev.OrderID, err)
				continue
			}
			log.Printf("updated payment status for order %s to %s", ev.OrderID, ev.Status)
		}
	}()

	return r
}

func parseUint(s string) (uint64, error) {
	var x uint64
	_, err := fmt.Sscan(s, &x)
	return x, err
}
