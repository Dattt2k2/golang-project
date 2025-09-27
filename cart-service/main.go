package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	logger "cart-service/log"
	"cart-service/repository"
	"cart-service/routes"
	"cart-service/service"

	pb "github.com/Dattt2k2/golang-project/module/gRPC-cart/service"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	logger.InitLogger()
	defer logger.Sync()

	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}

	// Setup DynamoDB with explicit credentials
	dynamoRegion := os.Getenv("DYNAMODB_REGION")
	if dynamoRegion == "" {
		dynamoRegion = "us-west-2"
	}

	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	log.Printf("AWS Region: %s", dynamoRegion)

	// Fix the slice bounds error
	if len(accessKey) >= 10 {
		log.Printf("AWS Access Key ID: %s", accessKey[:10]+"...")
	} else if len(accessKey) > 0 {
		log.Printf("AWS Access Key ID: %s", accessKey+"...")
	} else {
		log.Printf("AWS Access Key ID: not provided")
	}

	var cfg aws.Config

	if accessKey != "" && secretKey != "" {
		// Use explicit credentials
		cfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion(dynamoRegion),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		)
	} else {
		// Fallback to default config
		log.Printf("Using default AWS configuration")
		cfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion(dynamoRegion),
		)
	}

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

	// Create shared CartService
	tableName := os.Getenv("DYNAMODB_TABLE")
	if tableName == "" {
		tableName = "cart-table"
	}
	log.Printf("Using DynamoDB table: %s", tableName)

	cartRepo := repository.NewCartRepository(dynamoClient, tableName)
	cartSvc, err := service.NewCartService(cartRepo)
	if err != nil {
		logger.Logger.Fatal("Failed to create CartService: " + err.Error())
	}

	// Setup dependencies
	cartController, cartServer := routes.SetupCartDependencies(cartSvc)

	// Khởi tạo router
	router := gin.Default()

	// Thiết lập HTTP routes
	routes.CartRoutes(router, cartController)

	// Thiết lập gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterCartServiceServer(grpcServer, cartServer)

	// Khởi động gRPC server
	grpcPort := os.Getenv("gRPC_PORT")
	if grpcPort == "" {
		grpcPort = "8090"
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	go func() {
		log.Printf("Starting gRPC server at :%s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Xử lý tín hiệu để tắt server một cách graceful
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Khởi động HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	go func() {
		log.Printf("Starting HTTP server at :%s", port)
		if err := router.Run(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Đợi tín hiệu để dừng server
	<-quit
	log.Println("Shutting down server...")

	// Dừng gRPC server
	grpcServer.GracefulStop()
	log.Println("Server exited")
}
