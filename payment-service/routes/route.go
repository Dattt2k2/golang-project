package routes

import (
    "payment-service/repository"
    "payment-service/src/handlers"
    "payment-service/src/service"

    "github.com/gin-gonic/gin"
)

func SetupRoutes(repo *repository.PaymentRepository, webhookSecret string) *gin.Engine {
    r := gin.Default()

    // Initialize services
    paymentService := service.NewPaymentService(repo, webhookSecret)
    refundService := service.NewRefundService(repo)

    // Initialize handler
    handler := handlers.NewHandler(repo, paymentService, refundService, webhookSecret)

    // API routes
    api := r.Group("/api/v1")
    {
        // Payment routes
        api.POST("/payments", handler.ProcessPaymentHandler())
        api.GET("/payments/:order_id", handler.GetPaymentByOrderID())

        // Refund routes
        api.POST("/refunds", handler.ProcessRefundHandler())
        api.GET("/refunds/:refund_id", handler.GetRefundByRefundID())
    }

    // Webhook routes
    webhook := r.Group("/webhook")
    {
        webhook.POST("/payment", handler.PaymentWebhook())
        webhook.POST("/refund", handler.RefundWebhook())
    }

    return r
}