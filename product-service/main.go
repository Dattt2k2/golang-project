package main

import (
	"log"
	"net"
	"os"

	// database "github.com/Dattt2k2/golang-project/product-service/database"
	"github.com/Dattt2k2/golang-project/product-service/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	// "go.mongodb.org/mongo-driver/mongo"

	// pb "github.com/Dattt2k2/golang-project/module/gRPC-Product/service"
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


	router := gin.New()
	router.Use(gin.Logger())

	routes.ProductManagerRoutes(router)

	router.Run(":" + port)
	
}