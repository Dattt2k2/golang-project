package routes

import (
	"github.com/Dattt2k2/golang-project/product-service/controller"
	"github.com/gin-gonic/gin"
)

func UploadRoutes(router *gin.Engine) {
	uploadController := controllers.NewUploadController()

	uploadGroup := router.Group("/upload")
	{
		// Get presigned URL for direct S3 upload
		uploadGroup.POST("/presigned-url", uploadController.GetPresignedUploadURL)

		// Direct file upload through server (optional)
		uploadGroup.POST("/file", uploadController.UploadFile)
	}
}
