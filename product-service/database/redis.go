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

func InitRedis() {
	redisUrl := os.Getenv("REDIS_URL")
	// Chỉ xử lý URL nếu nó có format "redis://"
	if strings.HasPrefix(redisUrl, "redis://") {
		redisUrl = strings.TrimPrefix(redisUrl, "redis://")
	}

	// Nếu không có URL, sử dụng địa chỉ mặc định
	if redisUrl == "" {
		redisUrl = "redis:6379"
	}

	// Log địa chỉ Redis sẽ kết nối để debug
	log.Printf("Connecting to Redis at: %s", redisUrl)

	// Khởi tạo client với timeout dài hơn
	RedisClient = redis.NewClient(&redis.Options{
		Addr:         redisUrl,
		Password:     "", // Nếu Redis có password, thêm vào đây
		DB:           0,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10, // Số lượng kết nối trong pool
		PoolTimeout:  30 * time.Second,
	})

	// Tạo context với timeout dài hơn
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Kiểm tra kết nối
	pong, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Printf("Failed to connect to Redis: %v", err)
	} else {
		log.Printf("Connected to Redis successfully: %s", pong)
	}
}

func CloseRedis() {
	if RedisClient != nil {
		if err := RedisClient.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		} else {
			log.Printf("Redis connection closed successfully")
		}
	}
}
