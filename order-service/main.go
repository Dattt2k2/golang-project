package main

import (
	"log"
	"os"

	"github.com/Dattt2k2/golang-project/order-service/routes"
	"github.com/Dattt2k2/golang-project/order-service/service"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main(){
	err := godotenv.Load("github.com/Dattt2k2/golang-project/order-service/.env")
	if err != nil{
		log.Println("Warning: Error loading .env file")
	}

	mongodbURL := os.Getenv("MONGODB_URL")
	if mongodbURL == ""{
		log.Fatalf("MONGODB_URL variable is not set")
	}

	port := os.Getenv("PORT")

	if port == ""{
		port = "8084"
	}

	router := gin.New()
	router.Use(gin.Logger())

	service.CartServiceConnection()
	service.ProductServiceConnection()

	routes.OrderRotes(router)

	router.Run(":" + port)

	
}