package main

import (
	"log"
	"os"
	"payment-service/database"
	"payment-service/repository"
	"payment-service/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	db, err := database.NewPostgres()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate tables - you may want to add this function to database package
	// For now, we'll skip auto-migration

	// Initialize repositories
	paymentRepo := repository.NewPaymentRepository(db)
	vendorRepo := repository.NewVendorRepository(db)

	// Get webhook secret from environment
	webhookSecret := os.Getenv("WEBHOOK_SECRET")
	if webhookSecret == "" {
		log.Println("Warning: WEBHOOK_SECRET not set")
	}

	// Setup routes
	router := routes.SetupRoutes(paymentRepo, vendorRepo, webhookSecret)

	// Configure Gin mode
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	log.Printf("Payment service starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
