package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Dattt2k2/golang-project/product-service/database"
	"github.com/Dattt2k2/golang-project/product-service/models"
	"github.com/redis/go-redis/v9"
)

func CacheProductData(ctx context.Context, key string, data interface{}, duration time.Duration) error {
	if database.RedisClient == nil {
		return fmt.Errorf("Redis client is not initialized")
	}

	if _, err := database.RedisClient.Ping(ctx).Result(); err != nil {
		return fmt.Errorf("failed to ping Redis: %v", err)
	}

	cacheBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Errorf("failed to marshal data: %v", err)
	}
	err = database.RedisClient.Set(ctx, key, string(cacheBytes), duration).Err()
	if err != nil {
		return fmt.Errorf("Failed to set cache: %v", err)
	}

	log.Printf("Successfully cached data with key: %s", key)
	return nil 
}

func GetCachedProductData(ctx context.Context, key string, result interface{}) (bool, error ) {
	if database.RedisClient == nil {
		return false, fmt.Errorf("Redis client is not initialized")
	}
	
	if _, err := database.RedisClient.Ping(ctx).Result(); err != nil {
		return false, fmt.Errorf("Redis connecton failed: %v", err)
	}

	cachedData, err := database.RedisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil 
	} else if err != nil {
		return false, fmt.Errorf("failed to get cache: %v", err)
	}

	if err := json.Unmarshal([]byte(cachedData), result); err != nil {
		return false, fmt.Errorf("failed to unmarshal cached data: %v", err)
	}

	log.Printf("Cache hit for key: %s", key)
	return true, nil 
}

func InvalidateProductCache(ctx context.Context, pattern string) error {
    if database.RedisClient == nil {
        return fmt.Errorf("redis client is not initialized")
    }

    var cursor uint64
    var keys []string
    var err error

    for {
        var scanKeys []string
		scanKeys, cursor, err = database.RedisClient.Scan(ctx, cursor, pattern, 100).Result()
        if err != nil {
            return fmt.Errorf("error scanning Redis keys: %w", err)
        }
        
        keys = append(keys, scanKeys...)
        
        if cursor == 0 {
            break
        }
    }
    
    if len(keys) > 0 {
        err := database.RedisClient.Del(ctx, keys...).Err()
        if err != nil {
            return fmt.Errorf("error invalidating cache: %w", err)
        }
        log.Printf("Successfully invalidated %d cached keys with pattern %s", len(keys), pattern)
    }
    
    return nil
}

func GetAllProductsFromCache(ctx context.Context, page, limit int64) (*ProductsPageResult, bool, error) {
	cacheKey := fmt.Sprintf("products:page=%d&limit=%d", page, limit)
	
	var cachedResult ProductsPageResult
	found, err := GetCachedProductData(ctx, cacheKey, &cachedResult)
	if err != nil {
		log.Printf("Error getting cached data: %v", err)
		return nil, false, err 
	}

	if !found {
		return nil, false, nil 
	}

	return &cachedResult, true, nil
}

func CacheAllProducts(ctx context.Context, page, limit int64, products []models.Product, total int64, pages int, hasNext, hasPrev bool) error {
	cacheKey := fmt.Sprintf("products:page=%d&limit=%d", page, limit)

	cacheData := ProductsPageResult {
		Products: products,
		Total:   total,
		Pages: pages,
		HasNext: hasNext,
		HasPrev: hasPrev,
	}
	return CacheProductData(ctx, cacheKey, cacheData, 10*time.Minute)
}


type ProductsPageResult struct {
	Products []models.Product `json:"products"`
	Total    int64           `json:"total"`
	Pages    int             `json:"pages"`
	HasNext  bool            `json:"has_next"`
	HasPrev  bool            `json:"has_prev"`
}