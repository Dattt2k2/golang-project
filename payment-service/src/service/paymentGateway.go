package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"payment-service/models"
	"payment-service/repository"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/account"
	"github.com/stripe/stripe-go/v74/accountlink"
	pi "github.com/stripe/stripe-go/v74/paymentintent"
	"github.com/stripe/stripe-go/v74/payout"
	"github.com/stripe/stripe-go/v74/refund"
	"github.com/stripe/stripe-go/v74/transfer"
	"github.com/stripe/stripe-go/v74/webhook"
)

type PaymentMessage struct {
	OrderID         string  `json:"order_id"`
	Amount          float64 `json:"amount"`
	Status          string  `json:"status"`
	PaymentIntentID string  `json:"payment_intent_id"`
}

type PaymentService struct {
	Repo          *repository.PaymentRepository
	SigningSecret string
	Producer      *KafkaProducer
}

func NewPaymentService(repo *repository.PaymentRepository, signinSecret string) *PaymentService {
	// Use STRIPE_SECRET_KEY (sk_test_...) for API calls, NOT publishable key
	if k := os.Getenv("STRIPE_SECRET_KEY"); k != "" {
		stripe.Key = k
		fmt.Printf("[PaymentService] Stripe API key configured: %s... (length: %d)\n", k[:min(10, len(k))], len(k))
	} else {
		fmt.Println("[PaymentService] WARNING: STRIPE_SECRET_KEY not set!")
	}

	fmt.Printf("[PaymentService] Initializing with secret: %s... (length: %d)\n", signinSecret[:min(20, len(signinSecret))], len(signinSecret))

	// Initialize Kafka Producer for sending payment events to order-service
	var producer *KafkaProducer
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers != "" {
		brokers := strings.Split(kafkaBrokers, ",")
		producer = NewKafkaProducer(brokers)
		fmt.Printf("[PaymentService] Kafka producer initialized with brokers: %v\n", brokers)
	} else {
		fmt.Println("[PaymentService] Warning: KAFKA_BROKERS not set, Kafka events will not be sent")
	}

	return &PaymentService{
		Repo:          repo,
		SigningSecret: signinSecret,
		Producer:      producer,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Standard PaymentIntent creation (single vendor or no vendor)
func (s *PaymentService) CreatePaymentIntent(ctx context.Context, amount int64, currency, orderID string) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(amount),
		Currency:      stripe.String(currency),
		CaptureMethod: stripe.String(string(stripe.PaymentIntentCaptureMethodManual)),
	}
	params.AddMetadata("order_id", orderID)

	piObj, err := pi.New(params)
	if err != nil {
		return nil, err
	}

	_ = s.Repo.MarkAuthByOrderID(orderID, piObj.ID)
	if s.Producer != nil {
		_ = s.Producer.SendMessage(context.Background(), PaymentMessage{OrderID: orderID, Amount: float64(amount) / 100.0, Status: "authorized"})
	}
	return piObj, nil
}

// Stripe Connect PaymentIntent creation (multi-vendor support)
func (s *PaymentService) CreatePaymentIntentWithConnect(ctx context.Context, amount int64, currency, orderID, vendorStripeAccountID string, platformFeeAmount int64, vendorBreakdown string) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(amount),
		Currency:      stripe.String(currency),
		CaptureMethod: stripe.String(string(stripe.PaymentIntentCaptureMethodManual)), // Escrow
	}

	// Add metadata
	params.AddMetadata("order_id", orderID)
	params.AddMetadata("vendor_breakdown", vendorBreakdown)
	params.AddMetadata("platform_fee", fmt.Sprintf("%.2f", float64(platformFeeAmount)/100))

	// Stripe Connect configuration
	if vendorStripeAccountID != "" {
		// Transfer data for primary vendor
		params.TransferData = &stripe.PaymentIntentTransferDataParams{
			Destination: stripe.String(vendorStripeAccountID),
		}

		// Platform fee
		if platformFeeAmount > 0 {
			params.ApplicationFeeAmount = stripe.Int64(platformFeeAmount)
		}
	}

	piObj, err := pi.New(params)
	if err != nil {
		return nil, err
	}

	// Mark as authorized in DB
	_ = s.Repo.MarkAuthByOrderID(orderID, piObj.ID)
	if s.Producer != nil {
		_ = s.Producer.SendMessage(context.Background(), PaymentMessage{OrderID: orderID, Amount: float64(amount) / 100.0, Status: "authorized"})
	}

	return piObj, nil
}

