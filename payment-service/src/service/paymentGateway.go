package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"

	"payment-service/models"
	"payment-service/repository"

	"github.com/stripe/stripe-go/v74"
	pi "github.com/stripe/stripe-go/v74/paymentintent"
	"github.com/stripe/stripe-go/v74/refund"
	"github.com/stripe/stripe-go/v74/webhook"
)

type PaymentService struct {
	Repo          *repository.PaymentRepository
	SigningSecret string
	Producer      *KafkaProducer
}

func NewPaymentService(repo *repository.PaymentRepository, signinSecret string) *PaymentService {
	if k := os.Getenv("PAYMENT_GATEWAY_KEY"); k != "" {
		stripe.Key = k
	}

	return &PaymentService{
		Repo:          repo,
		SigningSecret: signinSecret,
		Producer:      nil,
	}
}

func (s *PaymentService) CreatePaymentIntent(ctx context.Context, amount int64, currency, orderID string) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(currency),
		// create intent without capturing immediately to hold funds (escrow)
		CaptureMethod: stripe.String(string(stripe.PaymentIntentCaptureMethodManual)),
	}
	params.AddMetadata("order_id", orderID)

	piObj, err := pi.New(params)
	if err != nil {
		return nil, err
	}
	// mark as authorized/held in our DB and record transaction
	_ = s.Repo.MarkAuthByOrderID(orderID, piObj.ID)
	if s.Producer != nil {
		_ = s.Producer.SendMessage(context.Background(), PaymentMessage{OrderID: orderID, Amount: float64(amount) / 100.0, Status: "authorized"})
	}
	return piObj, nil
}

// CapturePaymentIntent captures a previously authorized PaymentIntent (release funds to seller)
func (s *PaymentService) CapturePaymentIntent(ctx context.Context, paymentIntentID, orderID string) (*stripe.PaymentIntent, error) {
	// capture
	piObj, err := pi.Capture(paymentIntentID, nil)
	if err != nil {
		return nil, err
	}

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
	r, err := refund.New(params)
	if err != nil {
		return nil, err
	}

	// update refund record via repo (repo.UpdateRefundResult should be used by caller)
	if s.Producer != nil {
		// publish refund event
		_ = s.Producer.SendMessage(context.Background(), map[string]interface{}{"order_id": paymentIntentID, "status": "refund_initiated", "refund_id": refundID})
	}
	return r, nil
}

func (s *PaymentService) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	sigHeader := r.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, sigHeader, s.SigningSecret)
	if err != nil {
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}

	switch event.Type {
	case "payment_intent.succeeded", "payment_intent.captured":
		var piObj stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &piObj); err == nil {
			orderID := piObj.Metadata["order_id"]
			// mark captured in DB (implement MarkCapturedByOrderID in repo)
			_ = s.Repo.MarkCapturedByOrderID(orderID, piObj.ID)
			if s.Producer != nil {
				_ = s.Producer.SendMessage(context.Background(), PaymentMessage{OrderID: orderID, Amount: float64(piObj.Amount) / 100.0, Status: "captured"})
			}
		}
	case "payment_intent.payment_failed":
		var piObj stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &piObj); err == nil {
			orderID := piObj.Metadata["order_id"]
			_ = s.Repo.MarkFailedByOrderID(orderID, piObj.ID, piObj.LastPaymentError.Msg)
			if s.Producer != nil {
				_ = s.Producer.SendMessage(context.Background(), PaymentMessage{OrderID: orderID, Amount: float64(piObj.Amount) / 100.0, Status: "failed"})
			}
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

	// create intent with manual capture
	piObj, err := s.CreatePaymentIntent(context.Background(), int64(req.Amount*100), req.Currency, req.OrderID)
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
