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
		log.Printf("WARNING: Redis client is not initialized")
		return fmt.Errorf("Redis client is not initialized")
	}

	// Kiểm tra kết nối Redis
	pong, err := database.RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Printf("ERROR: Failed to ping Redis: %v", err)
		return fmt.Errorf("failed to ping Redis: %v", err)
	}
	log.Printf("Redis connection OK: %s", pong)

	// Chuyển đổi dữ liệu sang JSON
	cacheBytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("ERROR: Failed to marshal data for key %s: %v", key, err)
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	// Lưu vào Redis với log chi tiết
	log.Printf("Attempting to cache data with key: %s, TTL: %v", key, duration)
	err = database.RedisClient.Set(ctx, key, string(cacheBytes), duration).Err()
	if err != nil {
		log.Printf("ERROR: Failed to set cache for key %s: %v", key, err)
		return fmt.Errorf("Failed to set cache: %v", err)
	}

	// Kiểm tra xem key đã được lưu chưa
	ttl, err := database.RedisClient.TTL(ctx, key).Result()
	if err != nil {
		log.Printf("WARNING: Failed to check TTL for key %s: %v", key, err)
	} else {
		log.Printf("Successfully cached data with key: %s, TTL: %v", key, ttl)
	}

	// Liệt kê tất cả các keys để debug
	keys, err := database.RedisClient.Keys(ctx, "*").Result()
	if err != nil {
		log.Printf("WARNING: Failed to list keys: %v", err)
	} else {
		log.Printf("Current Redis keys: %v", keys)
	}

	return nil
}

func GetCachedProductData(ctx context.Context, key string, result interface{}) (bool, error) {
	log.Printf("Attempting to get cached data for key: %s", key)

	if database.RedisClient == nil {
		log.Printf("WARNING: Redis client is not initialized")
		return false, fmt.Errorf("Redis client is not initialized")
	}

	// Kiểm tra kết nối Redis
	pong, err := database.RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Printf("ERROR: Failed to ping Redis: %v", err)
		return false, fmt.Errorf("Redis connection failed: %v", err)
	}
	log.Printf("Redis connection OK: %s", pong)

	// Kiểm tra key có tồn tại không
	exists, err := database.RedisClient.Exists(ctx, key).Result()
	if err != nil {
		log.Printf("ERROR: Failed to check key existence: %v", err)
	} else if exists == 0 {
		log.Printf("Key not found in Redis: %s", key)
		return false, nil
	}

	// Lấy dữ liệu từ Redis
	cachedData, err := database.RedisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		log.Printf("Cache miss for key: %s (redis.Nil)", key)
		return false, nil
	} else if err != nil {
		log.Printf("ERROR: Failed to get cache for key %s: %v", key, err)
		return false, fmt.Errorf("failed to get cache: %v", err)
	}

	// Parse dữ liệu JSON
	if err := json.Unmarshal([]byte(cachedData), result); err != nil {
		log.Printf("ERROR: Failed to unmarshal cached data for key %s: %v", key, err)
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
	log.Printf("Preparing to cache all products for page=%d, limit=%d", page, limit)

	// Kiểm tra nếu danh sách sản phẩm trống
	if len(products) == 0 {
		log.Printf("WARNING: Skipping cache because product list is empty")
		return nil
	}

	// Log số lượng sản phẩm để debug
	log.Printf("Caching %d products with key: %s", len(products), cacheKey)

	cacheData := ProductsPageResult{
		Products: products,
		Total:    total,
		Pages:    pages,
		HasNext:  hasNext,
		HasPrev:  hasPrev,
	}

	// Thay đổi TTL từ 10 phút thành 1 giờ
	return CacheProductData(ctx, cacheKey, cacheData, 1*time.Hour)
}

type ProductsPageResult struct {
	Products []models.Product `json:"products"`
	Total    int64            `json:"total"`
	Pages    int              `json:"pages"`
	HasNext  bool             `json:"has_next"`
	HasPrev  bool             `json:"has_prev"`
}