// Create Stripe Connect account for vendor
func (s *PaymentService) CreateStripeConnectAccount(ctx context.Context, vendorID, vendorEmail, country string, businessProfile map[string]string) (*stripe.Account, error) {
	params := &stripe.AccountParams{
		Type:    stripe.String(string(stripe.AccountTypeExpress)),
		Country: stripe.String(country), // "US", "VN", etc.
		Email:   stripe.String(vendorEmail),
	}

	// Add metadata
	params.AddMetadata("vendor_id", vendorID)
	params.AddMetadata("created_by", "ecommerce_platform")

	// Business profile if provided
	if len(businessProfile) > 0 {
		params.BusinessProfile = &stripe.AccountBusinessProfileParams{}
		if name, ok := businessProfile["name"]; ok {
			params.BusinessProfile.Name = stripe.String(name)
		}
		if url, ok := businessProfile["url"]; ok {
			params.BusinessProfile.URL = stripe.String(url)
		}
		if mcc, ok := businessProfile["mcc"]; ok {
			params.BusinessProfile.MCC = stripe.String(mcc)
		}
	}

	account, err := account.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe Connect account: %w", err)
	}

	return account, nil
}

// Create account link for vendor onboarding
func (s *PaymentService) CreateAccountLink(ctx context.Context, accountID, refreshURL, returnURL string) (*stripe.AccountLink, error) {
	params := &stripe.AccountLinkParams{
		Account:    stripe.String(accountID),
		RefreshURL: stripe.String(refreshURL),
		ReturnURL:  stripe.String(returnURL),
		Type:       stripe.String("account_onboarding"),
	}

	link, err := accountlink.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create account link: %w", err)
	}

	return link, nil
}

// Transfer money to vendor account
func (s *PaymentService) CreateTransferToVendor(ctx context.Context, amount int64, vendorAccountID, orderID string) (*stripe.Transfer, error) {
	params := &stripe.TransferParams{
		Amount:      stripe.Int64(amount),
		Currency:    stripe.String("vnd"),
		Destination: stripe.String(vendorAccountID),
	}

	// Add metadata
	params.AddMetadata("order_id", orderID)
	params.AddMetadata("type", "vendor_payment")
	params.AddMetadata("timestamp", fmt.Sprintf("%d", time.Now().Unix()))

	transferObj, err := transfer.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create transfer: %w", err)
	}

	return transferObj, nil
}

// Get vendor account status
func (s *PaymentService) GetVendorAccountStatus(ctx context.Context, accountID string) (*stripe.Account, error) {
	account, err := account.GetByID(accountID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get account status: %w", err)
	}
	return account, nil
}

