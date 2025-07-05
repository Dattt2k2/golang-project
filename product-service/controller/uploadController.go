package controllers

import (
	"net/http"

	"product-service/models"
	"product-service/service"
	"github.com/gin-gonic/gin"
)

type UploadController struct {
	s3Service *service.S3Service
}

func NewUploadController() *UploadController {
	return &UploadController{
		s3Service: service.NewS3Service(),
	}
}

// GetPresignedUploadURL - API để lấy presigned URL cho upload
func (ctrl *UploadController) GetPresignedUploadURL(c *gin.Context) {
	var req models.PresignedUploadRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	presignedURL, publicURL, err := ctrl.s3Service.GeneratePresignedUploadURL(req.Filename, req.ContentType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to generate presigned URL",
			"details": err.Error(),
		})
		return
	}

	response := models.PresignedUploadResponse{
		PresignedURL: presignedURL,
		PublicURL:    publicURL,
		ExpiresIn:    900, // 15 minutes
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// Example usage for direct file upload (optional - keep for compatibility)
func (ctrl *UploadController) UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No file provided",
		})
		return
	}
	defer file.Close()

	imageURL, err := ctrl.s3Service.UploadFile(file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to upload file",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"image_url": imageURL,
	})
}

// CreateProductWithImage - Workflow 2: Tạo product và upload ảnh cùng lúc
func (ctrl *UploadController) CreateProductWithImage(c *gin.Context) {
	// Parse multipart form
	err := c.Request.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse form data",
		})
		return
	}

	// Get product data from form
	name := c.PostForm("name")
	category := c.PostForm("category")
	description := c.PostForm("description")

	if name == "" || category == "" || description == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required fields: name, category, description",
		})
		return
	}

	// Parse numeric fields
	quantity := c.PostForm("quantity")
	price := c.PostForm("price")

	if quantity == "" || price == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required fields: quantity, price",
		})
		return
	}

	// Handle file upload
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Image file is required",
		})
		return
	}
	defer file.Close()

	// Upload image to S3
	imageURL, err := ctrl.s3Service.UploadFile(file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to upload image",
			"details": err.Error(),
		})
		return
	}

	// Return product data with image URL - client can use this to create product
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Image uploaded successfully",
		"data": gin.H{
			"image_url":   imageURL,
			"name":        name,
			"category":    category,
			"description": description,
			"quantity":    quantity,
			"price":       price,
		},
		"next_step": "Use the image_url to create product via POST /products",
	})
}
