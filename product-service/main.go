package main

import (
	"log"
	"os"

	database "github.com/Dattt2k2/golang-project/product-service/database"
	"github.com/Dattt2k2/golang-project/product-service/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	// "go.mongodb.org/mongo-driver/mongo"
)


func main(){
	err := godotenv.Load("github.com/Dattt2k2/golang-project/product-service/.env")
    if err != nil {
        log.Println("Warning: Error loading .env file:", err)
    }
	mongodbURL := os.Getenv("MONGODB_URL")
	if mongodbURL == ""{
		log.Fatalf("MONGODB_URL is not set on .env file yet")
	}

	port := os.Getenv("PORT")
	if port == ""{
		port = "8082"
	}

	// controller.InitUserServiceConnection()

	router := gin.New()
	router.Use(gin.Logger())

	routes.ProductManagerRoutes(router, database.DB)

	router.Run(":" + port)
	
}