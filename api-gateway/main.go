package main

import (
	"fmt"
	"log"

	"github.com/Dattt2k2/golang-project/api-gateway/grpc"
	"github.com/Dattt2k2/golang-project/api-gateway/helpers"
	"github.com/Dattt2k2/golang-project/api-gateway/middleware"
	"github.com/Dattt2k2/golang-project/api-gateway/router"

	// "github.com/Dattt2k2/golang-project/api-gateway/redisdb"
	"github.com/gin-gonic/gin"
)

func main() {
	// redisdb.InitRedis()

	helpers.InitDotEnv()

	ginrouter := gin.Default()
	ginrouter.Use(middleware.CORSMiddleware())
	// ginrouter.Use(middleware.Authenticate())
	// ginrouter.Use(middleware.AuthMiddleware())

	router.SetupRouter(ginrouter)
	grpcClient.InitGrpcClient("localhost:8081")

	port := "8080"
	log.Printf("API gateway is running on port %s", port)

	if err := ginrouter.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}