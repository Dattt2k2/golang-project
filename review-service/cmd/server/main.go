package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	pb "module/gRPC-Order/service"
	"review-service/config"
	"review-service/internal/cron"
	"review-service/internal/handlers"
	"review-service/internal/kafka"
	"review-service/internal/repository"
	"review-service/internal/routes"
	"review-service/internal/services"
	logger "review-service/log"
)

func main() {
	logger.InitLogger()
	defer logger.Sync()

	_ = godotenv.Load(".env")

	dynamoRegion := config.Get("DYNAMODB_REGION", "")

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(dynamoRegion),
	)
	if err != nil {
		logger.Logger.Fatal("unable to load SDK config, " + err.Error())
	}

	dynamoClient := dynamodb.NewFromConfig(cfg)

	// Debug connection
	tables, err := dynamoClient.ListTables(context.Background(), &dynamodb.ListTablesInput{})
	if err != nil {
		log.Printf("Warning: Could not connect to DynamoDB: %v", err)
	} else {
		log.Printf("Connected to DynamoDB successfully")
		log.Printf("Available tables: %v", tables.TableNames)
	}

	reviewTable := config.Get("DYNAMODB_REVIEW_TABLE", "review-table")
	pendingTable := config.Get("DYNAMODB_PENDING_TABLE", "review-pending-table")

	repo := repository.NewReviewRepository(dynamoClient, reviewTable, pendingTable)
	service := services.NewReviewService(repo)

	 orderServiceAddress := config.Get("ORDER_SERVICE_ADDRESS", "order-service:8084")
    conn, err := grpc.Dial(orderServiceAddress, grpc.WithInsecure())
    if err != nil {
        logger.Logger.Fatal("Failed to connect to order-service: " + err.Error())
    }
    defer conn.Close()
	orderClient := pb.NewOrderServcieClient(conn)

	h := handlers.NewReviewHandler(service, orderClient)

	// Initialize Kafka Producer
	kafkaBrokers := strings.Split(config.Get("KAFKA_BROKER", ""), ",")
	kafkaProducer := kafka.NewProducer(kafkaBrokers, config.Get("KAFKA_RATING_TOPIC", "product_rating_updates"))
	defer kafkaProducer.Close()

	// Initialize Review Aggregator (sử dụng AWS SDK v2)
	aggregator := cron.NewPendingReviewAggregator(dynamoClient, kafkaProducer, pendingTable)

	// Start Scheduler
	scheduler := cron.NewScheduler(aggregator)
	scheduler.Start()
	defer scheduler.Stop()

	r := gin.Default()
	routes.RegisterRoutes(r, h)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8093"
	}

	log.Printf("review-service listening on :%s with cronjob scheduler", port)

	// Graceful shutdown
	go func() {
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("server exit: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down review service...")
	scheduler.Stop()
	kafkaProducer.Close()
	log.Println("Review service stopped")
}
