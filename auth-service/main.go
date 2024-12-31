package main

import (
	// "context"
	"log"
	"os"

	// "github.com/Dattt2k2/golang-project/auth-service/database"
	"github.com/Dattt2k2/golang-project/auth-service/helpers"
	"github.com/Dattt2k2/golang-project/auth-service/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)


func main(){
	err := godotenv.Load("./auth-service/.env")
    if err != nil {
        log.Println("Warning2 : Error loading .env file:", err)
    }
	mongodbURL := os.Getenv("MONGODB_URL")
	if mongodbURL == ""{
		log.Fatalf("MONGODB_URL is not set on .env file yet")
	}

	// mongoClient := database.DBinstance()
    // defer mongoClient.Disconnect(context.Background())

	port := os.Getenv("PORT")
	if port == ""{
		port = "8081"
	}
	helpers.InitDotEnv()

	router := gin.New()
	router.Use(gin.Logger())

	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	router.Run(":"+ port)
	
}