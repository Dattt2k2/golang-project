package main

import (
	"log"
	"os"

	"github.com/Dattt2k2/golang-project/auth-service/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)


func main(){
	err := godotenv.Load("github.com/Dattt2k2/golang-project/user-service/.env")
    if err != nil {
        log.Println("Warning: Error loading .env file:", err)
    }
	mongodbURL := os.Getenv("MONGODB_URL")
	if mongodbURL == ""{
		log.Fatalf("MONGODB_URL is not set on .env file yet")
	}

	port := os.Getenv("PORT")
	if port == ""{
		port = "8080"
	}

	router := gin.New()
	router.Use(gin.Logger())

	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	router.Run(":", port)
	
}