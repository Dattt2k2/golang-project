package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	controllers "product-service/controller"
	"product-service/database"
	"product-service/handler"
	"product-service/kafka"
	logger "product-service/log"
	"product-service/repository"
	"product-service/routes"
	service "product-service/service"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	dynamodbv1 "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	pb "module/gRPC-Product/service"
)

// Sửa toàn bộ function main:

func main() {
	logger.InitLogger()
	defer logger.Sync()

	err := godotenv.Load("./product-service/.env")
	if err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}

	dynamoRegion := os.Getenv("DYNAMODB_REGION")
	if dynamoRegion == "" {
		dynamoRegion = "us-west-2"
	}

	log.Printf("AWS Region: %s", dynamoRegion)

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(dynamoRegion),
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

	database.InitRedis()
	defer database.RedisClient.Close()
	log.Printf("Connected to Redis")

	// Tạo ProductService chung cho cả gRPC và HTTP
	tableName := os.Getenv("DYNAMODB_TABLE")
	if tableName == "" {
		tableName = "product-table"
	}
	log.Printf("Using DynamoDB table: %s", tableName)

	repo := repository.NewProductRepository(dynamoClient, tableName)
	productSvc := service.NewProductService(repo, service.NewS3Service())

	grpcReady := make(chan bool)

	go func() {
		grpcPort := os.Getenv("gRPC_PORT")
		if grpcPort == "" {
			grpcPort = "8089"
		}
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
		}

		// Sử dụng productSvc chung
		productServer := controllers.NewProductServer(productSvc)
		s := grpc.NewServer()

		pb.RegisterProductServiceServer(s, productServer)
		grpcReady <- true

		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to connect to gRPC Server: %v", err)
		}
	}()

	<-grpcReady

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	// Kafka setup - sử dụng productSvc chung
	kafkaHost := os.Getenv("KAFKA_URL")
	brokers := []string{kafkaHost}
	if kafkaHost == "" {
		brokers = []string{"kafka:9092"}
	}
	kafka.InitProductEventProducer(brokers)
	go kafka.ConsumeOrderSuccess(brokers, productSvc)
	go kafka.ConsumerOrderReturned(brokers, productSvc)

	// Send initial product events for search-service indexing
	go sendInitialProductEvents(productSvc)

	// Initialize AWS Session v1 for Rating Update Handler
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(dynamoRegion),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"",
		),
	})
	if err != nil {
		log.Printf("Warning: Failed to create AWS session for rating handler: %v", err)
	} else {
		dynamoDBClientV1 := dynamodbv1.New(sess)

		// Initialize Rating Update Handler
		ratingHandler := handler.NewRatingUpdateHandler(dynamoDBClientV1, tableName)

		// Initialize Kafka Rating Consumer
		kafkaBrokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
		if len(kafkaBrokers) == 0 || kafkaBrokers[0] == "" {
			kafkaBrokers = []string{"kafka:9092"}
		}
		ratingConsumer := kafka.NewRatingConsumer(
			kafkaBrokers,
			os.Getenv("KAFKA_RATING_TOPIC"),
			os.Getenv("KAFKA_CONSUMER_GROUP"),
			ratingHandler,
		)

		// Start Rating Consumer in goroutine
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go ratingConsumer.Start(ctx)
		defer ratingConsumer.Close()

		log.Println("Started Kafka rating consumer")
	}

	router := gin.New()
	router.Use(gin.Logger())

	// Pass productSvc to routes
	routes.ProductManagerRoutes(router, productSvc)
	routes.UploadRoutes(router)
	routes.ProductUploadRoutes(router)

	// Graceful shutdown
	go func() {
		router.Run(":" + port)
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down product service...")
	log.Println("Product service stopped")
}

func sendInitialProductEvents(svc service.ProductService) {
	log.Println("🔄 Starting to send initial product events to search-service...")

	products, err := svc.GetAllProductForIndex(context.Background())
	if err != nil {
		log.Printf("❌ Error fetching products for initial indexing: %v", err)
		return
	}

	log.Printf("📦 Found %d products to index", len(products))

	for _, product := range products {
		err := kafka.ProduceProductEvent(context.Background(), "INITIAL_SYNC", &product, product.ID)
		if err != nil {
			log.Printf("❌ Error sending product event for product ID %s: %v", product.ID, err)
		} else {
			log.Printf("✅ Sent product event for product ID: %s", product.ID)
		}
	}

	log.Println("✅ Finished sending initial product events to search-service")
}
