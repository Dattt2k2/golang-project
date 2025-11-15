package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	logger "order-service/log"

	"github.com/segmentio/kafka-go"
)

const (
	PaymentRequestTopic = "payment_requests"
	PaymentActionTopic  = "payment_actions"
	VendorPaymentTopic  = "vendor_payments"
)

var (
	paymentRequestWriter *kafka.Writer
	paymentActionWriter  *kafka.Writer
	vendorPaymentWriter  *kafka.Writer
)

type PaymentRequestEvent struct {
	OrderID       string  `json:"order_id"`
	UserID        string  `json:"user_id"`
	Amount        float64 `json:"amount"` // Amount in cents
	PaymentMethod string  `json:"payment_method"`
	Description   string  `json:"description"`
	Currency      string  `json:"currency"`
	Timestamp     int64   `json:"timestamp"`
	// New fields for Stripe Connect
	VendorID              string  `json:"vendor_id,omitempty"`
	VendorStripeAccountID string  `json:"vendor_stripe_account_id,omitempty"`
	VendorAmount          float64 `json:"vendor_amount"`
	PlatformFee           float64 `json:"platform_fee"`
	VendorBreakdown       string  `json:"vendor_breakdown,omitempty"` // JSON string with detailed breakdown
}

type PaymentCaptureEvent struct {
	OrderID   string  `json:"order_id"`
	PaymentID string  `json:"payment_id"`
	Amount    float64 `json:"amount"` // Amount in cents
	Timestamp int64   `json:"timestamp"`
}

type PaymentCancelEvent struct {
	OrderID   string `json:"order_id"`
	Reason    string `json:"reason"`
	Timestamp int64  `json:"timestamp"`
	PaymentID string `json:"payment_id"`
}

type VendorPaymentEvent struct {
	OrderID     string  `json:"order_id"`
	VendorID    string  `json:"vendor_id"`
	Amount      float64 `json:"amount"`
	PlatformFee float64 `json:"platform_fee"`
	ReleaseDate int64   `json:"release_date"`
	Timestamp   int64   `json:"timestamp"`
}

func InitPaymentProducer(broker []string) {
	paymentRequestWriter = &kafka.Writer{
		Addr:     kafka.TCP(broker...),
		Topic:    PaymentRequestTopic,
		Balancer: &kafka.LeastBytes{},
	}

	paymentActionWriter = &kafka.Writer{
		Addr:     kafka.TCP(broker...),
		Topic:    PaymentActionTopic,
		Balancer: &kafka.LeastBytes{},
	}

	vendorPaymentWriter = &kafka.Writer{
		Addr:     kafka.TCP(broker...),
		Topic:    VendorPaymentTopic,
		Balancer: &kafka.LeastBytes{},
	}
}

func ProducePaymentRequestEvent(ctx context.Context, request PaymentRequestEvent) error {
	if paymentRequestWriter == nil {
		return fmt.Errorf("Payment request producer not initialized")
	}

	if request.Timestamp == 0 {
		request.Timestamp = time.Now().Unix()
	}

	messagePayload, err := json.Marshal(request)
	if err != nil {
		logger.Err("Failed to marshal payment request event", err)
		return err
	}

	message := kafka.Message{
		Key:   []byte(request.OrderID),
		Value: messagePayload,
	}

	if err := paymentRequestWriter.WriteMessages(ctx, message); err != nil {
		logger.Err("Failed to write payment request message", err)
		return err
	}

	return nil
}

func ProducePaymentCaptureEvent(ctx context.Context, capture PaymentCaptureEvent) error {
	if paymentRequestWriter == nil {
		return fmt.Errorf("payment request producer not initialized")
	}

	// Set timestamp if not provided
	if capture.Timestamp == 0 {
		capture.Timestamp = time.Now().Unix()
	}

	// Create event with action type
	event := map[string]interface{}{
		"action": "capture",
		"data":   capture,
	}

	messagePayload, err := json.Marshal(event)
	if err != nil {
		logger.Err("Error marshalling payment capture event", err)
		return err
	}

	message := kafka.Message{
		Key:   []byte(capture.OrderID),
		Value: messagePayload,
	}

	if err := paymentActionWriter.WriteMessages(ctx, message); err != nil {
		logger.Err("Failed to write payment capture message", err)
		return err
	}

	logger.Info(fmt.Sprintf("✅ Payment capture event sent successfully for order: %s to topic: %s", capture.OrderID, PaymentActionTopic))
	return nil
}

func ProducePaymentCancelEvent(ctx context.Context, cancel PaymentCancelEvent) error {
	if paymentRequestWriter == nil {
		return fmt.Errorf("payment request producer not initialized")
	}

	if cancel.Timestamp == 0 {
		cancel.Timestamp = time.Now().Unix()
	}

	event := map[string]interface{}{
		"action": "cancel",
		"data":   cancel,
	}

	messagePayload, err := json.Marshal(event)
	if err != nil {
		logger.Err("Error marshalling payment cancel event", err)
		return err
	}

	message := kafka.Message{
		Key:   []byte(cancel.OrderID),
		Value: messagePayload,
	}

	if err := paymentActionWriter.WriteMessages(ctx, message); err != nil {
		logger.Err("Failed to write payment cancel message", err)
		return err
	}

	logger.Info(fmt.Sprintf("✅ Payment cancel event sent successfully for order: %s to topic: %s", cancel.OrderID, PaymentActionTopic))
	return nil
}

func ProduceVendorPaymentEvent(ctx context.Context, event VendorPaymentEvent) error {
	if vendorPaymentWriter == nil {
		return fmt.Errorf("vendor payment producer not initialized")
	}

	// Set timestamp if not provided
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().Unix()
	}

	messagePayload, err := json.Marshal(event)
	if err != nil {
		logger.Err("Error marshalling vendor payment event", err)
		return err
	}

	message := kafka.Message{
		Key:   []byte(event.OrderID),
		Value: messagePayload,
	}

	if err := vendorPaymentWriter.WriteMessages(ctx, message); err != nil {
		logger.Err("Failed to write vendor payment message", err)
		return err
	}
	return nil
}

func ClosePaymentRequestProducer() {
	if paymentRequestWriter != nil {
		paymentRequestWriter.Close()
	}
}
