package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"payment-service/models"
	"payment-service/repository"
	"payment-service/src/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	PaymentRepo    *repository.PaymentRepository
	PaymentService *service.PaymentService
	RefundService  *service.RefundService
	WebhookSecret  string
}

func NewHandler(repo *repository.PaymentRepository, paymentSvc *service.PaymentService, refundSvc *service.RefundService, secret string) *Handler {
	return &Handler{
		PaymentRepo:    repo,
		PaymentService: paymentSvc,
		RefundService:  refundSvc,
		WebhookSecret:  secret,
	}
}

// Payment handlers
func (h *Handler) ProcessPaymentHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.PaymentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		resp, err := h.PaymentService.ProcessPayment(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

func (h *Handler) GetPaymentByOrderID() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("order_id")

		payment, err := h.PaymentRepo.GetByOrderID(orderID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
			return
		}

		resp := models.PaymentResponse{
			Status:        payment.Status,
			TransactionID: payment.TransactionID,
		}

		c.JSON(http.StatusOK, resp)
	}
}

// Refund handlers
func (h *Handler) ProcessRefundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.RefundRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		resp, err := h.RefundService.ProcessRefund(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

func (h *Handler) GetRefundByRefundID() gin.HandlerFunc {
	return func(c *gin.Context) {
		refundID := c.Param("refund_id")

		refund, err := h.PaymentRepo.GetRefundByRefundID(refundID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Refund not found"})
			return
		}

		resp := models.RefundResponse{
			RefundID: refund.RefundID,
			Status:   refund.Status,
			Message:  "Refund found",
		}

		c.JSON(http.StatusOK, resp)
	}
}

// Webhook handlers

// InternalPaymentWebhook handles internal payment webhooks (not from Stripe)
// Uses custom X-Signature header for verification
func (h *Handler) InternalPaymentWebhook() gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
			return
		}

		// Use X-Signature header for internal webhooks
		signature := c.GetHeader("X-Signature")

		// Verify internal webhook signature
		if h.WebhookSecret != "" && !verifySignature(body, signature, h.WebhookSecret) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
			return
		}

		var payload struct {
			OrderID       string  `json:"order_id"`
			Status        string  `json:"status"`
			ProviderID    *string `json:"provider_id"`
			TransactionID *string `json:"transaction_id"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
			return
		}

		if err := h.PaymentRepo.UpdateStatus(payload.OrderID, payload.Status, payload.ProviderID, payload.TransactionID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
			return
		}

		c.Status(http.StatusOK)
	}
}

func (h *Handler) RefundWebhook() gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
			return
		}

		sig := c.GetHeader("X-Signature")
		if h.WebhookSecret != "" && !verifySignature(body, sig, h.WebhookSecret) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
			return
		}

		var payload struct {
			RefundID      string  `json:"refund_id"`
			Status        string  `json:"status"`
			ProviderRefID *string `json:"provider_ref_id"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
			return
		}

		if err := h.PaymentRepo.UpdateRefundResult(payload.RefundID, payload.Status, payload.ProviderRefID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update refund status"})
			return
		}

		c.Status(http.StatusOK)
	}
}

// StripeWebhook handles official Stripe webhook events
// This properly verifies Stripe signatures using webhook.ConstructEvent
// and processes events like payment_intent.succeeded, payment_intent.captured, etc.
func (h *Handler) StripeWebhook() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Delegate to PaymentService which has proper Stripe webhook handling
		h.PaymentService.HandleWebhook(c.Writer, c.Request)
	}
}

func verifySignature(payload []byte, signature, secret string) bool {
	if signature == "" || secret == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
