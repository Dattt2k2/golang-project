package main

import (
	"log"
	"net"
	"os"

	// "github.com/Dattt2k2/golang-project/cart-service/database"
	controllers "github.com/Dattt2k2/golang-project/cart-service/controller"
	"github.com/Dattt2k2/golang-project/cart-service/routes"
	pb "github.com/Dattt2k2/golang-project/module/gRPC-cart/service"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	// "go.mongodb.org/mongo-driver/mongo"
)


func main(){
	err := godotenv.Load("github.com/Dattt2k2/golang-project/cart-service/.env")
    if err != nil {
        log.Println("Warning: Error loading .env file:", err)
    }
	mongodbURL := os.Getenv("MONGODB_URL")
	if mongodbURL == ""{
		log.Fatalf("MONGODB_URL is not set on .env file yet")
	}

	controllers.InitProductServiceConnection()

	grpcReady := make(chan bool)

	go func(){
		grpcPort := os.Getenv("GRPC_PORT")
		if grpcPort == ""{
			grpcPort = "8090"
		}
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil{
			log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
		}

		s:= grpc.NewServer()

		pb.RegisterCartServiceServer(s, &controllers.CartServer{})

		grpcReady <- true

		if err := s.Serve(lis); err != nil{
			log.Fatalf("Failed to connect to gRPC Server: %v", err)
		}
	}()

	<-grpcReady

	port := os.Getenv("PORT")
	if port == ""{
		port = "8083"
	}

	// controller.InitProductServiceConnection()
	// controller.InitUserServiceConnection()

	router := gin.New()
	router.Use(gin.Logger())

	routes.CartRoutes(router)

	router.Run(":"+ port)
	
}