package routes

import (
	"context"
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
	cfg, err := config.LoadDefaultConfig(context.Background())
    if err != nil {
        logger.Logger.Fatal("unable to load SDK config, " + err.Error())
    }
    dynamoClient := dynamodb.NewFromConfig(cfg)
    productRepo := repository.NewProductRepository(dynamoClient, "products") // "products" là tên DynamoDB table
    return service.NewProductService(productRepo, service.NewS3Service())
}

func SetupProductController() *controller.ProductController {
	productSvc := NewProductService()
	s3Service := service.NewS3Service()
	return controller.NewProductController(productSvc, *s3Service)
}

func ProductManagerRoutes(incomingRoutes *gin.Engine) {
	productController := SetupProductController()
	authorized := incomingRoutes.Group("/")

	// add product to database
	authorized.POST("/products/add", productController.AddProduct())
	// Edit product from database
	authorized.PUT("/products/edit/:id", productController.EditProduct())
	// Delete product from databse
	authorized.DELETE("/products/delete/:id", productController.DeleteProduct())
	// Get all product from database
	authorized.GET("/products/get", productController.GetAllProducts())
	// search product
	// authorized.GET("/products/search", productController.GetProductByName())

	authorized.GET("/best-selling", productController.GetBestSellingProducts())
	// get product image
	// authorized.GET("/products/images/:filename", productController.GetProductImage())
	// authorized.GET("images/:filename", controller.GetProductImage)
	// authorized.GET("/verify", controller.VerifyImageExists)
	// Static file server for images
	incomingRoutes.StaticFS("/static-images", gin.Dir("product-service/uploads/images", true))
}
