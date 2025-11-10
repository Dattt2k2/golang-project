package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"payment-service/repository"
	"payment-service/routes"

	// "payment-service/src/config"
	"payment-service/src/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	// cfg, err := config.LoadConfig()
	// if err != nil {
	// 	log.Fatalf("Error loading configuration: %v", err)
	// }

	// Database configuration
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "password"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "payment_db"
	}

	// PostgreSQL DSN
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		dbHost, dbUser, dbPassword, dbName, dbPort)

	// Initialize database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL database: %v", err)
	}

	// Initialize repository
	paymentRepo := repository.NewPaymentRepository(db)
	if err := paymentRepo.Migrate(); err != nil {
		// Handle duplicate table errors gracefully in development
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") ||
			strings.Contains(strings.ToLower(err.Error()), "already exists") {
			log.Printf("Migration warning (tables may already exist): %v", err)
		} else {
			log.Fatalf("Failed to migrate database: %v", err)
		}
	}

	// Initialize vendor repository (required by routes.SetupRoutes)
	vendorRepo := repository.NewVendorRepository(db)
	// VendorRepository does not expose a Migrate method; if schema migration is required,
	// perform it using the repository package or gorm AutoMigrate directly.
	// For now, assume vendor tables are managed elsewhere or add a Migrate method to repository.VendorRepository.

	// Get webhook secret from environment (Stripe webhook secret)
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if webhookSecret == "" {
		log.Println("Warning: STRIPE_WEBHOOK_SECRET not set, using fallback")
		webhookSecret = os.Getenv("WEBHOOK_SECRET")
		if webhookSecret == "" {
			webhookSecret = "default-secret"
			log.Println("Warning: Using default webhook secret - NOT FOR PRODUCTION!")
		}
	}
	log.Printf("Webhook secret loaded: %s... (length: %d)", webhookSecret[:min(20, len(webhookSecret))], len(webhookSecret))

	paymentService := service.NewPaymentService(paymentRepo, webhookSecret)

	orderServiceURL := os.Getenv("ORDER_SERVICE_URL")
	if orderServiceURL == "" {
		orderServiceURL = "http://order-service:8087" // Default in Docker
	}

	kafkaBrokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	if len(kafkaBrokers) == 0 || kafkaBrokers[0] == "" {
		kafkaBrokers = []string{"kafka:9092"}
	}

	// Start payment consumer to handle payment requests from order-service
	if len(kafkaBrokers) > 0 && kafkaBrokers[0] != "" {
		paymentConsumer := service.NewPaymentConsumer(paymentService, orderServiceURL)

		// Start consumer in goroutine
		go paymentConsumer.StartConsumer(kafkaBrokers)
		log.Println("Payment consumer started for payment_requests topic")
	}

	// Setup routes using SetupRoutes function
	router := routes.SetupRoutes(paymentRepo, vendorRepo, webhookSecret)

	// Add health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "payment-service", "database": "postgresql"})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}

	log.Printf("Server starting on port %s with PostgreSQL", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
