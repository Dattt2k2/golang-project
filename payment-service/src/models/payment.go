package models

type PaymentRequest struct {
    Amount   float64 `json:"amount"`
    Currency string  `json:"currency"`
    Method   string  `json:"method"`
    OrderID  string  `json:"order_id"`
}

type PaymentResponse struct {
    Status  string `json:"status"`
    Message string `json:"message"`
    TransactionID string `json:"transaction_id"`
}