// Update vendor account
func (s *PaymentService) UpdateVendorAccount(ctx context.Context, accountID string, updates map[string]interface{}) (*stripe.Account, error) {
	params := &stripe.AccountParams{}

	// Add updates to params based on the updates map
	if email, ok := updates["email"].(string); ok {
		params.Email = stripe.String(email)
	}

	if businessProfile, ok := updates["business_profile"].(map[string]string); ok {
		params.BusinessProfile = &stripe.AccountBusinessProfileParams{}
		if name, exists := businessProfile["name"]; exists {
			params.BusinessProfile.Name = stripe.String(name)
		}
		if url, exists := businessProfile["url"]; exists {
			params.BusinessProfile.URL = stripe.String(url)
		}
	}

	account, err := account.Update(accountID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	return account, nil
}

// CapturePaymentIntent captures a previously authorized PaymentIntent (release funds to seller)
func (s *PaymentService) CapturePaymentIntent(ctx context.Context, paymentIntentID, orderID string) (*stripe.PaymentIntent, error) {
	// First, get the payment intent to check its status
	piObj, err := pi.Get(paymentIntentID, nil)
	if err != nil {
		return nil, err
	}

	// Check if already captured
	if piObj.Status == "succeeded" {
		log.Printf("✅ Payment %s already in 'succeeded' state for order %s - treating as captured", paymentIntentID, orderID)
		// Mark as captured in DB if not already done
		_ = s.Repo.MarkCapturedByOrderID(orderID, piObj.ID)

		// Still return the payment intent so vendor transfers can proceed
		return piObj, nil
	}

	// Check if payment is in a capturable state
	if piObj.Status != "requires_capture" {
		return nil, fmt.Errorf("payment intent status is %s, cannot capture", piObj.Status)
	}

	log.Printf("🔄 Capturing payment %s for order %s", paymentIntentID, orderID)

	// Capture the payment
	piObj, err = pi.Capture(paymentIntentID, nil)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ Payment %s captured successfully for order %s", paymentIntentID, orderID)

	// mark captured in DB
	_ = s.Repo.MarkCapturedByOrderID(orderID, piObj.ID)
	if s.Producer != nil {
		_ = s.Producer.SendMessage(context.Background(), PaymentMessage{OrderID: orderID, Amount: float64(piObj.Amount) / 100.0, Status: "captured"})
	}
	return piObj, nil
}

// RefundPayment triggers a refund for a captured or authorized payment
func (s *PaymentService) RefundPayment(ctx context.Context, paymentIntentID, refundID string, amount int64) (*stripe.Refund, error) {
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(paymentIntentID),
	}
	if amount > 0 {
		params.Amount = stripe.Int64(amount)
	}

	params.AddMetadata("refund_id", refundID)
	params.AddMetadata("timestamp", fmt.Sprintf("%d", time.Now().Unix()))

	r, err := refund.New(params)
	if err != nil {
		return nil, err
	}

	// update refund record via repo
	if s.Producer != nil {
		// publish refund event
		_ = s.Producer.SendMessage(context.Background(), map[string]interface{}{
			"order_id":  paymentIntentID,
			"status":    "refund_initiated",
			"refund_id": refundID,
			"amount":    float64(r.Amount) / 100.0,
		})
	}
	return r, nil
}

