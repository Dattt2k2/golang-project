package routes

import (
	controller "product-service/controller"
	"product-service/database"
	"product-service/repository"
	"product-service/service"
	"github.com/gin-gonic/gin"
	// "go.mongodb.org/mongo-driver/mongo"
)

// Hàm khởi tạo service riêng để dùng cho Kafka consumer hoặc các mục đích khác
func NewProductService() service.ProductService {
	productRepo := repository.NewProductRepository(database.OpenCollection(database.Client, "products"))
	return service.NewProductService(productRepo)
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
	authorized.GET("/products/search", productController.GetProductByName())

	authorized.GET("/best-selling", productController.GetBestSellingProducts())
	// get product image
	// authorized.GET("/products/images/:filename", productController.GetProductImage())
	// authorized.GET("images/:filename", controller.GetProductImage)
	// authorized.GET("/verify", controller.VerifyImageExists)
	// Static file server for images
	incomingRoutes.StaticFS("/static-images", gin.Dir("product-service/uploads/images", true))
}
