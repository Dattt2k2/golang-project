package handlers

import (
	"net/http"
	"payment-service/src/service"

	"github.com/gin-gonic/gin"
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
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		resp, err := h.VendorService.RegisterVendor(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, resp)
	}
}

// Get vendor information
func (h *VendorHandler) GetVendor() gin.HandlerFunc {
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
		_, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
			return
		}

		// TODO: Verify Stripe signature

		// Parse Stripe event
		var stripeEvent struct {
			Type string `json:"type"`
			Data struct {
				Object map[string]interface{} `json:"object"`
			} `json:"data"`
		}

		if err := c.ShouldBindJSON(&stripeEvent); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
			return
		}

		// Handle account.updated events
		if stripeEvent.Type == "account.updated" {
			accountID, ok := stripeEvent.Data.Object["id"].(string)
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
				return
			}

			err := h.VendorService.UpdateVendorFromStripe(c.Request.Context(), accountID, stripeEvent.Data.Object)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vendor"})
				return
			}
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
