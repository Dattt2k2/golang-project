package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"payment-service/repository"
	logger "payment-service/src/utils"

	"github.com/segmentio/kafka-go"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/refund"
)

// Payment request event từ order-service với Stripe Connect support
type PaymentRequestEvent struct {
	OrderID       string  `json:"order_id"`
	UserID        string  `json:"user_id"`
	Amount        float64 `json:"amount"`
	PaymentMethod string  `json:"payment_method"`
	Description   string  `json:"description"`
	Currency      string  `json:"currency"`
	Timestamp     int64   `json:"timestamp"`
	// Stripe Connect fields
	VendorID              string  `json:"vendor_id,omitempty"`
	VendorStripeAccountID string  `json:"vendor_stripe_account_id,omitempty"`
	VendorAmount          float64 `json:"vendor_amount"`
	PlatformFee           float64 `json:"platform_fee"`
	VendorBreakdown       string  `json:"vendor_breakdown,omitempty"`
}

// Payment events to send back to order-service
type PaymentStatusEvent struct {
	OrderID         string  `json:"order_id"`
	PaymentIntentID string  `json:"payment_intent_id"`
	Amount          float64 `json:"amount"`
	Status          string  `json:"status"` // "held", "captured", "failed", "cancelled"
	VendorAmount    float64 `json:"vendor_amount,omitempty"`
	PlatformFee     float64 `json:"platform_fee,omitempty"`
	Timestamp       int64   `json:"timestamp"`
	FailureReason   string  `json:"failure_reason,omitempty"`
}

// Vendor payment event from order-service
type VendorPaymentEvent struct {
	OrderID     string  `json:"order_id"`
	VendorID    string  `json:"vendor_id"`
	Amount      float64 `json:"amount"`
	PlatformFee float64 `json:"platform_fee"`
	ReleaseDate int64   `json:"release_date"`
	Timestamp   int64   `json:"timestamp"`
}

// Vendor payment event processed (sent back to order-service)
type VendorPaymentProcessedEvent struct {
	OrderID       string  `json:"order_id"`
	VendorID      string  `json:"vendor_id"`
	Amount        float64 `json:"amount"`
	PlatformFee   float64 `json:"platform_fee"`
	TransferID    string  `json:"transfer_id,omitempty"`
	Status        string  `json:"status"` // "transferred", "failed"
	FailureReason string  `json:"failure_reason,omitempty"`
	Timestamp     int64   `json:"timestamp"`
}

// Payment action events (capture, cancel)
type PaymentActionEvent struct {
	Action string      `json:"action"`
	Data   interface{} `json:"data"`
}

type PaymentCaptureData struct {
	OrderID   string  `json:"order_id"`
	PaymentID string  `json:"payment_id"`
	Amount    float64 `json:"amount"`
	Timestamp int64   `json:"timestamp"`
}

type PaymentCancelData struct {
	OrderID   string `json:"order_id"`
	Reason    string `json:"reason"`
	Timestamp int64  `json:"timestamp"`
	PaymentID string `json:"payment_id"`
}

type PaymentConsumer struct {
	paymentService  *PaymentService
	vendorRepo      *repository.VendorRepository
	orderServiceURL string
	kafkaProducer   *KafkaProducer
}

func NewPaymentConsumer(paymentService *PaymentService, vendorRepo *repository.VendorRepository, orderServiceURL string) *PaymentConsumer {
	return &PaymentConsumer{
		paymentService:  paymentService,
		vendorRepo:      vendorRepo,
		orderServiceURL: orderServiceURL,
		kafkaProducer:   NewKafkaProducer([]string{"kafka:9092"}), // Initialize producer
	}
}

func (pc *PaymentConsumer) StartConsumer(brokers []string) {
	// Start payment request consumer
	go pc.consumePaymentRequests(brokers)

	// Start payment action consumer (capture, cancel)
	go pc.consumePaymentActions(brokers)

	// Start vendor payment consumer
	go pc.consumeVendorPayments(brokers)

}

func (pc *PaymentConsumer) consumePaymentRequests(brokers []string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   "payment_requests",
		GroupID: "payment-service-requests",
	})
	defer reader.Close()

	for {
		message, err := reader.ReadMessage(context.Background())
		if err != nil {
			logger.Error("Error reading payment request: " + err.Error())
			continue
		}

		var paymentReq PaymentRequestEvent
		if err := json.Unmarshal(message.Value, &paymentReq); err != nil {
			logger.Error("Error unmarshalling payment request: " + err.Error())
			continue
		}

		pc.handlePaymentRequestWithConnect(paymentReq)
	}
}

