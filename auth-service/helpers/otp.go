package helpers

import (
	"auth-service/database"
	"auth-service/logger"
	"context"
	"errors"
	"math/rand"
	"time"

	"go.uber.org/zap"
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
	// Log OTP and email
	logger.Logger.Infof("Generated OTP for %s: %s", email, otp)
	logger.Logger.Infof("Storing OTP in Redis: key=otp:%s, value=%s", email, otp)
	return otp, nil
}

func GetOTP(email string) (string, error) {
	ctx := context.Background()
	otp, err := database.RedisClient.Get(ctx, "otp:"+email).Result()
	if err != nil {
		logger.Logger.Errorf("Failed to retrieve OTP for %s: %v", email, err)
		return "", err
	}
	// Log OTP and email
	logger.Logger.Infof("Retrieved OTP for %s: %s", email, otp)
	return otp, nil
}

func ResendOTP(email string) (string, error) {
	ctx := context.Background()
	resendKey := "resend_otp:" + email
	count, _ := database.RedisClient.Get(ctx, resendKey).Int()
	if count >= 5 {
		logger.Logger.Warn("Too many OTP resend requests", zap.String("email", email), zap.Int("count", count))
		return "", errors.New("too many OTP resend requests, please try again later")
	}
	database.RedisClient.Incr(ctx, resendKey)
	database.RedisClient.Expire(ctx, resendKey, 10*time.Minute)
	otp, err := GenerateAndStoreOTP(email, 5*time.Minute)
	if err != nil {
		return "", err
	}
	// Log new OTP
	logger.Logger.Infof("Resent OTP for %s: %s", email, otp)
	return otp, nil
}
