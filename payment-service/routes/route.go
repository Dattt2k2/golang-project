package routes

import (
	"payment-service/repository"
	"payment-service/src/handlers"
	"payment-service/src/service"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(repo *repository.PaymentRepository, vendorRepo *repository.VendorRepository, webhookSecret string) *gin.Engine {
	r := gin.Default()

	// Initialize services
	paymentService := service.NewPaymentService(repo, webhookSecret)
	refundService := service.NewRefundService(repo)
	vendorService := service.NewVendorService(vendorRepo, paymentService)

	// Initialize handlers
	handler := handlers.NewHandler(repo, paymentService, refundService, webhookSecret)
	vendorHandler := handlers.NewVendorHandler(vendorService)

	// API routes
	api := r.Group("/api/v1")
	{
		// Payment routes
		api.POST("/payments", handler.ProcessPaymentHandler())
		api.GET("/payments/:order_id", handler.GetPaymentByOrderID())

		// Refund routes
		api.POST("/refunds", handler.ProcessRefundHandler())
		api.GET("/refunds/:refund_id", handler.GetRefundByRefundID())

		// Vendor routes
		api.POST("/vendors/register", vendorHandler.RegisterVendor())
		api.GET("/vendors/:vendor_id", vendorHandler.GetVendor())
		api.POST("/vendors/:vendor_id/onboarding", vendorHandler.CreateOnboardingLink())
		api.GET("/vendors/:vendor_id/onboarding/status", vendorHandler.GetOnboardingStatus())
		api.GET("/vendors/:vendor_id/payout-methods", vendorHandler.GetPayoutMethods())
		api.PUT("/vendors/:vendor_id/bank-account", vendorHandler.UpdateBankAccount())
		api.GET("/vendors/:vendor_id/bank-account", vendorHandler.GetBankAccount())
	}

	// Public routes for Stripe redirects (no auth needed)
	public := r.Group("/public")
	{
		public.GET("/vendor/onboarding/success", vendorHandler.OnboardingSuccess())
		public.GET("/vendor/onboarding/refresh", vendorHandler.OnboardingRefresh())
	}

	// Webhook routes
	webhook := r.Group("/webhook")
	{
		webhook.POST("/payment", handler.PaymentWebhook())
		webhook.POST("/refund", handler.RefundWebhook())
		webhook.POST("/stripe/connect", vendorHandler.StripeConnectWebhook())
	}

	return r
}
