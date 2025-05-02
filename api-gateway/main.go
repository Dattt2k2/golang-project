package main

import (
	"fmt"

	"github.com/Dattt2k2/golang-project/api-gateway/grpc"
	"github.com/Dattt2k2/golang-project/api-gateway/helpers"
	"github.com/Dattt2k2/golang-project/api-gateway/logger"
	"github.com/Dattt2k2/golang-project/api-gateway/middleware"
	"github.com/Dattt2k2/golang-project/api-gateway/router"

	// "github.com/Dattt2k2/golang-project/api-gateway/redisdb"
	"github.com/gin-gonic/gin"
)

func main() {
	// redisdb.InitRedis()

	helpers.InitDotEnv()

	logger.InitLogger()
	defer logger.Sync()

	ginrouter := gin.Default()
	ginrouter.Use(middleware.CORSMiddleware())
	// ginrouter.Use(middleware.Authenticate())
	// ginrouter.Use(middleware.AuthMiddleware())

	router.SetupRouter(ginrouter)
	grpcClient.InitGrpcClient("localhost:8081")

	port := "8080"
	logger.Info(fmt.Sprintf("Starting API Gateway on port: %s", port))

	if err := ginrouter.Run(fmt.Sprintf(":%s", port)); err != nil {
		logger.Err("Failed to start API Gateway", err)
	}
}