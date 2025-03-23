package database

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis(){
	redisUrl := os.Getenv("REDIS_URL")
	redisUrl = strings.TrimPrefix(redisUrl, "redis://")
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
		log.Printf("Failed to connect to Redis: %v", err)
	} else {
		log.Printf("Connected to Redis")
	}
}

func CloseRedis(){
	if RedisClient != nil{
		if err := RedisClient.Close(); err != nil{
			log.Printf("Error closing Redis connection: %v", err)
		}
	}
}