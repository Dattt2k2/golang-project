package handlers

import (
	"log"
	"net/http"
	"payment-service/src/service"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v74/webhook"
)

type VendorHandler struct {
	VendorService *service.VendorService
}

func NewVendorHandler(vendorService *service.VendorService) *VendorHandler {
	return &VendorHandler{
		VendorService: vendorService,
	}
}

// Register new vendor
func (h *VendorHandler) RegisterVendor() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req service.VendorRegistrationRequest
		vendorID := c.GetHeader("X-User-ID")
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}
		req.VendorID = vendorID
		resp, err := h.VendorService.RegisterVendor(c.Request.Context(), req)
		if err != nil {
			log.Printf("Error registering vendor: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, resp)
	}
}

// Get vendor information
func (h *VendorHandler) GetVendor() gin.HandlerFunc {
	return func(c *gin.Context) {
		vendorID := c.GetHeader("X-User-ID")
		if vendorID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "vendor_id is required"})
			return
		}

		vendor, err := h.VendorService.GetVendor(c.Request.Context(), vendorID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Vendor not found"})
			return
		}

		c.JSON(http.StatusOK, vendor)
	}
}

// Create onboarding link
func (h *VendorHandler) CreateOnboardingLink() gin.HandlerFunc {
	return func(c *gin.Context) {
		vendorID := c.Param("vendor_id")
		if vendorID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "vendor_id is required"})
			return
		}

		onboardingURL, err := h.VendorService.CreateOnboardingLink(c.Request.Context(), vendorID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"vendor_id":      vendorID,
			"onboarding_url": onboardingURL,
			"message":        "Onboarding link created successfully",
		})
	}
}

// Handle Stripe Connect webhook for account updates
func (h *VendorHandler) StripeConnectWebhook() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read raw payload
		payload, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
			return
		}

		// Get Stripe signature from header
		sigHeader := c.GetHeader("Stripe-Signature")
		if sigHeader == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing Stripe-Signature header"})
			return
		}

		// Use PaymentService's signing secret to verify
		event, err := webhook.ConstructEventWithOptions(
			payload,
			sigHeader,
			h.VendorService.PaymentService.SigningSecret,
			webhook.ConstructEventOptions{
				IgnoreAPIVersionMismatch: true,
			},
		)
		if err != nil {
			log.Printf("Invalid webhook signature: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature"})
			return
		}

		log.Printf("Received Stripe Connect webhook: %s (ID: %s)", event.Type, event.ID)

		// Handle account.updated events
		if event.Type == "account.updated" {
			accountID, ok := event.Data.Object["id"].(string)
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
				return
			}

			log.Printf("Processing account.updated for account: %s", accountID)

			err := h.VendorService.UpdateVendorFromStripe(c.Request.Context(), accountID, event.Data.Object)
			if err != nil {
				log.Printf("Failed to update vendor from Stripe: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vendor"})
				return
			}

			log.Printf("Successfully updated vendor for account: %s", accountID)
		} else {
			log.Printf("Received unhandled webhook event type: %s", event.Type)
		}

		c.Status(http.StatusOK)
	}
}

// Update vendor bank account
func (h *VendorHandler) UpdateBankAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		vendorID := c.Param("vendor_id")
		if vendorID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "vendor_id is required"})
			return
		}

		var req service.UpdateBankAccountRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		err := h.VendorService.UpdateVendorBankAccount(c.Request.Context(), vendorID, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"vendor_id": vendorID,
			"message":   "Bank account updated successfully",
		})
	}
}

// Get vendor bank account for transfers
func (h *VendorHandler) GetBankAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		vendorID := c.Param("vendor_id")
		if vendorID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "vendor_id is required"})
			return
		}

		bankAccount, err := h.VendorService.GetVendorBankAccount(c.Request.Context(), vendorID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, bankAccount)
	}
}

// Handle onboarding success redirect from Stripe
func (h *VendorHandler) OnboardingSuccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		vendorID := c.Query("vendor_id")
		if vendorID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "vendor_id is required"})
			return
		}

		// Get vendor to check current status
		vendor, err := h.VendorService.GetVendor(c.Request.Context(), vendorID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Vendor not found"})
			return
		}

		// Return success page with current status
		c.JSON(http.StatusOK, gin.H{
			"vendor_id": vendorID,
			"status":    vendor.Status,
			"message":   "Onboarding process initiated. Please wait for Stripe to verify your account.",
			"next_step": "Your account will be activated once Stripe completes verification.",
		})
	}
}

// Handle onboarding refresh (when link expires)
func (h *VendorHandler) OnboardingRefresh() gin.HandlerFunc {
	return func(c *gin.Context) {
		vendorID := c.Query("vendor_id")
		if vendorID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "vendor_id is required"})
			return
		}

		// Create new onboarding link
		onboardingURL, err := h.VendorService.CreateOnboardingLink(c.Request.Context(), vendorID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Redirect to new onboarding URL
		c.Redirect(http.StatusTemporaryRedirect, onboardingURL)
	}
}

// Get onboarding status
func (h *VendorHandler) GetOnboardingStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		vendorID := c.Param("vendor_id")
		if vendorID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "vendor_id is required"})
			return
		}

		vendor, err := h.VendorService.GetVendor(c.Request.Context(), vendorID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Vendor not found"})
			return
		}

		status := map[string]interface{}{
			"vendor_id":            vendorID,
			"status":               vendor.Status,
			"onboarding_completed": vendor.OnboardingCompleted,
			"charges_enabled":      vendor.ChargesEnabled,
			"payouts_enabled":      vendor.PayoutsEnabled,
			"details_submitted":    vendor.DetailsSubmitted,
		}

		// Add onboarding URL if not completed
		if !vendor.OnboardingCompleted && vendor.OnboardingURL != nil {
			status["onboarding_url"] = *vendor.OnboardingURL
		}

		c.JSON(http.StatusOK, status)
	}
}

// Get available payout methods for vendor
func (h *VendorHandler) GetPayoutMethods() gin.HandlerFunc {
	return func(c *gin.Context) {
		vendorID := c.Param("vendor_id")
		if vendorID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "vendor_id is required"})
			return
		}

		methods, err := h.VendorService.GetAvailablePayoutMethods(c.Request.Context(), vendorID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, methods)
	}
}

// ProcessOrderCompletionPayout handles payout when order is completed
func (h *VendorHandler) ProcessOrderCompletionPayout() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			VendorID string  `json:"vendor_id" binding:"required"`
			OrderID  string  `json:"order_id" binding:"required"`
			Amount   float64 `json:"amount" binding:"required,gt=0"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		// Create payout
		err := h.VendorService.CreateVendorPayout(c.Request.Context(), req.VendorID, req.OrderID, req.Amount)
		if err != nil {
			log.Printf("Failed to create payout: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":   "Payout created successfully",
			"vendor_id": req.VendorID,
			"order_id":  req.OrderID,
			"amount":    req.Amount,
		})
	}
}