func (s *PaymentService) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("[Webhook] Failed to read body: %v\n", err)
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	// DEBUG: Log all headers
	fmt.Println("[Webhook DEBUG] === ALL HEADERS ===")
	for name, values := range r.Header {
		for _, value := range values {
			fmt.Printf("[Webhook DEBUG] %s: %s\n", name, value)
		}
	}

	sigHeader := r.Header.Get("Stripe-Signature")

	// DEBUG: Log signature and payload info
	fmt.Printf("[Webhook DEBUG] Signature header: '%s'\n", sigHeader)
	fmt.Printf("[Webhook DEBUG] Payload length: %d bytes\n", len(payload))
	if len(s.SigningSecret) > 20 {
		fmt.Printf("[Webhook DEBUG] Webhook secret (first 20 chars): %s...\n", s.SigningSecret[:20])
	} else {
		fmt.Printf("[Webhook DEBUG] Webhook secret length: %d\n", len(s.SigningSecret))
	}

	if sigHeader == "" {
		fmt.Println("[Webhook ERROR] No Stripe-Signature header found!")
		http.Error(w, "missing signature header", http.StatusBadRequest)
		return
	}

	if len(payload) > 0 && len(payload) < 500 {
		fmt.Printf("[Webhook DEBUG] Payload: %s\n", string(payload))
	}

	// Use ConstructEventWithOptions to ignore API version mismatch
	event, err := webhook.ConstructEventWithOptions(
		payload,
		sigHeader,
		s.SigningSecret,
		webhook.ConstructEventOptions{
			IgnoreAPIVersionMismatch: true,
		},
	)
	if err != nil {
		fmt.Printf("[Webhook ERROR] Invalid signature: %v\n", err)
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}

	fmt.Printf("[Webhook SUCCESS] ✅ Signature verified! Event: %s (ID: %s)\n", event.Type, event.ID)

	switch event.Type {
	case "payment_intent.succeeded":
		var piObj stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &piObj); err == nil {
			// DEBUG: Log metadata
			fmt.Printf("[Webhook DEBUG] PaymentIntent metadata: %+v\n", piObj.Metadata)

			orderID := piObj.Metadata["order_id"]
			if orderID == "" {
				fmt.Printf("[Webhook WARNING] No order_id in metadata. PaymentIntent may not be created by this service. PI: %s\n", piObj.ID)
				// Try to get order_id from description or other fields
				// For now, just log and return
				w.WriteHeader(http.StatusOK)
				return
			}

			fmt.Printf("[Webhook] Payment succeeded for order: %s, PaymentIntent: %s\n", orderID, piObj.ID)

			// Update payment status in database
			if err := s.Repo.MarkAuthByOrderID(orderID, piObj.ID); err != nil {
				fmt.Printf("[Webhook ERROR] Failed to update payment status: %v\n", err)
			}

			// Send event to order-service via Kafka
			if s.Producer != nil {
				event := PaymentMessage{
					OrderID: orderID,
					Amount:  float64(piObj.Amount) / 100.0,
					Status:  "succeeded",
				}
				if err := s.Producer.SendMessage(context.Background(), event); err != nil {
					fmt.Printf("[Webhook ERROR] Failed to send Kafka message: %v\n", err)
				} else {
					fmt.Printf("[Webhook] ✅ Kafka event sent to order-service: order=%s, status=succeeded\n", orderID)
				}
			} else {
				fmt.Println("[Webhook WARNING] Kafka producer not initialized, event not sent")
			}
		} else {
			fmt.Printf("[Webhook ERROR] Failed to unmarshal payment_intent: %v\n", err)
		}

	case "payment_intent.captured":
		var piObj stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &piObj); err == nil {
			orderID := piObj.Metadata["order_id"]
			if orderID == "" {
				fmt.Printf("[Webhook WARNING] No order_id in metadata for captured event. PI: %s\n", piObj.ID)
				w.WriteHeader(http.StatusOK)
				return
			}

			fmt.Printf("[Webhook] Payment captured for order: %s, PaymentIntent: %s\n", orderID, piObj.ID)

			// Update payment status in database
			if err := s.Repo.MarkCapturedByOrderID(orderID, piObj.ID); err != nil {
				fmt.Printf("[Webhook ERROR] Failed to update payment status: %v\n", err)
			}

			// Send event to order-service via Kafka
			if s.Producer != nil {
				event := PaymentMessage{
					OrderID: orderID,
					Amount:  float64(piObj.Amount) / 100.0,
					Status:  "captured",
				}
				if err := s.Producer.SendMessage(context.Background(), event); err != nil {
					fmt.Printf("[Webhook ERROR] Failed to send Kafka message: %v\n", err)
				} else {
					fmt.Printf("[Webhook] ✅ Kafka event sent to order-service: order=%s, status=captured\n", orderID)
				}
			} else {
				fmt.Println("[Webhook WARNING] Kafka producer not initialized, event not sent")
			}
		} else {
			fmt.Printf("[Webhook ERROR] Failed to unmarshal payment_intent: %v\n", err)
		}

	case "payment_intent.payment_failed":
		var piObj stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &piObj); err == nil {
			orderID := piObj.Metadata["order_id"]
			if orderID == "" {
				fmt.Printf("[Webhook WARNING] No order_id in metadata for failed event. PI: %s\n", piObj.ID)
				w.WriteHeader(http.StatusOK)
				return
			}

			errorMsg := "Unknown error"
			if piObj.LastPaymentError != nil {
				if piObj.LastPaymentError.Msg != "" {
					errorMsg = piObj.LastPaymentError.Msg
				} else {
					errorMsg = piObj.LastPaymentError.Error()
				}
			}
			fmt.Printf("[Webhook] Payment failed for order: %s, Error: %s\n", orderID, errorMsg)

			// Update payment status in database
			if err := s.Repo.MarkFailedByOrderID(orderID, piObj.ID, errorMsg); err != nil {
				fmt.Printf("[Webhook ERROR] Failed to update payment status: %v\n", err)
			}

			// Send event to order-service via Kafka
			if s.Producer != nil {
				event := PaymentMessage{
					OrderID: orderID,
					Amount:  float64(piObj.Amount) / 100.0,
					Status:  "failed",
				}
				if err := s.Producer.SendMessage(context.Background(), event); err != nil {
					fmt.Printf("[Webhook ERROR] Failed to send Kafka message: %v\n", err)
				} else {
					fmt.Printf("[Webhook] ✅ Kafka event sent to order-service: order=%s, status=failed\n", orderID)
				}
			} else {
				fmt.Println("[Webhook WARNING] Kafka producer not initialized, event not sent")
			}
		} else {
			fmt.Printf("[Webhook ERROR] Failed to unmarshal payment_intent: %v\n", err)
		}

	case "transfer.created":
		var transferObj stripe.Transfer
		if err := json.Unmarshal(event.Data.Raw, &transferObj); err == nil {
			orderID := transferObj.Metadata["order_id"]
			if s.Producer != nil {
				_ = s.Producer.SendMessage(context.Background(), map[string]interface{}{
					"order_id":    orderID,
					"transfer_id": transferObj.ID,
					"amount":      float64(transferObj.Amount) / 100.0,
					"status":      "transfer_created",
					"destination": transferObj.Destination,
				})
			}
		}

	case "transfer.paid":
		var transferObj stripe.Transfer
		if err := json.Unmarshal(event.Data.Raw, &transferObj); err == nil {
			orderID := transferObj.Metadata["order_id"]
			if s.Producer != nil {
				_ = s.Producer.SendMessage(context.Background(), map[string]interface{}{
					"order_id":    orderID,
					"transfer_id": transferObj.ID,
					"amount":      float64(transferObj.Amount) / 100.0,
					"status":      "transfer_paid",
					"destination": transferObj.Destination,
				})
			}
		}
	case "transfer.failed":
		var transferObj stripe.Transfer
		if err := json.Unmarshal(event.Data.Raw, &transferObj); err == nil {
			orderID := transferObj.Metadata["order_id"]
			if s.Producer != nil {
				data := map[string]interface{}{
					"order_id":    orderID,
					"transfer_id": transferObj.ID,
					"amount":      float64(transferObj.Amount) / 100.0,
					"status":      "transfer_failed",
					"destination": transferObj.Destination,
				}
				// If failure information is provided in metadata, include it.
				if v, ok := transferObj.Metadata["failure_reason"]; ok {
					data["failure_reason"] = v
				}
				_ = s.Producer.SendMessage(context.Background(), data)
			}
		}

	case "account.updated":
		var accountObj stripe.Account
		if err := json.Unmarshal(event.Data.Raw, &accountObj); err == nil {
			vendorID := accountObj.Metadata["vendor_id"]
			if s.Producer != nil && vendorID != "" {
				_ = s.Producer.SendMessage(context.Background(), map[string]interface{}{
					"vendor_id":         vendorID,
					"account_id":        accountObj.ID,
					"charges_enabled":   accountObj.ChargesEnabled,
					"payouts_enabled":   accountObj.PayoutsEnabled,
					"details_submitted": accountObj.DetailsSubmitted,
					"status":            "account_updated",
				})
			}
		}

	case "checkout.session.completed":
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err == nil {
			fmt.Printf("[Webhook DEBUG] Checkout Session metadata: %+v\n", session.Metadata)

			// Try both order_id (snake_case) and orderId (camelCase)
			orderID := session.Metadata["order_id"]
			if orderID == "" {
				orderID = session.Metadata["orderId"]
			}

			if orderID == "" {
				fmt.Printf("[Webhook WARNING] No order_id/orderId in Checkout Session metadata. Session: %s\n", session.ID)
				w.WriteHeader(http.StatusOK)
				return
			}

			fmt.Printf("[Webhook] Checkout completed for order: %s, Session: %s, PaymentIntent: %s\n", orderID, session.ID, session.PaymentIntent.ID)

			// Update or create payment status in database
			if session.PaymentIntent != nil {
				amount := float64(session.AmountTotal) / 100.0
				currency := string(session.Currency)

				if err := s.Repo.MarkAuthByOrderIDWithAmount(orderID, session.PaymentIntent.ID, amount, currency); err != nil {
					fmt.Printf("[Webhook ERROR] Failed to update payment status: %v\n", err)
				} else {
					fmt.Printf("[Webhook] ✅ Payment record created/updated: order=%s, amount=%.2f %s\n", orderID, amount, currency)
				}
			} // Send event to order-service via Kafka
			if s.Producer != nil {
				event := PaymentMessage{
					OrderID:         orderID,
					Amount:          float64(session.AmountTotal) / 100.0,
					PaymentIntentID: session.PaymentIntent.ID,
					Status:          "checkout_completed",
				}
				if err := s.Producer.SendMessage(context.Background(), event); err != nil {
					fmt.Printf("[Webhook ERROR] Failed to send Kafka message: %v\n", err)
				} else {
					fmt.Printf("[Webhook] ✅ Kafka event sent to order-service: order=%s, status=checkout_completed\n", orderID)
				}
			} else {
				fmt.Println("[Webhook WARNING] Kafka producer not initialized, event not sent")
			}
		} else {
			fmt.Printf("[Webhook ERROR] Failed to unmarshal checkout.session: %v\n", err)
		}

	case "payout.paid":
		var payoutObj stripe.Payout
		if err := json.Unmarshal(event.Data.Raw, &payoutObj); err == nil {
			fmt.Printf("[Webhook] Payout paid: ID=%s, Amount=%.2f %s, Status=%s\n",
				payoutObj.ID, float64(payoutObj.Amount)/100.0, payoutObj.Currency, payoutObj.Status)

			// TODO: Update payout status in database
			// This will be handled by VendorRepository.UpdatePayoutStatus
		}

	case "payout.failed":
		var payoutObj stripe.Payout
		if err := json.Unmarshal(event.Data.Raw, &payoutObj); err == nil {
			failureMsg := "Unknown error"
			if payoutObj.FailureMessage != "" {
				failureMsg = payoutObj.FailureMessage
			}
			fmt.Printf("[Webhook] Payout failed: ID=%s, Reason=%s\n", payoutObj.ID, failureMsg)

			// TODO: Update payout status in database and notify vendor
		}
	}

	w.WriteHeader(http.StatusOK)
}

