package models

import "gorm.io/gorm"

type Payment struct {
	gorm.Model
	OrderID       string `gorm:"index"`
	Amount        float64
	Status        string
	Currency      string
	TransactionID string
	ProviderID    *string
	PaymentMethod string
	ClientToken   string `gorm:"uniqueIndex"`
}

type PaymentRequest struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Method   string  `json:"method"`
	OrderID  string  `json:"order_id"`
}

type PaymentResponse struct {
	OrderID       string `json:"order_id"`
	ClientToken   string `json:"client_token"`
	Status        string `json:"status"`
	Message       string `json:"message"`
	TransactionID string `json:"transaction_id"`
}

type Refund struct {
	gorm.Model
	RefundID      string `gorm:"uniqueIndex"`
	OrderID       string `gorm:"index"`
	Amount        float64
	Currency      string
	Status        string
	ProviderID    *string
	ProviderRefID *string
	Reason        *string
}

type RefundRequest struct {
	OrderID string  `json:"order_id" binding:"required"`
	Amount  float64 `json:"amount" binding:"required"`
	Reason  *string `json:"reason"`
}

type RefundResponse struct {
	RefundID string `json:"refund_id"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}

// Transaction records payment lifecycle events (authorize, capture, refund)
type Transaction struct {
	gorm.Model
	TransactionID string `gorm:"uniqueIndex"`
	OrderID       string `gorm:"index"`
	ProviderID    *string
	Amount        float64
	Currency      string
	Status        string
	ProviderRefID *string
}
