package main

import (
	"log"
	"order-service/database"
	"order-service/models"
	// "order-service/repositories"
	"os"

	"order-service/kafka"
	logger "order-service/log"
	"order-service/routes"
	"order-service/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	logger.InitLogger()
	defer logger.Sync()

	// err := godotenv.Load("github.com/Dattt2k2/golang-project/order-service/.env")
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: Error loading .env file")
	}

	// mongodbURL := os.Getenv("MONGODB_URL")
	// if mongodbURL == ""{
	// 	log.Fatalf("MONGODB_URL variable is not set")
	// }
	db := database.InitDB()
	db.AutoMigrate(&models.Order{})

	port := os.Getenv("PORT")

	if port == "" {
		port = "8084"
	}

	router := gin.New()
	router.Use(gin.Logger())

	service.CartServiceConnection()
	service.ProductServiceConnection()

	// kafkaHost := os.Getenv("KAFKA_HOST")
	brokers := []string{"kafka:9092"}
	// brokers := kafkaHost
	kafka.InitOrderSuccessProducer(brokers)
	kafka.InitOrderReturnedProducer(brokers)
	kafka.InitPaymentProducer(brokers)
	// Start payment consumer to listen for payment status updates
	// orderRepo := repositories.NewOrderRepository(db)
	// kafka.StartPaymentConsumer(brokers, orderRepo)

	routes.OrderRoutes(router)

	router.Run(":" + port)

}
