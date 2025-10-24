// package main

// import (
// 	"log"
// 	"net"
// 	"order-service/database"
// 	"order-service/models"

// 	// "order-service/repositories"
// 	"os"

// 	pb "module/gRPC-Order/service"
// 	"order-service/kafka"
// 	logger "order-service/log"
// 	"order-service/routes"
// 	"order-service/service"

// 	"github.com/gin-gonic/gin"
// 	"github.com/joho/godotenv"
// 	"google.golang.org/grpc"
// )

// func main() {

// 	logger.InitLogger()
// 	defer logger.Sync()

// 	err := godotenv.Load(".env")
// 	if err != nil {
// 		log.Println("Warning: Error loading .env file")
// 	}

// 	db := database.InitDB()
// 	db.AutoMigrate(&models.Order{})

// 	port := os.Getenv("PORT")

// 	if port == "" {
// 		port = "8084"
// 	}

// 	router := gin.New()
// 	router.Use(gin.Logger())

// 	service.CartServiceConnection()
// 	service.ProductServiceConnection()

// 	grpcServer := grpc.NewServer()

// 	pb.RegisterOrderServcieServer(grpcServer, &pb.UnimplementedOrderServcieServer{})
// 	 go func() {
//         listener, err := net.Listen("tcp", ":"+port)
//         if err != nil {
//             log.Fatalf("Failed to listen on port %s: %v", port, err)
//         }
//         log.Printf("gRPC server is running on port %s...", port)
//         if err := grpcServer.Serve(listener); err != nil {
//             log.Fatalf("Failed to serve gRPC server: %v", err)
//         }
//     }()

// 	// kafkaHost := os.Getenv("KAFKA_HOST")
// 	brokers := []string{"kafka:9092"}
// 	// brokers := kafkaHost
// 	kafka.InitOrderSuccessProducer(brokers)
// 	kafka.InitOrderReturnedProducer(brokers)
// 	kafka.InitPaymentProducer(brokers)
// 	// Start payment consumer to listen for payment status updates
// 	// orderRepo := repositories.NewOrderRepository(db)
// 	// kafka.StartPaymentConsumer(brokers, orderRepo)

// 	routes.OrderRoutes(router)

// 	router.Run(":" + port)

// }

package main

import (
	"log"
	pb "module/gRPC-Order/service"
	"net"
	"order-service/database"
	logger "order-service/log"
	"order-service/models"
	"order-service/repositories"
	"order-service/routes"
	"order-service/service"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	// Initialize logger
	logger.InitLogger()
	defer logger.Sync()

	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: Error loading .env file")
	}

	// Initialize database
	db := database.InitDB()
	db.AutoMigrate(&models.Order{})

	// Get port from environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "8100"
	}
	grpcServer := grpc.NewServer()
	orderRepo := repositories.NewOrderRepository(db)
	orderServiceGRPC := &service.OrderServiceServer{
		OrderRepo: orderRepo,
	}
	pb.RegisterOrderServiceServer(grpcServer, orderServiceGRPC)

	// Start gRPC server in a separate goroutine
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

	// Initialize HTTP server (Gin)
	router := gin.New()
	router.Use(gin.Logger())

	// Initialize connections to other services
	service.CartServiceConnection()
	service.ProductServiceConnection()

	// Register HTTP routes
	routes.OrderRoutes(router)

	// Start HTTP server
	log.Printf("HTTP server is running on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to run HTTP server: %v", err)
	}
}
