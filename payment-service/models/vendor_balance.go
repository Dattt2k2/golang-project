package models

import (
	"time"

	"gorm.io/gorm"
)

// VendorBalance tracks the financial balance for each vendor
type VendorBalance struct {
	gorm.Model
	VendorID         string  `json:"vendor_id" gorm:"uniqueIndex;not null"`
	AvailableBalance float64 `json:"available_balance" gorm:"default:0"`
	PendingBalance   float64 `json:"pending_balance" gorm:"default:0"`
	TotalEarned      float64 `json:"total_earned" gorm:"default:0"`
	TotalPaidOut     float64 `json:"total_paid_out" gorm:"default:0"`
	Currency         string  `json:"currency" gorm:"default:'usd'"`
}

// VendorTransaction records all financial transactions for vendors
type VendorTransaction struct {
	gorm.Model
	VendorID       string     `json:"vendor_id" gorm:"index;not null"`
	OrderID        *string    `json:"order_id" gorm:"index"`
	Type           string     `json:"type" gorm:"not null"` // 'sale', 'payout', 'refund', 'fee', 'adjustment'
	Amount         float64    `json:"amount" gorm:"not null"`
	BalanceAfter   float64    `json:"balance_after"`
	Status         string     `json:"status" gorm:"default:'completed'"` // 'completed', 'pending', 'failed'
	Description    string     `json:"description"`
	StripePayoutID *string    `json:"stripe_payout_id" gorm:"index"`
	Metadata       *string    `json:"metadata"` // JSON string for additional data
	CompletedAt    *time.Time `json:"completed_at"`
}

// VendorBalanceResponse for API responses
type VendorBalanceResponse struct {
	VendorID         string  `json:"vendor_id"`
	AvailableBalance float64 `json:"available_balance"`
	PendingBalance   float64 `json:"pending_balance"`
	TotalEarned      float64 `json:"total_earned"`
	TotalPaidOut     float64 `json:"total_paid_out"`
	Currency         string  `json:"currency"`
}

// VendorTransactionResponse for API responses
type VendorTransactionResponse struct {
	ID             uint       `json:"id"`
	VendorID       string     `json:"vendor_id"`
	OrderID        *string    `json:"order_id"`
	Type           string     `json:"type"`
	Amount         float64    `json:"amount"`
	BalanceAfter   float64    `json:"balance_after"`
	Status         string     `json:"status"`
	Description    string     `json:"description"`
	StripePayoutID *string    `json:"stripe_payout_id"`
	CreatedAt      time.Time  `json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at"`
}

// PayoutRequest for creating payout
type PayoutRequest struct {
	VendorID string  `json:"vendor_id" validate:"required"`
	Amount   float64 `json:"amount" validate:"required,gt=0"`
	Currency string  `json:"currency" validate:"required"`
}

// PayoutResponse for payout operations
type PayoutResponse struct {
	PayoutID       string    `json:"payout_id"`
	VendorID       string    `json:"vendor_id"`
	Amount         float64   `json:"amount"`
	Currency       string    `json:"currency"`
	Status         string    `json:"status"`
	ExpectedDate   time.Time `json:"expected_date"`
	StripePayoutID string    `json:"stripe_payout_id"`
}
