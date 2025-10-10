package redisdb

import (
	"context"
	"fmt"

	cfg "api-gateway/config"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var Ctx = context.Background()

var RedisNil = redis.Nil

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     cfg.Get("REDIS_ADDR", "redis:6379"),
		Password: cfg.Get("REDIS_PASSWORD", ""),
		DB:       cfg.GetInt("REDIS_DB", 0),
	})

	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	}

	fmt.Println("Connected to Redis")
}
