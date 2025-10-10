package main

import (
	"fmt"
	"time"

	"api-gateway/helpers"
	"api-gateway/logger"
	"api-gateway/middleware"
	"api-gateway/redisdb"
	"api-gateway/router"

	// "github.com/Dattt2k2/golang-project/api-gateway/redisdb"
	// "github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	cfg "api-gateway/config"
)

func main() {
	// redisdb.InitRedis()
	logger.InitLogger()
	defer logger.Sync()

	helpers.InitDotEnv()
	_ = cfg.Load(".env")

	redisdb.InitRedis()
	defer redisdb.RedisClient.Close()
	logger.Info("Connected to Redis")

	ginrouter := gin.Default()
	ginrouter.Use(middleware.CORSMiddleware())
	ginrouter.Use(middleware.RateLimitMiddleware(redisdb.RedisClient, 100, 20*time.Second, "rate_limit"))
	// ginrouter.Use(cors.New(cors.Config{
	// 	AllowOrigins:     []string{"*"},
	// 	AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	// 	AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
	// 	AllowCredentials: true,
	// }))
	// ginrouter.Use(middleware.Authenticate())
	// ginrouter.Use(middleware.AuthMiddleware())

	router.SetupRouter(ginrouter)

	port := "8080"
	logger.Info(fmt.Sprintf("Starting API Gateway on port: %s", port))

	if err := ginrouter.Run(fmt.Sprintf(":%s", port)); err != nil {
		logger.Err("Failed to start API Gateway", err)
	}
}
