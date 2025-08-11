package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type PaymentRequest struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Method   string  `json:"method"`
	CardInfo struct {
		Number     string `json:"number"`
		Expiry     string `json:"expiry"`
		CVC        string `json:"cvc"`
	} `json:"card_info"`
}

type PaymentResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	TransactionID string `json:"transaction_id"`
}

const paymentGatewayURL = "https://api.paymentgateway.com/v1/payments"

func ProcessPayment(request PaymentRequest) (PaymentResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return PaymentResponse{}, fmt.Errorf("failed to marshal payment request: %w", err)
	}

	resp, err := http.Post(paymentGatewayURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return PaymentResponse{}, fmt.Errorf("failed to send payment request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PaymentResponse{}, fmt.Errorf("payment gateway returned non-200 status: %s", resp.Status)
	}

	var paymentResponse PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResponse); err != nil {
		return PaymentResponse{}, fmt.Errorf("failed to decode payment response: %w", err)
	}

	return paymentResponse, nil
}