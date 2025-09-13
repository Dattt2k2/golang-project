package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	logger "order-service/log"
	"order-service/models"

	"github.com/segmentio/kafka-go"
)

const (
	OrderSuccessTopic  = "order_success"
	OrderReturnedTopic = "order_returned"
)

var (
	orderSuccessWriter  *kafka.Writer
	orderReturnedWriter *kafka.Writer
)

// OrderSuccessEvent represents the structure of the order success event message
// that will be sent to the Kafka topic.
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

func InitOrderSuccessProducer(brokers []string) {
	orderSuccessWriter = &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    OrderSuccessTopic,
		Balancer: &kafka.LeastBytes{},
	}
	logger.Logger.Infof("Order success producer initialized with brokers: %v", brokers)
}

func ProduceOrderSuccessEvent(ctx context.Context, order models.Order) error {
	if orderSuccessWriter == nil {
		logger.Info("Order success producer not initialized")
		return fmt.Errorf("Order success producer not initialized")
	}

	var items []OrderItemInfo
	if err := json.Unmarshal(order.Items, &items); err != nil {
		return err
	}

	orderEvent := OrderSuccessEvent{
		OrderID:    strconv.FormatUint(uint64(order.ID), 10),
		UserID:     order.UserID,
		TotalPrice: order.TotalPrice,
		Items:      items,
	}

	messagePayload, err := json.Marshal(orderEvent)
	if err != nil {
		logger.Err("Error marshalling order event", err)
		return err
	}

	message := kafka.Message{
		Key:   []byte(strconv.FormatUint(uint64(order.ID), 10)),
		Value: messagePayload,
	}

	if err := orderSuccessWriter.WriteMessages(ctx, message); err != nil {
		logger.Err("Failed to write message", err)
		return err
	}

	logger.Logger.Infof("Order success event produced", orderEvent)
	return nil
}

func CloseOrderSuccessProducer() {
	if orderSuccessWriter != nil {
		orderSuccessWriter.Close()
	}
}

// OrderReturnedEvent represents the structure of the order returned event message
// that will be sent to the Kafka topic.

func InitOrderReturnedProducer(brokers []string) {
	orderReturnedWriter = &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    OrderReturnedTopic,
		Balancer: &kafka.LeastBytes{},
	}
	logger.Logger.Infof("Order returned producer initialized with brokers: %v", brokers)
}

func ProduceOrderReturnedEvent(ctx context.Context, order models.Order) error {
	if orderReturnedWriter == nil {
		logger.Err("Order returned producer not initialized", nil)
		return fmt.Errorf("Order returned producer not initialized")
	}

	var items []OrderItemInfo
	if err := json.Unmarshal(order.Items, &items); err != nil {
		return err
	}

	orderEvent := OrderSuccessEvent{
		OrderID:    strconv.FormatUint(uint64(order.ID), 10),
		UserID:     order.UserID,
		TotalPrice: order.TotalPrice,
		Items:      items,
	}

	messagePayLoad, err := json.Marshal(orderEvent)
	if err != nil {
		logger.Err("Error marshalling order event", err)
		return err
	}

	message := kafka.Message{
		Key:   []byte(strconv.FormatUint(uint64(order.ID), 10)),
		Value: messagePayLoad,
	}

	if err := orderReturnedWriter.WriteMessages(ctx, message); err != nil {
		logger.Err("Failed to write message", err)
		return err
	}

	logger.Logger.Infof("Order returned event produced", orderEvent)
	return nil
}
