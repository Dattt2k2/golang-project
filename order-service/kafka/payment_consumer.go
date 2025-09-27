package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

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
        Brokers:        brokers,
        Topic:          "payment",
        GroupID:        "order-service-group",
        MinBytes:       10e3, // 10KB
        MaxBytes:       10e6, // 10MB
        CommitInterval: time.Second,
    })

    go func() {
        defer r.Close()
        
        for {
            // Thêm timeout để tránh block vô hạn
            ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
            
            m, err := r.FetchMessage(ctx)
            cancel()
            
            if err != nil {
                log.Printf("error fetching payment message: %v", err)
                time.Sleep(5 * time.Second) // Delay trước khi retry
                continue
            }

            var ev PaymentEvent
            if err := json.Unmarshal(m.Value, &ev); err != nil {
                log.Printf("invalid payment event: %v", err)
                r.CommitMessages(context.Background(), m) // Commit để skip message lỗi
                continue
            }

            var oid uint64
            if oid, err = strconv.ParseUint(ev.OrderID, 10, 64); err != nil { // Sửa parseUint
                log.Printf("invalid order id in payment event: %v", err)
                r.CommitMessages(context.Background(), m)
                continue
            }

            if err := repo.UpdatePaymentStatus(context.Background(), uint(oid), ev.Status); err != nil {
                log.Printf("failed to update payment status for order %s: %v", ev.OrderID, err)
                // Không commit nếu DB update fail → sẽ retry
                continue
            }

            // Commit message sau khi xử lý thành công
            r.CommitMessages(context.Background(), m)
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
