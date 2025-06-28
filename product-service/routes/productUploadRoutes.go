package routes

import (
	"github.com/Dattt2k2/golang-project/product-service/controller"
	"github.com/gin-gonic/gin"
)

func ProductUploadRoutes(router *gin.Engine) {
	uploadController := controllers.NewUploadController()
	
	productGroup := router.Group("/products")
	{
		// Workflow 1: Get presigned URL first, then create product with image_path
		productGroup.POST("/upload/presigned-url", uploadController.GetPresignedUploadURL)
		
		// Workflow 2: Create product with image upload in one request
		productGroup.POST("/upload/with-image", uploadController.CreateProductWithImage)
	}
}

// Usage examples:
//
// Workflow 1 - Presigned URL:
// 1. POST /products/upload/presigned-url { "filename": "image.jpg", "content_type": "image/jpeg" }
// 2. Client uploads directly to S3 using presigned URL
// 3. POST /products { "name": "Product", "image_path": "s3_url", ... }
//
// Workflow 2 - Direct upload:
// 1. POST /products/upload/with-image (multipart form with image + product data)
