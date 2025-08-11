package handlers

import (
	"encoding/json"
	"net/http"

	"payment-service/src/service"
	"payment-service/src/models"
)

func ProcessPaymentHandler(w http.ResponseWriter, r *http.Request) {
	var paymentRequest models.PaymentRequest

	if err := json.NewDecoder(r.Body).Decode(&paymentRequest); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	paymentResponse, err := service.ProcessPayment(paymentRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(paymentResponse)
}