package helpers

import (
	"context"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

func generateOTP(length int) string {
	rand.Seed(time.Now().UnixNano())
	digits := "0123456789"
	otp := make([]byte, length)

	for i := range otp {
		otp[i] = digits[rand.Intn(len(digits))]
	}
	return string(otp)
}

func GenerateAndStoreOTP(rdb *redis.Client, email string, expire time.Duration) (string, error) {
	otp := generateOTP(6)
	ctx := context.Background()
	err := rdb.Set(ctx, email, otp, expire).Err()
	if err != nil {
		return "", err
	}
	return otp, nil
}

func GetOTP(rdb *redis.Client, email string) (string, error) {
	ctx := context.Background()
	return rdb.Get(ctx, email).Result()
}
