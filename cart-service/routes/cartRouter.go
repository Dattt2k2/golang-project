package routes

import (
	"log"

	"cart-service/controller"
	"cart-service/repository"
	"cart-service/service"
	"github.com/gin-gonic/gin"
)

// SetupCartDependencies thiết lập các dependencies theo mô hình 3 layer
func SetupCartDependencies() (*controller.CartController, *controller.CartServer) {
	// Setup repository layer
	cartRepo := repository.NewcartRepository()

	// Setup service layer
	cartService, err := service.NewCartService(cartRepo)
	if err != nil {
		log.Fatalf("Failed to initialize cart service: %v", err)
	}

	// Setup controller layers (HTTP & gRPC)
	cartController := controller.NewCartController(cartService)
	cartServer := controller.NewCartServer(cartService)

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
