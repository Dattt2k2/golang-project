package routes

import (
	// "context"
	// "log"
	// "os"

	"cart-service/controller"
	// logger "cart-service/log"
	// "cart-service/repository"
	"cart-service/service"

	// "github.com/aws/aws-sdk-go-v2/config"
	// "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
)

// SetupCartDependencies thiết lập các dependencies theo mô hình 3 layer
func SetupCartDependencies(cartSvc service.CartService) (*controller.CartController, *controller.CartServer) {
    // Sử dụng cartSvc được pass từ main.go
    cartController := controller.NewCartController(cartSvc)
    cartServer := controller.NewCartServer(cartSvc)
    
    return cartController, cartServer
}

// CartRoutes thiết lập các route HTTP cho cart service
func CartRoutes(router *gin.Engine, cartController *controller.CartController) {
	routes := router.Group("/cart")
	{
		routes.POST("/add/:id", cartController.AddToCart())
		routes.GET("/user/get", cartController.GetCart())
		routes.GET("/get", cartController.GetCartSeller())
		routes.DELETE("/delete/:id", cartController.DeleteProductFromCart())
		routes.DELETE("/clear", cartController.ClearCart())
	}
}