func (pc *PaymentConsumer) consumePaymentActions(brokers []string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   "payment_actions",
		GroupID: "payment-service-actions",
	})
	defer reader.Close()

	for {
		message, err := reader.ReadMessage(context.Background())
		if err != nil {
			logger.Error("Error reading payment action: " + err.Error())
			continue
		}

		var actionEvent PaymentActionEvent
		if err := json.Unmarshal(message.Value, &actionEvent); err != nil {
			logger.Error("Error unmarshalling payment action: " + err.Error())
			continue
		}

		pc.handleActionEvent(actionEvent)
	}
}

func (pc *PaymentConsumer) consumeVendorPayments(brokers []string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   "vendor_payments",
		GroupID: "payment-service-vendor-payments",
	})
	defer reader.Close()

	logger.Info("Started vendor payments consumer")

	for {
		message, err := reader.ReadMessage(context.Background())
		if err != nil {
			logger.Error("Error reading vendor payment: " + err.Error())
			continue
		}

		var vendorPayment VendorPaymentEvent
		if err := json.Unmarshal(message.Value, &vendorPayment); err != nil {
			logger.Error("Error unmarshalling vendor payment: " + err.Error())
			continue
		}

		logger.Info(fmt.Sprintf("Received vendor payment event for order %s, vendor %s, amount %.2f",
			vendorPayment.OrderID, vendorPayment.VendorID, vendorPayment.Amount))

		// The actual transfer was already done during payment capture
		// This event is just for tracking/logging purposes
		// We could update vendor balance or send notifications here
	}
}

func (pc *PaymentConsumer) handlePaymentRequestWithConnect(req PaymentRequestEvent) {

	if req.PaymentMethod != "stripe" {
		return
	}

	logger.Info(fmt.Sprintf("📥 Processing payment request - Order: %s, Amount: %.2f, VendorID: %s",
		req.OrderID, req.Amount, req.VendorID))

	ctx := context.Background()
	amountInCents := int64(req.Amount * 100)
	platformFeeInCents := int64(req.PlatformFee * 100)

	var paymentIntent *stripe.PaymentIntent
	var err error
	var vendorStripeAccountID string

	// Lookup vendor's Stripe account from database if VendorID is provided
	if req.VendorID != "" && pc.vendorRepo != nil {
		vendor, lookupErr := pc.vendorRepo.GetVendorByID(ctx, req.VendorID)
		if lookupErr != nil {
			logger.Error(fmt.Sprintf("⚠️  Failed to lookup vendor %s: %v. Creating standard payment.", req.VendorID, lookupErr))
		} else if vendor.StripeAccountID != "" {
			vendorStripeAccountID = vendor.StripeAccountID
			logger.Info(fmt.Sprintf("✅ Found Stripe account for vendor %s: %s", req.VendorID, vendorStripeAccountID))
		} else {
			logger.Error(fmt.Sprintf("⚠️  Vendor %s has no Stripe account. Creating standard payment.", req.VendorID))
		}
	}

	if vendorStripeAccountID != "" {
		// Multi-vendor payment với Stripe Connect
		logger.Info(fmt.Sprintf("💳 Creating Stripe Connect payment - Vendor: %s, Amount: %d cents, Fee: %d cents",
			vendorStripeAccountID, amountInCents, platformFeeInCents))

		paymentIntent, err = pc.paymentService.CreatePaymentIntentWithConnect(
			ctx,
			amountInCents,
			req.Currency,
			req.OrderID,
			vendorStripeAccountID,
			platformFeeInCents,
			req.VendorBreakdown,
		)
	} else {
		// Standard payment
		logger.Info(fmt.Sprintf("💳 Creating standard payment (no Connect) - Order: %s, Amount: %d cents",
			req.OrderID, amountInCents))

		paymentIntent, err = pc.paymentService.CreatePaymentIntent(
			ctx,
			amountInCents,
			req.Currency,
			req.OrderID,
		)
	}

	if err != nil {
		logger.Error("Failed to create PaymentIntent for order " + req.OrderID + ": " + err.Error())
		pc.notifyPaymentStatus(req.OrderID, "", req.Amount, "failed", req.VendorAmount, req.PlatformFee, err.Error())
		return
	}

	// Notify order service about payment held in escrow
	pc.notifyPaymentStatus(req.OrderID, paymentIntent.ID, req.Amount, "held", req.VendorAmount, req.PlatformFee, "")

	// TODO: Remove this simulation in production
	// Simulate payment success for testing
	go func() {
		time.Sleep(5 * time.Second)
		pc.notifyPaymentStatus(req.OrderID, paymentIntent.ID, req.Amount, "succeeded", req.VendorAmount, req.PlatformFee, "")
	}()
}