// ProcessPayment creates a payment record and a PaymentIntent in 'authorized' state (hold).
func (s *PaymentService) ProcessPayment(req models.PaymentRequest) (models.PaymentResponse, error) {
	// validate
	if req.OrderID == "" {
		return models.PaymentResponse{}, errors.New("order_id required")
	}

	// create DB payment (amount stored in DB as float64; convert)
	p := &models.Payment{
		OrderID:  req.OrderID,
		Amount:   req.Amount,
		Currency: req.Currency,
		Status:   "initiated",
	}
	if err := s.Repo.SavePayment(p); err != nil {
		return models.PaymentResponse{}, err
	}

	var piObj *stripe.PaymentIntent
	var err error

	// Check if this is a Connect payment
	if req.VendorStripeAccountID != "" {
		piObj, err = s.CreatePaymentIntentWithConnect(
			context.Background(),
			int64(req.Amount*100),
			req.Currency,
			req.OrderID,
			req.VendorStripeAccountID,
			int64(req.PlatformFee*100),
			req.VendorBreakdown,
		)
	} else {
		piObj, err = s.CreatePaymentIntent(context.Background(), int64(req.Amount*100), req.Currency, req.OrderID)
	}

	if err != nil {
		// mark failed
		_ = s.Repo.UpdateStatus(req.OrderID, "failed", nil, &piObj.ID)
		return models.PaymentResponse{}, err
	}

	// return client secret so frontend can confirm payment method
	resp := models.PaymentResponse{
		OrderID:       req.OrderID,
		ClientToken:   piObj.ClientSecret,
		Status:        "authorized",
		TransactionID: piObj.ID,
	}

	return resp, nil
}

