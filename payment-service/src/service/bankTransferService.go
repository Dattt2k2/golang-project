package service

import (
	"context"
	"errors"
	"fmt"
	"payment-service/models"
	"payment-service/repository"
	"time"
)

type BankTransferService struct {
	VendorRepo  *repository.VendorRepository
	PaymentRepo *repository.PaymentRepository
	Producer    *KafkaProducer
}

func NewBankTransferService(vendorRepo *repository.VendorRepository, paymentRepo *repository.PaymentRepository) *BankTransferService {
	return &BankTransferService{
		VendorRepo:  vendorRepo,
		PaymentRepo: paymentRepo,
	}
}

// Create direct bank transfer for vendor payout
func (s *BankTransferService) CreateVendorPayout(ctx context.Context, req VendorPayoutRequest) (*VendorPayoutResponse, error) {
	// Get vendor bank account info
	vendor, err := s.VendorRepo.GetVendorByID(ctx, req.VendorID)
	if err != nil {
		return nil, fmt.Errorf("vendor not found: %w", err)
	}

	// Validate vendor can receive payouts
	if vendor.Status != models.VendorAccountStatusActive {
		return nil, errors.New("vendor account is not active")
	}

	if vendor.BankAccountNumber == nil || *vendor.BankAccountNumber == "" {
		return nil, errors.New("vendor bank account not configured")
	}

	// Create payout record
	payout := &models.VendorPayout{
		VendorID:          req.VendorID,
		OrderID:           req.OrderID,
		Amount:            req.Amount,
		Currency:          req.Currency,
		Status:            "pending",
		PayoutMethod:      "bank_transfer",
		BankName:          safeStringValue(vendor.BankName),
		BankAccountName:   safeStringValue(vendor.BankAccountName),
		BankAccountNumber: safeStringValue(vendor.BankAccountNumber),
		BankRoutingNumber: safeStringValue(vendor.BankRoutingNumber),
		SwiftCode:         safeStringValue(vendor.SwiftCode),
		Description:       fmt.Sprintf("Payout for order %s", req.OrderID),
	}

	// Save payout record
	if err := s.PaymentRepo.CreateVendorPayout(ctx, payout); err != nil {
		return nil, fmt.Errorf("failed to create payout record: %w", err)
	}

	// Send to bank processing queue (via Kafka)
	if s.Producer != nil {
		payoutEvent := BankPayoutEvent{
			PayoutID: payout.ID,
			VendorID: req.VendorID,
			OrderID:  req.OrderID,
			Amount:   req.Amount,
			Currency: req.Currency,
			BankInfo: BankInfo{
				BankName:      payout.BankName,
				AccountName:   payout.BankAccountName,
				AccountNumber: payout.BankAccountNumber,
				RoutingNumber: payout.BankRoutingNumber,
				SwiftCode:     payout.SwiftCode,
				Country:       vendor.Country,
			},
			CreatedAt: time.Now(),
		}

		if err := s.Producer.SendBankPayoutEvent(ctx, payoutEvent); err != nil {
			// Log error but don't fail - payout can be processed manually
			fmt.Printf("Failed to send bank payout event: %v\n", err)
		}
	}

	return &VendorPayoutResponse{
		PayoutID:    payout.ID,
		VendorID:    req.VendorID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Status:      "pending",
		Message:     "Payout created successfully. Processing to bank account.",
		BankAccount: fmt.Sprintf("****%s", (*vendor.BankAccountNumber)[len(*vendor.BankAccountNumber)-4:]), // Masked account
	}, nil
}

// Update payout status (called by bank processing system)
func (s *BankTransferService) UpdatePayoutStatus(ctx context.Context, payoutID uint, status string, transactionRef *string, failureReason *string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if transactionRef != nil {
		updates["transaction_reference"] = *transactionRef
	}

	if failureReason != nil {
		updates["failure_reason"] = *failureReason
	}

	if status == "completed" {
		now := time.Now()
		updates["processed_at"] = &now
	}

	return s.PaymentRepo.UpdateVendorPayout(ctx, payoutID, updates)
}

// Get payout history for vendor
func (s *BankTransferService) GetVendorPayouts(ctx context.Context, vendorID string, limit int, offset int) ([]*models.VendorPayout, error) {
	return s.PaymentRepo.GetVendorPayouts(ctx, vendorID, limit, offset)
}

// Request/Response models
type VendorPayoutRequest struct {
	VendorID string  `json:"vendor_id" validate:"required"`
	OrderID  string  `json:"order_id" validate:"required"`
	Amount   float64 `json:"amount" validate:"required,gt=0"`
	Currency string  `json:"currency" validate:"required"`
}

type VendorPayoutResponse struct {
	PayoutID    uint    `json:"payout_id"`
	VendorID    string  `json:"vendor_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Status      string  `json:"status"`
	Message     string  `json:"message"`
	BankAccount string  `json:"bank_account"` // Masked
}

type BankInfo struct {
	BankName      string `json:"bank_name"`
	AccountName   string `json:"account_name"`
	AccountNumber string `json:"account_number"`
	RoutingNumber string `json:"routing_number,omitempty"`
	SwiftCode     string `json:"swift_code,omitempty"`
	Country       string `json:"country"`
}

type BankPayoutEvent struct {
	PayoutID  uint      `json:"payout_id"`
	VendorID  string    `json:"vendor_id"`
	OrderID   string    `json:"order_id"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	BankInfo  BankInfo  `json:"bank_info"`
	CreatedAt time.Time `json:"created_at"`
}
