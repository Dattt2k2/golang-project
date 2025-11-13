package main

import (
	"log"
	"net"
	"order-service/database"
	"order-service/models"
	"order-service/repositories"

	// "order-service/repositories"
	"os"

	pb "module/gRPC-Order/service"
	"order-service/kafka"
	logger "order-service/log"
	"order-service/routes"
	"order-service/service"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
)

func main() {

	logger.InitLogger()
	defer logger.Sync()

	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: Error loading .env file")
	}

	db := database.InitDB()
	db.AutoMigrate(&models.Order{})

	port := os.Getenv("PORT")

	if port == "" {
		port = "8084"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "8100"
	}

	keepAliveParams := keepalive.ServerParameters{
		MaxConnectionIdle:     15 * time.Minute, // Maximum idle time for a connection
		MaxConnectionAge:      30 * time.Minute, // Maximum age of a connection
		MaxConnectionAgeGrace: 5 * time.Minute,  // Grace period for closing connections
		Time:                  5 * time.Minute,  // Frequency of server pings
		Timeout:               20 * time.Second, // Timeout for client responses
	}

	// Enforcement policy to prevent resource exhaustion
	keepAliveEnforcementPolicy := keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second, // Minimum time a client should wait before sending a keepalive ping
		PermitWithoutStream: true,            // Allow pings even when there are no active streams
	}

	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(keepAliveParams),
		grpc.KeepaliveEnforcementPolicy(keepAliveEnforcementPolicy),
		grpc.MaxConcurrentStreams(1000),   // Increase max concurrent streams
		grpc.MaxRecvMsgSize(10*1024*1024), // 10MB max receive message size
		grpc.MaxSendMsgSize(10*1024*1024), // 10MB max send message size
		grpc.NumStreamWorkers(100),        // Increase number of workers
		// Add interceptors for rate limiting and logging
		grpc.UnaryInterceptor(service.UnaryServerInterceptor()),
		grpc.StreamInterceptor(service.StreamServerInterceptor()),
	)

	service.CartServiceConnection()
	service.ProductServiceConnection()

	// Initialize order repository and service
	orderRepo := repositories.NewOrderRepository(db)
	orderServiceGRPC := &service.OrderServiceServer{
		OrderRepo: orderRepo,
	}

	// Register the actual implementation instead of UnimplementedOrderServiceServer
	pb.RegisterOrderServiceServer(grpcServer, orderServiceGRPC)

	// Initialize and register health check service
	healthServer := service.InitHealthCheck()
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	// Start health monitoring in background
	go service.MonitorServiceHealth(healthServer)

	go func() {
		listener, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
		}
		log.Printf("gRPC server is running on port %s...", grpcPort)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v", err)
		}
	}()

	// kafkaHost := os.Getenv("KAFKA_HOST")
	brokers := []string{"kafka:9092"}
	// brokers := kafkaHost
	kafka.InitOrderSuccessProducer(brokers)
	kafka.InitOrderReturnedProducer(brokers)
	kafka.InitPaymentProducer(brokers)
	// Start payment consumer to listen for payment status updates
	orderService := service.NewOrderService(orderRepo)
	kafka.StartPaymentConsumer(brokers, orderRepo, orderService)

	router := gin.Default()
	routes.OrderRoutes(router)

	router.Run(":" + port)

}

// package main

// import (
// 	"log"
// 	pb "module/gRPC-Order/service"
// 	"net"
// 	"order-service/database"
// 	logger "order-service/log"
// 	"order-service/models"
// 	"order-service/repositories"
// 	"order-service/routes"
// 	"order-service/service"
// 	"os"
// 	"strings"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/joho/godotenv"
// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/keepalive"
// )

// func main() {
// 	// Initialize logger
// 	logger.InitLogger()
// 	defer logger.Sync()

// 	// Load environment variables
// 	err := godotenv.Load(".env")
// 	if err != nil {
// 		log.Println("Warning: Error loading .env file")
// 	}

// 	// Initialize database
// 	db := database.InitDB()
// 	db.AutoMigrate(&models.Order{})

// 	// Get port from environment variables
// 	port := os.Getenv("PORT")
// 	if port == "" {
// 		port = "8084"
// 	}
// 	grpcPort := os.Getenv("GRPC_PORT")
// 	if grpcPort == "" {
// 		grpcPort = "8100"
// 	}
// 	keepAliveParams := keepalive.ServerParameters{
// 		MaxConnectionIdle:     15 * time.Minute, // Maximum idle time for a connection
// 		MaxConnectionAge:      30 * time.Minute, // Maximum age of a connection
// 		MaxConnectionAgeGrace: 5 * time.Minute,  // Grace period for closing connections
// 		Time:                  5 * time.Minute,  // Frequency of server pings
// 		Timeout:               20 * time.Second, // Timeout for client responses
// 	}

// 	grpcServer := grpc.NewServer(
// 		grpc.KeepaliveParams(keepAliveParams),
// 	)

// 	orderRepo := repositories.NewOrderRepository(db)
// 	orderServiceGRPC := &service.OrderServiceServer{
// 		OrderRepo: orderRepo,
// 	}
// 	pb.RegisterOrderServiceServer(grpcServer, orderServiceGRPC)

// 	// Start gRPC server in a separate goroutine
// 	go func() {
// 		listener, err := net.Listen("tcp", ":"+grpcPort)
// 		if err != nil {
// 			log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
// 		}
// 		log.Printf("gRPC server is running on port %s...", grpcPort)
// 		if err := grpcServer.Serve(listener); err != nil {
// 			log.Fatalf("Failed to serve gRPC server: %v", err)
// 		}
// 	}()

// 	// Initialize Kafka consumer
// 	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
// 	if kafkaBrokers == "" {
// 		kafkaBrokers = "kafka:9092" // Default broker
// 	}
// 	brokers := strings.Split(kafkaBrokers, ",")
// 	topic := "payment_events"
// 	groupID := "order-service-group"

// 	orderService := service.NewOrderService(orderRepo)
// 	orderService.StartKafkaConsumer(brokers, topic, groupID)

// 	// Initialize HTTP server (Gin)
// 	router := gin.New()
// 	router.Use(gin.Logger())

// 	// Initialize connections to other services
// 	service.CartServiceConnection()
// 	service.ProductServiceConnection()

// 	// Register HTTP routes
// 	routes.OrderRoutes(router)

// 	// Start HTTP server
// 	log.Printf("HTTP server is running on port %s...", port)
// 	if err := router.Run(":" + port); err != nil {
// 		log.Fatalf("Failed to run HTTP server: %v", err)
// 	}
// }
