package main

import (
	// "context"
	"context"
	"os"
	"time"

	// "auth-service/database"
	"auth-service/database"
	"auth-service/helpers"
	"auth-service/kafka"
	"auth-service/logger"
	"auth-service/models"
	"auth-service/repository"
	"auth-service/routes"
	"auth-service/service"

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

	repo := repository.NewUserRepository()
	authSvc := service.NewAuthService(repo)
	go func() {
		reader := kafka.NewKafkaReader("kafka:9092", "user.deleted", "auth-service-group")
		defer reader.Close()
		kafka.ConsumeUserDeleted(reader, func(payload kafka.UserDeletedPayload) {
			logger.Info("Start delete account")
			err := authSvc.DeleteUser(context.Background(), payload.UserID)
			if err != nil {
				logger.Err("Error deleting user sessions for deleted user: "+payload.UserID, err)
			} else {
				logger.Info("Successfully deleted user sessions for deleted user: " + payload.UserID)
			}
		})
	}()

	go func() {
		reader := kafka.NewKafkaReader("kafka:9092", "user.disabled", "auth-service-group")
		defer reader.Close()
		kafka.ConsumeUserDisabled(reader, func(payload kafka.UserDisabledPayload) {
			logger.Info("Start disable account")
			if err := authSvc.UpdateUserDisabled(context.Background(), payload.UserID, payload.IsDisabled); err != nil {
				logger.Err("Failed to update user disabled flag by id for user: "+payload.UserID, err)
				if payload.Email != "" {
					user, err2 := authSvc.GetUserByEmail(context.Background(), payload.Email)
					if err2 != nil {
						logger.Err("Fallback: cannot find user by email in auth DB: "+payload.Email, err2)
						return
					}
					if user != nil {
						if err3 := authSvc.UpdateUserDisabled(context.Background(), user.ID.String(), payload.IsDisabled); err3 != nil {
							logger.Err("Fallback: failed to update disabled flag using found id: "+user.ID.String(), err3)
							return
						}
						logger.Info("Fallback: updated user disabled flag for user (by email->id): " + user.ID.String())
					}
					return
				}
				return
			}
			logger.Info("Updated user disabled flag for user: " + payload.UserID)

			if payload.IsDisabled {
				err := authSvc.LogoutAll(context.Background(), payload.UserID)
				if err != nil {
					logger.Err("Error disabling user sessions for user: "+payload.UserID, err)
				} else {
					logger.Info("Successfully disabled user sessions for user: " + payload.UserID)
				}
			} else {
				logger.Info("User enabled: " + payload.UserID)
			}
		})
	}()

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
