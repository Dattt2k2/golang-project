package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Dattt2k2/golang-project/order-service/models"
	"github.com/segmentio/kafka-go"
)

const (
	OrderSuccessTopic = "order_success"
)

var (
	orderSuccessWriter *kafka.Writer
)

type OrderSuccessEvent struct {
	OrderID string `json:"order_id"`
	UserID string `json:"user_id"`
	Items []OrderItemInfo `json:"items"`
	TotalPrice float64 `json:"total_price"`
}

type OrderItemInfo struct {
	ProductID string `json:"product_id"`
	Quantity int `json:"quantity"`
	Price float64 `json:"price"`
}

func InitOrderSuccessProducer(brokers []string) {
	orderSuccessWriter = &kafka.Writer{
		Addr:    kafka.TCP(brokers...),
		Topic:   OrderSuccessTopic,
		Balancer: &kafka.LeastBytes{},
	}
	log.Printf("Order success producer initialized with brokers: %v", brokers)
}

func ProduceOrderSuccessEvent(ctx context.Context, order models.Order) error {
	if orderSuccessWriter == nil {
		log.Printf("Order success producer not initialized")
		return fmt.Errorf("Order success producer not initialized")
	}

	orderEvent := OrderSuccessEvent{
		OrderID: order.ID.Hex(),
		UserID: order.UserID.Hex(),
		TotalPrice: order.TotalPrice,
		Items: make([]OrderItemInfo, len(order.Items)),
	}

	for i, item := range order.Items {
		orderEvent.Items[i] = OrderItemInfo{
			ProductID: item.ProductID.Hex(),
			Quantity: item.Quantity,
			Price: item.Price,
		}
	}

	messagePayload, err := json.Marshal(orderEvent)
	if err != nil{
		log.Printf("Error marshalling order event: %v", err)
		return err
	}

	message := kafka.Message{
		Key:  []byte(order.ID.Hex()),
		Value: messagePayload,
	}

	if err := orderSuccessWriter.WriteMessages(ctx, message); err != nil{
		log.Printf("Failed to write message: %v", err)
		return err
	}

	log.Printf("Order success event produced: %v", orderEvent)
	return nil
}


func CloseOrderSuccessProducer(){
	if orderSuccessWriter != nil{
		orderSuccessWriter.Close()
	}
}


