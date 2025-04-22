package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Dattt2k2/golang-project/cart-service/kafka"
	"github.com/Dattt2k2/golang-project/cart-service/routes"
	pb "github.com/Dattt2k2/golang-project/module/gRPC-cart/service"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	// Khởi tạo router
	router := gin.Default()

	// Thiết lập dependencies theo mô hình 3 layer
	cartController, cartServer := routes.SetupCartDependencies()

	// Thiết lập HTTP routes
	routes.CartRoutes(router, cartController)

	// Thiết lập gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterCartServiceServer(grpcServer, cartServer)

	// Khởi động gRPC server
	lis, err := net.Listen("tcp", ":8089")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	go func() {
		log.Println("Starting gRPC server at :8089")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Khởi động Kafka consumer
	kafkaBrokers := []string{"kafka:9092"} // Hoặc đọc từ config
	kafka.ConsumeOrderSuccess(kafkaBrokers, cartController)

	// Xử lý tín hiệu để tắt server một cách graceful
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Khởi động HTTP server
	go func() {
		log.Println("Starting HTTP server at :8088")
		if err := router.Run(":8088"); err != nil {
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
