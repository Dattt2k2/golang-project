package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"product-service/config"
	s3Client "product-service/s3"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Service struct {
	client *s3.S3
	config *config.S3Config
}

func NewS3Service() *S3Service {
	cfg := config.LoadS3Config()

	client := s3Client.NewS3Client(cfg)
	return &S3Service{
		client: client,
		config: cfg,
	}
}

func (s *S3Service) UploadFile(file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	// Validate file extension
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	ext = strings.TrimPrefix(ext, ".")

	isAllowed := false
	for _, allowedExt := range s.config.AllowedExts {
		if ext == allowedExt {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return "", fmt.Errorf("file extension .%s is not allowed", ext)
	}

	// Validate file size
	if fileHeader.Size > s.config.MaxFileSize {
		return "", fmt.Errorf("file size exceeds maximum limit of %d bytes", s.config.MaxFileSize)
	}

	// Generate unique filename
	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf("%d_%s", timestamp, fileHeader.Filename)
	key := fmt.Sprintf("%s/%s", s.config.Folder, filename)

	// Read file content
	buffer := make([]byte, fileHeader.Size)
	_, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	// Upload to S3
	_, err = s.client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s.config.BucketName),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(fileHeader.Header.Get("Content-Type")),
		ACL:         aws.String("public-read"), // Make file publicly accessible
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %v", err)
	}

	// Return URL
	if s.config.CloudFrontURL != "" {
		return fmt.Sprintf("%s/%s", s.config.CloudFrontURL, key), nil
	}

	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.config.BucketName, s.config.Region, key), nil
}

func (s *S3Service) DeleteFile(fileURL string) error {
	// Extract key from URL
	var key string
	if s.config.CloudFrontURL != "" && strings.Contains(fileURL, s.config.CloudFrontURL) {
		key = strings.TrimPrefix(fileURL, s.config.CloudFrontURL+"/")
	} else {
		// Parse S3 URL
		parts := strings.Split(fileURL, ".amazonaws.com/")
		if len(parts) == 2 {
			key = parts[1]
		} else {
			return fmt.Errorf("invalid S3 URL format")
		}
	}

	_, err := s.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	})

	return err
}

func (s *S3Service) getContentType(ext string) string {
	contentTypes := map[string]string{
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"png":  "image/png",
		"gif":  "image/gif",
		"webp": "image/webp",
	}

	if contentType, exists := contentTypes[ext]; exists {
		return contentType
	}

	return "application/octet-stream"
}

// GeneratePresignedUploadURL - Tạo presigned URL để upload trực tiếp lên S3
// func (s *S3Service) GeneratePresignedUploadURL(filename string, contentType string) (string, string, error) {
// 	// Validate file extension
// 	ext := strings.ToLower(filepath.Ext(filename))
// 	ext = strings.TrimPrefix(ext, ".")

// 	isAllowed := false
// 	for _, allowedExt := range s.config.AllowedExts {
// 		if ext == allowedExt {
// 			isAllowed = true
// 			break
// 		}
// 	}

// 	if !isAllowed {
// 		return "", "", fmt.Errorf("file extension .%s is not allowed", ext)
// 	}

// 	// Generate unique filename
// 	timestamp := time.Now().UnixNano()
// 	uniqueFilename := fmt.Sprintf("%d_%s", timestamp, filename)
// 	key := fmt.Sprintf("%s/%s", s.config.Folder, uniqueFilename)

// 	// Create presigned PUT request
// 	req, _ := s.client.PutObjectRequest(&s3.PutObjectInput{
// 		Bucket:      aws.String(s.config.BucketName),
// 		Key:         aws.String(key),
// 		ContentType: aws.String(contentType),
// 		ACL:         aws.String("public-read"),
// 	})

// 	// Generate presigned URL (valid for 15 minutes)
// 	presignedURL, err := req.Presign(15 * time.Minute)
// 	if err != nil {
// 		return "", "", fmt.Errorf("failed to generate presigned URL: %v", err)
// 	}

// 	// Generate final public URL
// 	var publicURL string
// 	if s.config.CloudFrontURL != "" {
// 		publicURL = fmt.Sprintf("%s/%s", s.config.CloudFrontURL, key)
// 	} else {
// 		publicURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.config.BucketName, s.config.Region, key)
// 	}

// 	return presignedURL, publicURL, nil
// }

func (s *S3Service) GeneratePresignedUploadURL(filename, contentType string) (string, string, error) {
	req, _ := s.client.PutObjectRequest(&s3.PutObjectInput{
		Bucket:      aws.String(s.config.BucketName),
		Key:         aws.String(filename),
		ContentType: aws.String(contentType),
		// ACL:         aws.String("public-read"),
	})
	urlStr, err := req.Presign(15 * time.Minute)
	if err != nil {
		return "", "", err
	}
	publicURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.config.BucketName, s.config.Region, filename)
	return urlStr, publicURL, nil
}

// GeneratePresignedDownloadURL - Tạo presigned URL để download (cho private files)
func (s *S3Service) GeneratePresignedDownloadURL(key string, expiration time.Duration) (string, error) {
	req, _ := s.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	})

	presignedURL, err := req.Presign(expiration)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned download URL: %v", err)
	}

	return presignedURL, nil
}
