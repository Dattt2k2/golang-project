package routes

import (
	"context"
	"os"
	controller "product-service/controller"
	logger "product-service/log"
	"product-service/repository"
	"product-service/service"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
)

// Hàm khởi tạo service riêng để dùng cho Kafka consumer hoặc các mục đích khác
func NewProductService() service.ProductService {
	dynamoRegion := os.Getenv("DYNAMODB_REGION")
	if dynamoRegion == "" {
		dynamoRegion = "us-west-2"
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(dynamoRegion),
	)
	if err != nil {
		logger.Logger.Fatal("unable to load SDK config, " + err.Error())
	}

	dynamoClient := dynamodb.NewFromConfig(cfg)

	tableName := os.Getenv("DYNAMODB_TABLE")
	if tableName == "" {
		tableName = "product-table"
	}

	productRepo := repository.NewProductRepository(dynamoClient, tableName)
	return service.NewProductService(productRepo, service.NewS3Service())
}

// Sửa function này để nhận productSvc từ main.go
func ProductManagerRoutes(incomingRoutes *gin.Engine, productSvc service.ProductService) {
	s3Service := service.NewS3Service()
	productController := controller.NewProductController(productSvc, *s3Service)

	authorized := incomingRoutes.Group("/")

	// add product to database
	authorized.POST("/products/add", productController.AddProduct())
	// Edit product from database
	authorized.PUT("/products/edit/:id", productController.EditProduct())
	// Delete product from databse
	authorized.DELETE("/products/delete/:id", productController.DeleteProduct())
	// Statistics product
	authorized.GET("/products/statistics", productController.GetProductStatistics())
	authorized.POST("/products/category", productController.AddProductCategory())
	authorized.GET("/products/get/category", productController.GetProductCategory())
	authorized.DELETE("/products/category/:id", productController.DeleteProductCategory())
	// Get all product from database
	authorized.GET("/products/get/all", productController.GetAllProducts())

	authorized.GET("/products/user", productController.GetProductByUserID())
	// search product
	// authorized.GET("/products/search", productController.GetProductByName())
	authorized.GET("/products/get/:id", productController.GetProductByID())

	authorized.GET("/products/get/best-selling", productController.GetBestSellingProducts())

	authorized.GET("/products/get/category/:category", productController.GetProductByCategory())

	// Static file server for images
	incomingRoutes.StaticFS("/static-images", gin.Dir("product-service/uploads/images", true))
}