func (pc *PaymentConsumer) handleActionEvent(event PaymentActionEvent) {
	switch event.Action {
	case "capture":
		pc.handleCaptureEvent(event.Data)
	case "cancel":
		pc.handleCancelEvent(event.Data)
	default:
		logger.Error("Unknown payment action: " + event.Action)
	}
}

func (pc *PaymentConsumer) handleCaptureEvent(data interface{}) {
	captureData, ok := data.(map[string]interface{})
	if !ok {
		logger.Error("Invalid capture event data")
		return
	}

	orderID := captureData["order_id"].(string)
	paymentID := captureData["payment_id"].(string)
	amount := captureData["amount"].(float64)

	logger.Info(fmt.Sprintf("🔄 Processing payment capture request for order: %s, payment: %s", orderID, paymentID))

	// Capture the payment (release funds from escrow)
	capturedPayment, err := pc.paymentService.CapturePaymentIntent(context.Background(), paymentID, orderID)
	if err != nil {
		logger.Error("❌ Failed to capture payment for order " + orderID + ": " + err.Error())
		pc.notifyPaymentStatus(orderID, paymentID, amount, "capture_failed", 0, 0, err.Error())
		return
	}

	logger.Info(fmt.Sprintf("✅ Payment captured successfully for order: %s, amount: %.2f VND", orderID, amount))

	// Notify successful capture
	pc.notifyPaymentStatus(orderID, paymentID, amount, "captured", 0, 0, "")

	// Process vendor transfers if this is a Connect payment
	if capturedPayment.TransferData != nil && capturedPayment.TransferData.Destination != nil {
		logger.Info(fmt.Sprintf("💰 Processing vendor transfer for order: %s", orderID))
		go pc.processVendorTransfers(orderID, capturedPayment)
	} else {
		logger.Info(fmt.Sprintf("ℹ️  No vendor transfer needed for order: %s (not a Connect payment)", orderID))
	}
}

func (pc *PaymentConsumer) handleCancelEvent(data interface{}) {
	cancelData, ok := data.(map[string]interface{})
	if !ok {
		logger.Error("Invalid cancel event data")
		return
	}

	orderID := cancelData["order_id"].(string)
	paymentID := cancelData["payment_id"].(string)
	reason := cancelData["reason"].(string)

	// Get payment intent to check status
	paymentIntent, err := pc.paymentService.GetPaymentIntentByID(context.Background(), paymentID)
	if err != nil {
		logger.Error("Failed to get payment intent for order " + orderID + ": " + err.Error())
		pc.notifyPaymentStatus(orderID, paymentID, 0, "cancel_failed", 0, 0, err.Error())
		return
	}

	// If payment succeeded, refund instead of cancel
	if paymentIntent.Status == "succeeded" {
		logger.Info("Payment already succeeded for order " + orderID + ", creating refund instead of cancel")

		refundParams := &stripe.RefundParams{
			PaymentIntent: stripe.String(paymentID),
			Reason:        stripe.String("requested_by_customer"),
		}

		_, err := refund.New(refundParams)
		if err != nil {
			logger.Error("Failed to refund payment for order " + orderID + ": " + err.Error())
			pc.notifyPaymentStatus(orderID, paymentID, 0, "refund_failed", 0, 0, err.Error())
			return
		}

		pc.notifyPaymentStatus(orderID, paymentID, float64(paymentIntent.Amount)/100, "refunded", 0, 0, reason)
		return
	}

	// Otherwise, cancel the payment intent
	err = pc.paymentService.CancelPaymentIntent(context.Background(), paymentID)
	if err != nil {
		logger.Error("Failed to cancel payment for order " + orderID + ": " + err.Error())
		pc.notifyPaymentStatus(orderID, paymentID, 0, "cancel_failed", 0, 0, err.Error())
		return
	}

	pc.notifyPaymentStatus(orderID, paymentID, 0, "cancelled", 0, 0, reason)
}

