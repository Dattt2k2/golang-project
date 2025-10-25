package database

import (
	"context"
	"os"
	"time"

	"auth-service/logger"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis(){
	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == ""{
		redisUrl = "redis:6379"
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr: redisUrl,
		Password: "",
		DB: 0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := RedisClient.Ping(ctx).Result(); err != nil{
		logger.Err("Failed to connect to Redis", err)
	}
}

func CloseRedis(){
	if RedisClient != nil{
		if err := RedisClient.Close(); err != nil{
			logger.Err("Error closing Redis connection", err)
		}
	}
}