package controllers

import (
	"auth-service/database"
	"auth-service/helpers"
	"auth-service/kafka"
	"auth-service/models"
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var (
	redisClient = redis.NewClient(&redis.Options{Addr: os.Getenv("REDIS_URL")})
	kafkaWriter = kafka.NewKafkaWriter(os.Getenv("KAFKA_BROKER"), "email-topic")
)

type OTPRequest struct {
	Email   string `json:"email" binding:"required,email"`
	OTPCode string `json:"otp_code,omitempty"`
}

func SendOTPHander(email, template string) error {
	expireTime := 3
	otpCode, err := helpers.GenerateAndStoreOTP(email, time.Duration(expireTime)*time.Minute)
	if err != nil {
		return err
	}
	data := make(map[string]interface{})

	data["OTP"] = otpCode
	data["ExpireIn"] = expireTime
	data["Email"] = email

	msg := kafka.EmailMessage{
		To:       email,
		Subject:  "Your OTP Code",
		Template: template,
		Data:     data,
	}
	return kafka.SendEmailMessage(kafkaWriter, msg)
}

func VerifyOTPHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req OTPRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
            return
        }

        otpCode, err := helpers.GetOTP(req.Email)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "Failed to retrieve OTP code"})
            return
        }

        if otpCode != req.OTPCode {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "Invalid OTP code",
            })
            return
        }

        // Update user record in Postgres (idempotent)
        ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
        defer cancel()

        db := database.DB.WithContext(ctx)
        if db == nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
            return
        }

        if err := db.Model(&models.User{}).
            Where("email = ?", req.Email).
            Update("is_verify", true).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user verification"})
            return
        }

        redisClient.Del(ctx, "otp:"+req.Email)
        c.JSON(http.StatusOK, gin.H{
            "message": "OTP code verified successfully",
        })
	}
}

func ResendOTPHander() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		var req OTPRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		resendKey := "resend_otp:" + req.Email
		count, _ := redisClient.Get(ctx, resendKey).Int()
		if count >= 5 {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests, please try again later"})
			return
		}

		redisClient.Incr(ctx, resendKey)
		redisClient.Expire(ctx, resendKey, 10*time.Minute)

		otpCode, err := helpers.GetOTP(req.Email)
		if err != nil {
			otpCode, err = helpers.GenerateAndStoreOTP(req.Email, 3*time.Minute)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP"})
				return
			}
		}

		data := map[string]interface{}{
			"otp_code":    otpCode,
			"expire_time": 3,
			"email":       req.Email,
		}

		msg := kafka.EmailMessage{
			To:       req.Email,
			Subject:  "Your OTP Code",
			Template: "otp_template.html",
			Data:     data,
		}

		err = kafka.SendEmailMessage(kafkaWriter, msg)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP email"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "OTP code resent successfully"})
	}
}
