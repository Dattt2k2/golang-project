package main

import (
	"log"
	"os"

	database "github.com/Dattt2k2/golang-project/database/databaseConnection.gp"
	routes "github.com/Dattt2k2/golang-project/routes"
	"github.com/gin-gonic/gin"
)	

func main(){
	port := os.Getenv("PORT")

	if port == ""{
		port = "8080"
	}

	mongodbURL := os.Getenv("MONGODB_URL")
	if mongodbURL == "" {
		log.Fatal("MONGODB_URL environment variable not set")
	}

	err := database.InitBucket()
	if err != nil{
		log.Fatal("Failed to initialize GridFS bucket:", err)
	}

	router := gin.New()
	router.Use(gin.Logger())
	

	routes.AuthRoutes(router)
	routes.UserRoutes(router)
	routes.ProductManagerRoutes(router, database.DB)

	router.GET("/api-1", func(c *gin.Context){
		c.JSON(200, gin.H{"success": "Access granted for api-1"})
	})

	router.GET("/api-2", func(c *gin.Context){
		c.JSON(200, gin.H{"success": "Access granted for api-2"})
	})

	router.Run(":" + port)

}