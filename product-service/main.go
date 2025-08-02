package main

import (
	"log"
	"net"
	"os"

	// database "github.com/Dattt2k2/golang-project/product-service/database"
	controllers "product-service/controller"
	"product-service/database"
	"product-service/kafka"
	logger "product-service/log"
	"product-service/repository"
	"product-service/routes"
	service "product-service/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	// "go.mongodb.org/mongo-driver/mongo"

	pb "module/gRPC-Product/service"
)

func main() {

	logger.InitLogger()
	defer logger.Sync()

	err := godotenv.Load("./product-service/.env")
	if err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}
	mongodbURL := os.Getenv("MONGODB_URL")
	if mongodbURL == "" {
		log.Fatalf("MONGODB_URL is not set on .env file yet")
	}

	database.InitRedis()
	defer database.RedisClient.Close()
	log.Printf("Connected to Redis")

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
		repo := repository.NewProductRepository(database.OpenCollection(database.Client, "products"))
		svc := service.NewProductService(repo, service.NewS3Service())
		productServer := controllers.NewProductServer(svc)
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

	routes.SetupProductController()
	productSvc := routes.NewProductService()
	kafkaHost := os.Getenv("KAFKA_URL")
	brokers := []string{kafkaHost}
	if kafkaHost == "" {
		brokers = []string{"localhost:9092"}
	}
	kafka.InitProductEventProducer(brokers)
	// go kafka.ConsumeOrderSuccess(brokers, controllers.ProductController{})
	// go kafka.ConsumerOrderReturned(brokers, controllers.ProductController{})
	go kafka.ConsumeOrderSuccess(brokers, productSvc)
	go kafka.ConsumerOrderReturned(brokers, productSvc)

	router := gin.New()
	router.Use(gin.Logger())

	routes.ProductManagerRoutes(router)
	routes.UploadRoutes(router)

	router.Run(":" + port)

}
