package models

import (
	"time"

	"gorm.io/gorm"
)

type Payment struct {
	gorm.Model
	OrderID       string  `json:"order_id" gorm:"uniqueIndex;not null"`
	Amount        float64 `json:"amount" gorm:"not null"`
	Currency      string  `json:"currency" gorm:"not null"`
	Status        string  `json:"status" gorm:"not null"`   // initiated, authorized, captured, failed, refunded
	ProviderID    *string `json:"provider_id" gorm:"index"` // Stripe PaymentIntent ID
	TransactionID string  `json:"transaction_id" gorm:"index"`

	// Stripe Connect fields
	VendorStripeAccountID *string `json:"vendor_stripe_account_id"`
	PlatformFee           float64 `json:"platform_fee" gorm:"default:0"`
	VendorAmount          float64 `json:"vendor_amount" gorm:"default:0"`
	VendorBreakdown       *string `json:"vendor_breakdown"` // JSON string

	// Additional fields
	PaymentMethod string  `json:"payment_method" gorm:"default:'stripe'"`
	Description   string  `json:"description"`
	FailureReason *string `json:"failure_reason"`
	RefundAmount  float64 `json:"refund_amount" gorm:"default:0"`

	// Timestamps
	AuthorizedAt *time.Time `json:"authorized_at"`
	CapturedAt   *time.Time `json:"captured_at"`
	FailedAt     *time.Time `json:"failed_at"`
}

type PaymentRequest struct {
	OrderID       string  `json:"order_id" validate:"required"`
	Amount        float64 `json:"amount" validate:"required,gt=0"`
	Currency      string  `json:"currency" validate:"required"`
	Description   string  `json:"description"`
	PaymentMethod string  `json:"payment_method" validate:"required"`

	// Stripe Connect fields
	VendorStripeAccountID string  `json:"vendor_stripe_account_id,omitempty"`
	PlatformFee           float64 `json:"platform_fee,omitempty"`
	VendorBreakdown       string  `json:"vendor_breakdown,omitempty"`
}