func (s *PaymentService) CapturePayment(ctx context.Context, paymentIntentID string) error {
	_, err := pi.Capture(paymentIntentID, nil)
	if err != nil {
		return fmt.Errorf("failed to capture payment: %w", err)
	}

	return nil
}

func (s *PaymentService) CancelPaymentIntent(ctx context.Context, paymentIntentID string) error {
	_, err := pi.Cancel(paymentIntentID, nil)
	if err != nil {
		return fmt.Errorf("failed to cancel payment intent: %w", err)
	}

	return nil
}

// Utility functions
func (s *PaymentService) ValidateWebhookSignature(payload []byte, signature, secret string) error {
	_, err := webhook.ConstructEvent(payload, signature, secret)
	return err
}

func (s *PaymentService) GetPaymentIntentByID(ctx context.Context, paymentIntentID string) (*stripe.PaymentIntent, error) {
	return pi.Get(paymentIntentID, nil)
}

// CreateStripePayout creates a payout to a connected account
func (s *PaymentService) CreateStripePayout(ctx context.Context, stripeAccountID string, amountInCents int64, currency string) (string, error) {
	params := &stripe.PayoutParams{
		Amount:   stripe.Int64(amountInCents),
		Currency: stripe.String(currency),
	}
	params.SetStripeAccount(stripeAccountID)

	payout, err := payout.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create payout: %w", err)
	}

	return payout.ID, nil
}
