package main

import (
	// "context"
	"os"
	"time"

	// "auth-service/database"
	"auth-service/database"
	"auth-service/helpers"
	"auth-service/logger"
	"auth-service/models"
	"auth-service/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

var userBloomFilter *helpers.BloomFilter

func initBloomFilter(db *gorm.DB) {
	userBloomFilter = helpers.CreateOptimalUserBloomFilter(100000)

	if err := userBloomFilter.Init(db); err != nil {
		logger.Error("Error initializing bloom filter", logger.ErrField(err))
	}

	go func() {
		for {
			time.Sleep(24 * time.Hour)

			if err := userBloomFilter.Init(db); err != nil {
				logger.Error("Error updating bloom filter", logger.ErrField(err))
			} 
		}
	}()
}

func main() {

	logger.InitLogger()
	defer logger.Sync()

	err := godotenv.Load(".env")
	if err != nil {
		logger.Logger.Warn("Warning: Error loading .env file:", err)
	}
	// mongodbURL := os.Getenv("MONGODB_URL")
	// if mongodbURL == "" {
	// 	logger.Logger.Error("MONGODB_URL is not set on .env file yet")
	// }

	database.InitRedis()
	defer database.RedisClient.Close()

	// userCollection := database.OpenCollection(database.Client, "user")
	db := database.InitDB()
	db.AutoMigrate(&models.User{})
	initBloomFilter(db)

	helpers.SetUserBloomFilter(userBloomFilter)

	// mongoClient := database.DBinstance()
	// defer mongoClient.Disconnect(context.Background())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	helpers.InitDotEnv()

	router := gin.New()
	router.Use(gin.Logger())

	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	router.Run(":" + port)

}
