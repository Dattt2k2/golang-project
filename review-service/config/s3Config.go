package config

import (
	"os"
	"strconv"
)

type S3Config struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Folder          string
	Endpoint        string
	CloudFrontURL   string
	MaxFileSize     int64
	AllowedExts     []string
}

func LoadS3Config() *S3Config {
	maxFileSize, _ := strconv.ParseInt(getEnv("MAX_FILE_SIZE", "10485760"), 10, 64) // Default 10MB

	return &S3Config{
		Region:          getEnv("AWS_REGION", "ap-southeast-1"),
		AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
		SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		BucketName:      getEnv("AWS_S3_BUCKET", ""),
		Folder:          getEnv("AWS_S3_FOLDER", "uploads"),
		Endpoint:        getEnv("S3_ENDPOINT", ""),
		CloudFrontURL:   getEnv("CLOUDFRONT_URL", ""),
		MaxFileSize:     maxFileSize,
		AllowedExts:     []string{"jpg", "jpeg", "png", "gif", "webp"},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
