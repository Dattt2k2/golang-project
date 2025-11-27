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
	OrderSuccessTopic    = "order_success"
	OrderReturnedTopic   = "order_returned"
	OrderDeleteItemTopic = "cart_delete_items"
)

var (
	orderSuccessWriter    *kafka.Writer
	orderReturnedWriter   *kafka.Writer
	orderDeleteItemWriter *kafka.Writer
)

// OrderSuccessEvent represents the structure of the order success event message
// that will be sent to the Kafka topic.
type OrderSuccessEvent struct {
	OrderID    string          `json:"order_id"`
	UserID     string          `json:"user_id"`
	Items      []OrderItemInfo `json:"items"`
	TotalPrice float64         `json:"total_price"`
}

type OrderReturnedEvent struct {
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

func InitOrderDeleteItemProducer(brokers []string) {
	orderDeleteItemWriter = &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    OrderDeleteItemTopic,
		Balancer: &kafka.LeastBytes{},
	}
}

func ProduceOrderDeleteItemEvent(ctx context.Context, userID string, productIDs []string) error {
	if orderDeleteItemWriter == nil {
		logger.Err("Order delete item producer not initialized", nil)
		return fmt.Errorf("Order delete item producer not initialized")
	}

	event := CartDeleteEvent{
		UserID:     userID,
		ProductIDs: productIDs,
	}

	messagePayload, err := json.Marshal(event)
	if err != nil {
		logger.Err("Error marshalling order delete item event", err)
		return err
	}

	logger.Info(fmt.Sprintf("Sending Kafka message to topic %s for cart delete: %s", OrderDeleteItemTopic, string(messagePayload)))

	message := kafka.Message{
		Key:   []byte(userID),
		Value: messagePayload,
	}

	if err := orderDeleteItemWriter.WriteMessages(ctx, message); err != nil {
		logger.Err("Failed to write delete item message", err)
		return err
	}

	logger.Info(fmt.Sprintf("✅ Successfully produced cart delete event for UserID=%s", userID))
	return nil
}

func CloseOrderDeleteItemProducer() {
	if orderDeleteItemWriter != nil {
		orderDeleteItemWriter.Close()
	}
}

func InitOrderSuccessProducer(brokers []string) {
	orderSuccessWriter = &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        OrderSuccessTopic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
	}
}

func ProduceOrderSuccessEvent(ctx context.Context, order models.Order) error {
	logger.Info("Start decresing product")
	if orderSuccessWriter == nil {
		return fmt.Errorf("Order success producer not initialized")
	}

	var items []OrderItemInfo
	if err := json.Unmarshal(order.Items, &items); err != nil {
		return err
	}

	orderEvent := OrderSuccessEvent{
		OrderID:    order.OrderID,
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

	orderEvent := OrderReturnedEvent{
		OrderID:    order.OrderID,
		UserID:     order.UserID,
		TotalPrice: order.TotalPrice,
		Items:      items,
	}

	messagePayLoad, err := json.Marshal(orderEvent)
	if err != nil {
		logger.Err("Error marshalling order returned event", err)
		return err
	}

	logger.Info(fmt.Sprintf("📨 Sending Kafka message to topic %s for order return: %s", OrderReturnedTopic, string(messagePayLoad)))

	message := kafka.Message{
		Key:   []byte(strconv.FormatUint(uint64(order.ID), 10)),
		Value: messagePayLoad,
	}

	if err := orderReturnedWriter.WriteMessages(ctx, message); err != nil {
		logger.Err("Failed to write returned message", err)
		return err
	}

	logger.Info(fmt.Sprintf("✅ Successfully produced order_returned event for OrderID=%s", order.OrderID))
	return nil
}
