package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	logger "payment-service/src/utils"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writers map[string]*kafka.Writer
}

func NewKafkaProducer(brokers []string) *KafkaProducer {
	writers := map[string]*kafka.Writer{
		"payment_events": {
			Addr:         kafka.TCP(brokers...),
			Topic:        "payment_events",
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireOne,
			Async:        false,
		},
		"vendor_payment_processed": {
			Addr:         kafka.TCP(brokers...),
			Topic:        "vendor_payment_processed",
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireOne,
			Async:        false,
		},
		"vendor_account_updates": {
			Addr:         kafka.TCP(brokers...),
			Topic:        "vendor_account_updates",
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireOne,
			Async:        false,
		},
	}

	return &KafkaProducer{
		writers: writers,
	}
}

func (kp *KafkaProducer) SendMessage(ctx context.Context, message interface{}) error {
	// Determine topic based on message type
	topic := ""
	switch message.(type) {
	case PaymentStatusEvent:
		topic = "payment_events"
	case VendorPaymentProcessedEvent:
		topic = "vendor_payment_processed"
	case map[string]interface{}:
		// For generic messages, try to determine topic from content
		if msg, ok := message.(map[string]interface{}); ok {
			if _, exists := msg["vendor_id"]; exists {
				topic = "vendor_account_updates"
			} else {
				topic = "payment_events"
			}
		}
	default:
		topic = "payment_events" // Default topic
	}

	return kp.SendMessageToTopic(ctx, topic, message)
}

func (kp *KafkaProducer) SendMessageToTopic(ctx context.Context, topic string, message interface{}) error {
	writer, exists := kp.writers[topic]
	if !exists {
		return fmt.Errorf("no writer found for topic: %s", topic)
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create message key for partitioning
	var messageKey []byte
	switch msg := message.(type) {
	case PaymentStatusEvent:
		messageKey = []byte(msg.OrderID)
	case VendorPaymentProcessedEvent:
		messageKey = []byte(msg.OrderID)
	default:
		messageKey = []byte(fmt.Sprintf("%d", time.Now().UnixNano()))
	}

	kafkaMessage := kafka.Message{
		Key:   messageKey,
		Value: messageBytes,
		Time:  time.Now(),
	}

	if err := writer.WriteMessages(ctx, kafkaMessage); err != nil {
		logger.Error(fmt.Sprintf("Failed to write message to topic %s: %v", topic, err))
		return err
	}

	return nil
}

func (kp *KafkaProducer) Close() error {
	for topic, writer := range kp.writers {
		if err := writer.Close(); err != nil {
			logger.Error(fmt.Sprintf("Failed to close writer for topic %s: %v", topic, err))
		}
	}
	return nil
}

// Specific methods for different event types
func (kp *KafkaProducer) SendPaymentEvent(ctx context.Context, event PaymentStatusEvent) error {
	return kp.SendMessageToTopic(ctx, "payment", event)
}

func (kp *KafkaProducer) SendVendorPaymentEvent(ctx context.Context, event VendorPaymentProcessedEvent) error {
	return kp.SendMessageToTopic(ctx, "vendor_payment_processed", event)
}

func (kp *KafkaProducer) SendVendorAccountUpdate(ctx context.Context, update map[string]interface{}) error {
	return kp.SendMessageToTopic(ctx, "vendor_account_updates", update)
}

func (kp *KafkaProducer) SendBankPayoutEvent(ctx context.Context, event interface{}) error {
	return kp.SendMessageToTopic(ctx, "bank_payouts", event)
}
