package helpers

import (
	"auth-service/database"
	"context"
	"errors"
	"math/rand"
	"time"
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

// Hàm này sẽ được gọi từ service, không cần truyền rdb
func GenerateOTP(email string) (string, error) {
	return GenerateAndStoreOTP(email, 5*time.Minute)
}

func GenerateAndStoreOTP(email string, expire time.Duration) (string, error) {
	otp := generateOTP(6)
	ctx := context.Background()
	err := database.RedisClient.Set(ctx, "otp:"+email, otp, expire).Err()
	if err != nil {
		return "", err
	}
	return otp, nil
}

func GetOTP(email string) (string, error) {
	ctx := context.Background()
	return database.RedisClient.Get(ctx, "otp:"+email).Result()
}

func ResendOTP(email string) (string, error) {
	ctx := context.Background()
	resendKey := "resend_otp:" + email
	count, _ := database.RedisClient.Get(ctx, resendKey).Int()
	if count >= 5 {
		return "", errors.New("too many OTP resend requests, please try again later")
	}
	database.RedisClient.Incr(ctx, resendKey)
	database.RedisClient.Expire(ctx, resendKey, 10*time.Minute)
	otp, err := GenerateAndStoreOTP(email, 5*time.Minute)
	if err != nil {
		return "", err
	}
	return otp, nil
}
