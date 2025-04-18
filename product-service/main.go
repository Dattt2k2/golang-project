package main

import (
	"log"
	"net"
	"os"

	// database "github.com/Dattt2k2/golang-project/product-service/database"
	"github.com/Dattt2k2/golang-project/auth-service/database"
	controllers "github.com/Dattt2k2/golang-project/product-service/controller"
	"github.com/Dattt2k2/golang-project/product-service/kafka"
	"github.com/Dattt2k2/golang-project/product-service/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	// "go.mongodb.org/mongo-driver/mongo"

	pb "github.com/Dattt2k2/golang-project/module/gRPC-Product/service"
)


func main(){
	err := godotenv.Load("./product-service/.env")
    if err != nil {
        log.Println("Warning: Error loading .env file:", err)
    }
	mongodbURL := os.Getenv("MONGODB_URL")
	if mongodbURL == ""{
		log.Fatalf("MONGODB_URL is not set on .env file yet")
	}

	database.InitRedis()
	defer database.RedisClient.Close()
	log.Printf("Connected to Redis")

	grpcReady := make(chan bool)

	go func(){
		grpcPort := os.Getenv("gRPC_PORT")
		if grpcPort == ""{
			grpcPort = "8089"
		}
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil{
			log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
		}

		s:= grpc.NewServer()
		
		pb.RegisterProductServiceServer(s, &controllers.ProductServer{})

		grpcReady <- true

		if err := s.Serve(lis); err != nil{
			log.Fatalf("Failed to connect to gRPC Server: %v", err)
		}
	}()

	<-grpcReady

	port := os.Getenv("PORT")
	if port == ""{
		port = "8082`"
	}


	uploadDir := "./uploads/images"
    if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
        log.Fatalf("Failed to create upload directory: %v", err)
    }
    
    // List all files in the directory
    files, err := os.ReadDir(uploadDir)
    if err != nil {
        log.Printf("Error reading upload directory: %v", err)
    } else {
        log.Printf("Files in upload directory:")
        for _, file := range files {
            log.Printf("- %s", file.Name())
        }
    }
	// controller.InitUserServiceConnection()

	// grpcPort := os.Getenv("gRPC_PORT")
	// if grpcPort == ""{
	// 	grpcPort = "8089"
	// }
	// lis, err := net.Listen("tcp", ":"+grpcPort)
	// if err != nil{
	// 	log.Fatalf("Failed to listen on port 8089: %v", err)
	// }

	// grpcServer := grpc.NewServer()

	// pb.RegisterProductServiceServer(grpcServer, &pb.UnimplementedProductServiceServer{})

	// log.Printf("gRPC server is running on port: %v", grpcPort)
	// if err := grpcServer.Serve(lis); err != nil{
	// 	log.Fatalf("Failed to serve gRPC server : %v", err)
	// }

	brokers := []string{"kafka:9092"}
	go kafka.ConsumeOrderSuccess(brokers, controllers.ProductController{})

	router := gin.New()
	router.Use(gin.Logger())

	routes.ProductManagerRoutes(router)

	router.Run(":" + port)
	
}