// Process vendor transfers after payment capture
func (pc *PaymentConsumer) processVendorTransfers(orderID string, paymentIntent *stripe.PaymentIntent) {

	// Parse vendor breakdown from metadata
	vendorBreakdownStr := paymentIntent.Metadata["vendor_breakdown"]
	if vendorBreakdownStr == "" {
		logger.Error("No vendor breakdown found for order: " + orderID)
		return
	}

	var vendorBreakdown map[string]map[string]float64
	if err := json.Unmarshal([]byte(vendorBreakdownStr), &vendorBreakdown); err != nil {
		logger.Error("Failed to parse vendor breakdown: " + err.Error())
		return
	}

	// Process transfer for each vendor
	for vendorID, amounts := range vendorBreakdown {
		vendorAmount := amounts["vendor_amount"]
		platformFee := amounts["platform_fee"]

		// Get vendor's Stripe account ID from database
		// TODO: Implement vendorRepo.GetVendorStripeAccountID(vendorID)
		vendorStripeAccountID := "acct_" + vendorID // Placeholder until vendor registration is implemented

		// Create transfer to vendor
		transferResult, err := pc.paymentService.CreateTransferToVendor(
			context.Background(),
			int64(vendorAmount*100),
			vendorStripeAccountID,
			orderID,
		)

		if err != nil {
			logger.Error(fmt.Sprintf("Failed to transfer to vendor %s for order %s: %v", vendorID, orderID, err))
			pc.notifyVendorPaymentResult(orderID, vendorID, vendorAmount, platformFee, "", "failed", err.Error())
			continue
		}

		pc.notifyVendorPaymentResult(orderID, vendorID, vendorAmount, platformFee, transferResult.ID, "transferred", "")
	}
}

// Notify order service about payment status changes
func (pc *PaymentConsumer) notifyPaymentStatus(orderID, paymentIntentID string, amount float64, status string, vendorAmount, platformFee float64, failureReason string) {
	event := PaymentStatusEvent{
		OrderID:         orderID,
		PaymentIntentID: paymentIntentID,
		Amount:          amount,
		Status:          status,
		VendorAmount:    vendorAmount,
		PlatformFee:     platformFee,
		Timestamp:       time.Now().Unix(),
		FailureReason:   failureReason,
	}

	if err := pc.produceEvent("payment_events", event); err != nil {
		logger.Error("Failed to produce payment event: " + err.Error())
	}
}

// Notify about vendor payment results
func (pc *PaymentConsumer) notifyVendorPaymentResult(orderID, vendorID string, amount, platformFee float64, transferID, status, failureReason string) {
	event := VendorPaymentProcessedEvent{
		OrderID:       orderID,
		VendorID:      vendorID,
		Amount:        amount,
		PlatformFee:   platformFee,
		TransferID:    transferID,
		Status:        status,
		FailureReason: failureReason,
		Timestamp:     time.Now().Unix(),
	}

	if err := pc.produceEvent("vendor_payment_processed", event); err != nil {
		logger.Error("Failed to produce vendor payment event: " + err.Error())
	}
}

// Produce event to Kafka
func (pc *PaymentConsumer) produceEvent(topic string, event interface{}) error {
	if pc.kafkaProducer == nil {
		logger.Error("Kafka producer not initialized")
		return fmt.Errorf("kafka producer not initialized")
	}

	return pc.kafkaProducer.SendMessage(context.Background(), event)
}

// Legacy notification methods for backward compatibility
func (pc *PaymentConsumer) notifyPaymentSuccess(orderID, paymentIntentID string) {
	url := fmt.Sprintf("%s/api/v1/orders/%s/payment/success", pc.orderServiceURL, orderID)

	payload := map[string]string{
		"payment_intent_id": paymentIntentID,
	}

	pc.sendHTTPNotification(url, payload)
}

func (pc *PaymentConsumer) notifyPaymentFailure(orderID, reason string) {
	url := fmt.Sprintf("%s/api/v1/orders/%s/payment/failure", pc.orderServiceURL, orderID)

	payload := map[string]string{
		"reason": reason,
	}

	pc.sendHTTPNotification(url, payload)
}

func (pc *PaymentConsumer) sendHTTPNotification(url string, payload interface{}) {
	payloadBytes, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		logger.Error("Failed to notify order-service: " + err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Order-service notification failed with status: " + strconv.Itoa(resp.StatusCode))
	}
}