type PaymentResponse struct {
	OrderID       string  `json:"order_id"`
	ClientToken   string  `json:"client_token"` // client_secret for frontend
	Status        string  `json:"status"`
	TransactionID string  `json:"transaction_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
}

type Refund struct {
	gorm.Model
	PaymentID     uint       `json:"payment_id" gorm:"not null"`
	Payment       Payment    `json:"payment" gorm:"foreignKey:PaymentID"`
	OrderID       string     `json:"order_id" gorm:"not null;index"`
	Amount        float64    `json:"amount" gorm:"not null"`
	Currency      string     `json:"currency" gorm:"not null"`
	Status        string     `json:"status" gorm:"not null"` // pending, succeeded, failed
	RefundID      string     `json:"refund_id" gorm:"uniqueIndex"`
	Reason        string     `json:"reason"`
	FailureReason *string    `json:"failure_reason"`
	ProcessedAt   *time.Time `json:"processed_at"`
}

type RefundRequest struct {
	OrderID string  `json:"order_id" validate:"required"`
	Amount  float64 `json:"amount" validate:"required,gt=0"`
	Reason  string  `json:"reason" validate:"required"`
}

type RefundResponse struct {
	RefundID string  `json:"refund_id"`
	Status   string  `json:"status"`
	Message  string  `json:"message"`
	Amount   float64 `json:"amount,omitempty"`
}

type Transaction struct {
	gorm.Model
	TransactionID string   `json:"transaction_id" gorm:"uniqueIndex;not null"`
	OrderID       string   `json:"order_id" gorm:"not null;index"`
	PaymentID     *uint    `json:"payment_id"`
	Payment       *Payment `json:"payment" gorm:"foreignKey:PaymentID"`

	Type     string  `json:"type" gorm:"not null"`   // payment, refund, transfer
	Status   string  `json:"status" gorm:"not null"` // pending, completed, failed
	Amount   float64 `json:"amount" gorm:"not null"`
	Currency string  `json:"currency" gorm:"not null"`

	// Provider details
	ProviderID   *string `json:"provider_id"`   // Stripe transaction ID
	ProviderData *string `json:"provider_data"` // JSON string of provider response

	// For transfers
	DestinationAccountID *string `json:"destination_account_id"`
	TransferID           *string `json:"transfer_id"`

	ProcessedAt   *time.Time `json:"processed_at"`
	FailureReason *string    `json:"failure_reason"`
}

// Vendor Connect Account model
type VendorAccount struct {
	gorm.Model
	VendorID        string `json:"vendor_id" gorm:"uniqueIndex;not null"`
	StripeAccountID string `json:"stripe_account_id" gorm:"uniqueIndex;not null"`
	Email           string `json:"email" gorm:"not null"`
	Country         string `json:"country" gorm:"not null"`

	// Account status
	ChargesEnabled   bool `json:"charges_enabled" gorm:"default:false"`
	PayoutsEnabled   bool `json:"payouts_enabled" gorm:"default:false"`
	DetailsSubmitted bool `json:"details_submitted" gorm:"default:false"`

	// Business information
	BusinessName *string `json:"business_name"`
	BusinessURL  *string `json:"business_url"`
	BusinessType *string `json:"business_type"`

	// Bank Account Information
	BankName          *string `json:"bank_name"`
	BankAccountName   *string `json:"bank_account_name"`
	BankAccountNumber *string `json:"bank_account_number"`
	BankRoutingNumber *string `json:"bank_routing_number"` // For international transfers
	SwiftCode         *string `json:"swift_code"`          // For international transfers

	// Onboarding
	OnboardingCompleted bool    `json:"onboarding_completed" gorm:"default:false"`
	OnboardingURL       *string `json:"onboarding_url"`

	// Status tracking
	Status      string    `json:"status" gorm:"default:'pending'"` // pending, active, restricted, inactive
	LastUpdated time.Time `json:"last_updated" gorm:"autoUpdateTime"`
}

// Payment status constants
const (
	PaymentStatusInitiated  = "initiated"
	PaymentStatusAuthorized = "authorized"
	PaymentStatusCaptured   = "captured"
	PaymentStatusFailed     = "failed"
	PaymentStatusRefunded   = "refunded"
	PaymentStatusCancelled  = "cancelled"
)

// Refund status constants
const (
	RefundStatusPending   = "pending"
	RefundStatusSucceeded = "succeeded"
	RefundStatusFailed    = "failed"
)

// Transaction type constants
const (
	TransactionTypePayment  = "payment"
	TransactionTypeRefund   = "refund"
	TransactionTypeTransfer = "transfer"
)

// Vendor Payout model for bank transfers
type VendorPayout struct {
	gorm.Model
	VendorID     string  `json:"vendor_id" gorm:"not null;index"`
	OrderID      string  `json:"order_id" gorm:"not null;index"`
	Amount       float64 `json:"amount" gorm:"not null"`
	Currency     string  `json:"currency" gorm:"not null"`
	Status       string  `json:"status" gorm:"not null"`        // pending, processing, completed, failed
	PayoutMethod string  `json:"payout_method" gorm:"not null"` // bank_transfer, stripe_connect

	// Bank details (copied from vendor at payout time)
	BankName          string `json:"bank_name"`
	BankAccountName   string `json:"bank_account_name"`
	BankAccountNumber string `json:"bank_account_number"`
	BankRoutingNumber string `json:"bank_routing_number"`
	SwiftCode         string `json:"swift_code"`

	// Processing info
	TransactionReference *string    `json:"transaction_reference"`
	Description          string     `json:"description"`
	FailureReason        *string    `json:"failure_reason"`
	ProcessedAt          *time.Time `json:"processed_at"`
}

// Vendor account status constants
const (
	VendorAccountStatusPending    = "pending"
	VendorAccountStatusActive     = "active"
	VendorAccountStatusRestricted = "restricted"
	VendorAccountStatusInactive   = "inactive"
)

// Payout status constants
const (
	PayoutStatusPending    = "pending"
	PayoutStatusProcessing = "processing"
	PayoutStatusCompleted  = "completed"
	PayoutStatusFailed     = "failed"
)
