package main

import (
	"context"
	"log"
	"net"
	"os"

	controllers "product-service/controller"
	"product-service/database"
	"product-service/kafka"
	logger "product-service/log"
	"product-service/repository"
	"product-service/routes"
	service "product-service/service"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
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

    uploadDir := "./uploads/images"
    if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
        log.Fatalf("Failed to create upload directory: %v", err)
    }

    files, err := os.ReadDir(uploadDir)
    if err != nil {
        log.Printf("Error reading upload directory: %v", err)
    } else {
        log.Printf("Files in upload directory:")
        for _, file := range files {
            log.Printf("- %s", file.Name())
        }
    }

    // Kafka setup - sử dụng productSvc chung
    kafkaHost := os.Getenv("KAFKA_URL")
    brokers := []string{kafkaHost}
    if kafkaHost == "" {
        brokers = []string{"localhost:9092"}
    }
    kafka.InitProductEventProducer(brokers)
    go kafka.ConsumeOrderSuccess(brokers, productSvc)
    go kafka.ConsumerOrderReturned(brokers, productSvc)

    router := gin.New()
    router.Use(gin.Logger())

    // Pass productSvc to routes
    routes.ProductManagerRoutes(router, productSvc)
    routes.UploadRoutes(router)

    router.Run(":" + port)
}